#!/bin/bash

source .env

PROMPT="$*"
TEMP_FILE=$(mktemp)
ANTHROPIC_MODEL=${MODEL:-"claude-sonnet-4-5-20250929"}
SYSTEM_PROMPT=${SYSTEM_PROMPT:-""}

# Initialize conversation history array
declare -a CONVERSATION_HISTORY=()

# Function to build messages JSON array from conversation history
build_messages_json() {
    local messages_json="["
    local first=true

    for ((i=0; i<${#CONVERSATION_HISTORY[@]}; i+=2)); do
        if [ "$first" = true ]; then
            first=false
        else
            messages_json+=","
        fi

        # Add user message
        local user_msg="${CONVERSATION_HISTORY[$i]}"
        user_msg="${user_msg//\\/\\\\}"  # Escape backslashes
        user_msg="${user_msg//\"/\\\"}"  # Escape quotes
        user_msg="${user_msg//$'\n'/\\n}"  # Escape newlines
        messages_json+="{\"role\":\"user\",\"content\":\"$user_msg\"}"

        # Add assistant message if it exists
        if [ $((i+1)) -lt ${#CONVERSATION_HISTORY[@]} ]; then
            local assistant_msg="${CONVERSATION_HISTORY[$((i+1))]}"
            assistant_msg="${assistant_msg//\\/\\\\}"
            assistant_msg="${assistant_msg//\"/\\\"}"
            assistant_msg="${assistant_msg//$'\n'/\\n}"
            messages_json+=",{\"role\":\"assistant\",\"content\":\"$assistant_msg\"}"
        fi
    done

    messages_json+="]"
    echo "$messages_json"
}

# Function to send message and get response
send_message() {
    local messages_json=$(build_messages_json)

    JSON_PAYLOAD=$(jq -n \
        --arg model "$ANTHROPIC_MODEL" \
        --arg system_prompt "$SYSTEM_PROMPT" \
        --argjson messages "$messages_json" \
        '{
            model: $model,
            system: $system_prompt,
            max_tokens: 2048,
            messages: $messages
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

    # Extract and display response
    local response=$(jq -r '.content[0].text' "$TEMP_FILE")
    echo "$response"

    # Add assistant's response to conversation history
    CONVERSATION_HISTORY+=("$response")
}

# Add initial user prompt to conversation history
CONVERSATION_HISTORY+=("$PROMPT")

# Send initial message
send_message

# Conversation loop
while true; do
    echo ""
    echo -ne "\033[32mYou (or 'exit' to quit): \033[0m"
    read -e user_input

    # Check if user wants to exit
    if [[ "$user_input" == "exit" ]] || [[ "$user_input" == "quit" ]] || [[ -z "$user_input" ]]; then
        echo "Goodbye!"
        break
    fi

    # Add user input to conversation history
    CONVERSATION_HISTORY+=("$user_input")

    # Send message and get response
    send_message
done

# Cleanup
rm -f "$TEMP_FILE"
