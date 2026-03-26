package store

import "github.com/TTitcombe/questlog/internal/model"

// ResourceFilter narrows ListResources results.
type ResourceFilter struct {
	Track  string
	Status model.Status
	Type   model.ResourceType
}

// Store is the interface all CLI commands use for data access.
type Store interface {
	// Track operations
	CreateTrack(track model.Track) error
	GetTrack(name string) (model.Track, error)
	ListTracks() ([]model.Track, error)

	// Resource operations
	SaveResource(resource model.Resource) error
	GetResource(id string) (model.Resource, error)
	DeleteResource(id string) error
	ListResources(filter ResourceFilter) ([]model.Resource, error)

	// Inbox
	ListInbox() ([]model.Resource, error)
	MoveToTrack(resourceID string, trackName string) error

	// Index
	RebuildIndex() error
	SearchIndex(query string) ([]model.IndexEntry, error)
	GetIndex() (model.Index, error)

	DataDir() string
}
