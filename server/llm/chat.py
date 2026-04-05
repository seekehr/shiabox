import os
from groq import Groq, Stream
from groq.types.chat import ChatCompletionSystemMessageParam, ChatCompletionUserMessageParam, ChatCompletion, \
    ChatCompletionChunk
from config.file_configs import get_chat_prompt
from config.constants import CHAT_MODEL


def _is_missing_source_value(value: object) -> bool:
    if value is None:
        return True
    s = str(value).strip()
    return s == "" or s.upper() == "N/A"


def _format_hadith_source(r: dict) -> str:
    """Build a single Source: line from whatever metadata the hit provides.

    Payloads may be minimal (e.g. chapter + hadith number + content only) or include
    page, sermon, or a preformatted ``source`` string.
    """
    raw = r.get("source")
    if isinstance(raw, str) and not _is_missing_source_value(raw):
        return f"Source: {raw.strip()}"

    parts: list[str] = []
    book = r.get("book")
    if not _is_missing_source_value(book):
        bs = str(book).strip()
        parts.append(f"Book {bs}" if bs.isdigit() else bs)
    chapter = r.get("chapter")
    if not _is_missing_source_value(chapter):
        parts.append(f"Chapter {chapter}")
    hadith_no = r.get("hadith_number")
    if not _is_missing_source_value(hadith_no):
        parts.append(f"Hadith #{hadith_no}")
    page = r.get("page")
    if not _is_missing_source_value(page):
        parts.append(f"Page {page}")
    sermon = r.get("sermon")
    if not _is_missing_source_value(sermon):
        parts.append(f"Sermon {sermon}")

    if not parts:
        return "Source: (see content)"
    return "Source: " + ", ".join(parts)


class ChatLLM:
    """Handle Groq chat completions."""

    def __init__(self, client: Groq):
        """Initialize the chat LLM helper.

        Args:
            client (Groq): The Groq API client.
        """
        self._client = client
        self._chat_prompt = get_chat_prompt()

    def prompt(self, user_prompt: str, search_results: list[dict] | None = None) -> ChatCompletion | Stream[ChatCompletionChunk]:
        """Request a chat completion from the Groq LLM.

        Args:
            user_prompt (str): The user message.
            search_results (list[dict] | None): Optional retrieval hits. If set, system content
                combines the chat prompt with formatted results; otherwise only the chat prompt is used.

        Returns:
            ChatCompletion | Stream[ChatCompletionChunk]: The completion response (streamed chunks when stream=True).
        """
        if search_results:
            system_content = self._build_rag_prompt(user_prompt, search_results)
        else:
            system_content = self._chat_prompt

        system_msg: ChatCompletionSystemMessageParam = {"role": "system", "content": system_content}
        user_msg: ChatCompletionUserMessageParam = {"role": "user", "content": user_prompt}
        return self._client.chat.completions.create(
            messages=[system_msg, user_msg],
            model=CHAT_MODEL,
            temperature=0.5,
            stream=True,
        )

    def _build_rag_prompt(self, question: str, results: list[dict]) -> str:
        """Build system prompt text that appends formatted search results for RAG.

        Args:
            question (str): The user question; substituted into the chat prompt template.
            results (list[dict]): Hit dicts with keys score, book, chapter, hadith_number, content.

        Returns:
            str: Chat prompt with the question substituted into the template, plus formatted hadith blocks.
        """
        candidates = []
        for i, r in enumerate(results, 1):
            score = float(r.get("score", 0.0))
            content = r.get("content") or ""
            candidates.append(
                f"Hadith {i} (Score: {score:.4f})\n"
                f"{_format_hadith_source(r)}\n"
                f"{content}"
            )
        candidates_text = "\n\n".join(candidates)
        return self._chat_prompt.replace("{InputText}", question) + "\n" + candidates_text


