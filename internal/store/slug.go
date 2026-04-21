package store

import (
	"fmt"
	"os"
	"strings"
)

// slugify converts a title to a URL-safe slug.
func slugify(title string) string {
	// Lowercase and replace non-alphanumeric runs with "-"
	s := strings.ToLower(title)
	var sb strings.Builder
	inDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
			inDash = false
		} else if !inDash {
			sb.WriteByte('-')
			inDash = true
		}
	}
	slug := strings.Trim(sb.String(), "-")

	// Truncate at 60 chars on a word boundary
	if len(slug) > 60 {
		cut := slug[:60]
		if idx := strings.LastIndex(cut, "-"); idx > 40 {
			cut = cut[:idx]
		}
		slug = cut
	}

	return slug
}

// uniqueSlug generates a slug that doesn't already exist as a .md file in dir.
func uniqueSlug(title, dir string) string {
	base := slugify(title)
	candidate := base
	for i := 2; ; i++ {
		path := fmt.Sprintf("%s/%s.md", dir, candidate)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
}

// Slugify is the exported variant of slugify for use outside this package.
func Slugify(title string) string {
	return slugify(title)
}
