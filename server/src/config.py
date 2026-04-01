from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent

# Asset directories
PDF_BOOKS_DIR = BASE_DIR / "assets" / "pdf_books"
UNPARSED_BOOKS_DIR = BASE_DIR / "assets" / "unparsed_books"
PARSED_BOOKS_DIR = BASE_DIR / "assets" / "parsed_books"
EMBEDDINGS_DIR = BASE_DIR / "assets" / "embeddings"

# Qdrant
COLLECTION_NAME = "shiabox"
VECTOR_SIZE = 1024
MAX_RESULTS_LIMIT = 5

# Embedding (Ollama)
OLLAMA_EMBED_URL = "http://localhost:11434/api/embed"
EMBED_MODEL = "mxbai-embed-large"
EMBED_BATCH_SIZE = 30
EMBED_WORKER_COUNT = 10

# Vector DB upsert
MAX_VECTORS_PER_BATCH = 50
MAX_VECTOR_WORKERS = 10

# Setup / chunking
RATELIMIT_SLEEP_SECONDS = 65
MAX_REQUESTS_PER_MIN = 15
CHUNK_SIZE_CHARACTERS = 50_000
OVERLAP_CHARACTERS = 2_500

# LLM models
CHUNKER_MODEL = "gemini-2.5-flash-lite-preview-06-17"
CHAT_MODEL = "meta-llama/llama-4-scout-17b-16e-instruct"

# Prompt files
CHAT_PROMPT_FILE = BASE_DIR / "assets" / "prompt.txt"
CHUNKER_PROMPT_FILE = BASE_DIR / "assets" / "books_parser_prompt.txt"

# Server
FRONTEND_URL = "http://localhost:5173"
