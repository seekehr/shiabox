package llms

// Parse responses returned/provided from/to groq_controller.go

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	StopFinishReason   FinishReasonType = "stop"
	LengthFinishReason FinishReasonType = "length"
	FilterFinishReason FinishReasonType = "content_filter"
)

type FinishReasonType string

// groqCompleteAIChoice - Returned in choices{} JSON object by Groq, when stream is false
type groqCompleteAIChoice struct {
	Index        int              `json:"index"`
	FinishReason FinishReasonType `json:"finish_reason"`
	Message      AIMessage        `json:"message"`
}

// groqStreamedAIChoice - Returned in choices{} JSON object by Groq, when stream is true
type groqStreamedAIChoice struct {
	Index        int              `json:"index"`
	FinishReason FinishReasonType `json:"finish_reason"`
	Delta        AIMessage        `json:"delta"`
}

// GroqCompleteAIResponse - API Response by Groq, when stream is false
type GroqCompleteAIResponse struct {
	Choices []groqCompleteAIChoice `json:"choices"`
}

// GroqStreamedAIResponse - API Response by Groq, when stream is true
type GroqStreamedAIResponse struct {
	Choices []groqStreamedAIChoice `json:"choices"`
}

type GroqParser struct{}

// ParseResponse - Parse the unstreamed response from the Groq API. ngl i dont like spamming pointers to structs but it's ideal for `nil` ig
func (GroqParser) ParseResponse(body io.ReadCloser, err error) (*GroqCompleteAIResponse, error) {
	if err != nil {
		return nil, err // top 10 worst things ive done
	}

	var response GroqCompleteAIResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ParseStreamedSSE - Allow SSE token streaming from the API. <-chan returns a read-only channel
func (GroqParser) ParseStreamedSSE(body io.ReadCloser) <-chan *GroqStreamedAIResponse {
	dataChan := make(chan *GroqStreamedAIResponse)

	go func() {
		defer func() {
			body.Close()
			close(dataChan)
		}()

		reader := bufio.NewReader(body)
		var buf bytes.Buffer

		for {
			// Read up to the next newline
			line, err := reader.ReadString('\n')
			if err != nil {
				// EOF or ntwk error ends the stream
				if !errors.Is(err, io.EOF) {
					fmt.Println("read error:", err)
				}
				return
			}

			// delimeter reached
			if line == "\n" || line == "\r\n" {
				event := buf.String()
				buf.Reset()

				if strings.HasPrefix(event, "data: ") {
					payload := strings.TrimPrefix(event, "data: ")
					payload = strings.TrimSpace(payload) // strip trailing \r
					// Groq ends the stream with: data: [DONE]
					if payload == "[DONE]" {
						return
					}

					var response GroqStreamedAIResponse
					if err := json.Unmarshal([]byte(payload), &response); err != nil {
						fmt.Println("unmarshal error:", err)
						continue
					}

					// create a copy of the response struct and send the pointer to the copy. (for performance). explanation below
					// explanation: channel send only moves the pointer which is cheaper than moving the whole ass struct
					respCopy := response
					dataChan <- &respCopy
				}
				// ignore other SSE fields like "event:" or "retry:"; not mentioned in the api doc
				continue
			}

			// Not a blank line yet so keep buffering
			buf.WriteString(line)
		}
	}()

	return dataChan
}

// Unused for now. Was used for Mistral API which sent constant JSON responses instead of text/event-stream
/*func ParseStreamedResponse(body io.ReadCloser) chan string {
	dataChan := make(chan string)
	go func() {
		defer func() {
			// this is deffered because our func returns IMMEDIATELY, which would close the body b4 we finished reading,
			// whereas the goroutine keeps running even if our function returns until its task is complete
			body.Close() // i got a bit confused, so im writing dis so i remember: interfaces are technically copied-by-value
			// but body.Close() would still work because basically the {pointerToTypeInformation, pointerToData) of
			// an interface is copied not the entire value (unlike structs)
			close(dataChan)
		}()
		var chunk AIResponse

		scanner := bufio.NewScanner(body)
		scanner.Buffer(make([]byte, 0, 64*1024), bufferSize)
		for scanner.Scan() {
			line := scanner.Text()
			if err := json.Unmarshal([]byte(line), &chunk); err == nil {
				dataChan <- chunk.Response
			} else {
				fmt.Println("Error reading response from LLM: ", err.Error())
			}
		}
	}()
	return dataChan
}*/
