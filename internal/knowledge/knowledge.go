// Package knowledge defines the backend-neutral model and the Store interface
// that every knowledge backend implements.
//
// The CLI, the MCP server and the Claude Code plugin talk only to Store, so a
// markdown vault, Notion or Confluence can sit behind the same commands.
package knowledge

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a source or entry does not exist.
var ErrNotFound = errors.New("not found")

// Summary identifies an entry without its content. It is what the
// table of contents is made of.
type Summary struct {
	// ID is stable and opaque to callers; each backend defines its shape
	// (a vault path, a Notion page id, a Confluence content id).
	ID    string `json:"id"`
	Title string `json:"title"`
	// Type groups entries, e.g. decisions, conventions or gotchas.
	Type string `json:"type,omitempty"`
}

// Entry is a summary plus its content.
type Entry struct {
	Summary
	// Source says where the entry lives: a path inside a vault or a URL.
	Source string `json:"source"`
	// Content is markdown.
	Content string `json:"content"`
}

// Hit is one search result.
type Hit struct {
	Summary
	// Score orders hits within one result set only; it is not comparable
	// across backends.
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet,omitempty"`
}

// ListOptions narrows List.
type ListOptions struct {
	// Type, when set, keeps only entries of that type.
	Type string
}

// SearchOptions narrows Search.
type SearchOptions struct {
	// Limit caps the number of hits; zero or less means no cap.
	Limit int
}

// Store is a read-only knowledge backend.
//
// Implementations must:
//   - return List results in a stable order and never fetch content for them;
//   - return an error wrapping ErrNotFound from Get for an unknown id;
//   - return Search hits best first, and an empty result (not an error) when
//     nothing matches.
//
// knowledgetest.Run checks these rules against any implementation.
type Store interface {
	List(ctx context.Context, opts ListOptions) ([]Summary, error)
	Get(ctx context.Context, id string) (Entry, error)
	Search(ctx context.Context, query string, opts SearchOptions) ([]Hit, error)
}
