package constants

type Chunk struct {
	Book    string
	Page    int
	Content string
}

type HadithChunk struct {
	Chunk
	Hadith int
}
