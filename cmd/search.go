package cmd

import (
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"juntolabs.dev/knowtorious/internal/knowledge"
)

func newSearchCmd(opts *options) *cobra.Command {
	var searchOpts knowledge.SearchOptions
	cmd := &cobra.Command{
		Use:   "search QUERY...",
		Short: "Find entries matching a query, best first",
		Args:  usageArgs(cobra.MinimumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			if strings.TrimSpace(query) == "" {
				return usageError{errors.New("query is empty")}
			}
			if searchOpts.Limit < 0 {
				return usageError{errors.New("--limit must be 0 or more")}
			}
			store, err := opts.open(cmd.Context())
			if err != nil {
				return err
			}
			hits, err := store.Search(cmd.Context(), query, searchOpts)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if opts.json {
				return writeJSON(out, hits)
			}
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			for _, h := range hits {
				fmt.Fprintf(tw, "%s\t%s\t%s\n", h.ID, h.Title, h.Snippet)
			}
			return tw.Flush()
		},
	}
	cmd.Flags().IntVar(&searchOpts.Limit, "limit", 10, "maximum results (0 for all)")
	return cmd
}
