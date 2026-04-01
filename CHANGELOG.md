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

### v1.3:
- [x] Set `stream: true` for Mistral to allow live streaming of tokens.
- [x] Introduced live streaming in cli aswell (idk why im calling it livestreaming xd).
- [x] Introduced SSE in API `/ai/request` allow live streaming of tokens over HTTP.

### v1.4:
- [x] 80-120X FASTER! 
- [x] Introduced the Groq API (using model `meta-llama/llama-4-scout-17b-16e-instruct`). 
- [x] Introduced `ParseStreamedSSE` and commented out the older model for now, to allow live-streaming of tokens.
- [x] Changed the `chan string` to `<-chan AIResponse` in parser.go to allow more information to be processed (and also made the channel <- read-only).
- [x] 30 responses per minute now only, but better than the 60+ seconds that requests used to take.
- [x] **Reason I switched from Mistral:** Only allowed 1 request per 60-120 seconds, as I could only run 1 instance on my PC which is designed to handle one thread only.

### v1.5:
- [x] Implement our changes on the backend server.
- [x] Make sure `controller` handles the new data format (`AIResponse` structure instead of a `string`) properly on the client side.
- [x] Make sure the actual page handles the new `controller` output properly. In the future, we'll also handle finish reasons such as `length`.
- [x] Update README.md to be more accurate.
- [x] Update INSTALLING.md to include setting up the `.env`.
