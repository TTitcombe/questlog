package model

import "time"

type Track struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Created     time.Time   `json:"created"`
	GoalSlug    string      `json:"goal_slug,omitempty"`
	DependsOn   []string    `json:"depends_on,omitempty"`
	Milestones  []Milestone `json:"milestones,omitempty"`
}
