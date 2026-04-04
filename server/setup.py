import asyncio
import json
from pathlib import Path
import re

import aiofiles

import config.constants
from vectordb.qdrant import setup_qdrant, embed_and_init_books

semaphore = asyncio.Semaphore(10)


async def load_json_file(path: str):
    """Load a single JSON file."""
    async with aiofiles.open(path, mode="r", encoding="utf-8") as f:
        content = await f.read()
        return json.loads(content)


async def load_json_file_limited(path: str):
    """Load a single JSON file with concurrent limit."""
    async with semaphore:
        return await load_json_file(path)


async def load_parsed_books() -> list[tuple[str, list[dict]]]:
    """Load all parsed book JSON files concurrently.

    Returns:
        List of (book_name, entries).
    """
    paths = list(Path(config.constants.PARSED_BOOKS_DIR).glob("*.json"))
    tasks = [load_json_file_limited(str(p)) for p in paths]
    books_data = await asyncio.gather(*tasks)
    return [(p.stem, data) for p, data in zip(paths, books_data)]

async def convert_pdf_to_text():
    """Convert all PDF books to text, and delete the ones successfully converted from pdf folder."""
    pdf_books = list(Path(config.constants.PDF_BOOKS_DIR).glob("*.pdf"))

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

async def main():
    """Ready-up all the PDF books, convert them into text if needed, chunk them
    into ahadith if needed, embed the ahadith, and save them all in a qdrant database.
    """
    print("Setting up... MAKE SURE PDFTOTEXT by Popper's Utils is installed and in the PATH.")
    print("Converting PDF books to text...")
    await convert_pdf_to_text()

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
