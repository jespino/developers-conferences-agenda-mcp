# Developers Conferences Agenda MCP Plugin

[MCP](https://github.com/metoro-io/mcp-golang) server that provides access to developer conferences and events data. This plugin fetches information from [developers.events](https://developers.events/) and makes it available through MCP-compatible clients like Claude.

## Feature Demo

![Demo GIF](https://example.com/placeholder-for-demo.gif)

*Demonstration of searching for events with upcoming CFP deadlines*

## Features

- Search for developer conferences and events with filtering options
- Find events with open Call for Papers (CFP)
- Get upcoming events
- Find events with approaching CFP deadlines
- Access event data through MCP resources

## Setup Guide

### Installation

Choose one of these installation methods:

```bash
# Using Go
go install github.com/jespino/developers-conferences-agenda-mcp@latest

# Using Docker
git clone https://github.com/jespino/developers-conferences-agenda-mcp.git
cd developers-conferences-agenda-mcp
docker build -t mcp/conferences .
docker run -i mcp/conferences
```

### Configuration and Usage

The plugin works automatically with no additional configuration required, as it fetches data from public sources.

#### Optional Arguments

- `--transport`: Choose transport type (`stdio` [default] or `sse`)
- `--port`: Port number for SSE transport (default: 8000)
- `--verbose`: Increase logging verbosity (can be used multiple times)
  - `-v` or `--verbose`: Set logging to INFO level
  - `-vv` or `--verbose --verbose`: Set logging to DEBUG level

## IDE Integration

### Claude Desktop Setup

Using Go:

```json
{
  "mcpServers": {
    "mcp-conferences": {
      "command": "developers-conferences-agenda-mcp"
    }
  }
}
```

Using Docker:

```json
{
  "mcpServers": {
    "mcp-conferences": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "mcp/conferences"
      ]
    }
  }
}
```

### Cursor IDE Setup

1. Open Cursor Settings
2. Navigate to `Features` > `MCP Servers` (or directly to `MCP`)
3. Click `+ Add new global MCP server`

#### JSON Configuration for stdio Transport

```json
{
  "mcpServers": {
    "mcp-conferences": {
      "command": "developers-conferences-agenda-mcp"
    }
  }
}
```

#### SSE Transport Configuration

First start the server:

```bash
developers-conferences-agenda-mcp --transport sse --port 9000
```

Then configure in Cursor:

```json
{
  "mcpServers": {
    "mcp-conferences-sse": {
      "url": "http://localhost:9000/sse"
    }
  }
}
```

## Resources

- `events://all`: All developer conferences and events
- `events://open-cfps`: Events with open Call for Papers

## Available Tools

| Tool Name | Description |
|-----------|-------------|
| `search_events` | Search for developer conferences and events with filtering options |
| `open_cfps` | Get events with open CFP (Call for Papers) |
| `upcoming_events` | Get upcoming developer conferences and events |
| `cfp_deadlines_soon` | Get events with CFP deadlines approaching within specified days |

## Examples

### Searching for events

```json
{
  "query": "golang",
  "location": "Europe",
  "fromDate": "2025-01-01",
  "toDate": "2025-12-31",
  "hasOpenCFP": true,
  "limit": 5
}
```

### Finding events with CFP deadlines soon

```json
{
  "days": 14
}
```

## Development & Debugging

### Local Development Setup

Clone the repository and run locally:

```bash
git clone https://github.com/jespino/developers-conferences-agenda-mcp.git
cd developers-conferences-agenda-mcp
go run main.go
```

### Running Tests

```bash
go test -v ./...
```

### Debugging Tools

```bash
# Using MCP Inspector
npx @modelcontextprotocol/inspector developers-conferences-agenda-mcp

# For local development version
npx @modelcontextprotocol/inspector go run main.go
```

## Security

- This plugin only accesses publicly available data
- No authentication credentials are required
- No sensitive data is collected or stored

## Data Source

Event data is fetched from [developers.events](https://developers.events/) which is a community-maintained list of developer conferences and events.

## License

Licensed under MIT - see [LICENSE](LICENSE) file.

## Topics

`mcp` `conferences` `events` `developer-conferences` `cfp`
