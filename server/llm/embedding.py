from ollama import embed

from config.constants import EMBEDDING_MODEL, TEXT_SIZE


def _truncate(text: str) -> str:
    if len(text) > TEXT_SIZE:
        return text[-TEXT_SIZE:]
    return text


def embed_text(text: str) -> tuple[list[float] | None, str | None]:
    try:
        response = embed(model=EMBEDDING_MODEL, input=_truncate(text))
        return response.embeddings[0], None
    except Exception as e:
        return None, str(e)


def embed_batch(texts: list[str]) -> list[tuple[list[float] | None, str | None]]:
    truncated = [_truncate(t) for t in texts]
    try:
        response = embed(model=EMBEDDING_MODEL, input=truncated)
        return [(emb, None) for emb in response.embeddings]
    except Exception as e:
        print(f"    Batch failed ({len(texts)} texts), falling back to individual: {e}")
        return [embed_text(t) for t in texts]