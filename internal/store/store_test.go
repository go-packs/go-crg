package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ajeet/go-crg/internal/parser"
)

func TestStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	filePath := "test.py"
	nodes := []parser.Node{
		{
			Kind:          parser.KindFunction,
			Name:          "test_func",
			QualifiedName: "test.py::test_func",
			FilePath:      filePath,
			UpdatedAt:     time.Now(),
		},
	}
	edges := []parser.Edge{
		{
			Kind:            parser.EdgeCalls,
			SourceQualified: "test.py::caller",
			TargetQualified: "test.py::test_func",
			FilePath:        filePath,
			UpdatedAt:       time.Now(),
		},
	}

	if err := s.SaveFileResults(filePath, nodes, edges); err != nil {
		t.Fatalf("SaveFileResults failed: %v", err)
	}

	// Verify Retrieval
	retrievedNodes, err := s.GetNodesByFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(retrievedNodes) != 1 || retrievedNodes[0].Name != "test_func" {
		t.Errorf("Expected 1 node test_func, got %d", len(retrievedNodes))
	}

	// Test deletion on re-save
	if err := s.SaveFileResults(filePath, []parser.Node{}, []parser.Edge{}); err != nil {
		t.Fatal(err)
	}
	retrievedNodes, _ = s.GetNodesByFile(filePath)
	if len(retrievedNodes) != 0 {
		t.Errorf("Expected 0 nodes after empty save, got %d", len(retrievedNodes))
	}
}
