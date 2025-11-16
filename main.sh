#!/bin/bash

source .env

PROMPT="$*"
TEMP_FILE=$(mktemp)
prompt=$(echo $PROMPT)
ANTHROPIC_MODEL=${MODEL:-"claude-sonnet-4-5-20250929"}
SYSTEM_PROMPT=${SYSTEM_PROMPT:-""}

JSON_PAYLOAD=$(jq -n \
  --arg model "$ANTHROPIC_MODEL" \
  --arg prompt "$PROMPT" \
  --arg system_prompt "$SYSTEM_PROMPT" \
  '{
    model: $model,
    system: $system_prompt,
    max_tokens: 2048,
    messages: [{role: "user", content: $prompt}]
  }')

# Show loading message
echo -ne "\033[36mFetching response from Claude...\033[0m"

curl -s https://api.anthropic.com/v1/messages \
     --header "x-api-key: $ANTHROPIC_API_KEY" \
     --header "anthropic-version: 2023-06-01" \
     --header "content-type: application/json" \
     --data "$JSON_PAYLOAD" > "$TEMP_FILE"

# Clear loading message
echo -ne "\r\033[K"

# Display response
jq -r '.content[0].text' "$TEMP_FILE"
