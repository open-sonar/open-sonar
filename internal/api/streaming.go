package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"open-sonar/internal/models"
	"open-sonar/internal/utils"
)

// StreamingResponse handles streaming responses for chat completions
type StreamingResponse struct {
	w         http.ResponseWriter
	flusher   http.Flusher
	requestID string
	model     string
	created   int64
	citations []string
}

// NewStreamingResponse creates a new streaming response handler
func NewStreamingResponse(w http.ResponseWriter, model string, requestID string, citations []string) (*StreamingResponse, error) {
	// Set appropriate headers for streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Get flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}

	return &StreamingResponse{
		w:         w,
		flusher:   flusher,
		requestID: requestID,
		model:     model,
		created:   time.Now().Unix(),
		citations: citations,
	}, nil
}

// SendChunk sends a content chunk in the stream
func (s *StreamingResponse) SendChunk(content string, index int, isFirst, isLast bool) error {
	delta := models.Delta{
		Content: content,
	}

	// Add role only to first message
	if isFirst {
		delta.Role = "assistant"
	}

	choice := models.Choice{
		Index: index,
		Delta: &delta,
	}

	// Set finish reason on last message
	if isLast {
		choice.FinishReason = "stop"
	}

	// Create response object
	response := models.ChatCompletionResponse{
		ID:      s.requestID,
		Model:   s.model,
		Object:  "chat.completion.chunk",
		Created: s.created,
		Choices: []models.Choice{choice},
	}

	// Add citations to the final chunk only
	if isLast && len(s.citations) > 0 {
		response.Citations = s.citations
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	// Write the data prefixed with "data: " and a double newline
	_, err = fmt.Fprintf(s.w, "data: %s\n\n", jsonData)
	if err != nil {
		return err
	}

	// Flush the response to ensure it's sent immediately
	s.flusher.Flush()
	return nil
}

// SendFinal sends the [DONE] message at the end of the stream
func (s *StreamingResponse) SendFinal() error {
	_, err := fmt.Fprintf(s.w, "data: [DONE]\n\n")
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// StreamTokens chunked output with artificial pauses for realistic streaming
func StreamTokens(streamer *StreamingResponse, content string, chunkSize int) error {
	// Split content into chunks
	var chunks []string
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		chunks = append(chunks, content[i:end])
	}

	// Stream each chunk with a small delay
	for i, chunk := range chunks {
		isFirst := i == 0
		isLast := i == len(chunks)-1

		err := streamer.SendChunk(chunk, 0, isFirst, isLast)
		if err != nil {
			utils.Error(fmt.Sprintf("Error streaming chunk: %v", err))
			return err
		}

		// Small delay between chunks for realistic streaming
		if !isLast {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Send the final [DONE] message
	return streamer.SendFinal()
}
