# Open Sonar

A web-search enhanced AI assistant that combines the power of web search with local LLM inference to provide accurate, up-to-date information with source citations.

Ported over as a Python and JS package.

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

We currently provide a Dockerfile and a Go Package.
Check package_test.sh and docker_test.sh for information on how to run the server.

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
