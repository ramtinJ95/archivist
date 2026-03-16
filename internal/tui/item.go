package tui

import (
	"fmt"
	"strings"

	"github.com/ramtinJ95/archivist/internal/adrlog"
)

type ADRItem struct {
	record *adrlog.Record
}

func (i ADRItem) Title() string {
	return fmt.Sprintf("%d. %s", i.record.Number, i.record.Title)
}

func (i ADRItem) Description() string {
	return i.record.Date + " | " + strings.Join(i.record.Status, ", ")
}

func (i ADRItem) FilterValue() string {
	return i.record.Title
}
