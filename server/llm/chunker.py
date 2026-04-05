import asyncio
import json
import re

from google import genai
from google.genai import types

from config.constants import CHUNKER_MODEL, CHUNK_SIZE, MAX_CHUNKING_JOBS_IN_ONE_MIN

# Numbered hadith lines (e.g. "1. Narrator…", "12. …"); split here instead of mid-hadith.
_HADITH_LINE_START = re.compile(r"(?m)^\s*\d+\.\s")


def _segments_at_hadith_starts(text: str) -> list[str]:
    """Split *text* into segments that each start at a numbered hadith line (or the preamble)."""
    matches = list(_HADITH_LINE_START.finditer(text))
    if not matches:
        return [text]

    segments: list[str] = []
    first = matches[0].start()
    if first > 0:
        segments.append(text[:first])
    for i, m in enumerate(matches):
        start = m.start()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(text)
        segments.append(text[start:end])
    return segments


def hadith_aware_pieces(text: str) -> list[str]:
    """Split *text* at numbered hadith lines; each piece is at most ``CHUNK_SIZE`` characters.

    Segments from boundaries are kept whole when they fit. Longer segments (or whole books with
    no hadith lines) are cut on a ``CHUNK_SIZE`` stride. Empty segments are dropped.
    """
    if not text:
        return []
    pieces: list[str] = []
    for s in _segments_at_hadith_starts(text):
        if not s:
            continue
        if len(s) <= CHUNK_SIZE:
            pieces.append(s)
        else:
            for i in range(0, len(s), CHUNK_SIZE):
                pieces.append(s[i : i + CHUNK_SIZE])
    return pieces


class ChunkerLLM:
    """A class to handle requests to the chunker (gemini) LLM.
    """
    def __init__(self):
        """Initialize the chunker LLM helper.

        Args:
            client (genai.Client): The gemini API client.
        """
        # The client gets the API key from the environment variable `GEMINI_API_KEY`.
        self.client = genai.Client()

    async def _chunk(self, system_prompt: str, text: str) -> str:
        """Chunk the text using the gemini LLM.

        Args:
            system_prompt (str): The system prompt to use.
            text (str): The text to chunk.

        Returns:
            str: The chunked text.
        """
        def _sync_generate() -> str:
            response = self.client.models.generate_content(
                model=CHUNKER_MODEL,
                config=types.GenerateContentConfig(
                    system_instruction=system_prompt,
                ),
                contents=text,
            )
            return response.text

        return await asyncio.to_thread(_sync_generate)

    async def start_chunking(self, text: str, system_prompt: str) -> str:
        """Process *text* in hadith-boundary pieces (capped by ``CHUNK_SIZE``) with bounded concurrency.

        Chunk results are ordered by piece index (0, 1, 2, …), concatenated, then JSON-encoded.

        Args:
            text (str): Full source text.
            system_prompt (str): System instruction passed to each ``_chunk`` call.

        Returns:
            str: JSON string of the concatenated model outputs (see ``json.dumps``).
        """
        if not text:
            return json.dumps("")

        pieces = hadith_aware_pieces(text)
        sem = asyncio.Semaphore(MAX_CHUNKING_JOBS_IN_ONE_MIN)

        async def _run_one(index: int, piece: str) -> tuple[int, str]:
            async with sem:
                out = await self._chunk(system_prompt, piece)
            return index, out

        indexed = await asyncio.gather(*(_run_one(i, p) for i, p in enumerate(pieces)))
        indexed.sort(key=lambda t: t[0])
        combined = "".join(part for _, part in indexed)
        return json.dumps(combined)
