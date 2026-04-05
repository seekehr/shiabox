from pathlib import Path

from config.constants import CHAT_PROMPT_PATH
from config.constants import CHUNKER_PROMPT_DIR

def get_chat_prompt() -> str:
    """Load the system chat prompt from disk.

    Returns:
        str: Full prompt file contents.

    Raises:
        FileNotFoundError: If CHAT_PROMPT_PATH is missing.
    """
    file_path = Path(CHAT_PROMPT_PATH)
    if not file_path.is_file():
        raise FileNotFoundError(f"Chat prompt file not found: {CHAT_PROMPT_PATH}")
    with open(file_path, encoding="utf-8") as file:
        return file.read()

def get_chunking_prompt(book_name: str) -> str:
    """Load the chunking prompt from book name.

    Args:
        book_name (str): The name of the book.

    Returns:
        str: The chunking prompt.
    """
    file_path = Path(CHUNKER_PROMPT_DIR / f"{book_name.lower()}.txt")
    if not file_path.is_file():
        raise FileNotFoundError(f"Chunking prompt file not found: {file_path}")
    with open(file_path, encoding="utf-8") as file:
        return file.read()
