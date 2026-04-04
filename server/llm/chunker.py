import asyncio

from google import genai
from google.genai import types

from config.constants import CHUNKER_MODEL, CHUNK_SIZE, MAX_CHUNKING_JOBS_IN_ONE_MIN

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
        response = self.client.models.generate_content(
            model=CHUNKER_MODEL,
            config=types.GenerateContentConfig(
                system_instruction=system_prompt,
            ),
            contents=text,
        )
        return response.text

    async def start_chunking(self, text: str, system_prompt: str) -> list[str]:
        """Process *text* in slices of ``CHUNK_SIZE`` with bounded concurrent ``chunk`` calls.

        Args:
            text (str): Full source text.
            system_prompt (str): System instruction passed to each ``chunk`` call.

        Returns:
            list[str]: Model outputs in the same order as the input slices.
        """
        if not text:
            return []

        pieces = [text[i : i + CHUNK_SIZE] for i in range(0, len(text), CHUNK_SIZE)]
        sem = asyncio.Semaphore(MAX_CHUNKING_JOBS_IN_ONE_MIN)

        async def _run_one(piece: str) -> str:
            async with sem:
                return await asyncio.to_thread(self._chunk, system_prompt, piece)

        return list(await asyncio.gather(*(_run_one(p) for p in pieces)))
