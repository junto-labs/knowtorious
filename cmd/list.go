package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"juntolabs.dev/knowtorious/internal/knowledge"
)

func newListCmd(opts *options) *cobra.Command {
	var listOpts knowledge.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List entry ids and titles (never content)",
		Args:  usageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := opts.open(cmd.Context())
			if err != nil {
				return err
			}
			items, err := store.List(cmd.Context(), listOpts)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if opts.json {
				return writeJSON(out, items)
			}
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			for _, it := range items {
				fmt.Fprintf(tw, "%s\t%s\n", it.ID, it.Title)
			}
			return tw.Flush()
		},
	}
	cmd.Flags().StringVar(&listOpts.Type, "type", "", "only list entries of this type (e.g. decisions, gotchas)")
	return cmd
}
