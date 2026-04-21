package model

import "time"

type Goal struct {
	Slug        string      `json:"slug"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Created     time.Time   `json:"created"`
	Milestones  []Milestone `json:"milestones,omitempty"`
}

type Milestone struct {
	ID               string     `json:"id"`
	Description      string     `json:"description"`
	Deadline         *time.Time `json:"deadline,omitempty"`
	ArtifactResourceID string   `json:"artifact_resource_id,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}
