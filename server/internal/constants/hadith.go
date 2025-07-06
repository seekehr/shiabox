package constants

type Chunk struct {
	Book    string `json:"Book"`
	Page    int    `json:"Page"`
	Content string `json:"Content"`
}

type HadithChunk struct {
	Chunk
	Hadith int `json:"Hadith"`
}

type HadithEmbedding struct {
	Hadith    int       `json:"Hadith"`
	Embedding []float32 `json:"Embedding"`
	Book      string    `json:"Book"`
	Page      int       `json:"Page"`
	Content   string    `json:"Content"`
}

type HadithEmbeddingResponse struct {
	HadithEmbedding
	Score float32 `json:"Score"`
}
