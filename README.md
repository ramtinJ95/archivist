# Archivist

Archivist is a production-focused ADR manager for teams that already use
[adr-tools](https://github.com/npryce/adr-tools) and want a better operator
experience without changing their repository format.

It works directly on existing ADR repositories, preserves the upstream file
contract, and adds a scriptable Go CLI plus a terminal UI for high-frequency
workflows like browsing, linking, superseding, validating, and generating
documentation.

## Why Archivist

- Drop-in compatibility with existing `adr-tools` repositories
- Faster day-to-day workflows through a native CLI and interactive TUI
- No migration step and no alternate metadata store
- Better validation and generation flows for operating ADRs as a living system
- Release automation and CI suitable for shipping tagged binaries

## Production readiness

Archivist is set up to ship as a product with:

- repository-level compatibility as the default behavior
- CI that runs tests, builds the CLI, and validates release config
- tagged GitHub release automation with versioned binaries and checksums
- version injection so `archivist version` reports the tagged release version

Current limitation:

- Official release targets are macOS and Linux. Windows is not yet a supported
  release target because editor and pager execution still rely on `sh -c`.

## Install

### Option 1: Download a release binary

Tagged binaries and checksums are published on the
[GitHub Releases](https://github.com/ramtinJ95/archivist/releases) page.

### Option 2: Install with Go

For reproducible installs, prefer an exact tag:

```bash
go install github.com/ramtinJ95/archivist/cmd/archivist@vX.Y.Z
```

To track the newest published version:

```bash
go install github.com/ramtinJ95/archivist/cmd/archivist@latest
```

### Option 3: Build from source

```bash
git clone https://github.com/ramtinJ95/archivist.git
cd archivist
go build -o archivist ./cmd/archivist
```

## Quick start

```bash
# Initialize a new ADR repository in doc/adr
archivist init

# Create a new ADR and open it in $VISUAL or $EDITOR
archivist new Use PostgreSQL for persistence

# Browse all ADRs
archivist list

# Show by number, filename, or partial filename
archivist show 2
archivist show use-postgresql

# Search across ADR content
archivist search "database"

# Validate ADR structure and references
archivist validate
```

## Command overview

| Command | Purpose |
|---|---|
| `archivist init [dir]` | Create a new ADR repository and seed the initial ADR |
| `archivist new TITLE...` | Create a new ADR and optionally open it in your editor |
| `archivist edit REF` | Open an existing ADR in your editor |
| `archivist list` | List ADRs in the discovered repository |
| `archivist show REF` | Print a full ADR, optionally through your pager |
| `archivist search PATTERN` | Search across ADR titles and content |
| `archivist link SOURCE LINK TARGET REVERSE-LINK` | Add reciprocal status links |
| `archivist validate` | Check ADRs for common structural issues |
| `archivist generate toc` | Generate a Markdown table of contents |
| `archivist generate graph` | Generate a DOT dependency graph |
| `archivist upgrade-repository` | Upgrade legacy date formatting |
| `archivist tui` | Launch the interactive terminal UI |
| `archivist version` | Print the current Archivist version |

## Common workflows

### Supersede an ADR

```bash
archivist new -s 2 Use SQLite instead
```

### Add reciprocal links

```bash
archivist link 3 "Amends" 1 "Amended by"
archivist new -l 1:Clarifies:Clarified-by Clarify rollout behavior
```

### Generate project documentation

```bash
archivist generate toc > doc/adr/README.md
archivist generate graph > doc/adr/graph.dot
dot -Tpng doc/adr/graph.dot -o doc/adr/graph.png
```

## Interactive TUI

Launch the TUI from anywhere inside an ADR repository:

```bash
archivist tui
```

The interface provides a split-pane list and preview, a full-detail view, a
validation report, and wizard flows with previews for create, supersede, link,
and generate operations.

| Key | Action |
|---|---|
| `j` / `k` or arrow keys | Navigate ADRs |
| `Enter` | Open full-detail view |
| `/` | Filter ADRs by title, path, or content |
| `n` | Create a new ADR |
| `e` | Edit the selected ADR |
| `s` | Supersede the selected ADR |
| `l` | Link the selected ADR |
| `v` | Open validation report |
| `g` | Generate TOC or graph |
| `?` | Open help |
| `q` | Quit |

## `adr-tools` compatibility

Archivist keeps the upstream repository contract intact:

1. It discovers ADR directories by walking upward for `.adr-dir`
2. It falls back to `doc/adr` when appropriate
3. It preserves filename numbering, slugging, templates, and generated output
4. It mutates link and supersede status lines using the same shape as
   `adr-tools`
5. It keeps the historical `Superceded` spelling for compatibility flows

Archivist is intended to operate in-place on existing repositories without a
conversion step.

The implementation contract for that behavior lives in
[docs/implementation-plan.md](docs/implementation-plan.md).

## Environment variables

| Variable | Purpose |
|---|---|
| `VISUAL` | Preferred editor command |
| `EDITOR` | Fallback editor command |
| `ADR_PAGER` | Preferred pager command for `show` |
| `PAGER` | Fallback pager command |
| `ADR_TEMPLATE` | Override template used for new ADRs |
| `ADR_DATE` | Override the generated ADR date |

## Release process

The repository includes:

- `.github/workflows/ci.yml` for test and build verification
- `.github/workflows/release.yml` for tag-driven GitHub releases
- `.goreleaser.yaml` for multi-archive packaging and checksum generation

To cut a release:

1. Push a semver tag such as `v0.1.0`
2. Let the release workflow publish binaries and checksums
3. Verify the tagged binary reports the expected value via `archivist version`

## Support

- Use [GitHub Issues](https://github.com/ramtinJ95/archivist/issues) for bugs,
  feature requests, and launch feedback
- Keep behavior and compatibility questions anchored to
  [docs/implementation-plan.md](docs/implementation-plan.md)

## Development

```bash
go test ./...
go build ./cmd/archivist
```

For a local TUI smoke test that drives the Bubble Tea interface through a
pseudo-terminal:

```bash
./scripts/tui-smoke.sh
```

This requires `expect`.
