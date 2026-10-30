package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"juntolabs.dev/knowtorious/internal/knowledge"
	"juntolabs.dev/knowtorious/internal/source"
)

var version = "0.1.0"

// Exit codes follow RFC 0001: 0 pass, 1 fail, 2 usage error.
const (
	exitFail  = 1
	exitUsage = 2
)

// usageError marks an error caused by how the command was called.
type usageError struct{ err error }

func (u usageError) Error() string { return u.err.Error() }
func (u usageError) Unwrap() error { return u.err }

type options struct {
	source string
	json   bool
}

// NewRootCmd builds the knowtorious command tree.
func NewRootCmd() *cobra.Command {
	opts := &options{}
	root := &cobra.Command{
		Use:           "knowtorious",
		Short:         "Read scoped knowledge from a markdown vault or another source",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&opts.source, "source", "", "knowledge source: a vault path or TYPE:LOCATION (default: $"+source.EnvVar+", then the nearest .knowtorious/)")
	root.PersistentFlags().BoolVar(&opts.json, "json", false, "write JSON instead of text")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError{err} })

	root.AddCommand(newListCmd(opts), newGetCmd(opts), newSearchCmd(opts))
	return root
}

// Execute runs the CLI and exits with the RFC 0001 exit code.
func Execute() {
	os.Exit(run(NewRootCmd(), os.Args[1:], os.Stderr))
}

func run(root *cobra.Command, args []string, stderr io.Writer) int {
	root.SetArgs(args)
	err := root.Execute()
	if err == nil {
		return 0
	}
	fmt.Fprintln(stderr, "error:", err)
	var u usageError
	if errors.As(err, &u) {
		return exitUsage
	}
	return exitFail
}

// usageArgs wraps a cobra argument validator so its errors exit with code 2.
func usageArgs(fn cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := fn(cmd, args); err != nil {
			return usageError{err}
		}
		return nil
	}
}

func (o *options) open(ctx context.Context) (knowledge.Store, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	store, err := source.Open(ctx, o.source, cwd)
	var unknown source.UnknownTypeError
	if errors.As(err, &unknown) {
		return nil, usageError{err}
	}
	return store, err
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
