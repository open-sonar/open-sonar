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


## Architecture

```bash
/open-sonar                      # Root
├── README.md                    # Project overview, setup instructions, and usage examples
├── LICENSE                      # Open source license file (e.g., MIT)
├── go.mod                       # Go module file (defines module and dependencies)
├── go.sum                       # Dependency checksums
├── cmd                          # Contains executable command-line applications
│   └── server                   # Main API server entry point
│       └── main.go              # Bootstraps the application (config, routes, logging, etc.)
├── internal                     # Core business logic and internal modules (not exposed externally)
│   ├── api                      # HTTP handlers and route definitions
│   │   ├── handlers_test.go          # Functions that handle individual API endpoints
│   │   ├── handlers.go          
│   │   └── routes.go            # Maps endpoints (e.g., /chat/completions) to handlers
│   ├── llm                      # LLM integration layer (adapter logic)
│   │   ├── adapter.go           # Defines the generic LLM interface for adapter implementations
│   │   ├── ollama.go
│   │   ├── openai.go            # Concrete adapter implementation for OpenAI
│   │   └── anthropic.go         # Concrete adapter implementation for Anthropic
│   ├── search                   # Real-time web search functionality
│   │   ├── web-scrape           # Web scraping functionality
│   │   │   ├── scraper.go
│   │   │   ├── Makefile
│   │   │   └── webscrape_test.go
│   │   ├── search.go            # Core functions to perform web searches (e.g., via Colly/goquery)
│   │   └── filters.go           # Functions to apply customizable search filters
│   ├── citations                # Citation extraction and formatting module
│   │   └── extractor.go         # Parses search results to extract and format citations
│   ├── cache                    # Caching and rate limiting functionality
│   │   ├── redis.go             # Integration with Redis (or similar) for caching responses
│   │   └── limiter.go           # Implements middleware for API rate limiting
│   ├── config                   # Configuration management module
│   │   └── config.go            # Loads and validates configuration files (YAML/JSON)
│   ├── models                   # Data models and API response schemas
│   │   └── response.go          # Defines structures (e.g., ChatResponse) for consistent JSON output
│   └── utils                    # Utility functions (logging, error handling, etc.)
│       └── logger.go            # Sets up structured logging and helper functions
├── docker                       # Containerization files
│   └── Dockerfile               # Dockerfile to build the API server image
├── scripts                      # Automation and deployment scripts
│   └── deploy.sh                # Script for deploying the application
└── docs                         # Additional documentation and guides
    └── API_DOCS.md              # Detailed API documentation and usage examples
```
