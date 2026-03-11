# archivist

A drop-in replacement for [adr-tools](https://github.com/npryce/adr-tools)
that adds a richer CLI and interactive TUI for managing Architecture Decision
Records. Works directly on existing adr-tools repositories without migration.

## Install

```bash
go install github.com/ramtinJ95/archivist/cmd/archivist@latest
```

Or build from source:

```bash
git clone https://github.com/ramtinJ95/archivist.git
cd archivist
go build -o archivist ./cmd/archivist
```

## Quick start

```bash
# Initialize a new ADR repository
archivist init

# Create a new ADR (opens $EDITOR after creation)
archivist new Use PostgreSQL for persistence

# List all ADRs
archivist list

# Show a specific ADR by number, filename, or partial match
archivist show 2
archivist show postgresql

# Supersede an existing ADR
archivist new -s 2 Use SQLite instead

# Link two ADRs with custom labels
archivist link 3 1 "Amends" "Amended by"

# Search across all ADRs
archivist search "database"

# Validate all ADRs for common issues
archivist validate

# Generate a table of contents
archivist generate toc

# Generate a Graphviz dependency graph
archivist generate graph | dot -Tpng -o graph.png

# Upgrade date format from DD/MM/YYYY to ISO 8601
archivist upgrade-repository
```

## Interactive TUI

```bash
archivist tui
```

Split-pane interface with ADR list and content preview:

| Key       | Action                    |
|-----------|---------------------------|
| j/k       | Navigate list             |
| Enter     | Full-screen detail view   |
| /         | Filter by title           |
| n         | Create new ADR wizard     |
| e         | Edit selected in $EDITOR  |
| s         | Supersede selected ADR    |
| l         | Link selected ADR         |
| ?         | Help                      |
| q         | Quit                      |

## adr-tools compatibility

Archivist discovers ADR directories the same way adr-tools does:

1. Walk upward from CWD looking for `.adr-dir`
2. Fall back to `doc/adr/` if present
3. Default to `doc/adr/` for new repositories

All file naming, templates, link formatting, and status mutations match
upstream `adr-tools` behavior, including the historical `Superceded`
spelling.

Custom templates are supported via `ADR_TEMPLATE` env var or
`<adr-dir>/templates/template.md`.

## Development

```bash
go test ./...
```
