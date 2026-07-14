# Scraper

TypeScript scraper for [thaqalayn.net](https://thaqalayn.net). Discovers books and volumes with Playwright, then fetches chapter/hadith HTML over HTTP and writes JSON for the shiabox RAG pipeline.

## Install

```bash
npm install
npx playwright install chromium
```

## Usage

```bash
npm start
```

Enter a book name when prompted (accent-insensitive substring match on the homepage, e.g. `kafi` or `al kafi`).

Output: `output/<slugified-book-name>.json` (e.g. `output/al-kafi.json`).

## How it works

1. **Playwright** — find the book link and any volume URLs (SPA UI).
2. **HTTP fetch** — load volume, chapter, and hadith pages (concurrency 12).
3. **Parse** — extract English, Arabic, chapter metadata, and source URL via `node-html-parser`.
4. **Write** — sort by chapter/hadith number and save a JSON array.

## Output schema

```json
{
  "book_name": "Al-Kāfi",
  "volume_number": 1,
  "chapter_number": 0,
  "hadith_number": 1,
  "content": "English translation…",
  "metadata": {
    "book_number": 0,
    "chapter_name": "The Book of Intelligence and Ignorance",
    "arabic": "Arabic text…",
    "url": "https://thaqalayn.net/hadith/1/1/0/1"
  }
}
```

## Using with the server

Copy scraper JSON into `server/assets/parsed_books/`, then run `uv run python setup.py` from `server/` (see [INSTALLING.md](../INSTALLING.md)).
