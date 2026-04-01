import time

from dotenv import load_dotenv

from src.llms.groq_llm import GroqLLM
from src.vector_db import VectorDB


def main() -> None:
    load_dotenv()

    vector_db = VectorDB.connect()
    groq = GroqLLM.create(vector_db)

    while True:
        try:
            prompt = input("Enter your prompt: ")
        except (EOFError, KeyboardInterrupt):
            print()
            break

        if not prompt.strip():
            continue

        try:
            stream = groq.handle_chat_request(prompt)
        except Exception as e:
            print(f"Error handling request: {e}")
            continue

        timer = time.time()
        print("\nModel Response: ", end="", flush=True)
        for chunk in stream:
            if chunk.choices:
                delta = chunk.choices[0].delta
                if delta and delta.content:
                    print(delta.content, end="", flush=True)
        print()
        print(f"Done in {time.time() - timer:.3f}s.")


if __name__ == "__main__":
    main()
