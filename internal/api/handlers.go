package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"open-sonar/internal/models"
	"open-sonar/internal/search/webscrape"
	"open-sonar/internal/utils"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"message": "open-sonar server is running.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	// Parse JSON
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

	// Set defaults if missing
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
		utils.Info("Performing web search before LLM call")

		results := webscrape.Scrape(chatReq.Query, pages, retries)

		resp := map[string]interface{}{
			"decision":   "search + LLM call",
			"pages_used": len(results),
			"message":    "Web scraped. LLM was called (stub).",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	utils.Info("No search needed, direct LLM call.")

	resp := map[string]interface{}{
		"decision": "direct LLM call",
		"message":  "Called the LLM directly with the provided query (stub).",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
