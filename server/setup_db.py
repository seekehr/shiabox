"""
Setup pipeline for shiabox: PDF -> TXT -> Chunk (Gemini) -> Postprocess -> Embed -> Qdrant.

Translates cmd/setup.go.
"""

import asyncio
import json
import re
import time
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

from dotenv import load_dotenv

from src.config import (
    CHUNK_SIZE_CHARACTERS,
    EMBEDDINGS_DIR,
    MAX_REQUESTS_PER_MIN,
    MAX_VECTOR_WORKERS,
    MAX_VECTORS_PER_BATCH,
    OVERLAP_CHARACTERS,
    PARSED_BOOKS_DIR,
    PDF_BOOKS_DIR,
    RATELIMIT_SLEEP_SECONDS,
    UNPARSED_BOOKS_DIR,
)
from src.embedding import embed_book, read_embedded_book
from src.llms.gemini_llm import GeminiLLM
from src.utils import pdf_to_txt, read_file_in_chunks
from src.vector_db import VectorDB

FLAG_INIT_BOOKS = 0
FLAG_POSTPROCESS_BOOKS = 1
FLAG_EMBED_BOOKS = 2
FLAG_INIT_VECTORS = 3
FLAG_INIT_BOTH = 4


# ---------------------------------------------------------------------------
# Step 0: PDF -> TXT
# ---------------------------------------------------------------------------

def pdf_to_txt_books() -> None:
    print("MAKE SURE YOU HAVE PDFTOTEXT BY POPPLER'S UTILS INSTALLED.")
    print("Converting PDF books to TXT format...")
    start = time.time()

    pdf_files = [f for f in PDF_BOOKS_DIR.iterdir() if f.suffix == ".pdf"]
    with ThreadPoolExecutor() as pool:
        futures = []
        for pdf in pdf_files:
            txt_out = UNPARSED_BOOKS_DIR / pdf.with_suffix(".txt").name
            futures.append(pool.submit(pdf_to_txt, pdf, txt_out))
        for fut in futures:
            fut.result()

    print(f"Converting books to TXT done in {time.time() - start:.3f}s.")


# ---------------------------------------------------------------------------
# Step 0 (cont.): Chunk books using Gemini
# ---------------------------------------------------------------------------

def chunk_books(gemini: GeminiLLM) -> None:
    start = time.time()
    txt_files = [f for f in UNPARSED_BOOKS_DIR.iterdir() if f.suffix == ".txt"]

    for txt_file in txt_files:
        _process_book(txt_file, gemini)

    print(f"Chunking done in {time.time() - start:.3f}s.")


def _process_book(book_path: Path, gemini: GeminiLLM) -> None:
    name = book_path.name
    out_path = PARSED_BOOKS_DIR / book_path.with_suffix(".json").name
    out_path.unlink(missing_ok=True)

    print(f"Sending requests to Gemini for book {name}.")

    chunks = list(
        read_file_in_chunks(book_path, CHUNK_SIZE_CHARACTERS, OVERLAP_CHARACTERS)
    )

    finished_jobs: list[tuple[int, str]] = []
    requests_counter = 0

    for index, chunk in enumerate(chunks):
        if requests_counter >= MAX_REQUESTS_PER_MIN:
            print(f"Sleeping {RATELIMIT_SLEEP_SECONDS}s to prevent rate limit.")
            time.sleep(RATELIMIT_SLEEP_SECONDS)
            requests_counter = 0

        print(f"Processing chunk {index}...")
        resp = gemini.send_prompt(chunk)
        print(f"Stop reason: {resp.finish_reason}")
        finished_jobs.append((index, resp.content))
        print(f"Received response from Gemini for chunk {index}.")
        requests_counter += 1

    finished_jobs.sort(key=lambda x: x[0])
    content = "".join(resp for _, resp in finished_jobs)

    out_path.write_text(content, encoding="utf-8")
    print(
        f"Written {len(content)} chars to {out_path.name}. "
        f"Finished processing LLM response for {name}."
    )


# ---------------------------------------------------------------------------
# Step 1: Postprocess JSON (clean up LLM artefacts)
# ---------------------------------------------------------------------------

def postprocess_books() -> None:
    start = time.time()
    json_files = [f for f in PARSED_BOOKS_DIR.iterdir() if f.suffix == ".json"]

    for json_file in json_files:
        _postprocess_file(json_file)

    print(f"Postprocessing books done in {time.time() - start:.3f}s.")


def _postprocess_file(path: Path) -> None:
    raw = path.read_text(encoding="utf-8")
    lines = raw.splitlines()

    cleaned: list[str] = []
    for line in lines:
        stripped = line.strip()
        if stripped in ("```", "```json"):
            continue
        if stripped in ("[", "]"):
            continue
        # Fix \" followed by , (Gemini sometimes escapes trailing quotes)
        stripped = re.sub(r'\\",', r'\\",",', stripped)
        cleaned.append(stripped)

    if cleaned and cleaned[0].strip() != "[":
        cleaned.insert(0, "[")
    if len(cleaned) > 1 and cleaned[-1].strip() != "]":
        cleaned.append("]")

    final = "\n".join(cleaned)
    # Clean hallucinated JSON array boundaries: ] [
    final = re.sub(r"}\s*\]\s*\[\s*\{", "},{", final)
    path.write_text(final, encoding="utf-8")


# ---------------------------------------------------------------------------
# Step 2: Embed books (Ollama mxbai-embed-large)
# ---------------------------------------------------------------------------

def embed_books() -> None:
    parsed_files = [f for f in PARSED_BOOKS_DIR.iterdir() if f.suffix == ".json"]
    for f in parsed_files:
        print(f"Embedding file: {f.name}")
        asyncio.run(embed_book(f, f.name))


# ---------------------------------------------------------------------------
# Step 3: Feed embeddings into Qdrant
# ---------------------------------------------------------------------------

def init_vectors(db: VectorDB) -> None:
    embedding_files = [f for f in EMBEDDINGS_DIR.iterdir() if f.suffix == ".json"]

    for embed_file in embedding_files:
        embed_data = read_embedded_book(embed_file)

        with ThreadPoolExecutor(max_workers=MAX_VECTOR_WORKERS) as pool:
            futures = []
            for i in range(0, len(embed_data), MAX_VECTORS_PER_BATCH):
                batch = embed_data[i : i + MAX_VECTORS_PER_BATCH]
                futures.append(pool.submit(db.add, batch))

            for fut in futures:
                fut.result()

    print(f"Vector count: {db.count()}")


# ---------------------------------------------------------------------------
# Main menu
# ---------------------------------------------------------------------------

def main() -> None:
    load_dotenv()

    db = VectorDB.connect()

    print(
        "Would you like to parse books or initialise the vector db?\n"
        "  0 = Full init (PDF->TXT, chunk, postprocess)\n"
        "  1 = Postprocess only\n"
        "  2 = Embed only\n"
        "  3 = Init vectors only\n"
        "  4 = Embed + init vectors\n"
    )
    flag = int(input("Enter flag: "))
    start = time.time()

    if flag == FLAG_INIT_BOOKS:
        pdf_to_txt_books()
        gemini = GeminiLLM.create()
        chunk_books(gemini)
        postprocess_books()
    elif flag == FLAG_POSTPROCESS_BOOKS:
        postprocess_books()
    elif flag == FLAG_EMBED_BOOKS:
        print("Generating embeddings...")
        embed_books()
    elif flag == FLAG_INIT_VECTORS:
        init_vectors(db)
    elif flag == FLAG_INIT_BOTH:
        print("Generating embeddings...")
        embed_books()
        init_vectors(db)
    else:
        raise ValueError(f"Invalid flag: {flag}")

    print(f"Done in {time.time() - start:.3f}s.")


if __name__ == "__main__":
    main()
