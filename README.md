# shiabox
Mistral-powered search engine for Shi'a ahadith, utilising a RAG architecture with Qdrant for the vector database and mxbai-embed-large for embeddings.

# How does it work?
#### `setup.go`

For now, the project only has a small part of AlKafi. What I did was take the PDF, convert it into .txt (all the formatting that we don't need is thus removed, and we can deal it with just raw content). After that, we must **CHUNK** the book, which means dividing it into small parts so we can feed it into the vector database later. The more isolated a chunk is (meaning, let's say, one hadith per chunk, rather than multiple ahadith or one page), the more **accurate** searching the vector database will be and the easier it will be to provide sources to the LLM (and also, the less tokens will be used; many benefits to having small, isolated chunks; e.g per hadith). These parsed books are stored in `assets/parsed_books` as JSON.

After it has been chunked, storing the raw text into the vector database is TECHNICALLY possible, but **not recommended** as vectors allow searching with **semantic** meanings and are thus basically able to represent information much better (in our case). Searching (which is what the main point of the program is) would be ineffective if we stored the raw text **for our use-case**. As such, we use Ollama's open-source `mxbai-embed-large` to convert each chunk's content (and only content, as the rest of the data is just saved as payload and not searched for similarities) into a bunch of vectors (e.g `(1.23232, 1.249999, 1.232111)` and so on). The embedded books are then stored in `assets/embeddings`, with the only difference from parsed_books being the embeddings data field being added.

After we have chunked it properly, we can feed the data to our **Qdrant** vector database, in the collection `shiabox`. It will store each chunk in a separate `point` (Qdrant open-source is single-node, meaning a single machine contains all the points). The main `data` that will be stored in the point is the hadith content, which includes the sanad (chain of narrators) and the matn (content). It would be ideal if I could separate the sanad from the matn but I have not achieved a scalable way of doing it yet unfortunately. Other important data (used by the LLM to provide the source) including `Hadith` (number of hadith), `Page`, and `Book` will be provided in the point `payload`. The id is basically `Book+Hadith` hashed using `md5`, e.g `AlKafi1030`, as it guarantees **uniqueness** and also could MAYBE allow searching the vector db for a hadith in the future using filters (i.e: Tell me what hadith 300 of AlKafi says).

#### `cli.go`

This is the main access point to the program for now (`server.go` will handle all the API stuff). When you enter your prompt, the prompt is first **converted into vectors**/**embedded** so we can search them with Qdrant easily (searching vectors (i.e `data`) using text would practically not go well at all). After that, the vector db is searched (using the `cosine` algorithm) and the top **10** results are returned with their `Score` (that determines similarity/relevance). 

These results are then fed into `Mistral`, with a prompt that is defined in `assets/prompt.txt`. As such, Mistral is provided the input text (as it cannot **convert** the vectors into text or text into vectors (it can do this but not as good as a dedicated embedding model like the one we use); it is **not specifically** trained for that, and so we **search using the vectors** (faster and more accurate) and THEN we take the raw text from the `top-k` results (from Qdrant) to feed into the Mistral LLM **for sorting/a more human-friendly answer**. Mistral then parses the given data according to our prompt and returns the answer. The benefits of using Mistral are that it can also look at the context and provide fallback results if our data does not contain the information asked for. There is a one-liner to prevent hallucinations but it is still a possibility. 

**Note:** Currently, EVERY hadith (sahih, da'if, hasan, or whatever else) is treated with the same level of accuracy. Implementing a `grading` for every ahadith, especially in a scalable/automated manner, will be very difficult and so is a "NOT COMING SOON" feature for now, 

# Installation Guide
Read [INSTALLATION GUIDE](INSTALLING.md).
# Preview
![alt text](https://github.com/seekehr/shiabox/blob/main/server/assets/images/readme_preview_1.png "Example 1")
