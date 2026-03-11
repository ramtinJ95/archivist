package adrlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const DefaultTemplate = `# NUMBER. TITLE

Date: DATE

## Status

STATUS

## Context

## Decision

## Consequences
`

const InitTemplate = `# 1. Record architecture decisions

Date: DATE

## Status

Accepted

## Context

We need to record the architectural decisions made on this project.

## Decision

We will use Architecture Decision Records, as [described by Michael Nygard](http://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions).

## Consequences

See Michael Nygard's article, linked above. For a lightweight ADR toolset, see Nat Pryce's [adr-tools](https://github.com/npryce/adr-tools).
`

func ResolveTemplate(adrAbsDir string) (string, error) {
	if envTpl := os.Getenv("ADR_TEMPLATE"); envTpl != "" {
		data, err := os.ReadFile(envTpl)
		if err != nil {
			return "", fmt.Errorf("read ADR_TEMPLATE %q: %w", envTpl, err)
		}
		return string(data), nil
	}

	repoTpl := filepath.Join(adrAbsDir, "templates", "template.md")
	if data, err := os.ReadFile(repoTpl); err == nil {
		return string(data), nil
	}

	return DefaultTemplate, nil
}

func ApplyTemplate(template string, number int, title, date, status string) string {
	numStr := strconv.Itoa(number)

	result := template
	result = strings.ReplaceAll(result, "NUMBER", numStr)
	result = strings.ReplaceAll(result, "TITLE", title)
	result = strings.ReplaceAll(result, "DATE", date)
	result = strings.ReplaceAll(result, "STATUS", status)
	return result
}

func PadNumber(n int) string {
	return fmt.Sprintf("%04d", n)
}
