// Package markdown is the default knowledge backend: a .knowtorious/ folder of
// markdown files.
//
// Every .md file under the root is an entry. Its id is the path relative to
// the root, with forward slashes and without the .md extension (e.g.
// "architecture/workflow-steps"). README.md files are folder stubs and are not
// entries; non-markdown files and dot-directories are ignored.
package markdown

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
	"juntolabs.dev/knowtorious/internal/knowledge"
)

// DirName is the vault directory name looked for by Find.
const DirName = ".knowtorious"

// Vault is a loaded markdown vault. It implements knowledge.Store.
type Vault struct {
	Root    string
	entries []knowledge.Entry
	byID    map[string]int
}

var _ knowledge.Store = (*Vault)(nil)

// Find walks up from dir looking for a .knowtorious directory.
func Find(dir string) (string, error) {
	start, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for dir = start; ; {
		candidate := filepath.Join(dir, DirName)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no %s directory in %s or any parent: %w", DirName, start, knowledge.ErrNotFound)
		}
		dir = parent
	}
}

// Open reads every entry under root.
func Open(root string) (*Vault, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("vault %s: %w", root, knowledge.ErrNotFound)
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault %s is not a directory", root)
	}

	v := &Vault{Root: root, byID: map[string]int{}}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() || filepath.Ext(path) != ".md" || strings.EqualFold(d.Name(), "README.md") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		e, err := parseEntry(filepath.ToSlash(rel), raw)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		v.entries = append(v.entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(v.entries, func(i, j int) bool { return v.entries[i].ID < v.entries[j].ID })
	for i, e := range v.entries {
		v.byID[e.ID] = i
	}
	return v, nil
}

// List returns entry summaries sorted by id.
func (v *Vault) List(_ context.Context, opts knowledge.ListOptions) ([]knowledge.Summary, error) {
	out := []knowledge.Summary{}
	for _, e := range v.entries {
		if opts.Type != "" && e.Type != opts.Type {
			continue
		}
		out = append(out, e.Summary)
	}
	return out, nil
}

// Get returns the entry with the given id. A trailing .md is accepted.
func (v *Vault) Get(_ context.Context, id string) (knowledge.Entry, error) {
	i, ok := v.byID[strings.TrimSuffix(id, ".md")]
	if !ok {
		return knowledge.Entry{}, fmt.Errorf("entry %q: %w", id, knowledge.ErrNotFound)
	}
	return v.entries[i], nil
}

type frontmatter struct {
	Title string `yaml:"title"`
	Type  string `yaml:"type"`
}

// parseEntry builds an entry from a file's relative path and contents.
// Optional YAML frontmatter may set title and type; otherwise the title is the
// first "# " heading, falling back to the file name, and the type is the
// top-level folder (or the file name for a top-level file such as
// conventions.md).
func parseEntry(rel string, raw []byte) (knowledge.Entry, error) {
	id := strings.TrimSuffix(rel, ".md")
	body := raw
	var fm frontmatter
	if rest, yml, ok := splitFrontmatter(raw); ok {
		if err := yaml.Unmarshal(yml, &fm); err != nil {
			return knowledge.Entry{}, fmt.Errorf("parsing frontmatter: %w", err)
		}
		body = rest
	}

	e := knowledge.Entry{
		Summary: knowledge.Summary{ID: id, Title: fm.Title, Type: fm.Type},
		Source:  rel,
		Content: string(body),
	}
	if e.Title == "" {
		e.Title = firstHeading(body)
	}
	if e.Title == "" {
		e.Title = titleFromName(id)
	}
	if e.Type == "" {
		e.Type, _, _ = strings.Cut(id, "/")
	}
	return e, nil
}

func splitFrontmatter(raw []byte) (body, yml []byte, ok bool) {
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
	if !bytes.HasPrefix(raw, []byte("---\n")) && !bytes.HasPrefix(raw, []byte("---\r\n")) {
		return raw, nil, false
	}
	_, rest, _ := bytes.Cut(raw, []byte("\n"))
	for off := 0; off < len(rest); {
		line, _, _ := bytes.Cut(rest[off:], []byte("\n"))
		if string(bytes.TrimRight(line, "\r")) == "---" {
			end := off + len(line) + 1
			if end > len(rest) {
				end = len(rest)
			}
			return rest[end:], rest[:off], true
		}
		off += len(line) + 1
	}
	return raw, nil, false
}

func firstHeading(body []byte) string {
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func titleFromName(id string) string {
	name := id[strings.LastIndex(id, "/")+1:]
	name = strings.NewReplacer("-", " ", "_", " ").Replace(name)
	if name == "" {
		return id
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
