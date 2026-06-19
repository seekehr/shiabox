import asyncio
import json
import os
import re

from openai import AsyncOpenAI

from config.constants import CHUNKER_MODEL, CHUNK_SIZE, MAX_CHUNKING_JOBS_IN_ONE_MIN, CHUNKER_TIMEOUT_SECS

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
    """Pack hadith segments greedily into chunks of at most ``CHUNK_SIZE`` characters.

    Segments that individually exceed ``CHUNK_SIZE`` are split on a stride. Empty segments
    are dropped.
    """
    if not text:
        return []

    raw: list[str] = []
    for s in _segments_at_hadith_starts(text):
        if not s:
            continue
        if len(s) <= CHUNK_SIZE:
            raw.append(s)
        else:
            for i in range(0, len(s), CHUNK_SIZE):
                raw.append(s[i : i + CHUNK_SIZE])

    # Greedily merge consecutive segments into chunks up to CHUNK_SIZE
    pieces: list[str] = []
    current = ""
    for seg in raw:
        if current and len(current) + len(seg) > CHUNK_SIZE:
            pieces.append(current)
            current = seg
        else:
            current += seg
    if current:
        pieces.append(current)
    return pieces


class ChunkerLLM:
    """A class to handle requests to the chunker (ollama) LLM."""

    def __init__(self):
        """Initialize the chunker LLM helper."""
        self.client = AsyncOpenAI(
            base_url="https://api.mistral.ai/v1",
            api_key=os.environ.get("MISTRAL_API_KEY"),
        )

    async def _chunk(self, system_prompt: str, text: str) -> str:
        strict_prefix = (
            "OUTPUT RULES (non-negotiable):\n"
            "- Output ONLY raw JSON objects, one per transmitted report, with NO other text.\n"
            "- Do NOT write explanations, headers, bullet points, markdown, or code fences.\n"
            "- If there are no transmitted reports in the text, output exactly: NO_RECORDS\n"
            "- Each JSON object must be on its own, separated by a blank line.\n\n"
        )
        response = await asyncio.wait_for(
            self.client.chat.completions.create(
                model=CHUNKER_MODEL,
                messages=[
                    {"role": "system", "content": strict_prefix + system_prompt},
                    {"role": "user", "content": text},
                ],
                max_completion_tokens=32000,
                temperature=0.2,
                top_p=1,
            ),
            timeout=CHUNKER_TIMEOUT_SECS,
        )
        return response.choices[0].message.content

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
                for attempt in range(1, 4):
                    try:
                        print(f"Chunking batch {index + 1}/{len(pieces)}..." + (f" (attempt {attempt})" if attempt > 1 else ""))
                        out = await self._chunk(system_prompt, piece)
                        preview = (out or "")[:200].replace("\n", " ")
                        print(f"  └─ batch {index + 1} output preview: {preview!r}")
                        return index, out
                    except asyncio.TimeoutError:
                        print(f"Batch {index + 1} timed out (attempt {attempt}/3), retrying...")
                raise RuntimeError(f"Batch {index + 1} failed after 3 attempts")

        indexed = await asyncio.gather(*(_run_one(i, p) for i, p in enumerate(pieces)))
        indexed.sort(key=lambda t: t[0])

        all_hadiths: list = []
        for _, part in indexed:
            if not part or "NO_RECORDS" in part:
                continue
            # extract every {...} block from free-form model output
            for match in re.finditer(r'\{[^{}]*\}', part, re.DOTALL):
                try:
                    obj = json.loads(match.group())
                    if isinstance(obj, dict):
                        all_hadiths.append(obj)
                except json.JSONDecodeError:
                    pass

        return json.dumps(all_hadiths)
