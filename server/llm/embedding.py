from ollama import embed

from config.constants import EMBEDDING_MODEL, TEXT_SIZE


def _truncate(text: str) -> str:
    if len(text) > TEXT_SIZE:
        return text[-TEXT_SIZE:]
    return text


def embed_text(text: str) -> tuple[list[float] | None, str | None]:
    """Embed a single text.

    Args:
        text (str): The text to embed.

    Returns:
        tuple[list[float] | None, str | None]: The embedding and any errors.
    """
    try:
        response = embed(model=EMBEDDING_MODEL, input=_truncate(text))
        # Pylint/astroid mis-infers `ollama.embed` as a generator; runtime value is EmbedResponse.
        return response.embeddings[0], None  # pylint: disable=no-member
    except Exception as e:
        return None, str(e)


def embed_batch(texts: list[str]) -> list[tuple[list[float] | None, str | None]]:
    """Embed a batch of text.

    Args:
        texts (list[str]): The text to embed.

    Returns:
        list[tuple[list[float] | None, str | None]]: The embeddings and any errors.
    """
    truncated = [_truncate(t) for t in texts]
    try:
        response = embed(model=EMBEDDING_MODEL, input=truncated)
        return [(emb, None) for emb in response.embeddings]  # pylint: disable=no-member
    except Exception as e:
        print(f"    Batch failed ({len(texts)} texts), falling back to individual: {e}")
        return [embed_text(t) for t in texts]
