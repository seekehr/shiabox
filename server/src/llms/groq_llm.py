import os
import time
from collections.abc import Generator

from groq import Groq
from groq.types.chat import ChatCompletionChunk

from src.config import CHAT_MODEL, CHAT_PROMPT_FILE
from src.embedding import embed_text
from src.llms.prompts import build_chat_prompt
from src.vector_db import VectorDB


class GroqLLM:
    def __init__(
        self,
        api_key: str,
        system_prompt: str,
        model: str,
        vector_db: VectorDB,
    ):
        self.client = Groq(api_key=api_key)
        self.system_prompt = system_prompt
        self.model = model
        self.vector_db = vector_db

    @classmethod
    def create(cls, vector_db: VectorDB) -> "GroqLLM":
        api_key = os.getenv("GROQ_API_KEY")
        if not api_key:
            raise RuntimeError("GROQ_API_KEY env var not set")

        system_prompt = CHAT_PROMPT_FILE.read_text(encoding="utf-8")
        return cls(
            api_key=api_key,
            system_prompt=system_prompt,
            model=CHAT_MODEL,
            vector_db=vector_db,
        )

    def send_prompt_stream(
        self, user_prompt: str
    ) -> Generator[ChatCompletionChunk, None, None]:
        """Send a prompt to Groq and yield streamed chunks."""
        stream = self.client.chat.completions.create(
            model=self.model,
            messages=[
                {"role": "system", "content": self.system_prompt},
                {"role": "user", "content": user_prompt},
            ],
            stream=True,
        )
        yield from stream

    def handle_chat_request(
        self, prompt: str
    ) -> Generator[ChatCompletionChunk, None, None]:
        """
        Full RAG pipeline: embed query -> search Qdrant -> build prompt -> stream from Groq.
        Synchronous wrapper that calls asyncio for the embedding step.
        """
        import asyncio

        start = time.time()
        prompt = prompt.strip()

        print("\n\n====\nEmbedding prompt...")
        vectors = asyncio.run(embed_text(prompt))
        print(
            f"Prompt embedded (vec length {len(vectors)}). "
            "Searching the vector db..."
        )

        found = self.vector_db.search(vectors)
        print(f"{len(found)} responses found.")

        print("Building prompt...")
        parsed_prompt = build_chat_prompt(prompt, found)
        elapsed = time.time() - start
        print(f"Prompt built and db searched in {elapsed:.3f}s.")
        print(f"Sending prompt... (chars: {len(parsed_prompt)})")

        return self.send_prompt_stream(parsed_prompt)
