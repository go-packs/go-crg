package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Snippet represents a fragment of source code with context.
type Snippet struct {
	FilePath  string `json:"file_path"`
	Content   string `json:"content"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
}

// ExtractSnippet reads a file and returns the lines between start and end (1-indexed).
// It adds 2 lines of context above and below if possible.
func ExtractSnippet(filePath string, start, end int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Add context lines
	ctxStart := start - 2
	if ctxStart < 1 {
		ctxStart = 1
	}
	ctxEnd := end + 2

	var result []string
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		if lineNum >= ctxStart && lineNum <= ctxEnd {
			prefix := "  "
			if lineNum >= start && lineNum <= end {
				prefix = "> " // Highlight the actual changed lines
			}
			result = append(result, fmt.Sprintf("%d%s%s", lineNum, prefix, scanner.Text()))
		}
		if lineNum > ctxEnd {
			break
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(result, "\n"), nil
}
