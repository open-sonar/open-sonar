package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// customFlusher implements the http.Flusher interface for testing
type customFlusher struct {
	flushed int
}

func (f *customFlusher) Flush() {
	f.flushed++
}

// customResponseWriter is used to capture the streaming output
type customResponseWriter struct {
	httptest.ResponseRecorder
	flusher *customFlusher
}

func (c *customResponseWriter) Flush() {
	c.flusher.Flush()
}

func newCustomResponseWriter() *customResponseWriter {
	return &customResponseWriter{
		ResponseRecorder: *httptest.NewRecorder(),
		flusher:          &customFlusher{},
	}
}

func (c *customResponseWriter) Header() http.Header {
	return c.ResponseRecorder.Header()
}

func (c *customResponseWriter) Write(data []byte) (int, error) {
	return c.ResponseRecorder.Write(data)
}

func (c *customResponseWriter) WriteHeader(statusCode int) {
	c.ResponseRecorder.WriteHeader(statusCode)
}

// Helper function to extract response chunks from streamed output
func extractChunks(body string) []map[string]interface{} {
	// Split by data: prefix and double newline
	lines := strings.Split(body, "data: ")

	var chunks []map[string]interface{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "[DONE]" {
			continue
		}

		// Parse JSON
		var chunk map[string]interface{}
		err := json.Unmarshal([]byte(line), &chunk)
		if err != nil {
			// Skip invalid JSON
			continue
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

func TestStreamingResponse(t *testing.T) {
	// Create custom response writer
	w := newCustomResponseWriter()

	// Create streaming response
	s, err := NewStreamingResponse(w, "test-model", "test-id", []string{"url1", "url2"})
	if err != nil {
		t.Fatalf("Error creating streaming response: %v", err)
	}

	// Verify headers
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type: text/event-stream, got: %s", contentType)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control: no-cache, got: %s", cacheControl)
	}

	// Send first chunk
	err = s.SendChunk("Hello", 0, true, false)
	if err != nil {
		t.Fatalf("Error sending first chunk: %v", err)
	}

	// Send middle chunk
	err = s.SendChunk(" world", 0, false, false)
	if err != nil {
		t.Fatalf("Error sending middle chunk: %v", err)
	}

	// Send final chunk
	err = s.SendChunk("!", 0, false, true)
	if err != nil {
		t.Fatalf("Error sending final chunk: %v", err)
	}

	// Send final message
	err = s.SendFinal()
	if err != nil {
		t.Fatalf("Error sending final message: %v", err)
	}

	// Check flush count (should be at least 4, one per message)
	if w.flusher.flushed < 4 {
		t.Errorf("Expected at least 4 flushes, got %d", w.flusher.flushed)
	}

	// Extract chunks
	body := w.Body.String()
	chunks := extractChunks(body)

	// Verify we got 3 chunks
	if len(chunks) != 3 {
		t.Fatalf("Expected 3 chunks, got %d. Body: %s", len(chunks), body)
	}

	// Verify first chunk has assistant role
	if choices, ok := chunks[0]["choices"].([]interface{}); ok {
		if len(choices) > 0 {
			choice := choices[0].(map[string]interface{})
			delta := choice["delta"].(map[string]interface{})
			if role, exists := delta["role"]; !exists || role != "assistant" {
				t.Errorf("Expected first chunk to have assistant role, got: %v", delta)
			}
			if content, exists := delta["content"]; !exists || content != "Hello" {
				t.Errorf("Expected first chunk to have content 'Hello', got: %v", content)
			}
		}
	} else {
		t.Error("Could not extract choices from first chunk")
	}

	// Verify last chunk has citations
	if citations, exists := chunks[2]["citations"].([]interface{}); !exists || len(citations) != 2 {
		t.Errorf("Expected last chunk to have 2 citations, got: %v", chunks[2]["citations"])
	}

	// Verify last chunk has stop finish_reason
	if choices, ok := chunks[2]["choices"].([]interface{}); ok {
		if len(choices) > 0 {
			choice := choices[0].(map[string]interface{})
			if reason, exists := choice["finish_reason"]; !exists || reason != "stop" {
				t.Errorf("Expected finish_reason to be 'stop', got: %v", reason)
			}
		}
	}

	// Verify final DONE message
	if !strings.Contains(body, "data: [DONE]") {
		t.Error("Expected [DONE] message in output")
	}
}

func TestStreamTokens(t *testing.T) {
	// Create custom response writer
	w := newCustomResponseWriter()

	// Create streaming response
	s, err := NewStreamingResponse(w, "test-model", "test-id", []string{"url1"})
	if err != nil {
		t.Fatalf("Error creating streaming response: %v", err)
	}

	// Stream content in chunks of 5 characters
	content := "This is a test of the streaming functionality."
	err = StreamTokens(s, content, 5)
	if err != nil {
		t.Fatalf("Error streaming tokens: %v", err)
	}

	// Extract chunks
	body := w.Body.String()
	chunks := extractChunks(body)

	// Expected number of chunks based on content length and chunk size
	expectedChunks := (len(content) + 4) / 5 // Ceiling division
	if len(chunks) != expectedChunks {
		t.Errorf("Expected %d chunks, got %d", expectedChunks, len(chunks))
	}

	// Reconstruct the content from chunks
	var reconstructed string
	for i, chunk := range chunks {
		if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
			choice := choices[0].(map[string]interface{})
			delta := choice["delta"].(map[string]interface{})
			if content, exists := delta["content"]; exists {
				reconstructed += content.(string)
			}

			// Check first chunk has role
			if i == 0 {
				if role, exists := delta["role"]; !exists || role != "assistant" {
					t.Errorf("Expected first chunk to have assistant role, got: %v", delta)
				}
			}

			// Check last chunk has finish reason
			if i == len(chunks)-1 {
				if reason, exists := choice["finish_reason"]; !exists || reason != "stop" {
					t.Errorf("Expected last chunk to have finish_reason 'stop', got: %v", choice)
				}
			}
		}
	}

	// Verify reconstructed content
	if reconstructed != content {
		t.Errorf("Reconstructed content mismatch. Expected: '%s', Got: '%s'", content, reconstructed)
	}

	// Verify DONE message is present
	if !strings.Contains(body, "data: [DONE]") {
		t.Error("Expected [DONE] message in output")
	}
}
