import os

from dotenv import load_dotenv
from groq import Groq

from llm.requests import LLMRequester
from llm.embedding import embeddings

def input_loop():
    load_dotenv()
    requester = LLMRequester(Groq(api_key=os.getenv("GROQ_API_KEY")))

    flag = input("Start server (0) or start chatting (1)? ")
    if flag == "0":
        pass #todo
    elif flag == "1":
        while True:
            user_prompt = input("Enter your prompt: ")
            stream = requester.prompt(user_prompt)
            for chunk in stream:
                content = chunk.choices[0].delta.content
                if content:
                    print(content, end="", flush=True)
            print()
    else:
        print("Unexpected input.")



if __name__ == '__main__':
    input_loop()
