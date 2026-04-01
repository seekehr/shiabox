import subprocess
import unicodedata
from collections.abc import Generator
from pathlib import Path

# Characters the Go CleanText considers "bad"
_BAD_CHARS = frozenset("GHELZ%_7qdCI:$\"(){}~Y#!")


def pdf_to_txt(pdf_path: Path, txt_path: Path) -> None:
    """Convert a PDF to cleaned plain-text using poppler's pdftotext."""
    subprocess.run(
        ["pdftotext", "-enc", "ASCII7", str(pdf_path), str(txt_path)],
        check=True,
    )

    lines: list[str] = []
    prev_empty = False
    with open(txt_path, encoding="utf-8", errors="replace") as f:
        for raw_line in f:
            cleaned = clean_text(raw_line.rstrip("\n"))
            if cleaned == "":
                if prev_empty:
                    continue
                prev_empty = True
            else:
                prev_empty = False
            lines.append(cleaned)

    txt_path.write_text("\n".join(lines), encoding="utf-8")


def _is_purely_numeric(s: str) -> bool:
    return s != "" and all(ch.isdigit() for ch in s)


def clean_text(text: str) -> str:
    """Port of the Go CleanText function -- removes garbage lines."""
    cleaned_lines: list[str] = []
    for line in text.split("\n"):
        trimmed = line.strip()
        if not trimmed:
            continue

        if _is_purely_numeric(trimmed):
            cleaned_lines.append(line)
            continue

        parts = trimmed.split(".", 1)
        if len(parts) > 1 and _is_purely_numeric(parts[0]):
            cleaned_lines.append(line)
            continue

        runes = list(trimmed)
        if not runes:
            continue

        bad_count = sum(1 for ch in runes if ch in _BAD_CHARS)
        alpha_count = sum(1 for ch in runes if unicodedata.category(ch).startswith("L"))
        has_space = any(ch.isspace() for ch in runes)

        if len(runes) < 10 and has_space and alpha_count < 4:
            continue
        if bad_count / len(runes) > 0.4:
            continue
        if len(runes) > 1 and alpha_count / len(runes) < 0.5:
            continue

        cleaned_lines.append(line)

    return "\n".join(cleaned_lines)


def read_file_in_chunks(
    file_path: Path,
    chunk_size: int,
    overlap_size: int,
) -> Generator[str, None, None]:
    """
    Yield successive chunks of *file_path* with overlap, prefixed with
    <NO_OVERLAP> for the first chunk and <END> appended to the last.

    Mirrors the Go ReadFileInChunks channel behaviour.
    """
    data = file_path.read_bytes()
    total = len(data)
    offset = 0
    first = True

    while offset < total:
        end = min(offset + chunk_size, total)
        chunk = data[offset:end].decode("utf-8", errors="replace")

        if first:
            chunk = "<NO_OVERLAP>" + chunk
            first = False

        is_last = end >= total
        if is_last:
            chunk += "<END>"

        yield chunk

        offset += chunk_size
        offset -= overlap_size
        if offset >= total:
            break
