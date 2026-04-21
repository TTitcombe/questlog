package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/TTitcombe/questlog/internal/model"
)

func (s *FSStore) goalDir(slug string) string {
	return filepath.Join(s.dataDir, "goals", slug)
}

func (s *FSStore) goalPath(slug string) string {
	return filepath.Join(s.goalDir(slug), "goal.json")
}

func (s *FSStore) SaveGoal(g model.Goal) error {
	if err := os.MkdirAll(s.goalDir(g.Slug), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.goalPath(g.Slug), data, 0644)
}

func (s *FSStore) LoadGoal(slug string) (model.Goal, error) {
	data, err := os.ReadFile(s.goalPath(slug))
	if err != nil {
		return model.Goal{}, err
	}
	var g model.Goal
	return g, json.Unmarshal(data, &g)
}

func (s *FSStore) ListGoals() ([]model.Goal, error) {
	dir := filepath.Join(s.dataDir, "goals")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var goals []model.Goal
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		g, err := s.LoadGoal(e.Name())
		if err != nil {
			continue
		}
		goals = append(goals, g)
	}
	return goals, nil
}
