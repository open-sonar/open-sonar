package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// local model
type OllamaClient struct{}

type OllamaRequest struct {
	Model string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func (o *OllamaClient) GenerateResponse(query string) (string, error) {
	reqBody, err := json.Marshal(OllamaRequest{
		Model: "mistral", // default 
		Prompt: query,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %v", err)
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("invalid Ollama response: %v", err)
	}

	return ollamaResp.Response, nil
}
