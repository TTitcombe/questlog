package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TTitcombe/questlog/internal/model"
)

// FSStore is the filesystem-backed Store implementation.
type FSStore struct {
	dataDir string
}

// New creates an FSStore rooted at dataDir (e.g. ~/.questlog).
// It creates the directory if it doesn't exist and auto-rebuilds
// index.json if missing.
func New(dataDir string) (*FSStore, error) {
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "tracks"),
		filepath.Join(dataDir, "inbox"),
		filepath.Join(dataDir, "goals"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("create data dir %s: %w", d, err)
		}
	}

	s := &FSStore{dataDir: dataDir}

	// Auto-rebuild index if missing
	if _, err := os.Stat(filepath.Join(dataDir, "index.json")); os.IsNotExist(err) {
		if err := s.RebuildIndex(); err != nil {
			return nil, fmt.Errorf("rebuild index: %w", err)
		}
	}

	return s, nil
}

func (s *FSStore) DataDir() string { return s.dataDir }

// --- Track operations ---

func (s *FSStore) CreateTrack(track model.Track) error {
	dir := filepath.Join(s.dataDir, "tracks", track.Name)
	if err := os.MkdirAll(filepath.Join(dir, "resources"), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(track, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "track.json"), data, 0644)
}

func (s *FSStore) GetTrack(name string) (model.Track, error) {
	path := filepath.Join(s.dataDir, "tracks", name, "track.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return model.Track{}, fmt.Errorf("track %q not found", name)
		}
		return model.Track{}, err
	}
	var t model.Track
	return t, json.Unmarshal(data, &t)
}

func (s *FSStore) ListTracks() ([]model.Track, error) {
	entries, err := os.ReadDir(filepath.Join(s.dataDir, "tracks"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var tracks []model.Track
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		t, err := s.GetTrack(e.Name())
		if err != nil {
			continue
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

func (s *FSStore) SaveTrack(t model.Track) error {
	path := filepath.Join(s.dataDir, "tracks", t.Name, "track.json")
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// --- Resource operations ---

func (s *FSStore) resourceDir(track string) string {
	if track == "inbox" {
		return filepath.Join(s.dataDir, "inbox")
	}
	return filepath.Join(s.dataDir, "tracks", track, "resources")
}

func (s *FSStore) SaveResource(r model.Resource) error {
	dir := s.resourceDir(r.Track)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if r.ID == "" {
		r.ID = uniqueSlug(r.Title, dir)
	}
	if r.Added.IsZero() {
		r.Added = time.Now()
	}
	if r.Status == "" {
		r.Status = model.StatusUnread
	}

	var filename string
	if r.Track == "inbox" {
		filename = fmt.Sprintf("%s-%s.md", r.Added.Format("2006-01-02"), r.ID)
	} else {
		filename = r.ID + ".md"
	}

	r.FilePath = filepath.Join(dir, filename)

	data, err := marshalMarkdown(r)
	if err != nil {
		return err
	}
	if err := os.WriteFile(r.FilePath, data, 0644); err != nil {
		return err
	}
	return s.upsertIndexEntry(r)
}

func (s *FSStore) findResourceFile(id string) (string, error) {
	// Search tracks
	tracksDir := filepath.Join(s.dataDir, "tracks")
	var found string
	_ = filepath.WalkDir(tracksDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		base := strings.TrimSuffix(filepath.Base(path), ".md")
		if base == id && strings.HasSuffix(path, ".md") {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if found != "" {
		return found, nil
	}

	// Search inbox (inbox files are prefixed with date)
	inboxDir := filepath.Join(s.dataDir, "inbox")
	entries, err := os.ReadDir(inboxDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("resource %q not found", id)
		}
		return "", err
	}
	for _, e := range entries {
		name := e.Name()
		// Strip date prefix: YYYY-MM-DD-<id>.md
		if strings.HasSuffix(name, "-"+id+".md") || strings.TrimSuffix(name, ".md") == id {
			return filepath.Join(inboxDir, name), nil
		}
	}
	return "", fmt.Errorf("resource %q not found", id)
}

func (s *FSStore) GetResource(id string) (model.Resource, error) {
	path, err := s.findResourceFile(id)
	if err != nil {
		return model.Resource{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Resource{}, err
	}
	return parseMarkdown(data, id, path)
}

func (s *FSStore) DeleteResource(id string) error {
	path, err := s.findResourceFile(id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return s.removeIndexEntry(id)
}

func (s *FSStore) ListResources(filter ResourceFilter) ([]model.Resource, error) {
	var resources []model.Resource

	collect := func(dir string) {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			// Derive ID from filename (strip date prefix for inbox)
			base := strings.TrimSuffix(filepath.Base(path), ".md")
			// inbox files: YYYY-MM-DD-<id> — the id is everything after 11 chars
			if len(base) > 11 && base[4] == '-' && base[7] == '-' && base[10] == '-' {
				base = base[11:]
			}
			r, parseErr := parseMarkdown(data, base, path)
			if parseErr != nil {
				return nil
			}
			if filter.Track != "" && r.Track != filter.Track {
				return nil
			}
			if filter.Status != "" && r.Status != filter.Status {
				return nil
			}
			if filter.Type != "" && r.Type != filter.Type {
				return nil
			}
			resources = append(resources, r)
			return nil
		})
	}

	if filter.Track == "" || filter.Track == "inbox" {
		collect(filepath.Join(s.dataDir, "inbox"))
	}

	tracksDir := filepath.Join(s.dataDir, "tracks")
	if filter.Track != "" && filter.Track != "inbox" {
		collect(filepath.Join(tracksDir, filter.Track, "resources"))
	} else if filter.Track == "" {
		trackEntries, err := os.ReadDir(tracksDir)
		if err == nil {
			for _, te := range trackEntries {
				if te.IsDir() {
					collect(filepath.Join(tracksDir, te.Name(), "resources"))
				}
			}
		}
	}

	return resources, nil
}

// SearchNotes does a case-insensitive full-text scan of all resource note bodies.
func (s *FSStore) SearchNotes(query string) ([]model.Resource, error) {
	all, err := s.ListResources(ResourceFilter{})
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var results []model.Resource
	for _, r := range all {
		if strings.Contains(strings.ToLower(r.Notes), q) {
			results = append(results, r)
		}
	}
	return results, nil
}

// --- Inbox ---

func (s *FSStore) ListInbox() ([]model.Resource, error) {
	return s.ListResources(ResourceFilter{Track: "inbox"})
}

func (s *FSStore) MoveToTrack(resourceID string, trackName string) error {
	r, err := s.GetResource(resourceID)
	if err != nil {
		return err
	}

	// Verify target track exists
	if _, err := s.GetTrack(trackName); err != nil {
		return fmt.Errorf("track %q does not exist; create it first with: qlog track new %s", trackName, trackName)
	}

	oldPath := r.FilePath

	// Remove from old location
	if err := os.Remove(oldPath); err != nil {
		return err
	}
	if err := s.removeIndexEntry(r.ID); err != nil {
		return err
	}

	// Save to new track
	r.Track = trackName
	r.FilePath = ""
	return s.SaveResource(r)
}
