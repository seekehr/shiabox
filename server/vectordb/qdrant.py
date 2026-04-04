import shutil
from pathlib import Path

from qdrant_client import QdrantClient
from qdrant_client.models import Distance, VectorParams, PointStruct

import config.constants
from llm.embedding import embed_batch, embed_text


def setup_qdrant() -> QdrantClient:
    qdrant_path = Path(config.constants.QDRANT_PATH)
    if qdrant_path.exists():
        shutil.rmtree(qdrant_path)
    client = QdrantClient(path=str(qdrant_path))
    client.create_collection(
        collection_name=config.constants.HADITHS_COLLECTION,
        vectors_config=VectorParams(
            size=config.constants.EMBEDDING_DIMENSIONS,
            distance=Distance.COSINE,
        ),
    )
    return client


def embed_and_init_books(client: QdrantClient, books: list[tuple[str, list[dict]]]) -> int:
    point_id = 0
    for book_name, hadiths in books:
        print(f"  Embedding '{book_name}' ({len(hadiths)} hadiths)...")
        for i in range(0, len(hadiths), config.constants.BATCH_SIZE):
            batch = hadiths[i : i + config.constants.BATCH_SIZE]
            texts = [h["Content"] for h in batch]
            embeddings = embed_batch(texts)

            points = []
            for h, (emb, err) in zip(batch, embeddings):
                if emb is None:
                    print(f"    Skipping hadith {h.get('Hadith', 'N/A')} ({len(h.get('Content', ''))} chars): {err}")
                    continue
                points.append(PointStruct(
                    id=point_id,
                    vector=emb,
                    payload={
                        "book": book_name,
                        "hadith_number": h.get("Hadith", "N/A"),
                        "chapter": h.get("Chapter", "N/A"),
                        "content": h.get("Content", "N/A"),
                    },
                ))
                point_id += 1

            if points:
                client.upsert(
                    collection_name=config.constants.HADITHS_COLLECTION,
                    points=points,
                )
        print(f"  Done with '{book_name}'!")
    return point_id


def get_qdrant_client() -> QdrantClient:
    qdrant_path = Path(config.constants.QDRANT_PATH)
    if not qdrant_path.exists():
        raise FileNotFoundError(
            f"Qdrant data not found at {qdrant_path}. Run setup.py first."
        )
    return QdrantClient(path=str(qdrant_path))


def search_ahadith(client: QdrantClient, query: str, top_k: int = 5) -> list[dict]:
    embedding, err = embed_text(query)
    if embedding is None:
        raise ValueError(f"Failed to embed query: {err}")

    results = client.query_points(
        collection_name=config.constants.HADITHS_COLLECTION,
        query=embedding,
        limit=top_k,
    )

    return [
        {
            "score": point.score,
            "book": point.payload["book"],
            "hadith_number": point.payload["hadith_number"],
            "chapter": point.payload["chapter"],
            "content": point.payload["content"],
        }
        for point in results.points
    ]
