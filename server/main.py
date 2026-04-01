"""FastAPI server for shiabox -- translates cmd/server.go."""

from dotenv import load_dotenv
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.config import FRONTEND_URL
from src.handlers.ai_handler import router as ai_router
from src.llms.groq_llm import GroqLLM
from src.vector_db import VectorDB

load_dotenv()

app = FastAPI(title="shiabox")

app.add_middleware(
    CORSMiddleware,
    allow_origins=[FRONTEND_URL],
    allow_methods=["GET", "POST", "OPTIONS"],
    allow_headers=["Origin", "Content-Type", "Accept", "Authorization"],
    allow_credentials=True,
)

app.include_router(ai_router)


@app.on_event("startup")
def startup() -> None:
    vector_db = VectorDB.connect()
    groq = GroqLLM.create(vector_db)
    app.state.groq = groq


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("main:app", host="0.0.0.0", port=1323, reload=True)
