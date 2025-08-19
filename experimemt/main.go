package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Tool represents a function or capability the agent can use.
type Tool struct {
	Name        string
	Description string
	Function    func(args map[string]interface{}) (string, error)
	Args        map[string]string // Maps argument names to their descriptions
}

// ToolInvocation represents the data extracted from the LLM's response
// to determine which tool to call.
type ToolInvocation struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"arguments"`
}

// OllamaRequest is the structure for a prompt sent to the Ollama API.
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse is the structure for the response from the Ollama API.
type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

// Agent represents our agentic system.
type Agent struct {
	OllamaURL string
	Model     string
	Tools     map[string]Tool
}

// NewAgent initializes a new Agent with the given configuration.
func NewAgent(ollamaURL, model string) *Agent {
	return &Agent{
		OllamaURL: ollamaURL,
		Model:     model,
		Tools:     make(map[string]Tool),
	}
}

// GetConversationHistory fetches the conversation history from a local file.
func (a *Agent) GetConversationHistory(filePath string) (string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File does not exist, return an empty history
		return "", nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read conversation history file: %v", err)
	}

	return string(data), nil
}

// SaveConversationHistory saves the conversation history to a local file.
func (a *Agent) SaveConversationHistory(filePath, history string) error {
	err := os.WriteFile(filePath, []byte(history), 0644)
	if err != nil {
		return fmt.Errorf("failed to save conversation history to file: %v", err)
	}
	return nil
}

// AddTool registers a new tool with the agent.
func (a *Agent) AddTool(tool Tool) {
	a.Tools[tool.Name] = tool
}

// GetToolsPrompt generates a string description of all available tools for the LLM.
func (a *Agent) GetToolsPrompt() string {
	var sb strings.Builder
	sb.WriteString("AVAILABLE TOOLS:\n")
	for _, tool := range a.Tools {
		sb.WriteString(fmt.Sprintf("Name: %s\n", tool.Name))
		sb.WriteString(fmt.Sprintf("Description: %s\n", tool.Description))
		sb.WriteString(fmt.Sprintf("Arguments: %v\n\n", tool.Args))
	}
	return sb.String()
}

// GeneratePrompt crafts the full prompt for the LLM, including user input, tool descriptions, and instructions.
func (a *Agent) GeneratePrompt(history, userInput string) string {
	toolsPrompt := a.GetToolsPrompt()
	return fmt.Sprintf(`
You are a helpful assistant. You have access to the following tools:

%s

The user has given you a task. You should think step-by-step and then decide to either use one of the tools or respond with the final answer.
Your final response should start with 'Final Answer:'.

IMPORTANT: If the user's query contains the phrase "chris tanti", you must use the `+"`chris_tanti`"+` tool.

Thought: You should always think about what to do first, before using a tool.
Action: To use a tool, you must use the following JSON format:
{ "name": "tool_name", "arguments": { "arg1": "value1", "arg2": "value2" } }
Observation: The result of the tool's action.

Current conversation history:
%s
User: %s`, toolsPrompt, history, userInput)
}

// Run executes the agentic loop for a given user input.
func (a *Agent) Run(historyFilePath, userInput string) (string, error) {
	// Load the history for this user
	history, err := a.GetConversationHistory(historyFilePath)
	if err != nil {
		return "", err
	}

	// --- NEW LOGIC: Pre-process the user's input to force tool use ---
	if strings.Contains(strings.ToLower(userInput), "chris tanti") {
		log.Println("--- User query contains 'chris tanti', directly invoking tool ---")
		tool := a.Tools["chris_tanti"]
		args := map[string]interface{}{"query": userInput}
		toolResult, err := tool.Function(args)
		if err != nil {
			log.Printf("Tool execution failed: %v\n", err)
			return "", err
		}
		// Return the result directly without a full LLM loop
		return toolResult, nil
	}
	// --- END NEW LOGIC ---

	for i := 0; i < 5; i++ { // Limit the number of steps to prevent infinite loops
		// 1. Plan: Get the LLM's next action
		prompt := a.GeneratePrompt(history, userInput)
		log.Println("--- Sending prompt to LLM ---")
		log.Println(prompt)
		response, err := a.CallOllama(prompt)
		if err != nil {
			return "", err
		}
		log.Println("--- Received response from LLM ---")
		log.Println(response)

		// 2. Act: Parse the response and execute the tool or provide the final answer.
		if strings.HasPrefix(response, "Final Answer:") {
			finalAnswer := strings.TrimSpace(strings.TrimPrefix(response, "Final Answer:"))
			history += "\nAssistant: " + finalAnswer
			a.SaveConversationHistory(historyFilePath, history)
			return finalAnswer, nil
		}

		// Use a regular expression to extract the JSON action
		re := regexp.MustCompile(`(?s)\{ "name": ".*?" \}`)
		matches := re.FindStringSubmatch(response)
		if len(matches) == 0 {
			return "", fmt.Errorf("could not find a valid tool action in the LLM's response")
		}

		var toolCall ToolInvocation
		// The JSON is likely part of a larger string, so we'll try to find the complete JSON object
		jsonString := matches[0]
		err = json.Unmarshal([]byte(jsonString), &toolCall)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal tool invocation JSON: %v", err)
		}

		tool, ok := a.Tools[toolCall.Name]
		if !ok {
			return "", fmt.Errorf("unknown tool: %s", toolCall.Name)
		}

		// 3. Reflect & Observe: Execute the tool and add the observation to the history.
		log.Printf("--- Calling tool: %s with arguments: %v ---\n", tool.Name, toolCall.Args)
		toolResult, err := tool.Function(toolCall.Args)
		if err != nil {
			log.Printf("Tool execution failed: %v\n", err)
			history += fmt.Sprintf("\nObservation: Tool execution failed with error: %v", err)
		} else {
			log.Printf("--- Tool result: %s ---\n", toolResult)
			history += fmt.Sprintf("\nObservation: %s", toolResult)
		}

		// Save the updated history for the next loop iteration or next run
		a.SaveConversationHistory(historyFilePath, history)
	}

	return "", fmt.Errorf("agent failed to find a final answer within the maximum number of steps")
}

// CallOllama sends a request to the Ollama server and returns the full response string.
func (a *Agent) CallOllama(prompt string) (string, error) {
	reqData := OllamaRequest{
		Model:  a.Model,
		Prompt: prompt,
		Stream: false, // For simplicity, we get the full response at once
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %v", err)
	}

	req, err := http.NewRequest("POST", a.OllamaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp OllamaResponse
	err = json.NewDecoder(resp.Body).Decode(&ollamaResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %v", err)
	}

	return ollamaResp.Response, nil
}

// Main function to run the agent.
func main() {
	// Set up the agent
	ollamaURL := "http://ollama.localhost:11434/api/generate"
	model := "deepseek-r1:1.5b"
	historyFilePath := "conversation_history.json"

	agent := NewAgent(ollamaURL, model)

	// Add the "calculator" tool
	agent.AddTool(Tool{
		Name:        "calculator",
		Description: "A tool that can perform basic arithmetic operations.",
		Args:        map[string]string{"operation": "string (e.g., 'add', 'subtract', 'multiply', 'divide')", "num1": "number", "num2": "number"},
		Function: func(args map[string]interface{}) (string, error) {
			op, ok := args["operation"].(string)
			if !ok {
				return "", fmt.Errorf("missing 'operation' argument")
			}
			num1, ok := args["num1"].(float64)
			if !ok {
				return "", fmt.Errorf("missing or invalid 'num1' argument")
			}
			num2, ok := args["num2"].(float64)
			if !ok {
				return "", fmt.Errorf("missing or invalid 'num2' argument")
			}

			var result float64
			switch op {
			case "add":
				result = num1 + num2
			case "subtract":
				result = num1 - num2
			case "multiply":
				result = num1 * num2
			case "divide":
				if num2 == 0 {
					return "", fmt.Errorf("division by zero")
				}
				result = num1 / num2
			default:
				return "", fmt.Errorf("unsupported operation: %s", op)
			}
			return fmt.Sprintf("%.2f", result), nil
		},
	})

	// Add a simple "web_search" tool
	agent.AddTool(Tool{
		Name:        "web_search",
		Description: "A tool that can search the internet for information.",
		Args:        map[string]string{"query": "string"},
		Function: func(args map[string]interface{}) (string, error) {
			query, ok := args["query"].(string)
			if !ok {
				return "", fmt.Errorf("missing 'query' argument")
			}
			// In a real application, this would call a search API.
			// For this example, we'll just return a mock response.
			return fmt.Sprintf("Search results for '%s': The weather is currently 75°F and sunny.", query), nil
		},
	})

	// Add a simple "chris_tanti" tool
	agent.AddTool(Tool{
		Name:        "chris_tanti",
		Description: "A tool that provides specific, predefined information about the individual Chris Tanti.",
		Args:        map[string]string{"query": "string"},
		Function: func(args map[string]interface{}) (string, error) {
			query, ok := args["query"].(string)
			if !ok {
				return "", fmt.Errorf("missing 'query' argument")
			}
			if strings.Contains(strings.ToLower(query), "chris tanti") {
				return "Chris Tanti is a prominent figure from cardiff that has made significant contributions to the fields of DEI in the workplace", nil
			}
			return "I cannot provide information about this person.", nil
		},
	})

	// Get user input from command line
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run main.go \"Your question here\"")
	}
	userInput := strings.Join(os.Args[1:], " ")

	// Run the agent
	log.Printf("Starting agent with prompt: %s\n", userInput)
	finalAnswer, err := agent.Run(historyFilePath, userInput)
	if err != nil {
		log.Fatalf("Agent failed with error: %v", err)
	}

	fmt.Println("\n--- Final Answer ---")
	fmt.Println(finalAnswer)
}
