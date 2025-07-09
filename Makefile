.PHONY: build clean install test run

# Build the MCP server
build:
	go build -o cmd/server/server ./cmd/server
	chmod +x cmd/server/server

# Clean build artifacts
clean:
	rm -f cmd/server/server

# Install dependencies
install:
	go mod download
	go mod tidy

# Test the server
test:
	go test ./...

# Run the server (for testing)
run: build
	./cmd/server/server

# Build for distribution
dist: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/server-linux-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/server-darwin-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/server-darwin-arm64 ./cmd/server
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/server-windows-amd64.exe ./cmd/server

# Setup Claude Desktop configuration
setup-claude:
	@echo "Add this to your Claude Desktop configuration:"
	@echo "File location: ~/Library/Application Support/Claude/claude_desktop_config.json"
	@cat claude_desktop_config.json

# Setup GitHub Copilot configuration
setup-copilot:
	@echo "For GitHub Copilot, you'll need to configure the MCP server in your editor"
	@echo "See the README for specific setup instructions"