package parser

import (
	"context"
	"fmt"
	"os"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

// PythonParser implements CodeParser for Python files.
type PythonParser struct {
	lang *sitter.Language
}

func NewPythonParser() *PythonParser {
	return &PythonParser{
		lang: python.GetLanguage(),
	}
}

func (p *PythonParser) DetectLanguage(path string) (string, bool) {
	return DetectLanguage(path)
}

func (p *PythonParser) ParseFile(path string) ([]Node, []Edge, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	psr := sitter.NewParser()
	psr.SetLanguage(p.lang)

	tree, err := psr.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, nil, err
	}

	nodes := []Node{}
	edges := []Edge{}

	// Root File node
	nodes = append(nodes, Node{
		Kind:      KindFile,
		Name:      path,
		FilePath:  path,
		LineStart: 1,
		Language:  "python",
		UpdatedAt: time.Now(),
	})

	// Walk the AST
	p.walk(tree.RootNode(), content, path, "", &nodes, &edges)

	return nodes, edges, nil
}

func (p *PythonParser) walk(n *sitter.Node, content []byte, filePath string, parentClass string, nodes *[]Node, edges *[]Edge) {
	nodeType := n.Type()

	switch nodeType {
	case "class_definition":
		name := p.getName(n, content)
		if name != "" {
			qualifiedName := fmt.Sprintf("%s::%s", filePath, name)
			if parentClass != "" {
				qualifiedName = fmt.Sprintf("%s::%s.%s", filePath, parentClass, name)
			}

			*nodes = append(*nodes, Node{
				Kind:          KindClass,
				Name:          name,
				QualifiedName: qualifiedName,
				FilePath:      filePath,
				LineStart:     int(n.StartPoint().Row) + 1,
				LineEnd:       int(n.EndPoint().Row) + 1,
				Language:      "python",
				ParentName:    parentClass,
				UpdatedAt:     time.Now(),
			})

			// CONTAINS edge
			source := filePath
			if parentClass != "" {
				source = fmt.Sprintf("%s::%s", filePath, parentClass)
			}
			*edges = append(*edges, Edge{
				Kind:            EdgeContains,
				SourceQualified: source,
				TargetQualified: qualifiedName,
				FilePath:        filePath,
				Line:            int(n.StartPoint().Row) + 1,
				UpdatedAt:       time.Now(),
			})

			// Recurse into class body
			for i := 0; i < int(n.ChildCount()); i++ {
				p.walk(n.Child(i), content, filePath, name, nodes, edges)
			}
			return
		}

	case "function_definition":
		name := p.getName(n, content)
		if name != "" {
			qualifiedName := fmt.Sprintf("%s::%s", filePath, name)
			if parentClass != "" {
				qualifiedName = fmt.Sprintf("%s::%s.%s", filePath, parentClass, name)
			}

			*nodes = append(*nodes, Node{
				Kind:          KindFunction,
				Name:          name,
				QualifiedName: qualifiedName,
				FilePath:      filePath,
				LineStart:     int(n.StartPoint().Row) + 1,
				LineEnd:       int(n.EndPoint().Row) + 1,
				Language:      "python",
				ParentName:    parentClass,
				UpdatedAt:     time.Now(),
			})

			// CONTAINS edge
			source := filePath
			if parentClass != "" {
				source = fmt.Sprintf("%s::%s", filePath, parentClass)
			}
			*edges = append(*edges, Edge{
				Kind:            EdgeContains,
				SourceQualified: source,
				TargetQualified: qualifiedName,
				FilePath:        filePath,
				Line:            int(n.StartPoint().Row) + 1,
				UpdatedAt:       time.Now(),
			})
		}
	}

	// Default recursion
	for i := 0; i < int(n.ChildCount()); i++ {
		p.walk(n.Child(i), content, filePath, parentClass, nodes, edges)
	}
}

func (p *PythonParser) getName(n *sitter.Node, content []byte) string {
	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		if child.Type() == "identifier" {
			return child.Content(content)
		}
	}
	return ""
}
