from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent

CHAT_PROMPT_PATH = BASE_DIR / "assets" / "prompt.txt"

PDF_BOOKS_DIR = BASE_DIR / "assets" / "pdf_books"
TXT_BOOKS_DIR = BASE_DIR / "assets" / "txt_books"
PARSED_BOOKS_DIR = BASE_DIR / "assets" / "parsed_books"
QDRANT_PATH = BASE_DIR / "assets" / "qdrant_data"

CHAT_MODEL = "meta-llama/llama-4-scout-17b-16e-instruct"

# EMBEDDING
EMBEDDING_MODEL = "qwen3-embedding:4b"
EMBEDDING_DIMENSIONS = 2560
BATCH_SIZE = 64
TEXT_SIZE = 640000

HADITHS_COLLECTION = "shiabox"