# Developers Conferences Agenda MCP Plugin

This is a plugin for the [MCP (Metoro Client Protocol)](https://github.com/metoro-io/mcp-golang) that provides access to upcoming developer conferences and events. It fetches data from [developers.events](https://developers.events/) and exposes it through various tools and resources.

## Features

- Search for developer conferences and events with filtering options
- Find events with open Call for Papers (CFP)
- Get upcoming events
- Find events with approaching CFP deadlines
- Access event data through MCP resources

## Installation

```bash
go install github.com/jespino/developers-conferences-agenda-mcp@latest
```

## Usage

This plugin works with MCP-compatible clients. Once registered with an MCP client, you can use the following tools and resources:

### Tools

| Tool Name | Description |
|-----------|-------------|
| `search_events` | Search for developer conferences and events with filtering options |
| `open_cfps` | Get events with open CFP (Call for Papers) |
| `upcoming_events` | Get upcoming developer conferences and events |
| `cfp_deadlines_soon` | Get events with CFP deadlines approaching within specified days |

### Resources

| Resource URI | Description |
|--------------|-------------|
| `events://all` | All developer conferences and events |
| `events://open-cfps` | Events with open Call for Papers |

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

## Data Source

Event data is fetched from [developers.events](https://developers.events/) which is a community-maintained list of developer conferences and events.

## Development

### Running tests

```bash
go test -v ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)
