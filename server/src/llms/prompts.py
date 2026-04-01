from src.models import HadithEmbeddingResponse


def build_chat_prompt(
    input_text: str,
    similar_hadith: list[HadithEmbeddingResponse],
) -> str:
    parts: list[str] = [f"InputText: {input_text}", "<START>"]
    for h in similar_hadith:
        parts.append(f"Hadith: {h.Hadith}")
        parts.append(f"Page: {h.Page}")
        parts.append(f"Book: {h.Book}")
        parts.append(f"Score: {h.Score}")
        parts.append(f"Content: {h.Content}")
        parts.append("\n=====")
    parts.append("<END>")
    return "\n".join(parts)


def build_chunker_prompt(input_text: str) -> str:
    return "\n" + input_text
