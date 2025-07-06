package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"server/internal/constants"
	"server/internal/utils"
	"strconv"
	"sync"
	"time"
)

const (
	url         = "http://localhost:11434/api/embed"
	model       = "mxbai-embed-large"
	batchSize   = 30
	workerCount = 10
)

type TextRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type BatchRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type response struct {
	Embedding [][]float32 `json:"embeddings"`
}

type batchResponse struct {
	Embedding [][]float32 `json:"embeddings"`
}

func EmbedText(text string, reuseClient *http.Client) ([]float32, error) {
	reqBody := TextRequest{
		Model: model,
		Input: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	resp, err := utils.MakePostRequest(url, bytes.NewReader(jsonData), reuseClient)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var embeddings response
	json.NewDecoder(resp.Body).Decode(&embeddings)
	if embeddings.Embedding == nil {
		return nil, fmt.Errorf("empty embedding")
	}
	return embeddings.Embedding[0], err
}

// EmbedBatch returns an array of ahadith, that have an array of embeddings
func EmbedBatch(contents []string, reuseClient *http.Client) ([][]float32, error) {
	reqBody := BatchRequest{
		Model: model,
		Input: contents,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	resp, err := utils.MakePostRequest(url, bytes.NewReader(jsonData), reuseClient)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var embeddingsBatch batchResponse
	json.NewDecoder(resp.Body).Decode(&embeddingsBatch)
	if len(embeddingsBatch.Embedding) < 2 {
		return nil, fmt.Errorf("empty embedding")
	}
	return embeddingsBatch.Embedding, err
}

func EmbedAhadith(chunk []constants.HadithChunk, reuseClient *http.Client) ([]constants.HadithEmbedding, error) {
	contents := make([]string, len(chunk))
	for i, c := range chunk {
		contents[i] = c.Content
	}
	embeddings, err := EmbedBatch(contents, reuseClient)
	if err != nil {
		return nil, err
	}

	if len(embeddings) != len(chunk) {
		return nil, fmt.Errorf("length mismatch during loop (" + strconv.Itoa(len(embeddings)) + " != " + strconv.Itoa(len(chunk)) + ")")
	}

	hadithAsEmbedded := make([]constants.HadithEmbedding, len(embeddings))
	for i, embed := range embeddings {
		if i >= len(chunk) {
			return nil, fmt.Errorf("length mismatch during loop (" + strconv.Itoa(i) + " >= " + strconv.Itoa(len(chunk)) + ")")
		}
		hadithAsEmbedded[i] = constants.HadithEmbedding{
			Hadith:    chunk[i].Hadith,
			Embedding: embed,
			Book:      chunk[i].Book,
			Page:      chunk[i].Page,
			Content:   chunk[i].Content,
		}
	}
	return hadithAsEmbedded, nil
}

func EmbedBook(path string, bookName string) error {
	timer := time.Now()
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var chunks []constants.HadithChunk
	err = json.Unmarshal(content, &chunks)
	if err != nil {
		return err
	}

	totalChunks := len(chunks)
	strLenChunks := strconv.Itoa(totalChunks)
	fmt.Println("Reached embedding phase of " + path + " in " + time.Since(timer).String() + ". Total chunks: " + strLenChunks)
	timer = time.Now()

	var (
		wg          sync.WaitGroup
		reuseClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	)
	workerChan := make(chan struct{}, workerCount) // need to enforce this to prevent too many http reqs
	embeddedAhadith := make([]constants.HadithEmbedding, totalChunks)

	for i := 0; i < totalChunks; i += batchSize {
		end := i + batchSize
		if end > totalChunks {
			end = totalChunks
		}
		batch := chunks[i:end]

		workerChan <- struct{}{} // an empty struct so we dont waste memory with an int or smth
		wg.Add(1)
		go func(start int, chunks []constants.HadithChunk) {
			defer func() { <-workerChan }() // remove the value from the channel as the thread is paused until the channel is unblocked
			defer wg.Done()                 // i luv defer
			embedding, err := EmbedAhadith(chunks, reuseClient)
			if err != nil {
				fmt.Println("Error embedding batch: " + err.Error())
				return
			}
			for offset, embeddedHadith := range embedding {
				// start+offset to calculate the actual value of the hadith in `chunks`
				// no mutex needed since we're targeting different values of the slice
				embeddedAhadith[start+offset] = embeddedHadith
			}
		}(i, batch)
	}

	wg.Wait()
	jsonData, err := json.Marshal(embeddedAhadith)
	if err != nil {
		return err
	}

	err = os.WriteFile(constants.EmbeddingsDir+bookName, jsonData, 0644)
	fmt.Println("Finished reading the entire book in " + time.Since(timer).String())
	return err
}

func ReadEmbeddedBook(embeddedBookPath string) ([]constants.HadithEmbedding, error) {
	file, err := os.Open(embeddedBookPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var chunks []constants.HadithEmbedding
	err = json.Unmarshal(content, &chunks)
	if err != nil {
		return nil, err
	}
	return chunks, nil
}
