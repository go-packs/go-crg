package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonParser(t *testing.T) {
	p := NewPythonParser()
	content := `
class MyClass:
    def method_a(self):
        print("Hello")
        self.method_b()

    def method_b(self):
        pass

def top_level_func():
    obj = MyClass()
    obj.method_a()
`
	tmpFile := filepath.Join(t.TempDir(), "test.py")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	nodes, edges, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify Nodes
	expectedNodes := map[NodeKind]int{
		KindFile:     1,
		KindClass:    1,
		KindFunction: 3, // method_a, method_b, top_level_func
	}

	counts := make(map[NodeKind]int)
	for _, n := range nodes {
		counts[n.Kind]++
	}

	for kind, count := range expectedNodes {
		if counts[kind] != count {
			t.Errorf("Expected %d nodes of kind %s, got %d", count, kind, counts[kind])
		}
	}

	// Verify Edges
	hasCall := false
	for _, e := range edges {
		if e.Kind == EdgeCalls && e.TargetQualified == "method_a" {
			hasCall = true
		}
	}
	if !hasCall {
		t.Error("Expected to find a CALLS edge to method_a")
	}
}

func TestGoParser(t *testing.T) {
	p := NewGoParser()
	content := `
package main

type MyStruct struct {}

func (m *MyStruct) MethodA() {
	m.MethodB()
}

func (m *MyStruct) MethodB() {}

func Main() {
	s := &MyStruct{}
	s.MethodA()
}
`
	tmpFile := filepath.Join(t.TempDir(), "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	nodes, edges, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify Nodes
	foundStruct := false
	foundMethodA := false
	for _, n := range nodes {
		if n.Kind == KindType && n.Name == "MyStruct" {
			foundStruct = true
		}
		if n.Kind == KindFunction && n.Name == "MethodA" && n.ParentName == "MyStruct" {
			foundMethodA = true
		}
	}

	if !foundStruct {
		t.Error("Expected to find Type node MyStruct")
	}
	if !foundMethodA {
		t.Error("Expected to find Function node MethodA with parent MyStruct")
	}

	// Verify Edges
	containsMethod := false
	for _, e := range edges {
		if e.Kind == EdgeContains && e.SourceQualified == tmpFile+"::MyStruct" && e.TargetQualified == tmpFile+"::MyStruct.MethodA" {
			containsMethod = true
		}
	}
	if !containsMethod {
		t.Error("Expected CONTAINS edge from MyStruct to MethodA")
	}
}
