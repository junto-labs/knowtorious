// Package knowledgetest checks that a knowledge.Store keeps the contract the
// CLI and MCP server rely on. Every backend runs it from its own tests.
package knowledgetest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"juntolabs.dev/knowtorious/internal/knowledge"
)

// Fixture describes content the store under test is known to hold.
type Fixture struct {
	// Query matches at least one entry.
	Query string
	// NoMatch matches no entry.
	NoMatch string
}

// Run checks store against the knowledge.Store contract. The store must hold
// at least two entries.
func Run(t *testing.T, store knowledge.Store, fx Fixture) {
	t.Helper()
	ctx := context.Background()

	t.Run("list is stable and has ids and titles", func(t *testing.T) {
		first, err := store.List(ctx, knowledge.ListOptions{})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(first), 2, "fixture needs at least two entries")
		second, err := store.List(ctx, knowledge.ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, first, second)
		for _, s := range first {
			assert.NotEmpty(t, s.ID)
			assert.NotEmpty(t, s.Title, s.ID)
		}
	})

	t.Run("list filters by type", func(t *testing.T) {
		all, err := store.List(ctx, knowledge.ListOptions{})
		require.NoError(t, err)
		typ := all[0].Type
		filtered, err := store.List(ctx, knowledge.ListOptions{Type: typ})
		require.NoError(t, err)
		require.NotEmpty(t, filtered)
		for _, s := range filtered {
			assert.Equal(t, typ, s.Type)
		}
	})

	t.Run("get returns every listed entry", func(t *testing.T) {
		all, err := store.List(ctx, knowledge.ListOptions{})
		require.NoError(t, err)
		for _, s := range all {
			e, err := store.Get(ctx, s.ID)
			require.NoError(t, err, s.ID)
			assert.Equal(t, s, e.Summary)
			assert.NotEmpty(t, e.Source, s.ID)
		}
	})

	t.Run("get unknown id is not found", func(t *testing.T) {
		_, err := store.Get(ctx, "knowtorious-test/does-not-exist")
		assert.ErrorIs(t, err, knowledge.ErrNotFound)
	})

	t.Run("search finds, ranks and limits", func(t *testing.T) {
		hits, err := store.Search(ctx, fx.Query, knowledge.SearchOptions{})
		require.NoError(t, err)
		require.NotEmpty(t, hits)
		for i := 1; i < len(hits); i++ {
			assert.GreaterOrEqual(t, hits[i-1].Score, hits[i].Score)
		}
		limited, err := store.Search(ctx, fx.Query, knowledge.SearchOptions{Limit: 1})
		require.NoError(t, err)
		assert.Len(t, limited, 1)
	})

	t.Run("search with no match is empty", func(t *testing.T) {
		hits, err := store.Search(ctx, fx.NoMatch, knowledge.SearchOptions{})
		require.NoError(t, err)
		assert.Empty(t, hits)
	})
}
