package parser

import (
	"context"
	"fmt"
	"os"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

type GoParser struct {
	lang *sitter.Language
}

func NewGoParser() *GoParser {
	return &GoParser{
		lang: golang.GetLanguage(),
	}
}

func (p *GoParser) DetectLanguage(path string) (string, bool) {
	return DetectLanguage(path)
}

func (p *GoParser) ParseFile(path string) ([]Node, []Edge, error) {
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
		Language:  "go",
		UpdatedAt: time.Now(),
	})

	p.walk(tree.RootNode(), content, path, "", "", &nodes, &edges)

	return nodes, edges, nil
}

func (p *GoParser) walk(n *sitter.Node, content []byte, filePath string, parentType string, enclosingFunc string, nodes *[]Node, edges *[]Edge) {
	nodeType := n.Type()

	switch nodeType {
	case "type_declaration":
		name := p.getName(n, content, "type_spec")
		if name != "" {
			qualifiedName := fmt.Sprintf("%s::%s", filePath, name)

			*nodes = append(*nodes, Node{
				Kind:          KindType,
				Name:          name,
				QualifiedName: qualifiedName,
				FilePath:      filePath,
				LineStart:     int(n.StartPoint().Row) + 1,
				LineEnd:       int(n.EndPoint().Row) + 1,
				Language:      "go",
				UpdatedAt:     time.Now(),
			})

			*edges = append(*edges, Edge{
				Kind:            EdgeContains,
				SourceQualified: filePath,
				TargetQualified: qualifiedName,
				FilePath:        filePath,
				Line:            int(n.StartPoint().Row) + 1,
				UpdatedAt:       time.Now(),
			})

			for i := 0; i < int(n.ChildCount()); i++ {
				p.walk(n.Child(i), content, filePath, name, "", nodes, edges)
			}
			return
		}

	case "function_declaration", "method_declaration":
		name := p.getName(n, content, "")
		if name != "" {
			// For Go methods, the receiver is the parent
			receiverType := ""
			if nodeType == "method_declaration" {
				receiverType = p.getReceiverType(n, content)
			}

			qualifiedName := fmt.Sprintf("%s::%s", filePath, name)
			if receiverType != "" {
				qualifiedName = fmt.Sprintf("%s::%s.%s", filePath, receiverType, name)
			}

			*nodes = append(*nodes, Node{
				Kind:          KindFunction,
				Name:          name,
				QualifiedName: qualifiedName,
				FilePath:      filePath,
				LineStart:     int(n.StartPoint().Row) + 1,
				LineEnd:       int(n.EndPoint().Row) + 1,
				Language:      "go",
				ParentName:    receiverType,
				UpdatedAt:     time.Now(),
			})

			source := filePath
			if receiverType != "" {
				source = fmt.Sprintf("%s::%s", filePath, receiverType)
			}
			*edges = append(*edges, Edge{
				Kind:            EdgeContains,
				SourceQualified: source,
				TargetQualified: qualifiedName,
				FilePath:        filePath,
				Line:            int(n.StartPoint().Row) + 1,
				UpdatedAt:       time.Now(),
			})

			for i := 0; i < int(n.ChildCount()); i++ {
				p.walk(n.Child(i), content, filePath, receiverType, name, nodes, edges)
			}
			return
		}

	case "call_expression":
		if enclosingFunc != "" {
			callName := p.getCallName(n, content)
			if callName != "" {
				caller := fmt.Sprintf("%s::%s", filePath, enclosingFunc)
				if parentType != "" {
					caller = fmt.Sprintf("%s::%s.%s", filePath, parentType, enclosingFunc)
				}
				*edges = append(*edges, Edge{
					Kind:            EdgeCalls,
					SourceQualified: caller,
					TargetQualified: callName,
					FilePath:        filePath,
					Line:            int(n.StartPoint().Row) + 1,
					UpdatedAt:       time.Now(),
				})
			}
		}
	}

	for i := 0; i < int(n.ChildCount()); i++ {
		p.walk(n.Child(i), content, filePath, parentType, enclosingFunc, nodes, edges)
	}
}

func (p *GoParser) getName(n *sitter.Node, content []byte, expectedParentType string) string {
	if expectedParentType != "" {
		for i := 0; i < int(n.ChildCount()); i++ {
			child := n.Child(i)
			if child.Type() == expectedParentType {
				return p.getName(child, content, "")
			}
		}
		return ""
	}

	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		if child.Type() == "identifier" || child.Type() == "type_identifier" || child.Type() == "field_identifier" {
			return child.Content(content)
		}
	}
	return ""
}

func (p *GoParser) getReceiverType(n *sitter.Node, content []byte) string {
	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		if child.Type() == "parameter_list" { // Method receiver
			for j := 0; j < int(child.ChildCount()); j++ {
				param := child.Child(j)
				if param.Type() == "parameter_declaration" {
					for k := 0; k < int(param.ChildCount()); k++ {
						t := param.Child(k)
						if t.Type() == "type_identifier" {
							return t.Content(content)
						} else if t.Type() == "pointer_type" {
							for l := 0; l < int(t.ChildCount()); l++ {
								if t.Child(l).Type() == "type_identifier" {
									return t.Child(l).Content(content)
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func (p *GoParser) getCallName(n *sitter.Node, content []byte) string {
	if n.ChildCount() == 0 {
		return ""
	}
	first := n.Child(0)

	if first.Type() == "identifier" {
		return first.Content(content)
	}

	if first.Type() == "selector_expression" {
		for i := int(first.ChildCount()) - 1; i >= 0; i-- {
			child := first.Child(i)
			if child.Type() == "field_identifier" {
				return child.Content(content)
			}
		}
	}
	return ""
}
