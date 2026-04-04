from pathlib import Path

from config.constants import CHAT_PROMPT_PATH


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