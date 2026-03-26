package model

import "time"

type Track struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Created     time.Time `json:"created"`
}
