package graph

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ajeet/go-crg/internal/parser"
	"github.com/ajeet/go-crg/internal/store"
)

func TestImpactAnalyzer(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Setup: A -> B -> C
	fileA := "a.py"
	fileB := "b.py"
	fileC := "c.py"

	nodeA := parser.Node{Kind: parser.KindFunction, Name: "A", QualifiedName: "a.py::A", FilePath: fileA, UpdatedAt: time.Now()}
	nodeB := parser.Node{Kind: parser.KindFunction, Name: "B", QualifiedName: "b.py::B", FilePath: fileB, UpdatedAt: time.Now()}
	nodeC := parser.Node{Kind: parser.KindFunction, Name: "C", QualifiedName: "c.py::C", FilePath: fileC, UpdatedAt: time.Now()}

	s.SaveFileResults(fileA, []parser.Node{nodeA}, nil)
	s.SaveFileResults(fileB, []parser.Node{nodeB}, []parser.Edge{{Kind: parser.EdgeCalls, SourceQualified: "b.py::B", TargetQualified: "a.py::A", FilePath: fileB, UpdatedAt: time.Now()}})
	s.SaveFileResults(fileC, []parser.Node{nodeC}, []parser.Edge{{Kind: parser.EdgeCalls, SourceQualified: "c.py::C", TargetQualified: "b.py::B", FilePath: fileC, UpdatedAt: time.Now()}})

	analyzer := NewImpactAnalyzer(s)

	// Test: Change file A, depth 2. Impact should include A, B, and C.
	results, err := analyzer.GetImpactRadius([]string{fileA}, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 nodes in impact radius, got %d", len(results))
	}

	foundC := false
	for _, n := range results {
		if n.QualifiedName == "c.py::C" {
			foundC = true
		}
	}
	if !foundC {
		t.Error("Expected to find node C in impact radius")
	}
}
