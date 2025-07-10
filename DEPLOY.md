# Deployment Instructions

## Deploy to Render

1. **Push to GitHub**: Ensure your code is pushed to a GitHub repository
2. **Create Render Service**: 
   - Go to [Render.com](https://render.com) and create a new Web Service
   - Connect your GitHub repository
   - Render will automatically detect the `render.yaml` configuration
3. **Note your deployment URL**: After deployment, you'll get a URL like `https://your-app-name.onrender.com`

## Connect to Claude Desktop

1. **Update claude_desktop_config.json**: Replace `your-app-name` with your actual Render app name in the configuration file
2. **Install MCP client**: Run `npm install -g @modelcontextprotocol/server-everything`
3. **Restart Claude Desktop**: Close and reopen Claude Desktop to pick up the new configuration

## HTTP MCP Protocol Details

The server now supports proper HTTP-based MCP communication:

### Endpoints:
- `POST /mcp` - Main MCP JSON-RPC endpoint
- `GET /mcp` - MCP notifications endpoint  
- `DELETE /mcp` - Session termination endpoint
- `GET /health` - Health check endpoint

### Transport Mode:
- **Local Development**: Uses stdio transport (traditional MCP)
- **Render Deployment**: Uses HTTP transport with StreamableHTTPServer
- **Automatic Detection**: Server detects mode based on `PORT` environment variable

### Configuration:
- **Stateless Mode**: Enabled for better scaling on Render
- **Endpoint Path**: `/mcp` for MCP protocol communication
- **Health Check**: `/health` for Render monitoring

## Configuration Files Created

- `render.yaml`: Render deployment configuration
- `Dockerfile`: Multi-stage Docker build for Go application
- `smithery.yaml`: Smithery configuration for MCP deployment
- `claude_desktop_config.json`: Updated with both local and remote server configurations

## Testing

- **Local**: Use the `340b-drugs-local` configuration (stdio mode)
- **Remote**: Use the `340b-drugs-remote` configuration after deployment (HTTP mode)

## How It Works

1. **Local Mode**: When `PORT` env var is not set, server runs in stdio mode for local Claude Desktop
2. **Render Mode**: When `PORT` env var is set, server runs HTTP MCP server on that port
3. **MCP Client**: `@modelcontextprotocol/server-everything` acts as a bridge between Claude Desktop (stdio) and your remote server (HTTP)

The server automatically detects the environment and chooses the appropriate transport method.