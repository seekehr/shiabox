# shiabox
Mistral-powered search engine for Shi'a ahadith, utilising a RAG architecture with Qdrant for the vector database and mxbai-embed-large for embeddings.

# Preview
![alt text](https://github.com/seekehr/shiabox/blob/main/server/assets/images/readme_preview_1.png "Example 1")

# How does it work?
#### `setup.go`

For now, the project only has a small part of AlKafi. What I did was take the PDF, convert it into .txt (all the formatting that we don't need is thus removed, and we can deal it with just raw content). After that, we must **CHUNK** the book, which means dividing it into small parts so we can feed it into the vector database later. The more isolated a chunk is (meaning, let's say, one hadith per chunk, rather than multiple ahadith or one page), the more **accurate** searching the vector database will be and the easier it will be to provide sources to the LLM (and also, the less tokens will be used; many benefits to having small, isolated chunks; e.g per hadith). These parsed books are stored in `assets/parsed_books`.

After it has been chunked, storing the raw text into the vector database is TECHNICALLY possible, but not recommended at all as vectors allow searching with semantic meanings and are basically able to represent information much better. Searching (which is what the main point of the program is) would be very ineffective if we stored the raw text. As such, we use Ollama's open-source `mxbai-embed-large` and 

After we have chunked it properly, we can feed the data to our **Qdrant** vector database, in the collection `shiabox`. It will store each chunk in a separate `point` (Qdrant open-source is single-node, meaning a single machine contains all the points). The main `data` that will be stored in the point is the hadith content, which includes the sanad (chain of narrators) and the matn (content). It would be ideal if I could separate the sanad from the matn but I have not achieved a scalable way of doing it yet. Other important data (used by the LLM to provide the source) including `Hadith` (number of hadith), `Page`, and `Book` will be provided in the point `payload`. The id is basically `Book+Hadith` hashed using `md5`, e.g `AlKafi1030`, as it guarantees **uniqueness** and also could allow searching the vector db for a hadith in the future (i.e: Tell me what hadith 300 of AlKafi says).

#### `cli.go`

This is the main access point to the program for now (`server.go` will handle all the API stuff). 
