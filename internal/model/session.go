package model

import "time"

// StatusChange records a resource status update made during a session.
type StatusChange struct {
	ResourceID string `json:"resource_id"`
	To         Status `json:"to"`
}

// Session records a completed focus session.
type Session struct {
	StartedAt     time.Time      `json:"started_at"`
	PlannedMins   int            `json:"planned_minutes"`
	ActualSecs    int            `json:"actual_seconds"`
	Track         string         `json:"track,omitempty"`
	Opened        []string       `json:"opened,omitempty"`        // resource IDs opened in browser
	StatusChanges []StatusChange `json:"status_changes,omitempty"` // status changes made during session
}
