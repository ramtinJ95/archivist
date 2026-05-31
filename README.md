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

## v1 release scope

Archivist v1 is intended for day-to-day use in both existing
`adr-tools` repositories and newly initialized ADR repositories. The release
scope is:

- repository-level `adr-tools` compatibility as the default behavior
- CLI workflows for init, create, list, show, search, edit, link, supersede,
  validate, generate, and upgrade
- an interactive TUI for browse, preview, edit, create, supersede, link,
  validate, and generate/export workflows
- CI and local release checks for unit tests, race tests, binary builds,
  `adr-tools` compatibility, and scripted TUI smoke coverage
- tagged GitHub release automation with versioned macOS/Linux binaries and
  checksums

Supported platforms:

- macOS and Linux are the official v1 release targets
- Windows is not yet supported because editor and pager execution still rely on
  `sh -c`

## Install

### Option 1: Install with Homebrew

On macOS or Linux, the recommended install path is the Archivist Homebrew tap:

```bash
brew install ramtinJ95/tap/archivist
```

Then verify the installed binary:

```bash
archivist version
```

To upgrade later:

```bash
brew upgrade archivist
```

### Option 2: Download a release binary

Tagged binaries and checksums are published on the
[GitHub Releases](https://github.com/ramtinJ95/archivist/releases) page.
Choose the archive for your platform:

| Platform | Archive |
|---|---|
| macOS Apple Silicon | `archivist_1.0.1_darwin_arm64.tar.gz` |
| macOS Intel | `archivist_1.0.1_darwin_amd64.tar.gz` |
| Linux x86-64 | `archivist_1.0.1_linux_amd64.tar.gz` |
| Linux ARM64 | `archivist_1.0.1_linux_arm64.tar.gz` |

Then unpack and place the binary on your `PATH`:

```bash
tar -xzf archivist_1.0.1_<os>_<arch>.tar.gz
sudo install -m 0755 archivist /usr/local/bin/archivist
archivist version
```

The release page also publishes `checksums.txt` for verifying downloads.

### Option 3: Install with Go

For reproducible installs, prefer an exact tag:

```bash
go install github.com/ramtinJ95/archivist/cmd/archivist@v1.0.1
```

To track the newest published version:

```bash
go install github.com/ramtinJ95/archivist/cmd/archivist@latest
```

Make sure your Go binary directory is on `PATH`. By default this is usually
`$HOME/go/bin` unless `GOBIN` is set.

### Option 4: Build from source

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
navigable validation report, and wizard flows with previews for create,
supersede, link, and generate/export operations. TUI create flows can add
supersede and relation links, then hand the created ADR to `$VISUAL` or
`$EDITOR` when configured.

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
| `g` | Generate TOC or graph preview/export |
| `?` | Open help |
| `q` | Quit |

In the generate screen, `t`/`d` preview TOC or graph output and `T`/`D` open
export wizards for writing TOC or graph files with generation options. In the
validation report, `j`/`k` select issues, `Enter` shows the affected ADR, and
`e` opens the affected ADR in your editor when possible.

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

1. Run the release checks:

   ```bash
   go test ./...
   go vet ./...
   go test -race -count=1 ./...
   go build ./cmd/archivist
   ./scripts/adr-tools-compare.sh
   ./scripts/tui-smoke.sh
   ```

2. Push the release tag:

   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```

3. Let the release workflow publish binaries and checksums
4. Verify a tagged binary reports the expected value:

   ```bash
   archivist version
   # archivist vX.Y.Z
   ```

## Support

- Use [GitHub Issues](https://github.com/ramtinJ95/archivist/issues) for bugs,
  feature requests, and launch feedback
- Keep behavior and compatibility questions anchored to
  [docs/implementation-plan.md](docs/implementation-plan.md)

## Development

```bash
go test ./...
go vet ./...
go test -race -count=1 ./...
go build ./cmd/archivist
```

To compare core compatibility behavior against upstream `adr-tools`:

```bash
./scripts/adr-tools-compare.sh
```

Set `ADR_TOOLS_DIR` to reuse an existing checkout, or let the script clone a
temporary copy of `npryce/adr-tools`.

To run an end-to-end scripted smoke test of the interactive TUI workflows:

```bash
./scripts/tui-smoke.sh
```

Set `KEEP_TUI_SMOKE_TMP=1` to keep the generated temporary repository for
inspection after the smoke test finishes.
