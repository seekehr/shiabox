import asyncio
import json
from pathlib import Path

import httpx

from src.config import (
    EMBED_BATCH_SIZE,
    EMBED_MODEL,
    EMBED_WORKER_COUNT,
    EMBEDDINGS_DIR,
    OLLAMA_EMBED_URL,
)
from src.models import HadithChunk, HadithEmbedding


async def embed_text(text: str, client: httpx.AsyncClient | None = None) -> list[float]:
    """Embed a single piece of text via Ollama."""
    payload = {"model": EMBED_MODEL, "input": text}
    if client is None:
        async with httpx.AsyncClient(timeout=30) as c:
            resp = await c.post(OLLAMA_EMBED_URL, json=payload)
    else:
        resp = await client.post(OLLAMA_EMBED_URL, json=payload)
    resp.raise_for_status()
    data = resp.json()
    embeddings = data.get("embeddings")
    if not embeddings:
        raise ValueError("Empty embedding returned from Ollama")
    return embeddings[0]


async def embed_batch(texts: list[str], client: httpx.AsyncClient) -> list[list[float]]:
    """Embed a batch of texts in a single Ollama request."""
    payload = {"model": EMBED_MODEL, "input": texts}
    resp = await client.post(OLLAMA_EMBED_URL, json=payload)
    resp.raise_for_status()
    data = resp.json()
    embeddings = data.get("embeddings")
    if not embeddings or len(embeddings) < 2:
        raise ValueError("Empty embedding batch returned from Ollama")
    return embeddings


async def embed_ahadith(
    chunks: list[HadithChunk],
    client: httpx.AsyncClient,
) -> list[HadithEmbedding]:
    """Embed a list of HadithChunks, returning HadithEmbeddings."""
    contents = [c.Content for c in chunks]
    embeddings = await embed_batch(contents, client)
    if len(embeddings) != len(chunks):
        raise ValueError(
            f"Length mismatch: {len(embeddings)} embeddings vs {len(chunks)} chunks"
        )
    return [
        HadithEmbedding(
            Hadith=chunks[i].Hadith,
            Embedding=embeddings[i],
            Book=chunks[i].Book,
            Page=chunks[i].Page,
            Content=chunks[i].Content,
        )
        for i in range(len(chunks))
    ]


async def embed_book(parsed_book_path: Path, output_name: str) -> None:
    """Read a parsed-book JSON file, embed every chunk, write to embeddings dir."""
    raw = parsed_book_path.read_text(encoding="utf-8")
    chunks = [HadithChunk(**obj) for obj in json.loads(raw)]
    total = len(chunks)
    print(f"Embedding {parsed_book_path.name}: {total} chunks")

    semaphore = asyncio.Semaphore(EMBED_WORKER_COUNT)
    embedded: list[HadithEmbedding] = [HadithEmbedding()] * total

    async with httpx.AsyncClient(timeout=30) as client:

        async def _process_batch(start: int, batch: list[HadithChunk]) -> None:
            async with semaphore:
                result = await embed_ahadith(batch, client)
                for offset, emb in enumerate(result):
                    embedded[start + offset] = emb

        tasks: list[asyncio.Task] = []
        for i in range(0, total, EMBED_BATCH_SIZE):
            batch = chunks[i : i + EMBED_BATCH_SIZE]
            tasks.append(asyncio.create_task(_process_batch(i, batch)))

        await asyncio.gather(*tasks)

    out_path = EMBEDDINGS_DIR / output_name
    out_path.write_text(
        json.dumps([e.model_dump() for e in embedded]),
        encoding="utf-8",
    )
    print(f"Finished embedding {parsed_book_path.name} -> {out_path}")


def read_embedded_book(path: Path) -> list[HadithEmbedding]:
    """Load a previously-embedded book JSON."""
    raw = path.read_text(encoding="utf-8")
    return [HadithEmbedding(**obj) for obj in json.loads(raw)]
