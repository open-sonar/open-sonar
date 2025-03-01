package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"open-sonar/internal/citations"
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

// ChatCompletionsHandler handles chat completion requests according to Perplexity API spec
func ChatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized: Missing or invalid Bearer token", http.StatusUnauthorized)
		return
	}

	// Parse request
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	var chatReq models.ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &chatReq); err != nil {
		utils.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(chatReq.Messages) == 0 {
		utils.Error("Missing 'messages' in request")
		http.Error(w, "'messages' field is required", http.StatusBadRequest)
		return
	}

	// Set default values if not provided
	if chatReq.Temperature == nil {
		defaultTemp := 0.2
		chatReq.Temperature = &defaultTemp
	}
	if chatReq.TopP == nil {
		defaultTopP := 0.9
		chatReq.TopP = &defaultTopP
	}
	if chatReq.FrequencyPenalty == nil && chatReq.PresencePenalty == nil {
		defaultFreqPenalty := 1.0
		chatReq.FrequencyPenalty = &defaultFreqPenalty
	}

	// Initialize LLM provider based on model
	modelName := chatReq.Model
	if modelName == "" {
		modelName = "sonar" // Default to sonar if not specified
	}

	provider, err := llm.NewLLMProvider(modelName)
	if err != nil {
		utils.Error(fmt.Sprintf("Invalid LLM provider: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract user query from last user message
	var userQuery string
	for i := len(chatReq.Messages) - 1; i >= 0; i-- {
		if chatReq.Messages[i].Role == "user" {
			userQuery = chatReq.Messages[i].Content
			break
		}
	}

	if userQuery == "" {
		utils.Error("No user message found in request")
		http.Error(w, "No user message found in request", http.StatusBadRequest)
		return
	}

	// Set up LLM options
	options := llm.LLMOptions{
		Temperature: *chatReq.Temperature,
		TopP:        *chatReq.TopP,
		TopK:        chatReq.TopK,
	}

	if chatReq.MaxTokens > 0 {
		options.MaxTokens = chatReq.MaxTokens
	} else {
		options.MaxTokens = 1024 // Default
	}

	if chatReq.PresencePenalty != nil {
		options.PresencePenalty = *chatReq.PresencePenalty
	}
	if chatReq.FrequencyPenalty != nil {
		options.FrequencyPenalty = *chatReq.FrequencyPenalty
	}

	// Format messages for LLM
	var messages []string
	for _, msg := range chatReq.Messages {
		messages = append(messages, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
	}

	// Search options
	searchOptions := webscrape.SearchOptions{
		MaxPages:            3, // Default
		MaxRetries:          2, // Default
		SearchDomainFilter:  chatReq.SearchDomainFilter,
		SearchRecencyFilter: chatReq.SearchRecencyFilter,
	}

	// Determine if this is a model that needs web search
	needsSearch := strings.HasPrefix(modelName, "sonar")

	var citationURLs []string

	// Perform web search for sonar models
	if needsSearch {
		utils.Info(fmt.Sprintf("Performing web search for query: %s", userQuery))
		results := webscrape.ScrapeWithOptions(userQuery, searchOptions)

		if len(results) > 0 {
			// Format search results for context
			searchContext := "Web Search Results:\n"
			for i, result := range results {
				if i >= 5 { // Limit to 5 results
					break
				}
				searchContext += fmt.Sprintf("[%d] %s - %s\n", i+1, result.Title, result.URL)
				if result.Summary != "" {
					searchContext += fmt.Sprintf("Summary: %s\n\n", result.Summary)
				}
			}

			// Add search context to the messages
			messages = append(messages, fmt.Sprintf("system: Use these search results to answer the user's query:\n%s", searchContext))

			// Extract citations
			citationURLs = citations.ExtractCitationURLs(results)
		}
	}

	// Generate response
	response, err := provider.GenerateResponseWithOptions(messages, options)
	if err != nil {
		utils.Error(fmt.Sprintf("LLM call failed: %v", err))
		http.Error(w, fmt.Sprintf("LLM processing error: %v", err), http.StatusInternalServerError)
		return
	}

	// Count tokens (simplified)
	promptText := strings.Join(messages, " ")
	promptTokens := utils.SimpleTokenCount(promptText)
	completionTokens := utils.SimpleTokenCount(response)

	// Prepare and send response
	completionResponse := models.ChatCompletionResponse{
		ID:        utils.GenerateUUID(),
		Model:     modelName,
		Object:    "chat.completion",
		Created:   time.Now().Unix(),
		Citations: citationURLs,
		Choices: []models.Choice{
			{
				Index:        0,
				FinishReason: "stop",
				Message: models.Message{
					Role:    "assistant",
					Content: response,
				},
			},
		},
		Usage: models.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completionResponse)
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
