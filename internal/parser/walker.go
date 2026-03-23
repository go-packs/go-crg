package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
)

type ResultStore interface {
	SaveFileResults(filePath string, nodes []Node, edges []Edge) error
}

type Walker struct {
	repoRoot string
	parsers  map[string]CodeParser
	store    ResultStore
}

func NewWalker(repoRoot string, store ResultStore) *Walker {
	return &Walker{
		repoRoot: repoRoot,
		parsers: map[string]CodeParser{
			"python": NewPythonParser(),
			"go":     NewGoParser(),
		},
		store: store,
	}
}

func (w *Walker) BuildGraph(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	files := make(chan string)

	// 1. Find all parseable files
	g.Go(func() error {
		defer close(files)
		return filepath.Walk(w.repoRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if we have a parser for this language
			if _, ok := DetectLanguage(path); ok {
				select {
				case files <- path:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	})

	// 2. Parse files in parallel (e.g., 20 workers)
	for i := 0; i < 20; i++ {
		g.Go(func() error {
			for path := range files {
				lang, _ := DetectLanguage(path)
				p, ok := w.parsers[lang]
				if !ok {
					continue
				}

				nodes, edges, err := p.ParseFile(path)
				if err != nil {
					fmt.Printf("Error parsing %s: %v\n", path, err)
					continue
				}

				if err := w.store.SaveFileResults(path, nodes, edges); err != nil {
					return fmt.Errorf("failed to save results for %s: %w", path, err)
				}
			}
			return nil
		})
	}

	return g.Wait()
}
