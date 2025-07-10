# 340B Drugs MCP Server
[![smithery badge](https://smithery.ai/badge/@pgiffy/local-340b-mcp)](https://smithery.ai/server/@pgiffy/local-340b-mcp)

A Model Context Protocol (MCP) server that provides access to 340B drug information and RxNorm API functionality. This server can be used with Claude Code, GitHub Copilot, and other MCP-compatible tools.

## Features

- **Get Related NDCs**: Find related National Drug Codes for drugs
- **RxNorm Information**: Get detailed drug information using RxCUI
- **340B Eligibility Check**: Check if drugs are eligible for 340B pricing (uses real ESP database)
- **Approximate Drug Matching**: Find approximate matches for drug names
- **Batch RxNorm Processing**: Process multiple drug names for RxNorm matches
- **Batch 340B Processing**: Process multiple NDC codes for 340B eligibility
- **Automatic Data Refresh**: Updates 340B data from ESP database every 24 hours

## Quick Start

### Installing via Smithery

To install 340B Drugs Server for Claude Desktop automatically via [Smithery](https://smithery.ai/server/@pgiffy/local-340b-mcp):

```bash
npx -y @smithery/cli install @pgiffy/local-340b-mcp --client claude
```

### 1. Build the Server

```bash
# Install dependencies
make install

# Build the server
make build
```

### 2. Setup for Claude Code

#### Option A: Using Claude MCP Command (Recommended)
```bash
# Build the server first
make build

# Register the MCP server with Claude Code
claude mcp add 340b-drugs /Users/petergifford/340b-mcp/cmd/server/server

# Verify registration
claude mcp list
```

#### Option B: Manual Configuration
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

### 5. generate_rxnorm_excel
Process multiple drug names and return RxNorm match results in Excel-like format.

**Parameters:**
- `drug_names` (required): JSON array of drug names to process

**Example:**
```json
{
  "tool": "generate_rxnorm_excel",
  "parameters": {
    "drug_names": "[\"aspirin\", \"ibuprofen\", \"acetaminophen\"]"
  }
}
```

### 6. is_340b_excel
Process multiple NDC codes and return 340B eligibility results in Excel-like format.

**Parameters:**
- `ndc_codes` (required): JSON array of NDC codes to check

**Example:**
```json
{
  "tool": "is_340b_excel",
  "parameters": {
    "ndc_codes": "[\"0009-3542-02\", \"0074-3368-02\", \"0074-6451-02\"]"
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
- **340B ESP Database**: `https://www.340besp.com/ndcs` (Excel format, auto-refreshed every 24 hours)

## Notes

- The 340B eligibility checking uses real data from the ESP (Enhanced Supplement Package) database
- Excel/CSV data is automatically downloaded and cached on startup
- Cache is refreshed every 24 hours to ensure up-to-date information
- All API calls include proper error handling and timeouts
- The server handles concurrent requests with thread-safe caching

## License

MIT License

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request
