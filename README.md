# 340B Drugs MCP Server

A Model Context Protocol (MCP) server that provides access to 340B drug information and RxNorm API functionality. This server can be used with Claude Code, GitHub Copilot, and other MCP-compatible tools.

## Features

- **Get Related NDCs**: Find related National Drug Codes for drugs
- **RxNorm Information**: Get detailed drug information using RxCUI
- **340B Eligibility Check**: Check if drugs are eligible for 340B pricing
- **Approximate Drug Matching**: Find approximate matches for drug names

## Quick Start

### 1. Build the Server

```bash
# Install dependencies
make install

# Build the server
make build
```

### 2. Setup for Claude Code

Add the following to your Claude Desktop configuration file:

**Location**: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "340b-drugs": {
      "command": "/Users/petergifford/340b-mcp/cmd/server/server",
      "args": [],
      "env": {}
    }
  }
}
```

Or use the provided configuration:
```bash
make setup-claude
```

### 3. Setup for GitHub Copilot / VS Code

For VS Code with MCP support, add the configuration from `vscode-settings.json` to your workspace or user settings:

```json
{
  "mcp.servers": {
    "340b-drugs": {
      "command": "/Users/petergifford/340b-mcp/cmd/server/server",
      "args": [],
      "env": {},
      "description": "340B drug information and RxNorm API access"
    }
  }
}
```

## Available Tools

### 1. get_related_ndcs
Get related NDCs for a given NDC, RxCUI, or drug name.

**Parameters:**
- `ndc` (optional): National Drug Code
- `rxcui` (optional): RxNorm Concept Unique Identifier  
- `name` (optional): Drug name

**Example:**
```json
{
  "tool": "get_related_ndcs",
  "parameters": {
    "ndc": "0009-3542-02"
  }
}
```

### 2. get_rx_info
Get detailed information for a given RxCUI.

**Parameters:**
- `rxcui` (required): RxNorm Concept Unique Identifier

**Example:**
```json
{
  "tool": "get_rx_info", 
  "parameters": {
    "rxcui": "161"
  }
}
```

### 3. check_340b_eligibility
Check if a drug is 340B eligible based on NDC, RxCUI, or drug name.

**Parameters:**
- `ndc` (optional): National Drug Code
- `rxcui` (optional): RxNorm Concept Unique Identifier
- `name` (optional): Drug name

**Example:**
```json
{
  "tool": "check_340b_eligibility",
  "parameters": {
    "ndc": "0009-3542-02"
  }
}
```

### 4. find_approximate_drug_match
Find approximate drug name matches using RxNorm API.

**Parameters:**
- `term` (required): Drug name to search for
- `max_entries` (optional): Maximum number of results to return (default: 1)

**Example:**
```json
{
  "tool": "find_approximate_drug_match",
  "parameters": {
    "term": "aspirin",
    "max_entries": 5
  }
}
```

## Development

### Build Commands

```bash
# Build for current platform
make build

# Build for all platforms
make dist

# Run tests
make test

# Run server locally
make run

# Clean build artifacts
make clean
```

### Project Structure

```
340b-mcp/
├── cmd/server/
│   ├── main.go          # Main server setup
│   ├── drugs.go         # Drug-related tools and functions
│   └── server           # Built binary
├── claude_desktop_config.json  # Claude Desktop configuration
├── vscode-settings.json        # VS Code MCP configuration
├── mcp-server-config.json     # MCP server metadata
├── package.json              # Node.js package metadata
├── Makefile                  # Build automation
└── README.md                # This file
```

## API Dependencies

This server connects to the following external APIs:
- **RxNorm REST API**: `https://rxnav.nlm.nih.gov/REST/`
- **RxTerms API**: For detailed drug information

## Notes

- The 340B eligibility checking is currently simplified and requires access to the ESP (Enhanced Supplement Package) database for full functionality
- All API calls include proper error handling and timeouts
- The server is stateless and can handle concurrent requests

## License

MIT License

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request