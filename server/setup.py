import asyncio
import json
from pathlib import Path
import re

import aiofiles
from dotenv import load_dotenv

import config.constants
from vectordb.qdrant import setup_qdrant, embed_and_init_books
from llm.chunker import ChunkerLLM
from config.file_configs import get_chunking_prompt

semaphore = asyncio.Semaphore(10)


async def load_json_file(path: str):
    """Load and parse one JSON file.

    Args:
        path (str): Filesystem path to the JSON file.

    Returns:
        object: Deserialized JSON (structure depends on the file).
    """
    async with aiofiles.open(path, mode="r", encoding="utf-8") as f:
        content = await f.read()
        return json.loads(content)


async def load_json_file_limited(path: str):
    """Load one JSON file under the module concurrency semaphore.

    Args:
        path (str): Filesystem path to the JSON file.

    Returns:
        object: Deserialized JSON (structure depends on the file).
    """
    async with semaphore:
        return await load_json_file(path)


async def load_parsed_books() -> list[tuple[str, list[dict]]]:
    """Load every *.json under PARSED_BOOKS_DIR concurrently.

    Returns:
        list[tuple[str, list[dict]]]: (stem filename, parsed hadith list) per book.
    """
    paths = list(Path(config.constants.PARSED_BOOKS_DIR).glob("*.json"))
    tasks = [load_json_file_limited(str(p)) for p in paths]
    books_data = await asyncio.gather(*tasks)
    return [(p.stem, data) for p, data in zip(paths, books_data)]

async def convert_pdf_to_text() -> list[str]:
    """Convert each PDF in PDF_BOOKS_DIR to text via pdftotext, strip Arabic, remove PDFs.

    Returns:
        list[str]: List of text files successfully converted.
    """
    pdf_books = list(Path(config.constants.PDF_BOOKS_DIR).glob("*.pdf"))
    text_files = []
    for pdf_book in pdf_books:
        print(f"Converting {pdf_book.stem} to text...")
        output_path = pdf_book.with_suffix(".txt")
        process = await asyncio.create_subprocess_exec(
            "pdftotext",
            str(pdf_book),
            str(output_path),
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )

        _, stderr = await process.communicate()

        if process.returncode != 0:
            print(f"Failed: {pdf_book.name}")
            print(stderr.decode())
            continue

        # Strip Arabic characters
        text = output_path.read_text(encoding="utf-8", errors="ignore")
        clean_text = re.sub(r'[\u0600-\u06FF]+', '', text)
        output_path.write_text(clean_text, encoding="utf-8")

        pdf_book.unlink()
        print(f"Converted: {output_path.name}")
        text_files.append(output_path)
    return text_files

async def main():
    """Convert PDFs, load parsed books, init Qdrant, embed hadiths, and upsert into the DB."""
    load_dotenv()
    print("Setting up... MAKE SURE PDFTOTEXT by Popper's Utils is installed and in the PATH.")
    print("Converting PDF books to text...")
    text_files = await convert_pdf_to_text()
    print(f"Need to chunk {len(text_files)} books into ahadith...")
    chunker = ChunkerLLM()
    for text_file in text_files:
        chunked = await chunker.start_chunking(text_file.read_text(encoding="utf-8"), get_chunking_prompt(text_file.stem))
        text_file.write_text(chunked, encoding="utf-8")
        print(f"Chunked: {text_file.name}")

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
