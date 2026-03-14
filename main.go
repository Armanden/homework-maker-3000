package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
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

	systemPrompt := ""

	if mode == "slides" {
		systemPrompt = "Generate a Typst slideshow using Polylux. Only output valid Typst code."
	} else {
		systemPrompt = "Generate a structured Typst homework document. Only output valid Typst code."
	}

	reqBody := Request{
		Model: "gpt-4.1-mini",
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result Response
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Choices[0].Message.Content, nil
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

	fmt.Println("Generating with AI...")

	typstCode, err := generateTypst(prompt, mode)
	if err != nil {
		panic(err)
	}

	file := "homework.typ"

	os.WriteFile(file, []byte(typstCode), 0644)

	fmt.Println("Typst file written.")

	cmd := exec.Command("typst", "compile", file, "homework.pdf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("Typst compile failed:", err)
		return
	}

	fmt.Println("Finished → homework.pdf")
}