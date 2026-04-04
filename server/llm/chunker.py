from google import genai
from google.genai import types

from config.constants import CHUNKER_MODEL

class ChunkerLLM:
    """A class to handle requests to the chunker (gemini) LLM.
    """
    def __init__(self):
        """Initialize the chunker LLM helper.

        Args:
            client (genai.Client): The gemini API client.
        """
        # The client gets the API key from the environment variable `GEMINI_API_KEY`.
        self.client = genai.Client()

    def chunk(self, system_prompt: str, text: str) -> str:
        """Chunk the text using the gemini LLM.

        Args:
            system_prompt (str): The system prompt to use.
            text (str): The text to chunk.

        Returns:
            str: The chunked text.
        """
        response = self.client.models.generate_content(
            model=CHUNKER_MODEL,
            config=types.GenerateContentConfig(
                system_instruction=system_prompt,
            ),
            contents=text,
        )
        return response.text
