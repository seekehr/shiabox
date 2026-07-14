# shiabox server

Python FastAPI backend: embeds scraped hadith JSON into on-disk Qdrant and serves RAG chat over Groq.

For the full pipeline overview, see the [root README](../README.md). For install steps, see [INSTALLING.md](../INSTALLING.md).

## Role in the stack

1. Load hadith JSON from `assets/parsed_books/` (produced by [`scraper/`](../scraper/)).
2. Embed with Ollama (`qwen3-embedding:4b`) into `assets/qdrant_data`.
3. On query: vector search → top ahadith → Groq ranking/response (system prompt in `assets/prompt.txt`).

## Entry points

| Command | Purpose |
|---------|---------|
| `uv run python setup.py` | Build or incrementally update the Qdrant index |
| `uv run python main.py` | `0` = FastAPI on `:1323`, `1` = CLI chat |

## Config

- `GROQ_API_KEY` in `.env`
- Embedding model and paths in [`config/constants.py`](config/constants.py)
