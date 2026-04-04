import os

from dotenv import load_dotenv
from groq import Groq

from llm.chat import ChatLLM
from vectordb.qdrant import get_qdrant_client, search_ahadith

TOP_K = 10


def input_loop():
    """Run an interactive CLI: server stub or RAG chat against Qdrant + Groq."""
    load_dotenv()
    requester = ChatLLM(Groq(api_key=os.getenv("GROQ_API_KEY")))

    flag = input("Start server (0) or start chatting (1)? ")
    if flag == "0":
        pass  # todo
    elif flag == "1":
        client = get_qdrant_client()
        try:
            while True:
                user_prompt = input("Enter your prompt: ")
                results = search_ahadith(client, user_prompt, top_k=TOP_K)
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
