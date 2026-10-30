# knowtorious

Scoped knowledge for people and agents, read from a `.knowtorious/` markdown vault. See [RFC 0001](https://github.com/junto-labs/junto/blob/rfc/0001-junto-platform/docs/rfcs/0001-junto-platform.md) for the design.

v0 is read-only: `list`, `get` and `search`. Agents load only the table of contents up front and fetch the entries they need, rather than reading knowledge out of `AGENTS.md` or `CLAUDE.md`.

## Install

```bash
go install juntolabs.dev/knowtorious@latest
```

## The vault

Every `.md` file under `.knowtorious/` is an entry. Its id is the path without `.md`. The layout matches Spektacular's knowledge folders:

```
.knowtorious/
  conventions.md            id: conventions
  architecture/storage.md   id: architecture/storage
  decisions/use-go.md       id: decisions/use-go
  gotchas/...
  glossary/...
  learnings/...
```

- **Title** comes from frontmatter `title:`, then the first `# ` heading, then the file name.
- **Type** comes from frontmatter `type:`, then the top-level folder (or the file name for a top-level file).
- `README.md` files, non-markdown files and dot-directories are ignored.

## Sources

The markdown vault is one backend. The commands only talk to the `knowledge.Store` interface (`List`, `Get`, `Search`), so another backend such as Notion or Confluence can sit behind the same CLI, MCP server and plugin.

Choose a source with `--source` or `$KNOWTORIOUS_SOURCE`. The default is the nearest `.knowtorious/` in the current directory or a parent.

```bash
knowtorious list --source ./docs/.knowtorious       # a vault path
knowtorious list --source markdown:./docs/.knowtorious
knowtorious list --source notion:<database-id>      # once a notion backend exists
```

To add a backend:

1. Implement `knowledge.Store` in its own package under `internal/`. It decides what an id looks like, and returns content as markdown.
2. Register it by type name in `internal/source`.
3. Run `knowledgetest.Run` from its tests to check the contract: stable list order, no content in `list`, `ErrNotFound` for unknown ids, ranked search, and an empty result rather than an error when nothing matches.

## Commands

```bash
knowtorious list [--type decisions]       # ids and titles, never content
knowtorious get ID [ID...]                # print entries
knowtorious search QUERY... [--limit 10]  # entries containing every term
```

Each command accepts `--json`. Exit codes: `0` success, `1` failure (e.g. unknown id, no vault), `2` usage error.
