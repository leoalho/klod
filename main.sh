#!/bin/bash

source .env

PROMPT="$*"
TEMP_FILE=$(mktemp)
prompt=$(echo $PROMPT)

JSON_PAYLOAD=$(jq -n \
  --arg model "claude-sonnet-4-5-20250929" \
  --arg prompt "$PROMPT" \
  '{
    model: $model,
    max_tokens: 1024,
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
rm "$TEMP_FILE"
