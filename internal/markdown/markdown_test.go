package markdown

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"juntolabs.dev/knowtorious/internal/knowledge"
	"juntolabs.dev/knowtorious/internal/knowledge/knowledgetest"
)

const testVault = "testdata/.knowtorious"

var ctx = context.Background()

func openTest(t *testing.T) *Vault {
	t.Helper()
	v, err := Open(testVault)
	require.NoError(t, err)
	return v
}

func TestStoreContract(t *testing.T) {
	knowledgetest.Run(t, openTest(t), knowledgetest.Fixture{Query: "e", NoMatch: "kubernetes"})
}

func TestListSkipsNonEntries(t *testing.T) {
	got, err := openTest(t).List(ctx, knowledge.ListOptions{})
	require.NoError(t, err)
	// README.md stubs, non-markdown files and dot-directories are skipped.
	assert.Equal(t, []knowledge.Summary{
		{ID: "architecture/vault-layout", Title: "Vault layout", Type: "architecture"},
		{ID: "conventions", Title: "Conventions", Type: "conventions"},                   // title from file name
		{ID: "decisions/use-go", Title: "Use Go for every component", Type: "decisions"}, // frontmatter wins over heading
		{ID: "gotchas/tabs-in-yaml", Title: "Tabs in YAML", Type: "gotchas"},             // first heading
	}, got)
}

func TestFrontmatterIsStrippedFromContent(t *testing.T) {
	e, err := openTest(t).Get(ctx, "decisions/use-go")
	require.NoError(t, err)
	assert.Equal(t, "decisions/use-go.md", e.Source)
	assert.NotContains(t, e.Content, "title:")
	assert.Contains(t, e.Content, "single binary")
}

func TestGetAcceptsMdSuffix(t *testing.T) {
	e, err := openTest(t).Get(ctx, "gotchas/tabs-in-yaml.md")
	require.NoError(t, err)
	assert.Equal(t, "gotchas/tabs-in-yaml", e.ID)
}

func TestGetIgnoredFilesAreNotFound(t *testing.T) {
	v := openTest(t)
	for _, id := range []string{"../vault", "architecture/README", ".hidden/secret"} {
		_, err := v.Get(ctx, id)
		assert.ErrorIs(t, err, knowledge.ErrNotFound, id)
	}
}

func TestFindWalksUp(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, DirName), 0o755))
	nested := filepath.Join(root, "a", "b")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	got, err := Find(nested)
	require.NoError(t, err)
	want, _ := filepath.EvalSymlinks(filepath.Join(root, DirName))
	got, _ = filepath.EvalSymlinks(got)
	assert.Equal(t, want, got)
}

func TestFindMissing(t *testing.T) {
	_, err := Find(t.TempDir())
	assert.ErrorIs(t, err, knowledge.ErrNotFound)
}

func TestOpenMissing(t *testing.T) {
	_, err := Open(filepath.Join(t.TempDir(), DirName))
	assert.ErrorIs(t, err, knowledge.ErrNotFound)
}

func TestSearch(t *testing.T) {
	v := openTest(t)
	search := func(q string) []knowledge.Hit {
		hits, err := v.Search(ctx, q, knowledge.SearchOptions{})
		require.NoError(t, err)
		return hits
	}

	hits := search("YAML")
	require.Len(t, hits, 1)
	assert.Equal(t, "gotchas/tabs-in-yaml", hits[0].ID)
	assert.Equal(t, "YAML rejects tab indentation. Use spaces.", hits[0].Snippet)

	// Every term must match.
	assert.Empty(t, search("yaml binary"))

	// A title match outranks a body-only match.
	hits = search("go")
	require.NotEmpty(t, hits)
	assert.Equal(t, "decisions/use-go", hits[0].ID)

	assert.Empty(t, search("   "))
}
