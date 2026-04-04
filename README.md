# shiabox
AI-powered (groq) search engine for Shi'a ahadith, utilising a RAG architecture with Qdrant for the vector database, `Gemini` for chunking text into ahadith and `qwen3-embedding:4b` for embeddings.

# How does it work?

First of all, we must understand what RAG is. RAG basically allows you to provide up-to-date, or selective information to the AI, and then generate a response based on that. The information, in this case, are the **ahadith** that we obtain.

### How do we obtain the ahadith?

First, we download the .pdf obtained from the internet, and convert it using `pdftotext` by Popper's Utils with a special encoding to remove any Arabic text (as the models we use are not specialized in Arabic and it would be a hassle and waste of tokens to keep). Then, we use `server/assets/current_book_parser_prompt.txt` as the prompt, and ask the Gemini [TODO: REVIEW MODEL LATER] chunking model to chunk our book into ahadith and store it in a `.json` format. And so, we can easily use the ahadith for whatever we want.

### Step 2: Embedding the ahadith and saving them in our Qdrant database

Just raw .json is not enough. It'd be a massive waste of tokens, and a performance liability if we were to just feed the entire .json to the chatting AI model. Not only that, it would also be not scalable at all because more books would slow down the performance significantly.

So, what's the solution? We send the chatting AI the most relevant ahadith (**10**, in our case) and let it pick the **3** most relevant ones from those. For this to happen, we must first convert our ahadith into matrices that preserve tone and grammar, that AI can easily understand, and save them in an advanced 'mathematical innovation' points database.

### Step 3: Find top 10 ahadith from vector, feed to AI, stream back response

First of all, we embed (i.e convert into matrices, like earlier) the prompt the user sends using our embedding model. Then, we compare this embedded prompt to other matrices in the vector database using built-in functions of Qdrant. The top 10 most similar points are sent back to us, and we send these 10 ahadith to the AI with a special system prompt written in `server/assets/prompt.txt`.

# Installation Guide
Read [INSTALLATION GUIDE](INSTALLING.md).
# Preview
![alt text](https://github.com/seekehr/shiabox/blob/main/server/assets/images/readme_preview_1.png "Example 1")
