# Klod

Klod (pronounced \klod\\), a simple Go-based CLI tool for interacting with the Anthropic API

## Motivation

Sometimes you just need to quickly check/ask an LLM something. I basically live in the terminal so I want to avoid the context switch to a browser. Claude code on the other hand is a bit too eager to help (which has its benefits but no necessarily for your quick short messages). Therefore I quickly stiched together klod, originally in bash, still living as [./main.sh](main.sh) at the root of the project.

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
KLOD_LOGS=false
KLOD_LOG_FILE=
EOF
```

### Configuration options:

- `ANTHROPIC_API_KEY` (required): Your Anthropic API key
- `MODEL` (optional): Model to use (default: claude-sonnet-4-5-20250929)
- `SYSTEM_PROMPT` (optional): Custom system prompt for Claude
- `KLOD_LOGS` (optional): Enable conversation logging (default: false, set to "true" or "1" to enable)
- `KLOD_LOG_FILE` (optional): Custom log file path (default: `$XDG_STATE_HOME/klod/conversations.log` or `~/.local/state/klod/conversations.log`)

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
