import os
from dataclasses import dataclass

from google import genai
from google.genai.types import Content, GenerateContentResponse, Part

from src.config import CHUNKER_MODEL, CHUNKER_PROMPT_FILE


@dataclass
class GeminiResponse:
    content: str
    finish_reason: str


class GeminiLLM:
    def __init__(self, system_prompt: str, model: str):
        self.client = genai.Client(api_key=os.getenv("GEMINI_API_KEY"))
        self.system_prompt = system_prompt
        self.model = model

    @classmethod
    def create(cls) -> "GeminiLLM":
        system_prompt = CHUNKER_PROMPT_FILE.read_text(encoding="utf-8")
        return cls(system_prompt=system_prompt, model=CHUNKER_MODEL)

    def send_prompt(self, user_prompt: str) -> GeminiResponse:
        """Send a prompt to Gemini and return the parsed response."""
        contents = [
            Content(parts=[Part(text=self.system_prompt)], role="user"),
            Content(parts=[Part(text=user_prompt)], role="user"),
        ]

        resp: GenerateContentResponse = self.client.models.generate_content(
            model=self.model,
            contents=contents,
        )

        if not resp.candidates:
            raise ValueError("No candidates returned from Gemini")

        candidate = resp.candidates[0]
        if not candidate.content or not candidate.content.parts:
            raise ValueError("No content returned from Gemini")

        return GeminiResponse(
            content=candidate.content.parts[0].text or "",
            finish_reason=str(candidate.finish_reason or ""),
        )
