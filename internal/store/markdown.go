package store

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/TTitcombe/questlog/internal/model"
)

// frontmatter is the YAML struct written to / read from the --- block.
type frontmatter struct {
	Title            string             `yaml:"title"`
	Type             model.ResourceType `yaml:"type"`
	URL              string             `yaml:"url,omitempty"`
	Tags             []string           `yaml:"tags,omitempty"`
	Track            string             `yaml:"track"`
	Added            string             `yaml:"added"` // YYYY-MM-DD
	EstimatedMinutes int                `yaml:"estimated_minutes,omitempty"`
	Status           model.Status       `yaml:"status"`
	Progress         int                `yaml:"progress,omitempty"`
	Priority         int                `yaml:"priority,omitempty"`
}

// parseMarkdown parses a resource markdown file into a model.Resource.
// id and filePath are set by the caller from the filename/path.
func parseMarkdown(content []byte, id, filePath string) (model.Resource, error) {
	s := string(content)

	if !strings.HasPrefix(s, "---") {
		return model.Resource{}, fmt.Errorf("no frontmatter found in %s", filePath)
	}

	// Find closing ---
	rest := s[3:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return model.Resource{}, fmt.Errorf("unclosed frontmatter in %s", filePath)
	}

	yamlBlock := rest[:end]
	body := strings.TrimPrefix(rest[end+4:], "\n")

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return model.Resource{}, fmt.Errorf("parse frontmatter %s: %w", filePath, err)
	}

	added, _ := time.Parse("2006-01-02", fm.Added)

	return model.Resource{
		ID:               id,
		Title:            fm.Title,
		Type:             fm.Type,
		URL:              fm.URL,
		Tags:             fm.Tags,
		Track:            fm.Track,
		Added:            added,
		EstimatedMinutes: fm.EstimatedMinutes,
		Status:           fm.Status,
		Progress:         fm.Progress,
		Priority:         fm.Priority,
		Notes:            body,
		FilePath:         filePath,
	}, nil
}

// marshalMarkdown serialises a model.Resource to markdown bytes.
func marshalMarkdown(r model.Resource) ([]byte, error) {
	fm := frontmatter{
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
	}

	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	buf.WriteString("---\n")
	if r.Notes != "" {
		buf.WriteByte('\n')
		buf.WriteString(r.Notes)
		if !strings.HasSuffix(r.Notes, "\n") {
			buf.WriteByte('\n')
		}
	}
	return buf.Bytes(), nil
}
