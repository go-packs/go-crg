package parser

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CodeParser defines the interface for language-specific parsing.
type CodeParser interface {
	DetectLanguage(path string) (string, bool)
	ParseFile(path string) ([]Node, []Edge, error)
}

// ExtToLanguage maps file extensions to language names.
var ExtToLanguage = map[string]string{
	".py":   "python",
	".go":   "go",
	".js":   "javascript",
	".ts":   "typescript",
	".tsx":  "tsx",
	".rs":   "rust",
	".java": "java",
}

// FileHash computes the SHA-256 hash of a file's content.
func FileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// DetectLanguage identifies the language based on file extension.
func DetectLanguage(path string) (string, bool) {
	ext := filepath.Ext(path)
	lang, ok := ExtToLanguage[ext]
	return lang, ok
}
