# Changelog
Tracking from v1.3 and onwards if God Wills.


### v2.0 (Python Revamp):
- [x] Rewrote the entire Go backend in Python.
- [x] Replaced Echo with **FastAPI** (`main.py`) for the HTTP server, with SSE streaming via `StreamingResponse`.
- [x] Replaced the Go gRPC Qdrant client with the **qdrant-client** Python SDK (HTTP mode, much simpler).
- [x] Replaced raw HTTP + SSE parsing for Groq with the **groq** Python SDK (`stream=True`).
- [x] Replaced `google.golang.org/genai` with the **google-genai** Python SDK for Gemini-based book chunking.
- [x] Replaced goroutines + WaitGroups with **asyncio** (`asyncio.gather`, `asyncio.Semaphore`) and `ThreadPoolExecutor` for concurrency.
- [x] Go structs with JSON tags replaced by **Pydantic** `BaseModel` classes.
- [x] All constants, prompt-building logic, text-cleaning heuristics, and the full setup pipeline faithfully ported.
- [x] Three clean entry points: `cli.py` (interactive CLI), `setup_db.py` (data pipeline), `main.py` (FastAPI server on :1323).
- [x] Now requires `GEMINI_API_KEY` in `.env` alongside `GROQ_API_KEY`.
