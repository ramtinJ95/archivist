package testutil

import (
	"path/filepath"
	"testing"
)

func SeedADR(t *testing.T, adrDir string, filename, content string) string {
	t.Helper()
	path := filepath.Join(adrDir, filename)
	WriteFile(t, path, content)
	return path
}

const SampleADR1 = `# 1. Record architecture decisions

Date: 2024-01-15

## Status

Accepted

## Context

We need to record the architectural decisions made on this project.

## Decision

We will use Architecture Decision Records, as described by Michael Nygard.

## Consequences

See Michael Nygard's article, linked above.
`

const SampleADR2 = `# 2. Use Go for implementation

Date: 2024-01-16

## Status

Accepted

## Context

We need a compiled language with good CLI support.

## Decision

We will use Go.

## Consequences

The team needs Go experience.
`

const SampleADR3 = `# 3. Use Cobra for CLI

Date: 2024-01-17

## Status

Accepted

## Context

We need a CLI framework.

## Decision

We will use Cobra.

## Consequences

Cobra is the de facto standard for Go CLIs.
`
