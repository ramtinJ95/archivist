package adrlog

import (
	"fmt"
	"os"
	"strings"
)

func addStatusLine(path string, line string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	loc := statusHeadingPattern.FindStringIndex(content)
	if loc == nil {
		return fmt.Errorf("no ## Status heading found in %s", path)
	}

	afterHeading := content[loc[1]:]
	nextLoc := nextHeadingPattern.FindStringIndex(afterHeading)

	var insertPos int
	if nextLoc != nil {
		insertPos = loc[1] + nextLoc[0]
	} else {
		insertPos = len(content)
	}

	before := strings.TrimRight(content[:insertPos], "\n")
	after := content[insertPos:]

	newContent := before + "\n\n" + line + "\n\n" + after

	return atomicWriteFile(path, []byte(newContent))
}

func removeStatusLine(path string, line string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	loc := statusHeadingPattern.FindStringIndex(content)
	if loc == nil {
		return fmt.Errorf("no ## Status heading found in %s", path)
	}

	afterHeading := content[loc[1]:]
	nextLoc := nextHeadingPattern.FindStringIndex(afterHeading)

	var sectionEnd int
	if nextLoc != nil {
		sectionEnd = loc[1] + nextLoc[0]
	} else {
		sectionEnd = len(content)
	}

	section := content[loc[1]:sectionEnd]
	lines := strings.Split(section, "\n")

	var kept []string
	for _, l := range lines {
		if strings.TrimSpace(l) != line {
			kept = append(kept, l)
		}
	}

	newSection := strings.Join(kept, "\n")
	newContent := content[:loc[1]] + newSection + content[sectionEnd:]

	return atomicWriteFile(path, []byte(newContent))
}

func AddLink(sourcePath, targetPath, forwardLabel, reverseLabel string) error {
	sourceRec, err := ParseRecord(sourcePath)
	if err != nil {
		return err
	}

	targetRec, err := ParseRecord(targetPath)
	if err != nil {
		return err
	}

	sourceBase := extractFilename(sourcePath)
	targetBase := extractFilename(targetPath)

	fwdLine := fmt.Sprintf("%s [%d. %s](%s)", forwardLabel, targetRec.Number, targetRec.Title, targetBase)
	if err := addStatusLine(sourcePath, fwdLine); err != nil {
		return err
	}

	revLine := fmt.Sprintf("%s [%d. %s](%s)", reverseLabel, sourceRec.Number, sourceRec.Title, sourceBase)
	return addStatusLine(targetPath, revLine)
}
