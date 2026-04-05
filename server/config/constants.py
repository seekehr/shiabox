from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent

CHAT_PROMPT_PATH = BASE_DIR / "assets" / "prompt.txt"
CHUNKER_PROMPT_DIR = BASE_DIR / "assets" / "chunking_prompts"

PDF_BOOKS_DIR = BASE_DIR / "assets" / "pdf_books"
TXT_BOOKS_DIR = BASE_DIR / "assets" / "txt_books"
PARSED_BOOKS_DIR = BASE_DIR / "assets" / "parsed_books"
QDRANT_PATH = BASE_DIR / "assets" / "qdrant_data"

CHAT_MODEL = "meta-llama/llama-4-scout-17b-16e-instruct"

# CHUNKING
CHUNKER_MODEL = "gemini-3-flash-preview"
CHUNK_SIZE = 10000
MAX_CHUNKING_JOBS_IN_ONE_MIN = 10

# EMBEDDING
EMBEDDING_MODEL = "qwen3-embedding:4b"
EMBEDDING_DIMENSIONS = 2560
BATCH_SIZE = 64
TEXT_SIZE = 640000

# API
MAX_REQUESTS_PER_MINUTE = 5

HADITHS_COLLECTION = "shiabox"
