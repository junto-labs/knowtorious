package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testVault = "../internal/markdown/testdata/.knowtorious"

func runCLI(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	root := NewRootCmd()
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	code = run(root, append(args, "--source", testVault), &errOut)
	return out.String(), errOut.String(), code
}

func TestListText(t *testing.T) {
	out, _, code := runCLI(t, "list")
	require.Equal(t, 0, code)
	assert.Regexp(t, `(?m)^gotchas/tabs-in-yaml +Tabs in YAML$`, out)
	assert.NotContains(t, out, "YAML rejects", "list must never print content")
}

func TestListJSON(t *testing.T) {
	out, _, code := runCLI(t, "list", "--json", "--type", "decisions")
	require.Equal(t, 0, code)
	var items []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &items))
	assert.Equal(t, []map[string]any{
		{"id": "decisions/use-go", "title": "Use Go for every component", "type": "decisions"},
	}, items)
}

func TestGetText(t *testing.T) {
	out, _, code := runCLI(t, "get", "gotchas/tabs-in-yaml")
	require.Equal(t, 0, code)
	assert.Equal(t, "# Tabs in YAML\n\nYAML rejects tab indentation. Use spaces.\n", out)
}

func TestGetJSONMultiple(t *testing.T) {
	out, _, code := runCLI(t, "get", "--json", "conventions", "decisions/use-go")
	require.Equal(t, 0, code)
	var entries []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &entries))
	require.Len(t, entries, 2)
	assert.Equal(t, "conventions", entries[0]["id"])
	assert.Equal(t, "decisions/use-go.md", entries[1]["source"])
	assert.Contains(t, entries[1]["content"], "single binary")
}

func TestGetUnknownFails(t *testing.T) {
	out, errOut, code := runCLI(t, "get", "nope")
	assert.Equal(t, exitFail, code)
	assert.Empty(t, out)
	assert.Contains(t, errOut, `entry "nope": not found`)
}

func TestSearchJSON(t *testing.T) {
	out, _, code := runCLI(t, "search", "--json", "tab", "spaces")
	require.Equal(t, 0, code)
	var hits []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &hits))
	require.Len(t, hits, 1)
	assert.Equal(t, "gotchas/tabs-in-yaml", hits[0]["id"])
}

func TestSearchNoHitsIsEmptyArray(t *testing.T) {
	out, _, code := runCLI(t, "search", "--json", "kubernetes")
	require.Equal(t, 0, code)
	assert.Equal(t, "[]\n", out)
}

func TestUsageErrorsExit2(t *testing.T) {
	for _, args := range [][]string{
		{"get"},
		{"search"},
		{"list", "extra"},
		{"list", "--bogus"},
		{"search", "x", "--limit", "-1"},
	} {
		_, _, code := runCLI(t, args...)
		assert.Equal(t, exitUsage, code, args)
	}
}

func TestMissingVaultFails(t *testing.T) {
	root := NewRootCmd()
	var errOut bytes.Buffer
	root.SetOut(&bytes.Buffer{})
	code := run(root, []string{"list", "--source", t.TempDir() + "/.knowtorious"}, &errOut)
	assert.Equal(t, exitFail, code)
	assert.Contains(t, errOut.String(), "not found")
}

func TestUnknownSourceTypeIsUsageError(t *testing.T) {
	root := NewRootCmd()
	var errOut bytes.Buffer
	root.SetOut(&bytes.Buffer{})
	code := run(root, []string{"list", "--source", "confluence:ENG"}, &errOut)
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut.String(), `unknown source type "confluence"`)
}
