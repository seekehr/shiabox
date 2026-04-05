import json
import os

from dotenv import load_dotenv
from fastapi import Request, APIRouter
from fastapi.responses import JSONResponse, StreamingResponse
from groq import Groq
from pydantic import BaseModel
from slowapi import Limiter
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded
from slowapi.middleware import SlowAPIMiddleware
from groq.types.chat import ChatCompletionChunk

from config.constants import MAX_REQUESTS_PER_MINUTE
from llm.chat import ChatLLM
from vectordb.qdrant import get_qdrant_client, search_ahadith

router = APIRouter()
limiter = Limiter(key_func=get_remote_address)

async def rate_limit_handler(request: Request, exc: RateLimitExceeded):
    """Custom error handler for rate limit exceeded.

    Args:
        request (Request): The request that exceeded the rate limit.
        exc (RateLimitExceeded): The exception that was raised.

    Returns:
        JSONResponse: A JSON response with the status code 429 and the message "Rate limit exceeded. Slow down."
    """
    return JSONResponse(
        status_code=429,
        content={"detail": "Rate limit exceeded. Slow down."},
    )


class AIRequest(BaseModel):
    """AI request body.

    Args:
        BaseModel (_type_): The base model.
        prompt (str): The prompt to send to the AI.
    """
    prompt: str

def _sse_chunk(message: str, index: int, finish_reason: str, role: str, content: str, done: bool) -> str:
    """Create a SSE chunk.

    Args:
        message (str): The message.
        index (int): The index.
        finish_reason (str): The finish reason.
        role (str): The role.
        content (str): The content.
        done (bool): Whether the chunk is done.

    Returns:
        str: The SSE chunk.
    """
    payload = {
        "message": message,
        "data": {
            "choices": [
                {
                    "index": index,
                    "finish_reason": finish_reason,
                    "delta": {"role": role, "content": content},
                }
            ]
        },
        "done": done,
    }
    return f"data: {json.dumps(payload, ensure_ascii=False)}\n\n"


def _groq_chunk_to_sse(chunk: ChatCompletionChunk) -> str | None:
    """Convert a Groq chunk to a SSE chunk.

    Args:
        chunk (ChatCompletionChunk): The Groq chunk.

    Returns:
        str | None: The SSE chunk.
    """
    if not chunk.choices:
        return None
    ch = chunk.choices[0]
    delta = ch.delta
    role = (delta.role or "assistant") if delta else "assistant"
    content = (delta.content or "") if delta else ""
    finish_reason = ch.finish_reason or ""
    return _sse_chunk(
        message="",
        index=ch.index,
        finish_reason=finish_reason,
        role=role,
        content=content,
        done=False,
    )


@router.post("/request")
@limiter.limit(f"{MAX_REQUESTS_PER_MINUTE}/minute")  # 5 requests per minute per IP
async def ai_request(request: Request, body: AIRequest):
    """AI request endpoint.

    Args:
        request (Request): The request object.
        body (AIRequest): The request body.

    Returns:
        StreamingResponse: A streaming response with the AI response.

    Yields:
        str: A SSE chunk of the AI response.
    """
    load_dotenv()

    def event_stream():
        qdrant = None
        try:
            qdrant = get_qdrant_client()
            results = search_ahadith(qdrant, body.prompt, top_k=10)
            groq_client = Groq(api_key=os.getenv("GROQ_API_KEY"))
            chat = ChatLLM(groq_client)
            stream = chat.prompt(body.prompt, results)
            for chunk in stream:
                line = _groq_chunk_to_sse(chunk)
                if line:
                    yield line
            yield _sse_chunk(
                message="",
                index=0,
                finish_reason="stop",
                role="assistant",
                content="",
                done=True,
            )
        except Exception as e:
            yield _sse_chunk(
                message=str(e),
                index=0,
                finish_reason="",
                role="assistant",
                content="",
                done=True,
            )
        finally:
            if qdrant is not None:
                qdrant.close()

    return StreamingResponse(
        event_stream(),
        media_type="text/event-stream; charset=utf-8",
    )
