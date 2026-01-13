# Klod

Klod (pronounced \klod\), A simple, fast CLI tool for chatting with Claude using the Anthropic API with real-time streaming responses.

## Features

- Maintains conversation history within a session
- Configurable model and system prompts

## Requirements

- Go 1.21 or higher
- Anthropic API key ([get one here](https://console.anthropic.com/))

## Installation

1. Clone and build:
```bash
git clone https://github.com/leoalho/klod.git
cd anthropic-cli
go build
```

2. Create a symlink for global access:
```bash
sudo ln -s $(pwd)/klod /usr/local/bin/klod
```

Or install via Go:
```bash
go install
```

## Configuration

The tool looks for configuration files in the following order:
1. `~/.config/klod/config` (XDG standard location)
2. `~/.klod.env` (home directory)
3. `.env` in the current directory (for project-specific overrides)

### Setup your config:

```bash
# Create the config directory
mkdir -p ~/.config/klod

# Create config file
cat > ~/.config/klod/config << EOF
ANTHROPIC_API_KEY=your-api-key-here
MODEL=claude-sonnet-4-5-20250929
SYSTEM_PROMPT=
EOF
```

### Configuration options:

- `ANTHROPIC_API_KEY` (required): Your Anthropic API key
- `MODEL` (optional): Model to use (default: claude-sonnet-4-5-20250929)
- `SYSTEM_PROMPT` (optional): Custom system prompt for Claude

## Usage

Start a conversation:
```bash
klod Hello, how are you?
```

This will:
1. Send your initial message to Claude
2. Stream the response in real-time
3. Enter an interactive chat mode where you can continue the conversation

Type `exit` or `quit` to end the conversation.

## Development

Run without building:
```bash
go run main.go your message here
```
