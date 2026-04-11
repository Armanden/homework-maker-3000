package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Response struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func generateTypst(prompt string, mode string) (string, error) {

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	var systemPrompt string

	mode = strings.ToLower(mode)

	if mode == "slides" {
		systemPrompt = `Generate a valid Typst slideshow using Polylux.
Only output Typst code.
Do not include explanations or markdown.
Ensure the code compiles without errors.`
	} else {
		systemPrompt = `Generate a structured Typst homework document.
Only output Typst code.
Do not include explanations or markdown.
Ensure the code compiles without errors.`
	}

	reqBody := Request{
		Model: "gpt-4.1-mini",
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", string(bodyBytes))
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	typstCode := result.Choices[0].Message.Content

	// Clean markdown wrappers if present
	typstCode = strings.TrimSpace(typstCode)
	typstCode = strings.TrimPrefix(typstCode, "```typst")
	typstCode = strings.TrimPrefix(typstCode, "```")
	typstCode = strings.TrimSuffix(typstCode, "```")
	typstCode = strings.TrimSpace(typstCode)

	return typstCode, nil
}

func main() {

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("AI Homework Maker")

	fmt.Print("Prompt: ")
	prompt, _ := reader.ReadString('\n')
	prompt = strings.TrimSpace(prompt)

	fmt.Print("Output (pdf/slides): ")
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)

	fmt.Print("Output filename (without extension): ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		name = "homework"
	}

	fmt.Println("Generating with AI...")

	typstCode, err := generateTypst(prompt, mode)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Debug preview (optional)
	fmt.Println("\n----- GENERATED CODE -----")
	fmt.Println(typstCode)
	fmt.Println("--------------------------\n")

	file := name + ".typ"

	err = os.WriteFile(file, []byte(typstCode), 0644)
	if err != nil {
		fmt.Println("Failed to write file:", err)
		return
	}

	fmt.Println("Typst file written:", file)

	// Check if typst exists
	if _, err := exec.LookPath("typst"); err != nil {
		fmt.Println("Typst is not installed or not in PATH")
		return
	}

	outputPDF := name + ".pdf"

	cmd := exec.Command("typst", "compile", file, outputPDF)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("Typst compile failed:", err)
		return
	}

	fmt.Println("Finished →", outputPDF)
}
