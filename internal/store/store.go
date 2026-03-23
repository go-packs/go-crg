package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ajeet/go-crg/internal/parser"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var schema = `
CREATE TABLE IF NOT EXISTS nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    qualified_name TEXT NOT NULL UNIQUE,
    file_path TEXT NOT NULL,
    line_start INTEGER,
    line_end INTEGER,
    language TEXT,
    parent_name TEXT,
    params TEXT,
    return_type TEXT,
    is_test INTEGER DEFAULT 0,
    file_hash TEXT,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS edges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    kind TEXT NOT NULL,
    source_qualified TEXT NOT NULL,
    target_qualified TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line INTEGER DEFAULT 0,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_nodes_file ON nodes(file_path);
CREATE INDEX IF NOT EXISTS idx_nodes_qualified ON nodes(qualified_name);
CREATE INDEX IF NOT EXISTS idx_edges_source ON edges(source_qualified);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges(target_qualified);
`

type Store struct {
	db *sqlx.DB
}

func NewStore(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveFileResults(filePath string, nodes []parser.Node, edges []parser.Edge) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old data for this file
	if _, err := tx.Exec("DELETE FROM nodes WHERE file_path = ?", filePath); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM edges WHERE file_path = ?", filePath); err != nil {
		return err
	}

	// Insert nodes
	for _, node := range nodes {
		_, err := tx.NamedExec(`
			INSERT INTO nodes (kind, name, qualified_name, file_path, line_start, line_end, language, parent_name, params, return_type, is_test, file_hash, updated_at)
			VALUES (:kind, :name, :qualified_name, :file_path, :line_start, :line_end, :language, :parent_name, :params, :return_type, :is_test, :file_hash, :updated_at)
		`, node)
		if err != nil {
			return fmt.Errorf("failed to insert node %s: %w", node.QualifiedName, err)
		}
	}

	// Insert edges
	for _, edge := range edges {
		_, err := tx.NamedExec(`
			INSERT INTO edges (kind, source_qualified, target_qualified, file_path, line, updated_at)
			VALUES (:kind, :source_qualified, :target_qualified, :file_path, :line, :updated_at)
		`, edge)
		if err != nil {
			return fmt.Errorf("failed to insert edge: %w", err)
		}
	}

	return tx.Commit()
}

func (s *Store) GetNodesByFile(filePath string) ([]parser.Node, error) {
	nodes := []parser.Node{}
	err := s.db.Select(&nodes, "SELECT * FROM nodes WHERE file_path = ?", filePath)
	return nodes, err
}

func (s *Store) GetEdgesBySource(qualifiedName string) ([]parser.Edge, error) {
	edges := []parser.Edge{}
	err := s.db.Select(&edges, "SELECT * FROM edges WHERE source_qualified = ?", qualifiedName)
	return edges, err
}

func (s *Store) GetEdgesByTarget(qualifiedName string) ([]parser.Edge, error) {
	edges := []parser.Edge{}
	err := s.db.Select(&edges, "SELECT * FROM edges WHERE target_qualified = ?", qualifiedName)
	return edges, err
}
