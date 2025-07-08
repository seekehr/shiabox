package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

const (
	bufferSize = 2 * 1024 * 1024 // 2 mib
)

type responseChunk struct {
	Response string `json:"response"`
}

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
