# Open Sonar

An open-source implementation of the Perplexity Sonar API that can be run on-premises.

## Features

- Chat completion API compatible with Perplexity's API
- Support for web searches to enhance responses
- Customizable parameters (temperature, top_p, etc.)
- Support for multiple LLM backends (OpenAI, Anthropic)
- Domain filtering and recency filtering for search results

## Quick Start

1. Clone the repository
2. Copy `example.env` to `.env` and fill in your API keys
3. Build and run the server:

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

