## STEP 1

So, basically. The first requirement is to extract just the text content (without formatting) from the PDFS.
Since we will rely on a vector database for accuracy, the second requirement
is to parse this text into `chunks` that can be stored in our `collection` (one collection)
only for now, of the vector database (qdrant).

## STEP 2

Now, we are not going to store the text directly because it wouldn't be very LLM-friendly.
So instead, we are going to generate EMBEDDINGS (float values) of the ahadith content so
information like semantics can be better represented, and this will be done through
another machine learning algorithm (TY ollama <3).

## STEP 3

When a user enters a prompt, e.g: "Who was Musa ibn Ja'far?", this input will be
sent to the local embedding LLM through an HTTP request, and a `[]float32` will be
returned (containing the embeddings). This embedded input will THEN be searched by
qdrant, the vector database, and the top similar chunks will be returned.

## STEP 4

The input (embedded) + the chunks returned by qdrant (also embedded) will then be
sent to the LLM (Mistral). Mistral will then handle sorting the chunks and basically
generating a human-readable/friendly statement that will return the semantics of the
best/most similar chunk(s).