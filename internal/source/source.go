// Package source turns a source string into a knowledge.Store, so commands
// never know which backend they are reading.
//
// A source is either a path to a markdown vault or "TYPE:LOCATION", where
// TYPE names a registered backend:
//
//	.knowtorious                  markdown vault at a path
//	markdown:../other/.knowtorious
//	notion:<database-id>          (future backend)
//
// An empty source falls back to $KNOWTORIOUS_SOURCE, then to the nearest
// .knowtorious/ in the working directory or a parent.
package source

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"juntolabs.dev/knowtorious/internal/knowledge"
	"juntolabs.dev/knowtorious/internal/markdown"
)

// EnvVar holds the default source when no flag is given.
const EnvVar = "KNOWTORIOUS_SOURCE"

// Opener opens a backend at a location (everything after "TYPE:").
type Opener func(ctx context.Context, location string) (knowledge.Store, error)

var openers = map[string]Opener{
	"markdown": func(_ context.Context, location string) (knowledge.Store, error) {
		return markdown.Open(location)
	},
}

// typePrefix matches a "TYPE:" prefix. It needs two or more characters so a
// Windows drive letter ("C:") reads as a path.
var typePrefix = regexp.MustCompile(`^([a-z][a-z0-9-]+):`)

// UnknownTypeError is returned for a "TYPE:" prefix with no backend.
type UnknownTypeError struct{ Type string }

func (e UnknownTypeError) Error() string {
	return fmt.Sprintf("unknown source type %q (supported: %s)", e.Type, strings.Join(Types(), ", "))
}

// Types lists the registered backend types.
func Types() []string {
	out := make([]string, 0, len(openers))
	for t := range openers {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// Open resolves src and opens its store. cwd is where an empty source starts
// looking for a vault.
func Open(ctx context.Context, src, cwd string) (knowledge.Store, error) {
	if src == "" {
		src = os.Getenv(EnvVar)
	}
	if src == "" {
		root, err := markdown.Find(cwd)
		if err != nil {
			return nil, err
		}
		return markdown.Open(root)
	}

	if m := typePrefix.FindStringSubmatch(src); m != nil {
		open, ok := openers[m[1]]
		if !ok {
			return nil, UnknownTypeError{Type: m[1]}
		}
		return open(ctx, strings.TrimPrefix(src, m[0]))
	}
	return markdown.Open(src)
}
