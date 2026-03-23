package graph

import (
	"github.com/ajeet/go-crg/internal/parser"
	"github.com/ajeet/go-crg/internal/store"
	"gonum.org/v1/gonum/graph/simple"
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
	
	qnToID := make(map[string]int64)
	idToQN := make(map[int64]string)
	idToNode := make(map[int64]parser.Node)
	var nextID int64

	getOrCreateID := func(qn string) int64 {
		if id, ok := qnToID[qn]; ok {
			return id
		}
		id := nextID
		qnToID[qn] = id
		idToQN[id] = qn
		dg.AddNode(simple.Node(id))
		nextID++
		return id
	}

	// 1. Fetch nodes for changed files and seed the traversal
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
		}
	}

	// 2. Load all edges to build the graph structure
	// Note: For large repos, loading all edges could be memory intensive, 
	// but it is the simplest approach for a complete BFS.
	// In a real optimized system, we would query the DB dynamically during traversal.
	// For this MVP, we assume a small enough graph or implement a lazy load.
	
	// BFS traversal
	impactedIDs := make(map[int64]bool)
	frontier := make([]int64, len(seeds))
	copy(frontier, seeds)

	for _, id := range seeds {
		impactedIDs[id] = true
	}

	depth := 0
	for len(frontier) > 0 && depth < maxDepth {
		var nextFrontier []int64
		for _, id := range frontier {
			qn := idToQN[id]
			
			// Forward edges
			outEdges, _ := a.store.GetEdgesBySource(qn)
			for _, e := range outEdges {
				targetID := getOrCreateID(e.TargetQualified)
				if !impactedIDs[targetID] {
					impactedIDs[targetID] = true
					nextFrontier = append(nextFrontier, targetID)
				}
			}

			// Reverse edges
			inEdges, _ := a.store.GetEdgesByTarget(qn)
			for _, e := range inEdges {
				sourceID := getOrCreateID(e.SourceQualified)
				if !impactedIDs[sourceID] {
					impactedIDs[sourceID] = true
					nextFrontier = append(nextFrontier, sourceID)
				}
			}
		}
		frontier = nextFrontier
		depth++
	}

	// Fetch full node data for impacted nodes that weren't in the seeds
	results := []parser.Node{}
	for id := range impactedIDs {
		node, ok := idToNode[id]
		if !ok {
			// Fetch from DB if we don't have it in memory yet
			qn := idToQN[id]
			// A real implementation would have a GetNode(qn) method in the store
			// For now, we simulate this by leaving it out of results if missing,
			// or returning a skeleton node.
			node = parser.Node{QualifiedName: qn}
		}
		results = append(results, node)
	}

	return results, nil
}
