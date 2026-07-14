# Installing shiabox

## Prerequisites

| Need | Link / notes |
|------|----------------|
| **Node.js** | [nodejs.org](https://nodejs.org/) — required for the scraper and optional web UI |
| Python **3.14+** | [python.org/downloads](https://www.python.org/downloads/) |
| **uv** (env + deps) | [docs.astral.sh/uv/getting-started/installation](https://docs.astral.sh/uv/getting-started/installation/) |
| **Ollama** (embeddings) | [ollama.com/download](https://ollama.com/download) |
| **Groq** API key (chat) | [console.groq.com/keys](https://console.groq.com/keys) |

**Optional:** [Git](https://git-scm.com/downloads)

## 1. Scraper (`scraper/`)

Scrapes ahadith from [thaqalayn.net](https://thaqalayn.net) into JSON.

```bash
cd scraper
npm install
npx playwright install chromium
npm start
```

When prompted, enter a book name (substring match, e.g. `al kafi`). Output is written to `scraper/output/<slug>.json`.

Copy (or transform) that JSON into `server/assets/parsed_books/` before building the vector index. Details: [`scraper/README.md`](scraper/README.md).

## 2. Backend (`server/`)

Vectors are stored with **embedded Qdrant** (`server/assets/qdrant_data`); you do **not** run a separate Qdrant server.

1. Clone the repo ([GitHub: cloning a repository](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository)).
2. Pull the embedding model (must match `EMBEDDING_MODEL` in [`server/config/constants.py`](server/config/constants.py), currently `qwen3-embedding:4b`):

   ```bash
   ollama pull qwen3-embedding:4b
   ```

3. Install Python dependencies:

   ```bash
   cd server
   uv sync
   ```

4. Add `server/.env`:

   ```env
   GROQ_API_KEY=your_key_here
   ```

5. Ensure hadith JSON is under `server/assets/parsed_books/` (from the scraper step above).
6. Build the vector index:

   ```bash
   uv run python setup.py
   ```

   Enter `0` to rebuild from scratch, or `1` to only embed new books.

7. Run the app:

   ```bash
   uv run python main.py
   ```

   Enter **`0`** for the FastAPI server on `:1323`, or **`1`** for the interactive CLI chat.

## 3. Optional: web client (`client/`)

Stack: [Vite](https://vite.dev/) + React + TypeScript + Tailwind.

```bash
cd client
npm install
npm run dev
```

Open the URL Vite prints. If your API is not on `http://localhost:1323`, change `BACKEND_URL` in [`client/src/controller/controllers.ts`](client/src/controller/controllers.ts).
