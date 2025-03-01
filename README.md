# Open Sonar

A web-search enhanced AI assistant that combines the power of web search with local LLM inference to provide accurate, up-to-date information with source citations.

## Features

- 🔍 **Web Search Integration**: Augments LLM responses with real-time web search results
- 🧠 **Local Inference**: Uses Ollama for running LLMs locally
- 🌐 **OpenAI Compatible API**: Drop-in replacement for OpenAI's Chat Completions API
- 📑 **Citation Support**: Provides citations for information sourced from the web
- 🔧 **Customizable**: Support for different search providers and LLM backends

## Quick Start

### Prerequisites

1. Install [Go](https://golang.org/doc/install) (version 1.20 or later)
2. Install [Ollama](https://ollama.ai/):
   ```bash
   curl https://ollama.ai/install.sh | sh
   ```
3. Pull the deepseek-r1:1.5b model:
   ```bash
   ollama pull deepseek-r1:1.5b
   ```

### Running the Server

The easiest way to run the server is with our dev script:

```
# Clone the repository
git clone https://github.com/yourusername/open-sonar.git
cd open-sonar

# Install dependencies
make deps

# Start the development server
./scripts/dev_server.sh
```

This will start the server on port 8080 (default). You can customize settings in the .env file.

## Using the API
### Simple Chat API
```
curl -X POST "http://localhost:8080/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is quantum computing?",
    "needSearch": true
  }'
```

### OpenAI Compatible Chat Completions API
```
curl -X POST "http://localhost:8080/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "model": "sonar",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "What are recent developments in renewable energy?"}
    ],
    "temperature": 0.7,
    "search_domain_filter": [".edu", ".gov", ".org"]
  }'
```

## Development
### Running Tests
```
# Run unit tests (uses mocks)
make test

# Run E2E tests with mock LLM
make e2e-test

# Run E2E tests with real LLM (requires Ollama)
make e2e-test-prod

# Run integration tests specifically for Ollama
make integration-test
```

### Structure
/cmd/server - Server entry point
/internal - Internal packages
/api - API handlers and routing
/cache - Caching layer
/citations - Citation extraction and formatting
/llm - LLM provider interfaces
/models - Data structures and models
/search - Search abstraction
/webscrape - Web scraping implementations
/utils - Shared utilities

## Config
Copy example.env to .env and adjust the settings:
```
# Server configuration
PORT=8080
LOG_LEVEL=INFO  # DEBUG, INFO, WARN, ERROR

# Authentication
AUTH_TOKEN=your-auth-token-here

# LLM configuration
OLLAMA_MODEL=deepseek-r1:1.5b
OLLAMA_HOST=http://localhost:11434

# Alternative LLM providers (optional)
OPENAI_API_KEY=your-openai-api-key-here
OPENAI_MODEL=gpt-3.5-turbo
```

# Plan

## Goals

- Replicate all key features of Sonar:
  - Real-time web search and data retrieval
  - Citation extraction and attribution
  - Structured JSON output with optional chain-of-thought traces
  - Pluggable LLM integration (OpenAI, Anthropic, Ollama, etc.)
- Enable seamless switching between LLM providers via an adapter layer
- Provide high scalability, low latency, and developer-friendly API endpoints

## Pipeline 

- **API Request Handling:**
  - Exposes RESTful endpoints (e.g., /test, /chat, /chat/completions).
  - Validate and parse incoming requests (ChatRequest with fields like query, need_search, pages, retries, and provider).

- **Decision Filter:**
  - Determine need for real-time web search vs. direct LLM query
  - If required, performs a real-time web search to gather additional context before calling an LLM.

- **Search & Citation Extraction Module:**
  - Uses a DuckDuckGo-based scraper built with GoQuery and go-readability.
  - Randomizes User-Agent strings to reduce blocking.
  - Extracts search result links and summarizes page content for citation extraction.

- **LLM Adapter Layer:**
  - Provides a unified interface (LLMProvider) for multiple LLM integrations.
  - Implements adapters for OpenAI (via Langchaingo) and Anthropic, with placeholders for Ollama.
  - Exposes a direct LLM API endpoint (/chat/completions) for queries that don’t require a web search.

- **Response Aggregation & Formatting:**
  - Combine LLM output with extra context (search results, citations)
  - Format final answer as structured JSON (optional chain-of-thought)
  - Incremental/streaming responses via websockets or HTTP/2

- **Caching & Rate Limiting Module:**
  - Cache frequent queries/responses (e.g., Redis, in-memory)
  - Enforce rate limiting to protect providers

- **Return Response:**
  - Send aggregated, JSON-formatted answer back to client

- **Containerization & Deployment:**
  - Dockerfiles and CI/CD pipelines.
  - Plan for horizontal scaling?

