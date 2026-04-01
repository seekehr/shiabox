import hashlib

from qdrant_client import QdrantClient
from qdrant_client.models import (
    Distance,
    PointStruct,
    VectorParams,
)

from src.config import COLLECTION_NAME, MAX_RESULTS_LIMIT, VECTOR_SIZE
from src.models import HadithEmbedding, HadithEmbeddingResponse


class VectorDB:
    def __init__(self, client: QdrantClient):
        self.client = client

    @classmethod
    def connect(cls, host: str = "localhost", port: int = 6333) -> "VectorDB":
        """Connect to Qdrant and auto-create the collection if it doesn't exist."""
        client = QdrantClient(host=host, port=port)

        if not client.collection_exists(COLLECTION_NAME):
            print(f"Collection '{COLLECTION_NAME}' not found, creating it.")
            client.create_collection(
                collection_name=COLLECTION_NAME,
                vectors_config=VectorParams(
                    size=VECTOR_SIZE,
                    distance=Distance.COSINE,
                ),
            )

        return cls(client)

    @staticmethod
    def generate_uuid(hadith: HadithEmbedding) -> str:
        raw = f"{hadith.Book}{hadith.Hadith}"
        return hashlib.md5(raw.encode()).hexdigest()

    def add(self, embeddings: list[HadithEmbedding]) -> None:
        """Upsert a batch of hadith embeddings into Qdrant."""
        points = [
            PointStruct(
                id=self.generate_uuid(h),
                vector=h.Embedding,
                payload={
                    "Book": h.Book,
                    "Page": h.Page,
                    "Hadith": h.Hadith,
                    "Content": h.Content,
                },
            )
            for h in embeddings
        ]
        self.client.upsert(collection_name=COLLECTION_NAME, points=points)

    def search(self, query_embedding: list[float]) -> list[HadithEmbeddingResponse]:
        """Search for the most similar hadith embeddings."""
        results = self.client.query_points(
            collection_name=COLLECTION_NAME,
            query=query_embedding,
            with_payload=True,
            with_vectors=True,
            limit=MAX_RESULTS_LIMIT,
        )

        responses: list[HadithEmbeddingResponse] = []
        for point in results.points:
            payload = point.payload or {}
            responses.append(
                HadithEmbeddingResponse(
                    Hadith=int(payload.get("Hadith", 0)),
                    Embedding=point.vector if isinstance(point.vector, list) else [],
                    Book=str(payload.get("Book", "")),
                    Page=int(payload.get("Page", 0)),
                    Content=str(payload.get("Content", "")),
                    Score=point.score,
                )
            )
        return responses

    def count(self) -> int:
        info = self.client.get_collection(COLLECTION_NAME)
        return info.points_count
