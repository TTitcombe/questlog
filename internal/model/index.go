package model

type IndexEntry struct {
	ID               string       `json:"id"`
	Title            string       `json:"title"`
	Type             ResourceType `json:"type"`
	URL              string       `json:"url,omitempty"`
	Tags             []string     `json:"tags,omitempty"`
	Track            string       `json:"track"`
	Added            string       `json:"added"` // YYYY-MM-DD string
	EstimatedMinutes int          `json:"estimated_minutes,omitempty"`
	Status           Status       `json:"status"`
	Progress         int          `json:"progress,omitempty"`
	Priority         int          `json:"priority,omitempty"` // 1 (highest) – 5 (lowest), 0 = unset
	FilePath         string       `json:"file_path"` // relative to data dir
}

type Index struct {
	Entries   []IndexEntry `json:"entries"`
	UpdatedAt string       `json:"updated_at"`
}
