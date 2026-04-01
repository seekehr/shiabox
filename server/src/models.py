from pydantic import BaseModel


class HadithChunk(BaseModel):
    Book: str = ""
    Page: int = 0
    Content: str = ""
    Hadith: int = 0


class HadithEmbedding(BaseModel):
    Hadith: int = 0
    Embedding: list[float] = []
    Book: str = ""
    Page: int = 0
    Content: str = ""


class HadithEmbeddingResponse(HadithEmbedding):
    Score: float = 0.0
