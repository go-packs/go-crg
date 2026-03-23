package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ajeet/go-crg/internal/graph"
	"github.com/ajeet/go-crg/internal/mcp"
	"github.com/ajeet/go-crg/internal/store"
	mcp_server "github.com/mark3labs/mcp-go/server"
)

func main() {
	// 1. Initialize DB Path (default to ~/.crg-go/graph.db)
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".crg-go", "graph.db")

	// 2. Open Store
	s, err := store.NewStore(dbPath)
	if err != nil {
		log.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	// 3. Initialize Impact Analyzer
	analyzer := graph.NewImpactAnalyzer(s)

	// 4. Create MCP Server
	srv := mcp.NewMCPServer(s, analyzer)

	// 5. Run Server (Standard IO)
	log.Println("Code Review Graph Go server starting on stdio...")
	if err := mcp_server.ServeStdio(srv); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
