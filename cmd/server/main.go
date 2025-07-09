package main

import (
	"fmt"
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

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}
}
