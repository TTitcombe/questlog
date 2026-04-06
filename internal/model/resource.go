package model

import "time"

type ResourceType string

const (
	TypePaper   ResourceType = "paper"
	TypeVideo   ResourceType = "video"
	TypeBook    ResourceType = "book"
	TypeArticle ResourceType = "article"
	TypeNote    ResourceType = "note"
	TypeIdea    ResourceType = "idea"
)

var AllTypes = []ResourceType{TypePaper, TypeVideo, TypeBook, TypeArticle, TypeNote, TypeIdea}

// DefaultMinutes returns a sensible default estimated_minutes for a resource type.
func (t ResourceType) DefaultMinutes() int {
	switch t {
	case TypePaper:
		return 45
	case TypeVideo:
		return 20
	case TypeBook:
		return 60
	case TypeArticle:
		return 15
	case TypeNote, TypeIdea:
		return 5
	default:
		return 0
	}
}

type Status string

const (
	StatusUnread     Status = "unread"
	StatusInProgress Status = "in-progress"
	StatusDone       Status = "done"
)

var AllStatuses = []Status{StatusUnread, StatusInProgress, StatusDone}

type Resource struct {
	ID               string       // slug derived from title, e.g. "attention-is-all-you-need"
	Title            string       `yaml:"title"`
	Type             ResourceType `yaml:"type"`
	URL              string       `yaml:"url,omitempty"`
	Tags             []string     `yaml:"tags,omitempty"`
	Track            string       `yaml:"track"` // "inbox" or track name
	Added            time.Time    `yaml:"added"`
	EstimatedMinutes int          `yaml:"estimated_minutes,omitempty"`
	Status           Status       `yaml:"status"`
	Progress         int          `yaml:"progress,omitempty"`
	Priority         int          `yaml:"priority,omitempty"` // 1 (highest) – 5 (lowest), 0 = unset
	Rating           *int         // nil = unset, -1/0/1
	Notes            string       // markdown body after frontmatter
	FilePath         string       // absolute path on disk, not persisted
}
