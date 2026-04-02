import ollama
from ollama import embed

from config.constants import EMBEDDING_MODEL


def embed_text(text: str) -> ollama.EmbedResponse|None:
    try:
        response = embed(
            model=EMBEDDING_MODEL,
            prompt=text
        )
        return response["embedding"]
    except Exception as e:
        print(f"Exception encountered while embedding text: {len(text)}. Error: {e}")
        return None