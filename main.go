package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ollama/ollama/api"
)

// Action represents a command that can be executed.
type Action struct {
	Label   string `json:"label"`
	Command string `json:"command"`
}

// StructuredResponse represents a response from the agent that includes actions.
type StructuredResponse struct {
	Text    string   `json:"text"`
	Actions []Action `json:"actions"`
}

// This program demonstrates a basic AI agent that interacts with the gemma:270mb model
// using the official Ollama Go library.

// Prerequisites:
// 1. Install Ollama from https://ollama.com/.
// 2. Start the Ollama server by running `ollama serve` in your terminal.
// 3. Pull the gemma:270mb model by running `ollama pull gemma:270mb`.
// 4. Set up your Go module: `go mod init gemma_agent`
// 5. Get the Ollama API library: `go get github.com/ollama/ollama/api`

func main() {
	fmt.Println("Welcome! I am an agent powered by the gemma:270mb model.")
	fmt.Println("Type 'exit' or 'quit' to end the conversation.")

	url := &url.URL{
		Scheme: "http",
		Host:   "ollama.localhost",
		Path:   "/",
	}

	// Create a new Ollama API client.

	httpClient := http.DefaultClient
	client := api.NewClient(url, httpClient)

	// Store the conversation history. This is crucial for the agent to remember context.
	var messages []api.Message

	// Add a system message to instruct the model on the expected JSON format.
	systemMessage := `You are a helpful assistant. When the user does asks for a command-line action, you must respond with a JSON object with the following structure Otherwise, you should respond as a normal chatbot:
	{
	  "text": "Your response text",
	  "actions": [
	    {
	      "label": "A short description of the action",
	      "command": "The command to execute"
	    }
	  ]
	}
	.`
	messages = append(messages, api.Message{
		Role:    "system",
		Content: systemMessage,
	})

	// Create a context for the chat request.
	ctx := context.Background()

	// Use bufio.NewScanner to read user input line by line.
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break // End of input
		}
		user_input := scanner.Text()

		// Check for exit commands
		if user_input == "exit" || user_input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Add the user's message to the conversation history
		messages = append(messages, api.Message{
			Role:    "user",
			Content: user_input,
		})

		// Send the conversation history to the model for a response.
		// We use a handler function to process the streamed response.
		fmt.Println("Thinking...")
		fmt.Print("Agent: ")

		// Create a new request with the current conversation history.
		req := &api.ChatRequest{
			Model:    "gemma3:270m",
			Messages: messages,
		}

		// The Chat function is a streaming function, so we need to collect all chunks.
		// The Chat function is a streaming function. We'll print the content
		// as it comes in and also collect it for the history.
		var fullResponse string
		handler := func(resp api.ChatResponse) error {
			fmt.Print(resp.Message.Content)
			fullResponse += resp.Message.Content
			return nil
		}

		err := client.Chat(ctx, req, handler)
		if err != nil {
			log.Println("An error occurred with Ollama:", err)
			log.Println("Please ensure the Ollama server is running and the 'gemma:270mb' model is available.")
			// Optionally, break here if you want to stop on error.
			continue
		}
		fmt.Println() // Newline after the agent's response is complete.

		// Clean up the response, removing markdown code blocks if present.
		responseStr := fullResponse
		if strings.HasPrefix(responseStr, "```json") {
			responseStr = strings.TrimPrefix(responseStr, "```json")
			responseStr = strings.TrimSuffix(responseStr, "```")
		}
		responseStr = strings.TrimSpace(responseStr)

		// Try to parse the response as a structured response with actions.
		var structuredResp StructuredResponse
		err = json.Unmarshal([]byte(responseStr), &structuredResp)
		if err == nil && len(structuredResp.Actions) > 0 {
			fmt.Println(structuredResp.Text)
			for i, action := range structuredResp.Actions {
				fmt.Printf("%d: %s\n", i+1, action.Label)
			}
			fmt.Print("Choose an action to execute (or press Enter to continue): ")

			scanner.Scan()
			choiceStr := scanner.Text()
			if choiceStr != "" {
				choice, err := strconv.Atoi(choiceStr)
				if err == nil && choice > 0 && choice <= len(structuredResp.Actions) {
					selectedAction := structuredResp.Actions[choice-1]
					fmt.Printf("Executing: %s\n", selectedAction.Command)
					cmd := exec.Command("bash", "-c", selectedAction.Command)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					if err != nil {
						log.Printf("Error executing command: %v\n", err)
					}
				} else {
					fmt.Println("Invalid choice.")
				}
			}
		} else {
			// Print the agent's full response as plain text.
			fmt.Printf("Agent: %s\n", fullResponse)
		}

		// Add the agent's full response to the conversation history to maintain context.
		messages = append(messages, api.Message{
			Role:    "assistant",
			Content: fullResponse,
		})
	}
}
