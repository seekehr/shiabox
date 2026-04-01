import json
from collections.abc import AsyncGenerator

from fastapi import APIRouter, Request
from fastapi.responses import StreamingResponse
from pydantic import BaseModel

from src.llms.groq_llm import GroqLLM

router = APIRouter(prefix="/ai")


class RequestBody(BaseModel):
    prompt: str


def _sse_event(message: str = "", data: object = None, done: bool = False) -> str:
    payload = json.dumps({"message": message, "data": data, "done": done})
    return f"data: {payload}\n\n"


async def _stream_response(
    groq: GroqLLM, prompt: str
) -> AsyncGenerator[str, None]:
    try:
        stream = groq.handle_chat_request(prompt)
    except Exception as e:
        message = f"Error handling request. Error: {e}"
        if "ratelimit" in str(e).lower():
            message = (
                "30 msgs/s rate-limit reached of server. "
                "Please donate to help increase our rate-limit."
            )
        yield _sse_event(message=message, done=True)
        return

    for chunk in stream:
        if chunk.choices:
            choice = chunk.choices[0]
            yield _sse_event(
                data={
                    "choices": [
                        {
                            "index": choice.index,
                            "finish_reason": choice.finish_reason,
                            "delta": {
                                "role": choice.delta.role if choice.delta else None,
                                "content": choice.delta.content if choice.delta else None,
                            },
                        }
                    ]
                }
            )

    yield _sse_event(done=True)


@router.post("/request")
async def post_request(body: RequestBody, request: Request) -> StreamingResponse:
    groq: GroqLLM = request.app.state.groq

    if len(body.prompt) > 500:
        async def _err() -> AsyncGenerator[str, None]:
            yield _sse_event(message="Max prompt length is 500 characters.", done=True)

        return StreamingResponse(_err(), media_type="text/event-stream")

    return StreamingResponse(
        _stream_response(groq, body.prompt),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
        },
    )
