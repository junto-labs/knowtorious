package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"juntolabs.dev/knowtorious/internal/knowledge"
	"juntolabs.dev/knowtorious/internal/markdown"
)

const testVault = "../markdown/testdata/.knowtorious"

var ctx = context.Background()

func vaultRoot(t *testing.T, s knowledge.Store) string {
	t.Helper()
	v, ok := s.(*markdown.Vault)
	require.True(t, ok, "expected a markdown vault, got %T", s)
	return v.Root
}

func TestOpenPathAndTypedSource(t *testing.T) {
	want, err := filepath.Abs(testVault)
	require.NoError(t, err)
	for _, src := range []string{testVault, "markdown:" + testVault} {
		s, err := Open(ctx, src, t.TempDir())
		require.NoError(t, err, src)
		assert.Equal(t, want, vaultRoot(t, s), src)
	}
}

func TestOpenUnknownType(t *testing.T) {
	_, err := Open(ctx, "notion:abc123", t.TempDir())
	var unknown UnknownTypeError
	require.ErrorAs(t, err, &unknown)
	assert.Equal(t, "notion", unknown.Type)
	assert.Contains(t, err.Error(), "supported: markdown")
}

func TestOpenDefaultsToEnvThenDiscovery(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, markdown.DirName), 0o755))

	t.Setenv(EnvVar, "")
	s, err := Open(ctx, "", dir)
	require.NoError(t, err)
	want, _ := filepath.EvalSymlinks(filepath.Join(dir, markdown.DirName))
	got, _ := filepath.EvalSymlinks(vaultRoot(t, s))
	assert.Equal(t, want, got)

	t.Setenv(EnvVar, "markdown:"+testVault)
	s, err = Open(ctx, "", dir)
	require.NoError(t, err)
	abs, _ := filepath.Abs(testVault)
	assert.Equal(t, abs, vaultRoot(t, s))
}
