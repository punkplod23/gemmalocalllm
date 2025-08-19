package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// ==============================================================================
// Ollama Agent Executor in Go
//
// This program acts as the secure intermediary between a local Ollama LLM
// and your system's commands. It sends a user's prompt to Ollama, receives a
// tool call, and executes the corresponding local script.
// ==============================================================================

// Tool represents a structured definition of a tool for the LLM.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// OllamaRequest is the structure for the API call to Ollama's /api/generate endpoint.
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Tools  []Tool `json:"tools"`
	Stream bool   `json:"stream"`
	Raw    bool   `json:"raw"`
	Format string `json:"format"`
}

// OllamaResponse is the structure for the API's streaming response.
type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Done      bool   `json:"done"`
	Content   string `json:"content"`
}

// ToolCall represents the structured JSON output from the Ollama LLM.
type ToolCall struct {
	ToolName   string   `json:"tool_name"`
	Parameters struct{} `json:"parameters"` // A simple empty struct for this tool
}

// ToolResult represents the structured JSON output to be sent back to the LLM.
type ToolResult struct {
	ToolName   string `json:"tool_name"`
	Status     string `json:"status"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ReturnCode int    `json:"return_code"`
}

// executeToolCall parses a tool call from Ollama and executes the script.
// It now contains the logic to check and clean up disk space natively in Go.
func executeToolCall(toolCallJSON string) ([]byte, error) {
	var callData ToolCall
	err := json.Unmarshal([]byte(toolCallJSON), &callData)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON format for tool call: %w", err)
	}

	if callData.ToolName == "run_disk_agent" {
		fmt.Println("Executing tool: run_disk_agent")

		// --- BEGIN NATIVE GO LOGIC ---
		// Instead of calling an external script, we perform the actions directly.
		var usage int
		var finalOutput bytes.Buffer

		switch runtime.GOOS {
		case "linux":
			// Step 1: Check disk usage on Linux using 'df -h /'
			cmd := exec.Command("df", "-h", "/")
			output, err := cmd.CombinedOutput()

			if err != nil {
				return nil, fmt.Errorf("failed to check disk usage: %w", err)
			}

			// A simplified way to parse the output of 'df -h' for the root filesystem.
			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 5 {
					usageStr := strings.TrimRight(fields[4], "%")
					usage, _ = strconv.Atoi(usageStr)
				}
			}

		case "windows":
			// Step 1: Check disk usage on Windows using 'wmic logicaldisk'
			cmd := exec.Command("cmd", "/c", "wmic logicaldisk get size,freespace,caption")
			output, err := cmd.CombinedOutput()

			if err != nil {
				return nil, fmt.Errorf("failed to check disk usage: %w", err)
			}

			// Parse the WMIC output
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, ":") {
					fields := strings.Fields(line)
					if len(fields) >= 3 {
						size, _ := strconv.ParseFloat(fields[1], 64)
						free, _ := strconv.ParseFloat(fields[2], 64)
						if size > 0 {
							usage = int(((size - free) / size) * 100)
							break // We just care about the first disk for this example
						}
					}
				}
			}

		default:
			errorResult := map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Unsupported operating system: %s", runtime.GOOS),
			}
			resultJSON, _ := json.MarshalIndent(errorResult, "", "  ")
			return resultJSON, fmt.Errorf("unsupported OS")
		}

		// Step 2: Decide whether to prune based on the usage threshold (e.g., 75%)
		finalOutput.WriteString(fmt.Sprintf("Current disk usage on root ('/') is %d%%.\n", usage))

		if usage > 75 {
			finalOutput.WriteString("Usage is high, starting Docker system prune.\n")

			// Step 3: Run the prune command if the condition is met.
			// The use of hard-coded commands here is a form of sanitization.
			// There is no user input to inject malicious commands.
			pruneCmd := exec.Command("docker-compose", "system", "prune", "-f")
			pruneOutput, pruneErr := pruneCmd.CombinedOutput()

			if pruneErr != nil {
				finalOutput.WriteString(fmt.Sprintf("Error during prune: %s\n", pruneErr.Error()))
			}

			finalOutput.Write(pruneOutput)
			finalOutput.WriteString("\nDocker prune complete.\n")
		} else {
			finalOutput.WriteString("Usage is within a safe limit. No action required.\n")
		}

		// --- END NATIVE GO LOGIC ---

		result := ToolResult{
			ToolName:   "run_disk_agent",
			Status:     "success",
			Stdout:     finalOutput.String(),
			Stderr:     "",
			ReturnCode: 0,
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			// Handle the marshaling error gracefully
			return nil, fmt.Errorf("failed to marshal JSON result: %w", err)
		}
		return resultJSON, nil
	} else {
		errorResult := map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Unknown tool '%s'.", callData.ToolName),
		}
		resultJSON, _ := json.MarshalIndent(errorResult, "", "  ")
		return resultJSON, fmt.Errorf("unknown tool")
	}
}

// main function demonstrates the full agentic loop.
func main() {
	// 1. Define the tool to make Ollama aware of our script.
	toolDef := Tool{
		Name:        "run_disk_agent",
		Description: "Checks the server's disk usage on the root filesystem and performs an automatic Docker system prune if space is low.",
		Parameters:  map[string]interface{}{},
	}

	// 2. Prepare the API request with a user prompt and the tool definition.
	requestData := OllamaRequest{
		Model:  "phi4-mini-reasoning:latest",
		Prompt: "Check the disk space and perform a cleanup if necessary.",
		Tools:  []Tool{toolDef},
		Stream: false,
		Format: "json",
		Raw:    true,
	}

	jsonBytes, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("Error marshaling JSON request:", err)
		return
	}

	// 3. Send the request to the local Ollama API.
	resp, err := http.Post("http://ollama.localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fmt.Println("Error making API request:", err)
		return
	}
	defer resp.Body.Close()

	// 4. Parse the response from Ollama.
	var ollamaResponse OllamaResponse
	err = json.NewDecoder(resp.Body).Decode(&ollamaResponse)
	if err != nil {
		fmt.Println("Error decoding API response:", err)
		return
	}

	// --- FIX: Add a check for an empty response before attempting to execute. ---
	if ollamaResponse.Content == "" {
		fmt.Println("\nModel did not return a tool call. Please ensure your Ollama model is properly configured and can perform tool-use reasoning.")
		fmt.Println("------------------------------------------------------")
		return
	}

	// 5. The Ollama response is a JSON string containing the tool call.
	// We pass this content to our executor function.
	executionResult, err := executeToolCall(ollamaResponse.Content)
	if err != nil {
		fmt.Printf("Error executing tool: %v\n", err)
	}

	fmt.Printf("\n--- Result sent back to the LLM for final response ---\n%s\n", string(executionResult))
	fmt.Println("------------------------------------------------------")
}
