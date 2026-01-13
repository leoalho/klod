# prompt

A simple, fast CLI tool for chatting with Claude using the Anthropic API with real-time streaming responses.

## Features

- Interactive chat with Claude in your terminal
- Real-time streaming responses (see text as Claude types)
- Maintains conversation history within a session
- Configurable model and system prompts
- Multi-location config file support

## Requirements

- Go 1.21 or higher
- Anthropic API key ([get one here](https://console.anthropic.com/))

## Installation

1. Clone and build:
```bash
git clone <your-repo-url>
cd anthropic-cli
go build -o prompt
```

2. Create a symlink for global access:
```bash
sudo ln -s $(pwd)/prompt /usr/local/bin/prompt
```

Or install via Go:
```bash
go install
```

## Configuration

The tool looks for configuration files in the following order:
1. `.env` in the current directory (for project-specific configs)
2. `~/.config/prompt/config` (XDG standard location)
3. `~/.prompt.env` (home directory)

### Setup your config:

```bash
# Create the config directory
mkdir -p ~/.config/prompt

# Create config file
cat > ~/.config/prompt/config << EOF
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
prompt "Hello, how are you?"
```

This will:
1. Send your initial message to Claude
2. Stream the response in real-time
3. Enter an interactive chat mode where you can continue the conversation

Type `exit` or `quit` to end the conversation.

## Examples

```bash
# Ask a quick question
prompt "What is the capital of France?"

# Start a coding session
prompt "Help me write a Python function to calculate fibonacci numbers"

# Use a different model (set in config)
# Edit your config file and change MODEL=claude-opus-4-5-20251101
prompt "Explain quantum computing"
```

## Development

Run without building:
```bash
go run main.go "your message here"
```

## License

MIT
