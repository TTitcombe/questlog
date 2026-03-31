package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/TTitcombe/questlog/internal/model"
)

func (s *FSStore) sessionsPath() string {
	return filepath.Join(s.dataDir, "sessions.json")
}

func (s *FSStore) loadSessions() ([]model.Session, error) {
	data, err := os.ReadFile(s.sessionsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var sessions []model.Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *FSStore) saveSessions(sessions []model.Session) error {
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.sessionsPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.sessionsPath())
}

// AppendSession adds a session record to the session log.
func (s *FSStore) AppendSession(session model.Session) error {
	sessions, err := s.loadSessions()
	if err != nil {
		return err
	}
	sessions = append(sessions, session)
	return s.saveSessions(sessions)
}

// ListSessions returns all recorded sessions, oldest first.
func (s *FSStore) ListSessions() ([]model.Session, error) {
	return s.loadSessions()
}

// SessionsOnDate returns sessions that started on the given calendar date (local time).
func (s *FSStore) SessionsOnDate(date time.Time) ([]model.Session, error) {
	all, err := s.loadSessions()
	if err != nil {
		return nil, err
	}
	target := date.Local().Format("2006-01-02")
	var out []model.Session
	for _, sess := range all {
		if sess.StartedAt.Local().Format("2006-01-02") == target {
			out = append(out, sess)
		}
	}
	return out, nil
}

// Streak counts consecutive days with at least one session, counting back from today.
// If today has no session, it counts back from yesterday instead.
func (s *FSStore) Streak(today time.Time) (int, error) {
	sessions, err := s.loadSessions()
	if err != nil {
		return 0, err
	}

	dates := map[string]bool{}
	for _, sess := range sessions {
		dates[sess.StartedAt.Local().Format("2006-01-02")] = true
	}

	start := today.Local()
	if !dates[start.Format("2006-01-02")] {
		start = start.AddDate(0, 0, -1)
	}

	count := 0
	for d := start; ; d = d.AddDate(0, 0, -1) {
		if dates[d.Format("2006-01-02")] {
			count++
		} else {
			break
		}
	}
	return count, nil
}
