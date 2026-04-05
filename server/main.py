import os

from dotenv import load_dotenv
from groq import Groq

from llm.chat import ChatLLM
from vectordb.qdrant import get_qdrant_client, search_ahadith
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from slowapi.errors import RateLimitExceeded
from slowapi.middleware import SlowAPIMiddleware
from api.ai_routes import limiter, rate_limit_handler, router as ai_router
import uvicorn

def input_loop():
    """Run an interactive CLI: server stub or RAG chat against Qdrant + Groq."""
    load_dotenv()
    requester = ChatLLM(Groq(api_key=os.getenv("GROQ_API_KEY")))

    flag = input("Start server (0) or start chatting (1)? ")
    if flag == "0":
        app = FastAPI()
        app.state.limiter = limiter
        app.add_middleware(SlowAPIMiddleware)
        app.add_middleware(
            CORSMiddleware,
            allow_origins=[
                "http://localhost:5173",
                "http://127.0.0.1:5173",
            ],
            allow_methods=["*"],
            allow_headers=["*"],
        )
        app.add_exception_handler(RateLimitExceeded, rate_limit_handler)
        # Routes
        app.include_router(ai_router, prefix="/ai")
        uvicorn.run(app, host="0.0.0.0", port=1323)

    elif flag == "1":
        client = get_qdrant_client()
        try:
            while True:
                user_prompt = input("Enter your prompt: ")
                results = search_ahadith(client, user_prompt, top_k=10)
                print(f"Found {len(results)} candidate ahadith, filtering...\n")
                stream = requester.prompt(user_prompt, search_results=results)
                for chunk in stream:
                    content = chunk.choices[0].delta.content
                    if content:
                        print(content, end="", flush=True)
                print()
        finally:
            client.close()
    else:
        print("Unexpected input.")



if __name__ == '__main__':
    input_loop()
