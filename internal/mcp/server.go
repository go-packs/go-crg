package mcp

import (
	"context"
	"fmt"

	"github.com/ajeet/go-crg/internal/graph"
	"github.com/ajeet/go-crg/internal/parser"
	"github.com/ajeet/go-crg/internal/store"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewMCPServer(s *store.Store, analyzer *graph.ImpactAnalyzer) *server.MCPServer {
	srv := server.NewMCPServer("Code Review Graph Go", "1.0.0")

	// 1. Tool: build_graph
	srv.AddTool(mcp.Tool{
		Name:        "build_graph",
		Description: "Build or update the code knowledge graph for a repository.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"repo_root": map[string]interface{}{
					"type": "string", 
					"description": "The absolute path to the repository root.",
				},
			},
			Required: []string{"repo_root"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}
		repoRoot, ok := args["repo_root"].(string)
		if !ok {
			return nil, fmt.Errorf("repo_root must be a string")
		}
		
		w := parser.NewWalker(repoRoot, s)
		if err := w.BuildGraph(ctx); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Build failed: %v", err)), nil
		}
		
		return mcp.NewToolResultText(fmt.Sprintf("Successfully built graph for %s", repoRoot)), nil
	})

	// 2. Tool: get_impact_radius
	srv.AddTool(mcp.Tool{
		Name:        "get_impact_radius",
		Description: "Calculate the impact radius from a set of changed files.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"changed_files": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{"type": "string"},
					"description": "List of changed file paths relative to repo root.",
				},
				"max_depth": map[string]interface{}{
					"type": "integer", 
					"description": "Max hops to traverse (default 2).",
				},
			},
			Required: []string{"changed_files"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}
		
		changedFilesRaw, ok := args["changed_files"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("changed_files must be an array")
		}

		changedFiles := []string{}
		for _, f := range changedFilesRaw {
			if s, ok := f.(string); ok {
				changedFiles = append(changedFiles, s)
			}
		}

		maxDepth := 2
		if m, ok := args["max_depth"].(float64); ok {
			maxDepth = int(m)
		}

		results, err := analyzer.GetImpactRadius(changedFiles, maxDepth)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error calculating impact: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Calculated impact radius. Found %d affected nodes.", len(results))), nil
	})

	return srv
}
