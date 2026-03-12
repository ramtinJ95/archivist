# Archivist Implementation Plan

Status: approved for implementation planning as of 2026-03-11

This document is the authoritative implementation plan for Archivist.
It is intended to be self-contained: the team should be able to implement
the initial product from this document alone without needing to reconstruct
prior chat context.

## 1. Product Decision

Archivist will be a workflow-first hybrid product:

- a scriptable Go CLI
- a Bubble Tea TUI for the highest-value interactive workflows
- a shared compatibility-first core that operates directly on ADR repositories

The implementation strategy is a parity-first commit train:

- preserve the `adr-tools` repository contract first
- prove that Archivist can operate safely inside existing `adr-tools` repos
- only then add Archivist-specific workflow improvements

This is the central product rule:

Archivist must work directly on repositories that already use `adr-tools`
without requiring an import, migration, or file layout conversion.

If a user runs `archivist` from a repo or subdirectory where `adr-tools` has
already been used, Archivist must discover the existing ADR log and continue
working with it in place.

## 2. Goals

### Primary goals

- Be a drop-in repository-level replacement for `adr-tools`
- Preserve `adr-tools` file layout, discovery rules, numbering, templates,
  and generated outputs closely enough that existing repos continue to work
- Provide a richer CLI than `adr-tools` for reading, searching, validating,
  and editing decision logs
- Provide a TUI for browsing, previewing, creating, superseding, and linking
  ADRs
- Keep one shared domain core so CLI and TUI use exactly the same semantics

### Non-goals for the first release train

- No forced migration of existing ADR files
- No reformatting pass that rewrites all ADR documents into a new style
- No alternate storage backend; Markdown files remain the source of truth
- No TUI-only mutations that bypass core compatibility logic
- No speculative metadata sidecars or database indexes in v1

## 3. Product Principles

### 3.1 Compatibility is the default, not a mode

Archivist should not require a `--compat` flag to read `adr-tools` repos.
Compatibility must be the normal behavior of the tool.

### 3.2 Improvements must be additive

Archivist-specific improvements should add commands and workflows, not change
the underlying ADR repository contract unless explicitly versioned later.

### 3.3 Mutate narrowly

Commands that update ADRs must modify only the intended sections and lines.
Archivist should not rewrite entire files if a targeted section update is
sufficient.

### 3.4 One core, two frontends

Filesystem discovery, parsing, numbering, template resolution, linking,
superseding, generation, and validation must live in a shared core package.
CLI and TUI are thin frontends over that core.

### 3.5 Keep each commit reviewable

The work must be split into small, coherent commits. Every commit should leave
the tree in a buildable, testable state. Avoid one large feature dump.

## 4. Compatibility Contract With `adr-tools`

This section defines the behaviors Archivist must preserve for repository-level
compatibility.

### 4.1 ADR directory discovery

Archivist must preserve the upstream ADR directory lookup behavior:

- Start from the current working directory
- Walk upward toward `/`
- If `.adr-dir` is found, read its contents and use that relative path
- Otherwise, if `doc/adr` exists at that level, use `doc/adr`
- If nothing is found all the way to `/`, fall back to `doc/adr`

Important details:

- Returned paths must behave like upstream paths, which are relative to the
  caller's current directory rather than normalized absolute paths
- Running inside nested subdirectories of an ADR repo must still locate the
  ADR directory correctly
- A repo created by `adr-tools` with `.adr-dir` must continue to work without
  any conversion step

Examples:

- Running from the repo root may resolve to `doc/adr`
- Running from `services/foo` may resolve to `../../doc/adr`
- Running from a repo with `.adr-dir` pointing at `architecture-log` must use
  that location

### 4.2 `init` semantics

Archivist must preserve the essential behavior of `adr init`:

- `init` with no argument uses `doc/adr`
- `init <dir>` creates that directory and writes the literal path to `.adr-dir`
- `init` creates the initial ADR using the built-in init template
- `init` must not open an interactive editor for that first ADR

Compatibility note:

Upstream achieves the non-interactive behavior by forcing `VISUAL=true` and
reusing `adr new`. Archivist does not need to mimic the shell implementation,
but it must match the resulting behavior.

### 4.3 Template precedence

Archivist must preserve template precedence:

1. `ADR_TEMPLATE` environment variable
2. `<adr-dir>/templates/template.md`
3. bundled default template

The initial ADR created by `init` uses a separate bundled init template.

### 4.4 Template substitution behavior

The compatibility path should preserve the upstream placeholder model:

- `NUMBER`
- `TITLE`
- `DATE`
- `STATUS`

These should be replaced literally when generating a new ADR from a template.

This does not prevent Archivist from gaining richer templating later, but the
compatibility path must preserve the simple existing behavior.

### 4.5 Numbering and filename generation

Archivist must preserve the upstream numbering and filename generation rules:

- Determine the next ADR number as `max(existing leading digits) + 1`
- Format the visible filename number as four digits, such as `0001`
- Use a lowercase slug
- Collapse runs of non-alphanumeric characters into `-`
- Trim non-alphanumeric characters from the start and end

The next created ADR in an existing `adr-tools` repo must receive the same
filename that upstream would have produced.

### 4.6 ADR reference resolution

Archivist must preserve upstream reference resolution closely enough for user
habits to carry over:

- ADR references may be given as a number or partial filename
- The first matching entry wins
- Existing commands like supersede and link must accept the same reference
  shapes users already rely on

Implementation note:

The core may internally implement this in a cleaner way than `grep | head -1`,
but the externally observable matching behavior should stay compatible.

### 4.7 Default Markdown shape

Archivist must preserve the default ADR shape for newly created compatibility
ADRs:

```md
# NUMBER. TITLE

Date: DATE

## Status

STATUS

## Context

...

## Decision

...

## Consequences

...
```

The bundled init ADR must also preserve the same semantic shape as upstream.

### 4.8 Supersede semantics

`new -s <ref>` must preserve the upstream mutation behavior:

- Add `Superceded by [...]` to the old ADR
- Remove only the literal status line `Accepted` from the old ADR
- Add `Supercedes [...]` to the new ADR

Important compatibility detail:

The upstream strings are misspelled as `Superceded` and `Supercedes`. Archivist
should preserve those exact default relation labels in compatibility flows so
existing repos and generated outputs remain consistent.

### 4.9 Link semantics

Both manual linking and inline linking during creation must preserve the
upstream shape:

- `link SOURCE LINK TARGET REVERSE-LINK` adds reciprocal relation lines
- `new -l target:link:reverse-link` does the same while creating the ADR
- Insert relation lines in the `## Status` section
- Place them before the next `##` heading
- Keep a blank line after insertion

### 4.10 Status section parsing

Archivist should assume the same mutation anchor used by upstream:

- the status section heading is exactly `## Status`

Compatibility consequence:

- reading can be tolerant where safe
- mutation commands must not guess if the expected heading is absent
- Archivist-specific `validate` should report malformed or nonstandard files

### 4.11 `list` semantics

Archivist must preserve the user-visible `list` behavior:

- list only files matching the ADR filename pattern
- sort lexicographically
- error if the discovered ADR directory does not exist

### 4.12 `generate toc` semantics

Archivist must preserve the generated TOC shape:

- exact heading `# Architecture Decision Records`
- one bullet per ADR
- bullet title derived from the ADR title
- bullet target uses the file basename
- support intro, outro, and link-prefix options

### 4.13 `generate graph` semantics

Archivist must preserve the generated Graphviz semantics:

- default extension `.html`
- `-p` prefixes URLs
- `-e` overrides link extension
- nodes use `_N` numbering
- chronological dotted edges connect adjacent ADRs
- relation edges use `weight=0`
- reverse-link labels ending in ` by` are excluded from relation edges

### 4.14 Environment variables and help behavior

Archivist should preserve the most important environment behavior:

- `VISUAL` takes precedence over `EDITOR`
- if neither is set, creation still succeeds without opening a real editor
- `ADR_DATE` overrides the date string used when creating ADRs
- `ADR_TEMPLATE` overrides template selection
- `ADR_PAGER` should take precedence over `PAGER` for compatibility help flows

### 4.15 Upgrade behavior

Archivist must support the existing repository upgrade behavior:

- convert `Date: DD/MM/YYYY` to `Date: YYYY-MM-DD`

Upstream only documents and implements that specific date upgrade path.
Archivist should preserve that compatibility behavior before adding broader
upgrade or linting capabilities.

## 5. Scope Of Archivist-Specific Additions

Archivist is not intended to stop at parity.
After the compatibility core is in place, the tool adds higher-level workflows
that are useful but do not break the repository contract.

These additions are in scope:

- `show` for a focused ADR view
- `search` across titles and contents
- `validate` for repository health checks
- `edit` as a convenience wrapper around editor launch
- `tui` as the primary interactive decision-log workflow

These additions must remain additive:

- they must operate on the same discovered ADR repository
- they must reuse the same core parsing and mutation logic
- they must not create a second on-disk source of truth

## 6. Command Surface

### 6.1 Compatibility commands

These commands preserve or closely mirror the upstream `adr-tools` behaviors.

#### `archivist init [DIRECTORY]`

Behavior:

- initialize an ADR repository in `doc/adr` or the provided directory
- create `.adr-dir` if a custom directory is provided
- create the initial ADR using the bundled init template
- do not launch an interactive editor

#### `archivist new [-s SUPERCEDED] [-l TARGET:LINK:REVERSE-LINK] TITLE...`

Behavior:

- discover ADR directory
- determine next ADR number
- create slugged filename
- apply template and substitutions
- apply supersede and link mutations if requested
- launch the editor according to `VISUAL` / `EDITOR`
- print the resulting path

#### `archivist link SOURCE LINK TARGET REVERSE-LINK`

Behavior:

- resolve both ADR references
- add reciprocal relation lines to both ADRs

#### `archivist list`

Behavior:

- list discovered ADR files in compatibility order

#### `archivist generate toc [-i INTRO] [-o OUTRO] [-p LINK_PREFIX]`

Behavior:

- write Markdown TOC to stdout

#### `archivist generate graph [-p LINK_PREFIX] [-e LINK_EXTENSION]`

Behavior:

- write Graphviz DOT to stdout

#### `archivist upgrade-repository`

Behavior:

- run the compatibility date-format upgrade

Compatibility note:

The exact upstream command name should be supported.
Archivist may also offer `upgrade` as an alias, but `upgrade-repository`
should exist for parity.

### 6.2 Archivist additive commands

These commands are intentionally better-product additions and are not required
for strict upstream parity, but they are in scope for the chosen direction.

#### `archivist show <REF>`

Behavior:

- resolve an ADR by compatibility reference rules
- print a focused view of the ADR or its raw contents

#### `archivist edit <REF>`

Behavior:

- resolve an ADR by compatibility reference rules
- open it in `VISUAL` / `EDITOR`

#### `archivist search <QUERY>`

Behavior:

- search titles and file contents
- output match references in a scriptable format

#### `archivist validate`

Behavior:

- check repository structure and ADR discoverability
- report missing `## Status` headings
- report broken relation targets
- report invalid numbering or malformed ADR filenames
- report duplicate or ambiguous references where practical

#### `archivist tui`

Behavior:

- launch the interactive terminal UI using the same repository discovery and
  mutation core as the CLI

### 6.3 Help and completion

Archivist should provide:

- Cobra-generated help for the command tree
- shell completions
- concise examples for compatibility commands

Shell completion support is useful but should not delay the compatibility core.

## 7. TUI Requirements

### 7.1 TUI product role

The TUI is not a separate product. It is the primary human-facing workflow
surface built on top of the same core used by the CLI.

### 7.2 Initial TUI scope

The first TUI milestone should be browse-first and safe:

- ADR list
- detail preview
- search/filter
- open current ADR in editor

The second TUI milestone adds write workflows:

- create ADR wizard
- supersede selected ADR
- link ADRs

### 7.3 TUI screens

#### Decision log view

- left pane: ADR list
- right pane: current ADR preview
- visible metadata: number, title, date, status links

#### Search/filter view

- fuzzy or substring search over titles and optionally contents
- keyboard-driven filtering

#### New ADR wizard

- title entry
- optional supersede targets
- optional relation targets
- preview of resulting filename and destination path

#### Link flow

- choose source ADR
- choose relation label
- choose target ADR
- choose reverse label

#### Generate/export actions

- TOC preview or export
- graph output preview or export path

### 7.4 TUI keybindings

Initial recommended bindings:

- `j` / `k` or arrow keys: move
- `enter`: open or focus detail
- `/`: search
- `n`: new ADR
- `s`: supersede current ADR
- `l`: create link
- `e`: open in editor
- `g`: generation actions
- `?`: help
- `q`: back or quit

These bindings are recommendations, not compatibility requirements.

### 7.5 TUI safety rules

- all writes must go through the shared core
- preview changes before destructive or multi-file actions
- surface validation failures directly in the UI
- do not maintain TUI-only file representations

## 8. Exact Package Layout

The repo should use this package layout:

```text
cmd/archivist/main.go

internal/adrlog/
  discover.go
  discover_test.go
  list.go
  list_test.go
  refs.go
  refs_test.go
  templates.go
  templates_test.go
  create.go
  create_test.go
  mutate.go
  mutate_test.go
  parse.go
  parse_test.go
  generate.go
  generate_test.go
  upgrade.go
  upgrade_test.go
  repository.go
  types.go

internal/editor/
  editor.go
  editor_test.go

internal/cli/
  root.go
  init.go
  new.go
  link.go
  list.go
  generate.go
  upgrade.go
  show.go
  edit.go
  search.go
  validate.go
  tui.go

internal/tui/
  app.go
  model.go
  styles.go
  screen_list.go
  screen_detail.go
  screen_new.go
  screen_link.go
  screen_search.go
  actions.go

internal/testutil/
  fixtures.go
  golden.go
  repo.go

testdata/
  compat/
    init/
    create/
    list/
    link/
    supersede/
    generate/
    upgrade/
    discover/
  tui/
    ...

docs/
  implementation-plan.md
```

Note: the TUI file listing above is aspirational. The implemented layout
consolidates into fewer files (`tui.go`, `item.go`, `styles.go`, `wizard.go`)
which is acceptable as long as the package responsibilities in section 8.1 are
preserved.

### 8.1 Package responsibilities

#### `internal/adrlog`

This is the authoritative domain core.
It owns:

- repo discovery
- ADR listing
- ADR reference resolution
- template resolution
- creation
- targeted file mutations
- generators
- upgrade behavior
- validation helpers

It must not depend on Cobra or Bubble Tea.

#### `internal/editor`

This package owns external editor launch behavior and environment precedence.

#### `internal/cli`

This package owns Cobra command definitions, help text, flag parsing, and
presentation formatting. It must call into `internal/adrlog`.

#### `internal/tui`

This package owns Bubble Tea models, views, update handlers, and user
interaction flows. It must call into `internal/adrlog`.

#### `internal/testutil`

This package owns fixture setup, temp repo creation, and golden test helpers.

## 9. Core Data Model

The core should model at least the following concepts:

- ADR repository
- ADR record
- ADR reference
- relation entry
- template source
- generated report output
- validation issue

Suggested shapes:

```go
type Repository struct {
    CWD        string
    ADRDir     string
    RootHint   string
}

type Record struct {
    Number   int
    Filename string
    Path     string
    Title    string
    Date     string
    Status   []string
    Content  string
}

type Relation struct {
    SourceRef    string
    TargetRef    string
    ForwardLabel string
    ReverseLabel string
}

type ValidationIssue struct {
    Path     string
    Severity string
    Message  string
}
```

These are conceptual only. The exact field names can evolve if the semantics do
not.

## 10. File Mutation Rules

Archivist must not do broad Markdown reserialization for compatibility flows.

Rules:

- preserve the existing file as much as possible
- use targeted section edits when adding links or removing `Accepted`
- avoid changing headings, spacing, or unrelated sections
- use atomic write patterns: temp file plus rename
- keep write behavior deterministic for golden tests

If an ADR does not contain an exact `## Status` heading:

- compatibility mutations should fail clearly rather than invent a new layout
- `validate` should report the issue

## 11. Testing Strategy

### 11.1 Compatibility-first testing

The test strategy must mirror the parity-first implementation strategy.

First-class tests:

- discovery tests
- filename and slug tests
- template precedence tests
- init tests
- new tests
- link tests
- supersede tests
- list tests
- generator tests
- upgrade tests

### 11.2 Upstream-fixture mindset

Where possible, create fixtures and golden outputs that match the repository
shapes and command outcomes of `adr-tools`.

The intent is not to copy the shell implementation. The intent is to preserve
the observable repository behavior.

### 11.3 Test layers

#### Unit tests

- slug generation
- reference resolution
- template precedence
- graph and TOC generation

#### Integration tests

- full temp-repo command flows against fixture repos
- nested-directory discovery
- editor environment precedence
- supersede and link mutations across multiple files

#### Golden tests

- generated TOC
- generated graph
- resulting ADR file content after create/link/supersede

### 11.4 CI expectations

Each commit in the train should keep these green:

- `go test ./...`
- `go build ./cmd/archivist`

If TUI tests require special handling, they should still run in a noninteractive
mode in CI.

## 12. Milestones

### Milestone 0: Bootstrap

Objective:

- create the Go module, command entrypoint, and test harness

Exit criteria:

- `go build ./cmd/archivist`
- `go test ./...`
- fixture helpers exist

### Milestone 1: Compatibility discovery and listing core

Objective:

- discover ADR repos exactly like upstream
- list ADR files
- resolve ADR references

Exit criteria:

- running in nested directories works
- list output matches compatibility expectations
- reference resolution supports numbers and partial filenames

### Milestone 2: Compatibility creation core

Objective:

- create ADRs with correct numbering, slugging, template precedence, and editor
  behavior

Exit criteria:

- `init` works
- `new` works
- initial ADR shape matches plan
- custom template precedence works

### Milestone 3: Compatibility mutations and generators

Objective:

- support linking, superseding, TOC generation, graph generation, and upgrade

Exit criteria:

- compatibility golden tests pass for link/supersede/generate/upgrade

### Milestone 4: Compatibility CLI

Objective:

- wire compatibility commands through Cobra

Exit criteria:

- `init`, `new`, `link`, `list`, `generate`, and `upgrade-repository` are
  available and tested through the CLI surface

### Milestone 5: Archivist workflow commands

Objective:

- add `show`, `edit`, `search`, and `validate`

Exit criteria:

- commands work against existing `adr-tools` repos without migration

### Milestone 6: Read-only TUI

Objective:

- provide browse, preview, search, and open-in-editor

Exit criteria:

- TUI launches in existing ADR repos
- browse and preview operate on discovered ADRs

### Milestone 7: Write-capable TUI

Objective:

- create ADRs, supersede ADRs, and link ADRs interactively

Exit criteria:

- TUI writes produce the same repository results as CLI core operations

## 13. Commit Train

The implementation must be split into small, coherent commits.
Recommended commit sequence:

### Commit 1

`chore: bootstrap Go module, command entrypoint, and test harness`

Include:

- `go.mod`
- `cmd/archivist/main.go`
- test helpers
- empty or minimal Cobra root

Do not include:

- compatibility logic
- TUI

### Commit 2

`feat(core): discover ADR repositories and list records`

Include:

- discovery logic
- ADR listing
- discovery and list tests

### Commit 3

`feat(core): resolve ADR references and read ADR metadata`

Include:

- reference resolution
- title and status extraction helpers
- metadata parsing tests

### Commit 4

`feat(core): create ADRs with compatible numbering and templates`

Include:

- slugging
- next-number logic
- template precedence
- `new`
- create-path tests

### Commit 5

`feat(core): initialize ADR repositories and support editor handoff`

Include:

- `init`
- init template
- editor launch precedence
- init and editor tests

### Commit 6

`feat(core): add link and supersede mutations`

Include:

- targeted status-section mutations
- reciprocal links
- supersede behavior
- mutation golden tests

### Commit 7

`feat(core): generate toc, graph, and repository upgrades`

Include:

- TOC generator
- graph generator
- date upgrade
- generator and upgrade tests

### Commit 8

`feat(cli): wire compatibility commands through Cobra`

Include:

- `init`
- `new`
- `link`
- `list`
- `generate`
- `upgrade-repository`
- help text

### Commit 9

`feat(cli): add Archivist workflow commands`

Include:

- `show`
- `edit`
- `search`
- `validate`

### Commit 10

`feat(tui): add read-only decision log browser`

Include:

- list view
- detail preview
- search/filter
- open-in-editor

### Commit 11

`feat(tui): add ADR creation and relationship workflows`

Include:

- new ADR wizard
- supersede flow
- link flow

### Commit 12

`docs: refresh README and usage guidance`

Include:

- README alignment
- examples for existing `adr-tools` repos
- developer notes if needed

## 14. Definition Of Done

Archivist v1 of this plan is done when:

- it can be run inside an existing `adr-tools` repo and discover the ADR log
- compatibility commands operate on that repo without migration
- additive commands operate on the same repo safely
- TUI browse and write workflows reuse the same core
- the command tree is documented
- the compatibility test suite passes

## 15. Risks And Guardrails

### Risk: accidental incompatibility through cleanup

Guardrail:

- treat upstream oddities as part of the compatibility contract
- do not "clean up" spellings or matching semantics in the compatibility path

### Risk: broad file rewrites

Guardrail:

- use targeted text mutations
- add golden tests for all mutation commands

### Risk: CLI and TUI drift

Guardrail:

- keep all repo semantics inside `internal/adrlog`
- forbid business logic in CLI or TUI packages

### Risk: custom templates break mutations

Guardrail:

- require exact `## Status` for compatibility mutations
- report issues via `validate`
- do not attempt heuristic rewrites in v1

## 16. Optional Future Work Beyond This Plan

These are intentionally out of scope for the initial train:

- packaging a compatibility symlink or alternate binary name `adr`
- richer template engines
- repository dashboards or metrics
- multi-repo ADR search
- import/export beyond existing Markdown and generation flows

## 17. Upstream Behavior Reference Summary

This plan is based on the observed behavior of upstream `adr-tools` rather than
just its README summary.

The most important upstream artifacts reviewed were:

- `README.md`
- `src/adr-init`
- `src/adr-new`
- `src/adr-link`
- `src/adr-list`
- `src/adr-upgrade-repository`
- `src/_adr_dir`
- `src/_adr_file`
- `src/_adr_links`
- `src/_adr_add_link`
- `src/_adr_remove_status`
- `src/_adr_status`
- `src/_adr_title`
- `src/_adr_generate_toc`
- `src/_adr_generate_graph`
- `src/template.md`
- `src/init.md`
- upstream expected-output tests for discovery, creation, linking,
  superseding, generation, upgrade, and slug edge cases

Official upstream repo:

- https://github.com/npryce/adr-tools

Relevant Go ecosystem docs used during planning:

- Cobra user guide:
  https://github.com/spf13/cobra/blob/main/site/content/user_guide.md
- Bubble Tea basics:
  https://github.com/charmbracelet/bubbletea/blob/main/tutorials/basics/README.md
- Bubbles:
  https://github.com/charmbracelet/bubbles
- Lip Gloss:
  https://github.com/charmbracelet/lipgloss
