from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent

CHAT_PROMPT_PATH = BASE_DIR / "assets" / "prompt.txt"

PDF_BOOKS_DIR = BASE_DIR / "assets" / "pdf_books"
TXT_BOOKS_DIR = BASE_DIR / "assets" / "txt_books"
PARSED_BOOKS_DIR = BASE_DIR / "assets" / "parsed_books"

CHAT_MODEL = "meta-llama/llama-4-scout-17b-16e-instruct"
EMBEDDING_MODEL = "nomic-embed-text"