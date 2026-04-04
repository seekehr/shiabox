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

    def prompt(self, user_prompt: str, search_results: list[dict] | None = None) -> ChatCompletion | Stream[ChatCompletionChunk]:
        if search_results:
            system_content = self._build_rag_prompt(user_prompt, search_results)
        else:
            system_content = self._chat_prompt

        system_msg: ChatCompletionSystemMessageParam = {"role": "system", "content": system_content}
        user_msg: ChatCompletionUserMessageParam = {"role": "user", "content": user_prompt}
        return self._client.chat.completions.create(
            messages=[system_msg, user_msg],
            model=CHAT_MODEL,
            temperature=0.5,
            stream=True,
        )

    def _build_rag_prompt(self, question: str, results: list[dict]) -> str:
        candidates = []
        for i, r in enumerate(results, 1):
            candidates.append(
                f"Hadith {i} (Score: {r['score']:.4f})\n"
                f"Source: Book {r['book']}, Chapter {r['chapter']}, "
                f"Hadith #{r['hadith_number']}\n"
                f"{r['content']}"
            )
        candidates_text = "\n\n".join(candidates)
        return self._chat_prompt.replace("{InputText}", question) + "\n" + candidates_text


