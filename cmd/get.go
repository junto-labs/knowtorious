package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"juntolabs.dev/knowtorious/internal/knowledge"
)

func newGetCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "get ID...",
		Short: "Print one or more entries",
		Args:  usageArgs(cobra.MinimumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := opts.open(cmd.Context())
			if err != nil {
				return err
			}
			entries := make([]knowledge.Entry, 0, len(args))
			for _, id := range args {
				e, err := store.Get(cmd.Context(), id)
				if err != nil {
					return err
				}
				entries = append(entries, e)
			}

			out := cmd.OutOrStdout()
			if opts.json {
				return writeJSON(out, entries)
			}
			for i, e := range entries {
				if i > 0 {
					fmt.Fprintln(out)
				}
				if len(entries) > 1 {
					fmt.Fprintf(out, "<!-- %s -->\n", e.ID)
				}
				fmt.Fprint(out, e.Content)
				if !strings.HasSuffix(e.Content, "\n") {
					fmt.Fprintln(out)
				}
			}
			return nil
		},
	}
}
