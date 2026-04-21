package adrlog

import (
	"fmt"
	"os"
	"strings"
)

func addStatusLineContent(content string, line string) (string, error) {
	loc := statusHeadingPattern.FindStringIndex(content)
	if loc == nil {
		return "", fmt.Errorf("no ## Status heading found")
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

	return before + "\n\n" + line + "\n\n" + after, nil
}

func addStatusLine(path string, line string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent, err := addStatusLineContent(string(data), line)
	if err != nil {
		return fmt.Errorf("%w in %s", err, path)
	}

	return atomicWriteFile(path, []byte(newContent))
}

func removeStatusLineContent(content string, line string) (string, error) {
	loc := statusHeadingPattern.FindStringIndex(content)
	if loc == nil {
		return "", fmt.Errorf("no ## Status heading found")
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
	return content[:loc[1]] + newSection + content[sectionEnd:], nil
}

func removeStatusLine(path string, line string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent, err := removeStatusLineContent(string(data), line)
	if err != nil {
		return fmt.Errorf("%w in %s", err, path)
	}

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
	revLine := fmt.Sprintf("%s [%d. %s](%s)", reverseLabel, sourceRec.Number, sourceRec.Title, sourceBase)

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	sourceUpdated, err := addStatusLineContent(string(sourceData), fwdLine)
	if err != nil {
		return fmt.Errorf("%w in %s", err, sourcePath)
	}

	if sourcePath == targetPath {
		sourceUpdated, err = addStatusLineContent(sourceUpdated, revLine)
		if err != nil {
			return fmt.Errorf("%w in %s", err, targetPath)
		}
		return atomicWriteFile(sourcePath, []byte(sourceUpdated))
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return err
	}

	targetUpdated, err := addStatusLineContent(string(targetData), revLine)
	if err != nil {
		return fmt.Errorf("%w in %s", err, targetPath)
	}

	if err := atomicWriteFile(sourcePath, []byte(sourceUpdated)); err != nil {
		return err
	}
	if err := atomicWriteFile(targetPath, []byte(targetUpdated)); err != nil {
		if rollbackErr := atomicWriteFile(sourcePath, sourceData); rollbackErr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", err, rollbackErr)
		}
		return err
	}

	return nil
}
