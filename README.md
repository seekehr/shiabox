# shiabox

AI-powered (Groq) search engine for Shi'a ahadith, using a RAG architecture with Qdrant for the vector database and `qwen3-embedding:4b` for embeddings.

# How does it work?

RAG lets you ground an AI response in a selective corpus — here, **ahadith** scraped from [thaqalayn.net](https://thaqalayn.net) and indexed in Qdrant.

### Step 1: Scrape ahadith

The TypeScript scraper in [`scraper/`](scraper/) opens thaqalayn.net with Playwright to find a book (and its volumes), then fetches chapter and hadith pages over HTTP and writes structured JSON to `scraper/output/`.

Each hadith includes English text (`content`), Arabic (`metadata.arabic`), chapter/volume numbers, and a source URL. See [`scraper/README.md`](scraper/README.md) for install and usage.

### Step 2: Embed the ahadith into Qdrant

Raw JSON alone is too large to send to the chat model on every query. Instead, each hadith is embedded with Ollama (`qwen3-embedding:4b`) and stored in an on-disk Qdrant collection under `server/assets/qdrant_data`.

At query time we retrieve the top **10** nearest ahadith and let the chat model pick the **3** most relevant ones.

### Step 3: Search, rank, and stream a response

The user's prompt is embedded the same way, compared against the vector index, and the top matches are sent to Groq with the system prompt in [`server/assets/prompt.txt`](server/assets/prompt.txt). The reply is streamed back over HTTP (or the CLI).

# Installation Guide

See [INSTALLING.md](INSTALLING.md).

# Preview

![alt text](https://github.com/seekehr/shiabox/blob/main/server/assets/images/readme_preview_1.png "Example 1")
