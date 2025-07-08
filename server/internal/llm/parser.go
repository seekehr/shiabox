package llm

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
	bufferSize = 2 * 1024 * 1024 // 2 mib
)

type responseChunk struct {
	Response string `json:"response"`
}

// ParseStreamedResponse - Deprecated
func ParseStreamedResponse(body io.ReadCloser) chan string {
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
		var chunk responseChunk

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
}

func ParseStreamedSSE(body io.ReadCloser) <-chan string {
	dataChan := make(chan string)

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

			// delimeter reaached
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

					var chunk responseChunk
					if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
						fmt.Println("unmarshal error:", err)
						continue
					}
					dataChan <- chunk.Response
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
