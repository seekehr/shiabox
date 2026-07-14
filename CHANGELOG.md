# Changelog
Tracking from v1.3 and onwards if God Wills.

### v2.1 (Thaqalayn scraper):
- [x] Replaced the PDF → `pdftotext` → Gemini/Mistral LLM chunking pipeline with a direct HTML scraper for [thaqalayn.net](https://thaqalayn.net). Books are now manually maintained as pre-parsed JSON in `server/assets/parsed_books/`.
- [x] Removed `server/llm/chunker.py` and all related constants (`CHUNKER_PROMPT_DIR`, `PDF_BOOKS_DIR`, `TXT_BOOKS_DIR`, `CHUNKED_BOOKS_DIR`, `CHUNKER_MODEL`, etc.). `GEMINI_API_KEY` no longer required.
- [x] Scraper split into four focused files: `types.ts` (interfaces), `utils.ts` (`prompt`, `normalize`, `slugify`, `fetchHtml` with retry, `pool`), `parser.ts` (HTML parsing), `browser.ts` (Playwright).
- [x] Book lookup: Playwright opens the homepage, diacritic-insensitive fuzzy match on book cards.
- [x] Volume detection: Playwright clicks the volume combobox (client-side SPA, not in SSR HTML) and captures each volume's URL; single-volume books get `volume_number: 1`.
- [x] Chapters and ahadith fetched with plain `fetch()` at 12-way concurrency; 3-attempt exponential backoff for HTTP 500 / network errors.
- [x] Chapter names parsed from `div.text-2xl.mt-2` (actual chapter title, not the book-level `h1`).
- [x] Arabic/English parsed via `p[dir='rtl']` and `p.nassim:not([dir='rtl'])` selectors directly, replacing the heuristic paragraph scan.
- [x] Output schema: `book_name`, `volume_number`, `chapter_number`, `hadith_number`, `content` (English) at top level; `book_number`, `chapter_name`, `arabic`, `url` nested under `metadata`.

### v2.0 (Python Revamp):
- [x] Rewrote the entire Go backend in Python.
- [x] Replaced Echo with **FastAPI** (`main.py`) for the HTTP server, with SSE streaming via `StreamingResponse`.
- [x] Replaced the Go gRPC Qdrant client with the **qdrant-client** Python SDK (HTTP mode, much simpler).
- [x] Replaced raw HTTP + SSE parsing for Groq with the **groq** Python SDK (`stream=True`).
- [x] Replaced goroutines + WaitGroups with **asyncio** (`asyncio.gather`, `asyncio.Semaphore`) and `ThreadPoolExecutor` for concurrency.
- [x] Go structs with JSON tags replaced by **Pydantic** `BaseModel` classes.
- [x] Constants, prompt-building logic, and the setup pipeline ported to Python.
- [x] Entry points: `setup.py` (vector index) and `main.py` (FastAPI server / CLI on :1323).
- [x] Requires `GROQ_API_KEY` in `.env`.
