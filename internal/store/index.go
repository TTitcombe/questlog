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

func (s *FSStore) loadIndex() (model.Index, error) {
	data, err := os.ReadFile(s.indexPath())
	if err != nil {
		if os.IsNotExist(err) {
			return model.Index{}, nil
		}
		return model.Index{}, err
	}
	var idx model.Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return model.Index{}, err
	}
	return idx, nil
}

func (s *FSStore) saveIndex(idx model.Index) error {
	idx.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.indexPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.indexPath())
}

func (s *FSStore) upsertIndexEntry(r model.Resource) error {
	idx, err := s.loadIndex()
	if err != nil {
		return err
	}
	entry := resourceToEntry(r, s.dataDir)
	for i, e := range idx.Entries {
		if e.ID == r.ID {
			idx.Entries[i] = entry
			return s.saveIndex(idx)
		}
	}
	idx.Entries = append(idx.Entries, entry)
	return s.saveIndex(idx)
}

func (s *FSStore) removeIndexEntry(id string) error {
	idx, err := s.loadIndex()
	if err != nil {
		return err
	}
	entries := idx.Entries[:0]
	for _, e := range idx.Entries {
		if e.ID != id {
			entries = append(entries, e)
		}
	}
	idx.Entries = entries
	return s.saveIndex(idx)
}

// RebuildIndex walks all markdown files and rebuilds index.json from scratch.
func (s *FSStore) RebuildIndex() error {
	var entries []model.IndexEntry

	tracksDir := filepath.Join(s.dataDir, "tracks")
	if _, err := os.Stat(tracksDir); err == nil {
		_ = filepath.WalkDir(tracksDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			id := strings.TrimSuffix(filepath.Base(path), ".md")
			r, parseErr := parseMarkdown(data, id, path)
			if parseErr != nil {
				return nil
			}
			entries = append(entries, resourceToEntry(r, s.dataDir))
			return nil
		})
	}

	inboxDir := filepath.Join(s.dataDir, "inbox")
	if _, err := os.Stat(inboxDir); err == nil {
		_ = filepath.WalkDir(inboxDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			id := strings.TrimSuffix(filepath.Base(path), ".md")
			r, parseErr := parseMarkdown(data, id, path)
			if parseErr != nil {
				return nil
			}
			entries = append(entries, resourceToEntry(r, s.dataDir))
			return nil
		})
	}

	return s.saveIndex(model.Index{Entries: entries})
}

func (s *FSStore) GetIndex() (model.Index, error) {
	return s.loadIndex()
}

// SearchIndex performs an in-memory search over the index.
func (s *FSStore) SearchIndex(query string) ([]model.IndexEntry, error) {
	idx, err := s.loadIndex()
	if err != nil {
		return nil, err
	}

	tokens := strings.Fields(strings.ToLower(query))
	if len(tokens) == 0 {
		return idx.Entries, nil
	}

	var results []model.IndexEntry
	for _, e := range idx.Entries {
		searchStr := strings.ToLower(e.Title + " " + strings.Join(e.Tags, " ") + " " + e.Track)
		match := true
		for _, tok := range tokens {
			if !strings.Contains(searchStr, tok) {
				match = false
				break
			}
		}
		if match {
			results = append(results, e)
		}
	}
	return results, nil
}

func resourceToEntry(r model.Resource, dataDir string) model.IndexEntry {
	relPath, err := filepath.Rel(dataDir, r.FilePath)
	if err != nil {
		relPath = r.FilePath
	}
	return model.IndexEntry{
		ID:               r.ID,
		Title:            r.Title,
		Type:             r.Type,
		URL:              r.URL,
		Tags:             r.Tags,
		Track:            r.Track,
		Added:            r.Added.Format("2006-01-02"),
		EstimatedMinutes: r.EstimatedMinutes,
		Status:           r.Status,
		Progress:         r.Progress,
		Priority:         r.Priority,
		FilePath:         relPath,
	}
}

func (s *FSStore) indexPath() string {
	return fmt.Sprintf("%s/index.json", s.dataDir)
}
