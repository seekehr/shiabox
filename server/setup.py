import asyncio
import json
from pathlib import Path
import re

import aiofiles
from dotenv import load_dotenv

import config.constants
from qdrant_client import QdrantClient

from vectordb.qdrant import setup_qdrant, embed_and_init_books, get_qdrant_client
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
    """Load every *.json under CHUNKED_BOOKS_DIR concurrently.

    Returns:
        list[tuple[str, list[dict]]]: (stem filename, parsed hadith list) per book.
    """
    paths = list(Path(config.constants.PARSED_BOOKS_DIR).glob("*.json"))
    tasks = [load_json_file_limited(str(p)) for p in paths]
    books_data = await asyncio.gather(*tasks)
    return [(p.stem, data) for p, data in zip(paths, books_data)]

async def convert_pdf_to_text() -> list[Path]:
    """Convert each PDF in PDF_BOOKS_DIR to text in TXT_BOOKS_DIR via pdftotext, strip Arabic, remove PDFs.
    Any .txt files mistakenly placed in PDF_BOOKS_DIR are moved to TXT_BOOKS_DIR.

    Returns:
        list[Path]: Paths to text files successfully converted.
    """
    txt_dir = Path(config.constants.TXT_BOOKS_DIR)
    txt_dir.mkdir(parents=True, exist_ok=True)

    for misplaced in Path(config.constants.PDF_BOOKS_DIR).glob("*.txt"):
        dest = txt_dir / misplaced.name
        misplaced.rename(dest)
        print(f"Moved {misplaced.name} to txt_books/")

    pdf_books = list(Path(config.constants.PDF_BOOKS_DIR).glob("*.pdf"))
    text_files = []
    for pdf_book in pdf_books:
        print(f"Converting {pdf_book.stem} to text...")
        output_path = txt_dir / f"{pdf_book.stem}.txt"
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


def chunked_book_path(name: str) -> Path:
    return Path(config.constants.PARSED_BOOKS_DIR) / f"{name}.json"


def chunked_book_exists(name: str) -> bool:
    return chunked_book_path(name).is_file()


def parsed_book_path(name: str) -> Path:
    return Path(config.constants.PARSED_BOOKS_DIR) / f"{name}.json"


def unchunked_text_books() -> list[Path]:
    """Return .txt books that have not yet been chunked."""
    seen: set[str] = set()
    books: list[Path] = []
    for directory in (config.constants.TXT_BOOKS_DIR, config.constants.PDF_BOOKS_DIR):
        for path in sorted(Path(directory).glob("*.txt")):
            if path.stem in seen or chunked_book_exists(path.stem):
                continue
            seen.add(path.stem)
            books.append(path)
    return books


def all_text_books() -> list[Path]:
    """Return all .txt books regardless of chunked status."""
    seen: set[str] = set()
    books: list[Path] = []
    for directory in (config.constants.TXT_BOOKS_DIR, config.constants.PDF_BOOKS_DIR):
        for path in sorted(Path(directory).glob("*.txt")):
            if path.stem not in seen:
                seen.add(path.stem)
                books.append(path)
    return books


def save_chunked_book(name: str, chunked: str) -> list[dict]:
    data = json.loads(chunked)
    path = chunked_book_path(name)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, indent=2), encoding="utf-8")
    return data


def open_qdrant_for_incremental() -> tuple[QdrantClient, int]:
    qdrant_path = Path(config.constants.QDRANT_PATH)
    if qdrant_path.exists():
        client = get_qdrant_client()
        start_id = client.get_collection(config.constants.HADITHS_COLLECTION).points_count or 0
        return client, start_id
    return setup_qdrant(), 0


async def main():
    """Convert PDFs, load parsed books, init Qdrant, embed hadiths, and upsert into the DB."""
    load_dotenv()
    only_new = input("Only parse new books? (0|1): ").strip() == "1"
    print("Setting up... MAKE SURE PDFTOTEXT by Popper's Utils is installed and in the PATH.")
    print("Converting PDF books to text...")
    await convert_pdf_to_text()

    chunker = ChunkerLLM()
    newly_chunked: list[tuple[str, list[dict]]] = []

    if only_new:
        to_chunk = unchunked_text_books()
        print(f"Found {len(to_chunk)} unchunked text book(s) to chunk.")
        for text_file in to_chunk:
            print(f"Chunking book in batch: {text_file.stem}...")
            chunked = await chunker.start_chunking(
                text_file.read_text(encoding="utf-8"),
                get_chunking_prompt(text_file.stem),
            )
            hadiths = save_chunked_book(text_file.stem, chunked)
            newly_chunked.append((text_file.stem, hadiths))
            print(f"Chunked: {text_file.stem}")

        if not newly_chunked:
            print("No new books to chunk. Loading all chunked books for embedding...")
            books = await load_parsed_books()
        else:
            books = newly_chunked

        print("Opening Qdrant for incremental update...")
        client, start_id = open_qdrant_for_incremental()
    else:
        to_chunk = unchunked_text_books()
        print(f"Need to chunk {len(to_chunk)} book(s) into ahadith...")
        for text_file in to_chunk:
            print(f"Chunking book in batch: {text_file.stem}...")
            chunked = await chunker.start_chunking(
                text_file.read_text(encoding="utf-8"),
                get_chunking_prompt(text_file.stem),
            )
            save_chunked_book(text_file.stem, chunked)
            print(f"Chunked: {text_file.stem}")

        print("Loading chunked books...")
        books = await load_parsed_books()
        print("Initializing Qdrant...")
        client = setup_qdrant()
        start_id = 0

    total_hadiths = sum(len(hadiths) for _, hadiths in books)
    print(f"Embedding {len(books)} book(s) ({total_hadiths} hadiths total)...")
    next_id = embed_and_init_books(client, books, start_id=start_id)
    print(f"Successfully embedded {next_id - start_id} hadiths into Qdrant!")

    client.close()


if __name__ == "__main__":
    asyncio.run(main())
