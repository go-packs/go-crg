package graph

import (
	"github.com/ajeet/go-crg/internal/parser"
	"github.com/ajeet/go-crg/internal/store"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/traverse"
)

type ImpactAnalyzer struct {
	store *store.Store
}

func NewImpactAnalyzer(s *store.Store) *ImpactAnalyzer {
	return &ImpactAnalyzer{store: s}
}

// NodeWrapper is used to map qualified names to Gonum's integer IDs.
type NodeWrapper struct {
	id int64
	qn string
}

func (n NodeWrapper) ID() int64 { return n.id }

func (a *ImpactAnalyzer) GetImpactRadius(changedFiles []string, maxDepth int) ([]parser.Node, error) {
	dg := simple.NewDirectedGraph()
	
	// Map to keep track of nodes by QualifiedName
	qnToID := make(map[string]int64)
	idToNode := make(map[int64]parser.Node)
	var nextID int64

	// Helper to get or create ID
	getOrCreateID := func(qn string) int64 {
		if id, ok := qnToID[qn]; ok {
			return id
		}
		id := nextID
		qnToID[qn] = id
		nextID++
		return id
	}

	// In a real implementation, we would only load relevant edges.
	// For simplicity, let's assume we have a method to get all edges or load them incrementally.
	// Here we just fetch nodes from changed files to start with.
	seeds := []int64{}
	for _, file := range changedFiles {
		nodes, err := a.store.GetNodesByFile(file)
		if err != nil {
			continue
		}
		for _, node := range nodes {
			id := getOrCreateID(node.QualifiedName)
			idToNode[id] = node
			seeds = append(seeds, id)
			dg.AddNode(simple.Node(id))
		}
	}

	// BFS traversal
	impactedIDs := make(map[int64]bool)
	bfs := traverse.BreadthFirst{
		Visit: func(n dg.Node) {
			impactedIDs[n.ID()] = true
		},
	}

	// This is a simplified skeleton. 
	// To actually find impact, we need to load edges from the DB into the 'dg' graph
	// and then run the traversal.
	
	// TODO: Load edges into 'dg'
	
	results := []parser.Node{}
	for id := range impactedIDs {
		if node, ok := idToNode[id]; ok {
			results = append(results, node)
		}
	}

	return results, nil
}
