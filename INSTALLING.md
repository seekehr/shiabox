# Installation Guide
The project is split into 2 main parts; the `client/` and the Python backend. The backend uses Python 3.14+ with FastAPI, while the client uses React (Typescript) + Tailwind (and utilises Vite as a build tool). The backend contains: 
- **(A)** code that is relevant to the AI, which uses `Qdrant` (via the `qdrant-client` Python SDK) and `HTTP` to communicate with Groq and the embedding model.
- **(B)** code to help `client/` interact with the server (including the AI), which uses `FastAPI`.

In other words, you **DO NOT** need to install `client/` if you don't want to use the website. You can just use `python cli.py` to talk to the AI directly in your terminal.

## Requirements
#### Backend
- [Python](https://www.python.org/downloads/): We use version `3.14`.
- [Qdrant](https://github.com/qdrant/qdrant/releases/tag/v1.14.1): Our vector database. You must keep it running (e.g by opening another terminal/command prompt and typing qdrant inside).
- [Ollama](https://ollama.com/download/windows): LLM manager that allows us to download our other LLMs like Mistral easily.
- **Embedding model:** Simply run `ollama pull mxbai-embed-large` after installing ollama. We use the latest version.

#### `client/`
- [NodeJS](https://nodejs.org/en/download)
- After installing node and `cloning` this repo, simply navigate (via cd command) to the `client/` folder and run `npm install`. After it is completed, run `npm run dev` and copy paste
the URL provided in the terminal/command prompt inside your browser.

## Setting Up

**First**, make sure you complete all the requirements mentioned. After that, **make sure to clone the github repo using `git clone` first!** 
Either that, or you can Download as ZIP on github, then extract the source code in a folder.

Install the Python dependencies:
```
pip install uv
uv sync
```

Create a `.env` file in the project root, and it should contain your [Groq API key](https://groq.com/) and [Gemini API key](https://ai.google.dev/) (Gemini is only needed for chunking new books). Like: 
```
GROQ_API_KEY="api key here"
GEMINI_API_KEY="api key here"
```

Then, run `python setup_db.py`. When it asks you for input, first enter `0`, and after it is finished, use `python setup_db.py`
again, and this time enter `2`, then again with `3` (not the most intuitive design I know lol).

After that, if you intend to just talk to the AI directly in the terminal/command prompt, run `python cli.py`. If you're also running your frontend in the `client/` folder (using
`npm run dev` like it is mentioned in the requirements), then run `python main.py` for the frontend to be able to communicate with the backend.
