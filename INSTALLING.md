# Installing shiabox

## Prerequisites

| Need | Link / notes |
|------|----------------|
| Python **3.14+** | [python.org/downloads](https://www.python.org/downloads/) |
| **uv** (env + deps) | [docs.astral.sh/uv/getting-started/installation](https://docs.astral.sh/uv/getting-started/installation/) |
| **Ollama** (embeddings) | [ollama.com/download](https://ollama.com/download) |
| **Groq** API key (chat) | [console.groq.com/keys](https://console.groq.com/keys) |

**Optional:** [Git](https://git-scm.com/downloads) · [Node.js](https://nodejs.org/) (web UI) · [Google AI Studio / Gemini](https://ai.google.dev/) (only if you use your own book-chunking flow; not wired in `server/` today)

## Backend (CLI)

Vectors are stored with **embedded Qdrant** (`server/assets/qdrant_data`); you do **not** run a separate Qdrant server.

1. Clone the repo ([GitHub: create a repo / clone](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository)).
2. Pull the embedding model (must match `EMBEDDING_MODEL` in [`server/config/constants.py`](server/config/constants.py), currently `qwen3-embedding:4b`):

   ```bash
   ollama pull qwen3-embedding:4b
   ```

3. Install Python dependencies from `server/`:

   ```bash
   cd server
   uv sync
   ```

4. Add `server/.env`:

   ```env
   GROQ_API_KEY=your_key_here
   ```

5. Put hadith JSON under `server/assets/parsed_books/` (`.json` per book; the repo may already include samples).
6. Build the vector index:

   ```bash
   uv run python setup.py
   ```

7. Run the chat loop:

   ```bash
   uv run python main.py
   ```

   Enter **`1`** when prompted to chat.

## Optional: web client (`client/`)

Stack: [Vite](https://vite.dev/) + React + TypeScript + Tailwind.

```bash
cd client
npm install
npm run dev
```

Open the URL Vite prints. If your API is not on `http://localhost:1323`, change `BACKEND_URL` in [`client/src/controller/controllers.ts`](client/src/controller/controllers.ts).
