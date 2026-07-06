import asyncio
import json
from pathlib import Path

import aiofiles
from dotenv import load_dotenv

import config.constants
from qdrant_client import QdrantClient

from vectordb.qdrant import setup_qdrant, embed_and_init_books, get_qdrant_client

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


def open_qdrant_for_incremental() -> tuple[QdrantClient, int]:
    qdrant_path = Path(config.constants.QDRANT_PATH)
    if qdrant_path.exists():
        client = get_qdrant_client()
        start_id = client.get_collection(config.constants.HADITHS_COLLECTION).points_count or 0
        return client, start_id
    return setup_qdrant(), 0


async def main():
    load_dotenv()
    only_new = input("Only embed new books? (0|1): ").strip() == "1"

    if only_new:
        print("Opening Qdrant for incremental update...")
        client, start_id = open_qdrant_for_incremental()
    else:
        print("Initializing Qdrant...")
        client = setup_qdrant()
        start_id = 0

    print("Loading parsed books...")
    books = await load_parsed_books()

    total_hadiths = sum(len(hadiths) for _, hadiths in books)
    print(f"Embedding {len(books)} book(s) ({total_hadiths} hadiths total)...")
    next_id = embed_and_init_books(client, books, start_id=start_id)
    print(f"Successfully embedded {next_id - start_id} hadiths into Qdrant!")

    client.close()


if __name__ == "__main__":
    asyncio.run(main())
