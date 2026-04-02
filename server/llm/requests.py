import os
from groq import Groq, Stream
from groq.types.chat import ChatCompletionSystemMessageParam, ChatCompletionUserMessageParam, ChatCompletion, \
    ChatCompletionChunk
from config.file_configs import get_chat_prompt
from config.constants import CHAT_MODEL

class LLMRequester:
    def __init__(self, client: Groq):
        print(os.getenv("GROQ_API_KEY"))
        self._client = client
        self._chat_prompt = get_chat_prompt()

    def prompt(self, user_prompt: str) -> ChatCompletion | Stream[ChatCompletionChunk]:
        system_msg: ChatCompletionSystemMessageParam = {"role": "system", "content": self._chat_prompt}
        user_msg: ChatCompletionUserMessageParam = {"role": "user", "content": user_prompt}
        return self._client.chat.completions.create(
            messages=[system_msg, user_msg],
            model=CHAT_MODEL,
            # Controls randomness
            temperature=0.5,
            # If set, partial message deltas will be sent.
            stream=True,
        )


