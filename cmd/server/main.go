package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Initialize NDC cache
	fmt.Fprintln(os.Stderr, "Initializing NDC cache...")
	if err := InitNDCCache(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize NDC cache: %v\n", err)
		os.Exit(1)
	}
	
	// Set up graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Fprintln(os.Stderr, "Shutting down...")
		StopNDCCache()
		os.Exit(0)
	}()

	s := server.NewMCPServer(
		"340b-drugs",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	relatedNDCsTool, relatedNDCsHandler := CreateRelatedNDCsTool()
	s.AddTool(relatedNDCsTool, relatedNDCsHandler)

	rxInfoTool, rxInfoHandler := CreateRxInfoTool()
	s.AddTool(rxInfoTool, rxInfoHandler)

	is340BTool, is340BHandler := Create340BEligibilityTool()
	s.AddTool(is340BTool, is340BHandler)

	approximateMatchTool, approximateMatchHandler := CreateApproximateMatchTool()
	s.AddTool(approximateMatchTool, approximateMatchHandler)

	generateRxNormExcelTool, generateRxNormExcelHandler := CreateGenerateRxNormExcelTool()
	s.AddTool(generateRxNormExcelTool, generateRxNormExcelHandler)

	is340BExcelTool, is340BExcelHandler := CreateIs340BExcelTool()
	s.AddTool(is340BExcelTool, is340BExcelHandler)

	// Check if we're running in HTTP mode (has PORT env var) or stdio mode
	if port := os.Getenv("PORT"); port != "" {
		// HTTP mode for Render deployment
		fmt.Fprintf(os.Stderr, "Starting HTTP MCP server on port %s\n", port)
		
		// Create HTTP multiplexer
		mux := http.NewServeMux()
		
		// Add health check endpoint
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		
		// Add MCP endpoint - simple JSON-RPC handler
		mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Content-Type", "application/json")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			if r.Method == "GET" {
				// Return server info for GET requests
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"jsonrpc":"2.0","result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{"listChanged":false},"prompts":{"listChanged":false},"resources":{"listChanged":false,"subscribe":false}},"serverInfo":{"name":"340b-drugs","version":"1.0.0"}}}`))
				return
			}
			
			if r.Method != "POST" {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			
			// For now, return a simple error response for POST requests
			// This prevents the hanging issue
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not implemented in HTTP mode. Please use stdio mode."},"id":null}`))
		})
		
		// Start HTTP server
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}
	} else {
		// stdio mode for local development
		fmt.Fprintln(os.Stderr, "Starting stdio MCP server...")
		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}
}
