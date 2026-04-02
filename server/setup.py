import asyncio
import json
from pathlib import Path

import aiofiles

import config.constants
from vectordb.qdrant import setup_qdrant, embed_and_init_books

semaphore = asyncio.Semaphore(10)


async def load_json_file(path: str):
    async with aiofiles.open(path, mode="r", encoding="utf-8") as f:
        content = await f.read()
        return json.loads(content)


async def load_json_file_limited(path: str):
    async with semaphore:
        return await load_json_file(path)


async def load_parsed_books() -> list[tuple[str, list[dict]]]:
    paths = list(Path(config.constants.PARSED_BOOKS_DIR).glob("*.json"))
    tasks = [load_json_file_limited(str(p)) for p in paths]
    books_data = await asyncio.gather(*tasks)
    return [(p.stem, data) for p, data in zip(paths, books_data)]


async def main():
    print("Setting up...")
    print("Loading parsed books...")
    books = await load_parsed_books()
    total_hadiths = sum(len(hadiths) for _, hadiths in books)
    print(f"Loaded {len(books)} books ({total_hadiths} hadiths total)")

    print("Initializing Qdrant...")
    client = setup_qdrant()

    print("Embedding hadiths...")
    embedded = embed_and_init_books(client, books)
    print(f"Successfully embedded {embedded} hadiths into Qdrant!")

    client.close()


if __name__ == "__main__":
    asyncio.run(main())
