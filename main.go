package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/joho/godotenv"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIRequest struct {
	Model     string    `json:"model"`
	System    string    `json:"system,omitempty"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
}

type ContentBlock struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type APIResponse struct {
	Content []ContentBlock `json:"content"`
	Error   *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type StreamEvent struct {
	Type  string       `json:"type"`
	Delta *StreamDelta `json:"delta,omitempty"`
	Error *StreamError `json:"error,omitempty"`
}

type StreamDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type StreamError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

var conversationHistory []Message
var loggingEnabled bool
var logFilePath string

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getTerminalWidth() int {
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return 80 // Default fallback
	}
	return int(ws.Col)
}

func printSeparator() {
	width := getTerminalWidth()
	fmt.Println(strings.Repeat("─", width))
}

func logConversation(message Message) error {
	if !loggingEnabled {
		return nil
	}

	// Ensure log directory exists
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file in append mode
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Write timestamp and message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s\n", timestamp, strings.ToUpper(message.Role), message.Content)

	if _, err := file.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}

func loadConfig() error {
	// Try loading config from multiple locations in order
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPaths := []string{
		filepath.Join(homeDir, ".config", "klod", "config"), // XDG standard
		filepath.Join(homeDir, ".klod.env"),                 // Home directory
		".env",                                              // Current directory (for project-specific overrides)
	}

	var lastErr error
	for _, path := range configPaths {
		err := godotenv.Load(path)
		if err == nil {
			return nil // Successfully loaded
		}
		lastErr = err
	}

	return lastErr
}

func main() {
	// Load config file
	err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config file: %v\n", err)
	}

	// Get configuration from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: ANTHROPIC_API_KEY not set in .env file")
		os.Exit(1)
	}

	model := os.Getenv("MODEL")
	if model == "" {
		model = "claude-sonnet-4-5-20250929"
	}

	systemPrompt := os.Getenv("SYSTEM_PROMPT")

	// Configure logging
	loggingEnabled = os.Getenv("KLOD_LOGS") == "true" || os.Getenv("KLOD_LOGS") == "1"
	logFilePath = os.Getenv("KLOD_LOG_FILE")
	if logFilePath == "" {
		// Use XDG_STATE_HOME or default to ~/.local/state per XDG spec
		stateDir := os.Getenv("XDG_STATE_HOME")
		if stateDir == "" {
			homeDir, _ := os.UserHomeDir()
			stateDir = filepath.Join(homeDir, ".local", "state")
		}
		// Create a unique log file for this session based on timestamp
		sessionTime := time.Now().Format("2006-01-02_15-04-05")
		logFilePath = filepath.Join(stateDir, "klod", "sessions", sessionTime+".log")
	}

	// Get initial prompt from command-line arguments
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: klod <your prompt>")
		os.Exit(1)
	}

	initialPrompt := strings.Join(os.Args[1:], " ")

	// Add initial user message to conversation history
	userMsg := Message{
		Role:    "user",
		Content: initialPrompt,
	}
	conversationHistory = append(conversationHistory, userMsg)
	logConversation(userMsg)

	// Send initial message
	printSeparator()
	fmt.Print("\033[34mAssistant: \033[0m")
	response, err := sendMessage(apiKey, model, systemPrompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	printSeparator()

	// Add assistant's response to conversation history
	assistantMsg := Message{
		Role:    "assistant",
		Content: response,
	}
	conversationHistory = append(conversationHistory, assistantMsg)
	logConversation(assistantMsg)

	// Interactive conversation loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\033[32mYou (or 'exit' to quit): \033[0m")

		userInput, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		userInput = strings.TrimSpace(userInput)

		// Check if user wants to exit
		if userInput == "exit" || userInput == "quit" || userInput == "" {
			fmt.Println("Goodbye!")
			break
		}

		// Add user input to conversation history
		userMsg := Message{
			Role:    "user",
			Content: userInput,
		}
		conversationHistory = append(conversationHistory, userMsg)
		logConversation(userMsg)

		// Send message and get response
		printSeparator()
		fmt.Print("\033[34mAssistant: \033[0m")
		response, err := sendMessage(apiKey, model, systemPrompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		printSeparator()

		// Add assistant's response to conversation history
		assistantMsg := Message{
			Role:    "assistant",
			Content: response,
		}
		conversationHistory = append(conversationHistory, assistantMsg)
		logConversation(assistantMsg)
	}
}

func sendMessage(apiKey, model, systemPrompt string) (string, error) {
	// Build API request with streaming enabled
	request := APIRequest{
		Model:     model,
		System:    systemPrompt,
		MaxTokens: 2048,
		Messages:  conversationHistory,
		Stream:    true,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API call
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read and parse streaming response
	var fullResponse strings.Builder
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to read stream: %w", err)
		}

		line = strings.TrimSpace(line)

		// Skip empty lines and non-data lines
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract JSON data
		data := strings.TrimPrefix(line, "data: ")

		// Skip the [DONE] message
		if data == "[DONE]" {
			break
		}

		// Parse the event
		var event StreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue // Skip malformed events
		}

		// Check for errors
		if event.Error != nil {
			return "", fmt.Errorf("API error: %s - %s", event.Error.Type, event.Error.Message)
		}

		// Print text deltas as they arrive
		if event.Type == "content_block_delta" && event.Delta != nil && event.Delta.Type == "text_delta" {
			fmt.Print(event.Delta.Text)
			fullResponse.WriteString(event.Delta.Text)
		}
	}

	fmt.Println() // Add newline after streaming response

	return fullResponse.String(), nil
}
