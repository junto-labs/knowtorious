package markdown

import (
	"context"
	"sort"
	"strings"

	"juntolabs.dev/knowtorious/internal/knowledge"
)

// titleWeight makes a term in the title count for more than one in the body.
const titleWeight = 10

const maxSnippet = 160

// Search returns entries containing every term in query, case-insensitively,
// ranked by how often the terms appear (title matches weigh more).
func (v *Vault) Search(_ context.Context, query string, opts knowledge.SearchOptions) ([]knowledge.Hit, error) {
	hits := []knowledge.Hit{}
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return hits, nil
	}

	for _, e := range v.entries {
		title := strings.ToLower(e.Title + " " + e.ID)
		body := strings.ToLower(e.Content)
		score := 0
		for _, t := range terms {
			n := strings.Count(title, t)*titleWeight + strings.Count(body, t)
			if n == 0 {
				score = 0
				break
			}
			score += n
		}
		if score == 0 {
			continue
		}
		hits = append(hits, knowledge.Hit{
			Summary: e.Summary,
			Score:   float64(score),
			Snippet: snippet(e.Content, terms),
		})
	}

	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if opts.Limit > 0 && len(hits) > opts.Limit {
		hits = hits[:opts.Limit]
	}
	return hits, nil
}

// snippet returns the first non-heading line that contains a term, trimmed.
func snippet(content string, terms []string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lower := strings.ToLower(line)
		for _, t := range terms {
			if strings.Contains(lower, t) {
				return truncate(line, maxSnippet)
			}
		}
	}
	return ""
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
