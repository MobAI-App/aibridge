# AiBridge

A CLI tool that wraps AI coding assistants (Claude Code, Codex, Gemini CLI) with a PTY and exposes an HTTP API for external text injection.

AiBridge enables desktop applications, browser extensions, and other tools to inject context into terminal-based AI assistants without manual copy-paste.

## Features

- **PTY Wrapper** - Full terminal emulation with raw mode support
- **HTTP API** - Simple REST API for text injection
- **Idle Detection** - Auto-detects when the AI assistant is ready for input
- **Injection Queue** - FIFO queue with priority support
- **Multi-tool Support** - Built-in patterns for Claude Code, Codex, and Gemini CLI
- **Paranoid Mode** - Inject text without auto-submitting for review

## Installation

### Quick Install (macOS/Linux/Windows WSL)

```bash
curl -fsSL https://raw.githubusercontent.com/MobAI-App/aibridge/main/install.sh | bash
```

### Windows Native
Download the latest Windows release from [GitHub Releases] (https://github.com/MobAI-App/aibridge/releases).

### Download Binary

| Platform | Architecture | Download |
|----------|--------------|----------|
| macOS | Apple Silicon | `aibridge_*_darwin_arm64.tar.gz` |
| macOS | Intel | `aibridge_*_darwin_amd64.tar.gz` |
| Linux | x86_64 | `aibridge_*_linux_amd64.tar.gz` |
| Linux | ARM64 | `aibridge_*_linux_arm64.tar.gz` |
| Windows | x86_64 | `aibridge_*_windows_amd64.zip` |
| Windows | ARM64 | `aibridge_*_windows_arm64.zip` |

### Go Install

```bash
go install github.com/MobAI-App/aibridge/cmd/aibridge@latest
```

### From Source

```bash
git clone https://github.com/MobAI-App/aibridge.git
cd aibridge
go build -o aibridge ./cmd/aibridge
```

## Usage

### Basic Usage

```bash
# Wrap Claude Code
aibridge claude

# Wrap with custom port
aibridge -p 8080 claude

# Paranoid mode - inject without auto-submit
aibridge --paranoid claude
```

### Injecting Text

Once aibridge is running, inject text via HTTP:

```bash
# Simple injection
curl -X POST http://localhost:9999/inject \
  -H "Content-Type: application/json" \
  -d '{"text": "explain this code"}'

# Priority injection (skips to front of queue)
curl -X POST http://localhost:9999/inject \
  -H "Content-Type: application/json" \
  -d '{"text": "urgent request", "priority": true}'

# Synchronous injection (blocks until injected)
curl -X POST "http://localhost:9999/inject?sync=true" \
  -H "Content-Type: application/json" \
  -d '{"text": "wait for this"}'
```

### CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--port` | `-p` | 9999 | HTTP server port |
| `--host` | | 127.0.0.1 | HTTP server host |
| `--busy-pattern` | | (auto) | Custom busy detection regex |
| `--timeout` | `-t` | 300 | Sync injection timeout (seconds) |
| `--verbose` | `-v` | false | Enable verbose logging |
| `--paranoid` | | false | Inject text without hitting Enter |
| `--version` | | | Print version and exit |

## HTTP API

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/status` | GET | Bridge status |
| `/inject` | POST | Queue text injection |
| `/queue` | DELETE | Clear pending injections |

### GET /health

Returns health status.

```json
{"status": "ok", "version": "1.0.0"}
```

### GET /status

Returns bridge status.

```json
{
  "idle": true,
  "queue_length": 0,
  "child_running": true,
  "child_tool": "claude",
  "uptime_seconds": 123.45
}
```

### POST /inject

Queue text for injection.

**Request:**
```json
{
  "text": "your prompt here",
  "priority": false
}
```

**Response:**
```json
{
  "id": "uuid",
  "queued": true,
  "position": 1
}
```

**Query Parameters:**
- `sync=true` - Block until text is injected

**Error Codes:**
- `400` - Invalid JSON or empty text
- `408` - Sync injection timeout
- `429` - Queue full (max 100 items)
- `503` - Child process not running

### DELETE /queue

Clear all pending injections.

```json
{"cleared": 5}
```

## Busy Detection

AiBridge detects when the AI assistant is busy using regex patterns matched against terminal output. When the pattern stops appearing for 500ms, the tool is considered idle.

### Built-in Patterns

| Tool | Pattern |
|------|---------|
| Claude Code | `esc to interrupt` |
| Codex | `esc to interrupt` |
| Gemini | `esc to cancel` |

### Custom Patterns

```bash
aibridge --busy-pattern 'processing' some-tool
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    AiBridge                          │
├─────────────────────────────────────────────────────┤
│  HTTP Server (localhost:9999)                        │
│  ├── GET  /health                                    │
│  ├── GET  /status                                    │
│  ├── POST /inject                                    │
│  └── DELETE /queue                                   │
├─────────────────────────────────────────────────────┤
│  Injection Queue (FIFO + Priority)                   │
├─────────────────────────────────────────────────────┤
│  Busy Detector (Regex Pattern Matching)              │
├─────────────────────────────────────────────────────┤
│  PTY Manager (Raw Mode, Window Resize)               │
├─────────────────────────────────────────────────────┤
│  Child Process (claude, codex, gemini, etc.)         │
└─────────────────────────────────────────────────────┘
```

## Development

### Requirements

- Go 1.22+

### Building

```bash
go build -o aibridge ./cmd/aibridge
```

### Testing

```bash
go test ./...
```

### Dependencies

- `github.com/creack/pty` - PTY support
- `github.com/spf13/cobra` - CLI framework
- `github.com/google/uuid` - Injection IDs
- `golang.org/x/term` - Raw terminal mode

## Security

- **Localhost only** - HTTP server binds to 127.0.0.1 by default
- **No authentication** - Designed for local development use
- **CORS enabled** - Allows requests from any origin for browser extensions

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
