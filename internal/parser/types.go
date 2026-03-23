package parser

import "time"

// NodeKind represents the type of code entity.
type NodeKind string

const (
	KindFile     NodeKind = "File"
	KindFunction NodeKind = "Function"
	KindClass    NodeKind = "Class"
	KindType     NodeKind = "Type"
	KindTest     NodeKind = "Test"
)

// Node represents a structural code entity (e.g., function, class).
type Node struct {
	ID            int64     `json:"id" db:"id"`
	Kind          NodeKind  `json:"kind" db:"kind"`
	Name          string    `json:"name" db:"name"`
	QualifiedName string    `json:"qualified_name" db:"qualified_name"`
	FilePath      string    `json:"file_path" db:"file_path"`
	LineStart     int       `json:"line_start" db:"line_start"`
	LineEnd       int       `json:"line_end" db:"line_end"`
	Language      string    `json:"language" db:"language"`
	ParentName    string    `json:"parent_name" db:"parent_name"`
	Params        string    `json:"params" db:"params"`
	ReturnType    string    `json:"return_type" db:"return_type"`
	IsTest        bool      `json:"is_test" db:"is_test"`
	FileHash      string    `json:"file_hash" db:"file_hash"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// EdgeKind represents the relationship between nodes.
type EdgeKind string

const (
	EdgeCalls        EdgeKind = "CALLS"
	EdgeContains     EdgeKind = "CONTAINS"
	EdgeImportsFrom  EdgeKind = "IMPORTS_FROM"
	EdgeInherits     EdgeKind = "INHERITS"
	EdgeImplements   EdgeKind = "IMPLEMENTS"
	EdgeTestedBy     EdgeKind = "TESTED_BY"
	EdgeDependsOn    EdgeKind = "DEPENDS_ON"
)

// Edge represents a relationship between two nodes.
type Edge struct {
	ID              int64     `json:"id" db:"id"`
	Kind            EdgeKind  `json:"kind" db:"kind"`
	SourceQualified string    `json:"source" db:"source_qualified"`
	TargetQualified string    `json:"target" db:"target_qualified"`
	FilePath        string    `json:"file_path" db:"file_path"`
	Line            int       `json:"line" db:"line"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
