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

// HealthCheckHandler returns a simple JSON indicating service health.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"status":    "OK",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}

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

		// Create a search timer
		searchTimer := utils.NewTimer("Web search")

		// Extract search queries from the user message
		searchQueries := extractSearchQueries(userQuery)

		// Perform searches for each extracted query
		var allResults []webscrape.PageInfo
		for _, query := range searchQueries {
			results := webscrape.ScrapeWithOptions(query, searchOptions)
			allResults = append(allResults, results...)
		}

		// Score and rank results by relevance to the original query
		rankedResults := rankResultsByRelevance(allResults, userQuery)

		// Limit to most relevant results
		maxResults := 8
		if len(rankedResults) > maxResults {
			rankedResults = rankedResults[:maxResults]
		}

		searchTimer.Stop()

		if len(rankedResults) > 0 {
			// Create system prompt with search context
			systemPrompt := createSearchPromptTemplate(userQuery, rankedResults)
			messages = append(messages, fmt.Sprintf("system: %s", systemPrompt))

			// Extract citations
			citationURLs = citations.ExtractCitationURLs(rankedResults)

			// Log the extracted citations for debugging
			if len(citationURLs) > 0 {
				utils.Info(fmt.Sprintf("Extracted %d citations", len(citationURLs)))
			} else {
				utils.Warn("No citations were extracted from search results")
			}
		} else {
			// No results found, let LLM know
			messages = append(messages, "system: No relevant search results were found for this query. Please respond based on your training data.")
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

// extractSearchQueries breaks down a complex query into search-friendly queries
func extractSearchQueries(query string) []string {
	// For simple implementation, just return the original query
	// In a more advanced implementation, we could use NLP to extract key topics
	return []string{query}
}

// rankResultsByRelevance scores and ranks results by relevance to the query
func rankResultsByRelevance(results []webscrape.PageInfo, query string) []webscrape.PageInfo {
	// Simple relevance scoring based on keyword presence
	type scoredResult struct {
		result webscrape.PageInfo
		score  float64
	}

	// Create a list to hold scored results
	scoredResults := make([]scoredResult, 0, len(results))

	// Convert query to lowercase for case-insensitive matching
	lowercaseQuery := strings.ToLower(query)
	queryTerms := strings.Fields(lowercaseQuery)

	for _, result := range results {
		// Base score
		score := 1.0

		// Title match weight is higher
		lowercaseTitle := strings.ToLower(result.Title)
		for _, term := range queryTerms {
			if strings.Contains(lowercaseTitle, term) {
				score += 2.0
			}
		}

		// Content match
		lowercaseContent := strings.ToLower(result.Content)
		for _, term := range queryTerms {
			if strings.Contains(lowercaseContent, term) {
				score += 1.0
			}
		}

		// Domain credibility bonus (simple version)
		if strings.Contains(result.URL, ".edu") ||
			strings.Contains(result.URL, ".gov") ||
			strings.Contains(result.URL, "wikipedia.org") {
			score += 1.5
		}

		// Add to scored results
		scoredResults = append(scoredResults, scoredResult{
			result: result,
			score:  score,
		})
	}

	// Sort results by score (descending)
	utils.SortScored(scoredResults, func(i, j int) bool {
		return scoredResults[i].score > scoredResults[j].score
	})

	// Extract just the results
	rankedResults := make([]webscrape.PageInfo, 0, len(scoredResults))
	for _, scored := range scoredResults {
		rankedResults = append(rankedResults, scored.result)
	}

	return rankedResults
}

// createSearchPromptTemplate creates a system prompt incorporating search results
func createSearchPromptTemplate(query string, results []webscrape.PageInfo) string {
	promptTemplate := `I'll help answer the question based on the web search results provided below.

USER QUERY: %s

WEB SEARCH RESULTS:
%s

INSTRUCTIONS:
1. Use ONLY the information from these search results to answer the user's query
2. If the search results don't contain relevant information, admit that you don't have enough information
3. Provide a comprehensive answer that synthesizes information from multiple sources
4. Include specific facts and details from the sources
5. Cite sources using [1], [2], etc., corresponding to the search result numbers
6. DO NOT make up or include information not present in these search results
7. Maintain a helpful, informative, and accurate tone

Your answer should be well-structured, accurate, and directly address the user's query.`

	// Format the search results
	searchResultsText := ""
	for i, result := range results {
		searchResultsText += fmt.Sprintf("[%d] %s\nURL: %s\n", i+1, result.Title, result.URL)
		if result.Summary != "" {
			searchResultsText += fmt.Sprintf("Summary: %s\n\n", result.Summary)
		} else if result.Content != "" {
			// Use truncated content if summary not available
			content := result.Content
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			searchResultsText += fmt.Sprintf("Content: %s\n\n", content)
		}
	}

	return fmt.Sprintf(promptTemplate, query, searchResultsText)
}

// formatEnhancedSearchResults creates a better formatted context for the LLM
func formatEnhancedSearchResults(results []webscrape.PageInfo, query string) string {
	formattedResults := fmt.Sprintf("Web search results for query: \"%s\"\n\n", query)

	for i, result := range results {
		formattedResults += fmt.Sprintf("[%d] %s\n", i+1, result.Title)
		formattedResults += fmt.Sprintf("URL: %s\n", result.URL)

		if result.Summary != "" && len(result.Summary) > 0 {
			// Truncate summary if it's too long
			summary := result.Summary
			if len(summary) > 300 {
				summary = summary[:300] + "..."
			}
			formattedResults += fmt.Sprintf("Summary: %s\n", summary)
		} else if result.Content != "" {
			// Use content if summary isn't available
			content := result.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			formattedResults += fmt.Sprintf("Content: %s\n", content)
		}

		formattedResults += "\n"
	}

	return formattedResults
}

// ChatHandler handles legacy chat requests
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
