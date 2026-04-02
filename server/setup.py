import asyncio
import json
from pathlib import Path

import aiofiles
import config.constants

semaphore = asyncio.Semaphore(10)

async def load_json_file(path: str):
    async with aiofiles.open(path, mode="r", encoding="utf-8") as f:
        content = await f.read()
        return json.loads(content)

async def load_json_file_limited(path: str):
    async with semaphore:
        return await load_json_file(path)

async def load_parsed_books():
    paths = list(Path(config.constants.PARSED_BOOKS_DIR).glob("*.json"))
    tasks = [load_json_file_limited(str(p)) for p in paths]
    return await asyncio.gather(*tasks)

async def main():
    print("Setting up...")
    print("Loading parsed books...")
    books = await load_parsed_books()
    print(f"Loaded {len(books)} parsed books!")
    print("Embedding books...")



if __name__ == "__main__":
    asyncio.run(main()) # creates a new event loop and waits for it