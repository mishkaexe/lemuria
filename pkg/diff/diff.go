package diff

import (
	"fmt"
	"strings"
)

type Diff struct{}

type DiffResult struct {
	differences bool
	output      string
}

type diffLine struct {
	text string
	op   rune // ' ' for context, '-' for deletion, '+' for addition
}

func New() *Diff {
	return &Diff{}
}

func (d *Diff) CompareStrings(old, new string) (*DiffResult, error) {
	if old == new {
		return &DiffResult{
			differences: false,
			output:      "",
		}, nil
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	// Simple line-by-line diff algorithm
	diffs := d.computeDiff(oldLines, newLines)
	
	if len(diffs) == 0 {
		return &DiffResult{
			differences: false,
			output:      "",
		}, nil
	}

	output := d.formatUnifiedDiff(diffs, oldLines, newLines)
	return &DiffResult{
		differences: true,
		output:      output,
	}, nil
}

func (d *Diff) computeDiff(oldLines, newLines []string) []diffLine {
	var diffs []diffLine
	
	// Simple diff algorithm: compare line by line with context
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}
	
	oldIndex := 0
	newIndex := 0
	
	for oldIndex < len(oldLines) || newIndex < len(newLines) {
		if oldIndex < len(oldLines) && newIndex < len(newLines) {
			if oldLines[oldIndex] == newLines[newIndex] {
				// Lines match - add as context
				diffs = append(diffs, diffLine{text: oldLines[oldIndex], op: ' '})
				oldIndex++
				newIndex++
			} else {
				// Lines differ - check if it's a substitution or separate add/delete
				if d.findLineInSlice(oldLines[oldIndex], newLines[newIndex:]) != -1 {
					// Old line exists later in new - this is an addition
					diffs = append(diffs, diffLine{text: newLines[newIndex], op: '+'})
					newIndex++
				} else if d.findLineInSlice(newLines[newIndex], oldLines[oldIndex:]) != -1 {
					// New line exists later in old - this is a deletion
					diffs = append(diffs, diffLine{text: oldLines[oldIndex], op: '-'})
					oldIndex++
				} else {
					// Simple substitution
					diffs = append(diffs, diffLine{text: oldLines[oldIndex], op: '-'})
					diffs = append(diffs, diffLine{text: newLines[newIndex], op: '+'})
					oldIndex++
					newIndex++
				}
			}
		} else if oldIndex < len(oldLines) {
			// Remaining old lines are deletions
			diffs = append(diffs, diffLine{text: oldLines[oldIndex], op: '-'})
			oldIndex++
		} else {
			// Remaining new lines are additions
			diffs = append(diffs, diffLine{text: newLines[newIndex], op: '+'})
			newIndex++
		}
	}
	
	return diffs
}

func (d *Diff) findLineInSlice(line string, slice []string) int {
	for i, l := range slice {
		if l == line {
			return i
		}
	}
	return -1
}

func (d *Diff) formatUnifiedDiff(diffs []diffLine, oldLines, newLines []string) string {
	var result strings.Builder
	
	// Add unified diff header
	result.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", 1, len(oldLines), 1, len(newLines)))
	
	// Add diff lines
	for _, diff := range diffs {
		result.WriteString(fmt.Sprintf("%c%s\n", diff.op, diff.text))
	}
	
	return strings.TrimSuffix(result.String(), "\n")
}

func (d *Diff) CompareManifests(oldManifest, newManifest string) (*DiffResult, error) {
	// For now, treat manifests as regular strings
	// In the future, we could add YAML-aware comparison here
	return d.CompareStrings(oldManifest, newManifest)
}

func (d *Diff) CompareMultipleManifests(oldManifests, newManifests string) (*DiffResult, error) {
	// For now, treat multiple manifests as regular strings
	// In the future, we could parse and compare individual manifests
	return d.CompareStrings(oldManifests, newManifests)
}

func (r *DiffResult) HasDifferences() bool {
	return r.differences
}

func (r *DiffResult) String() string {
	return r.output
}