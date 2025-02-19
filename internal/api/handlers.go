package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"open-sonar/internal/llm"
	"open-sonar/internal/models"
	"open-sonar/internal/search/webscrape"
	"open-sonar/internal/utils"
)

// TestHandler returns a simple status message.
func TestHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"message": "open-sonar server is running.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ChatCompletionsHandler handles direct LLM calls (no search).
func ChatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	var chatReq models.ChatRequest
	if err := json.Unmarshal(bodyBytes, &chatReq); err != nil {
		utils.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if chatReq.Query == "" {
		utils.Error("Missing 'query' in request")
		http.Error(w, "Query field is required", http.StatusBadRequest)
		return
	}

	// Select the correct LLM provider (OpenAI, Anthropic, Ollama, etc.)
	provider, err := llm.NewLLMProvider(chatReq.Provider)
	if err != nil {
		utils.Error(fmt.Sprintf("Invalid LLM provider: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the LLM
	response, err := provider.GenerateResponse(chatReq.Query)
	if err != nil {
		utils.Error(fmt.Sprintf("LLM call failed: %v", err))
		http.Error(w, fmt.Sprintf("LLM processing error: %v", err), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	jsonResponse := map[string]string{"response": response}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonResponse)
}

// ChatHandler decides whether to perform a web search before calling the LLM.
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	var chatReq models.ChatRequest
	if err := json.Unmarshal(bodyBytes, &chatReq); err != nil {
		utils.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if chatReq.Query == "" {
		utils.Error("Missing 'query' in request")
		http.Error(w, "Query field required", http.StatusBadRequest)
		return
	}

	utils.Info(fmt.Sprintf("Received query: %q (NeedSearch=%v, Pages=%d, Retries=%d)",
		chatReq.Query, chatReq.NeedSearch, chatReq.Pages, chatReq.Retries,
	))

	pages := chatReq.Pages
	if pages <= 0 {
		pages = 3
	}
	retries := chatReq.Retries
	if retries <= 0 {
		retries = 2
	}

	// Decision Engine
	if chatReq.NeedSearch {
		utils.Info("Performing web search before calling LLM...")

		// Scrape for relevant context
		results := webscrape.Scrape(chatReq.Query, pages, retries)

		// Initialize LLM provider
		provider, err := llm.NewLLMProvider(chatReq.Provider)
		if err != nil {
			utils.Error(fmt.Sprintf("Invalid LLM provider: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Format top 3 search results to pass as additional context
		searchContext := formatSearchResults(results)

		// Combine user's query with search context
		response, err := provider.GenerateResponse(chatReq.Query + "\n\n" + searchContext)
		if err != nil {
			utils.Error(fmt.Sprintf("LLM call failed: %v", err))
			http.Error(w, fmt.Sprintf("LLM processing error: %v", err), http.StatusInternalServerError)
			return
		}

		jsonResponse := map[string]interface{}{
			"decision":   "search + LLM call",
			"pages_used": len(results),
			"response":   response,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jsonResponse)
		return
	}

	utils.Info("No search needed, calling LLM directly.")

	// Initialize LLM provider
	provider, err := llm.NewLLMProvider(chatReq.Provider)
	if err != nil {
		utils.Error(fmt.Sprintf("Invalid LLM provider: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call LLM directly
	response, err := provider.GenerateResponse(chatReq.Query)
	if err != nil {
		utils.Error(fmt.Sprintf("LLM call failed: %v", err))
		http.Error(w, fmt.Sprintf("LLM processing error: %v", err), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	jsonResponse := map[string]interface{}{
		"decision": "direct LLM call",
		"response": response,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonResponse)
}

// formatSearchResults formats up to 3 results for LLM context.
func formatSearchResults(results []webscrape.PageInfo) string {
	formatted := ""
	for i, res := range results {
		formatted += fmt.Sprintf("- %s (%s)\n", res.Title, res.URL)
		if i == 2 {
			break
		}
	}
	return formatted
}
