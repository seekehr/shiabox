package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

type Header struct {
	Key   string
	Value string
}

func MakePostRequest(url string, data *bytes.Reader, reuseClient *http.Client) (*http.Response, error) {
	if reuseClient == nil {
		resp, err := http.Post(url, "application/json", data)
		return resp, err
	} else {
		req, err := http.NewRequest("POST", url, data)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := reuseClient.Do(req)
		return resp, err
	}
}

// why doesn't http.Post have an option for headers? le dummys

// MakeHeadersRequest - Improve the stupid http.Post/http.Get (imo uwu) format. Does not close body.
func MakeHeadersRequest(url string, body io.Reader, client *http.Client, headers ...Header) (*http.Response, error) {
	if client == nil {
		return nil, fmt.Errorf("nil http client")
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	for _, header := range headers {
		req.Header.Set(header.Key, header.Value)
	}

	return client.Do(req)
}

func ReadTextFromFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ReadFileBuffered - For memory usage as we open many files at once
func ReadFileBuffered(path string) <-chan string {
	dataStream := make(chan string)
	go func() {
		file, err := os.Open(path)
		if err != nil {
			fmt.Println("Error opening file: ", err)
			return
		}

		defer func() {
			file.Close()
			close(dataStream)
		}()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			dataStream <- line
		}
	}()

	return dataStream
}

// ReadFileInChunks - Get a channel to chunkSize portions of your file at a time. More memory efficient + overlap support to avoid cut-offs.
func ReadFileInChunks(file *os.File, chunkSize int, overlapSize int64) (<-chan string, error) {
	out := make(chan string)

	go func() {
		defer close(out)
		defer file.Close()

		buf := make([]byte, chunkSize)
		offset := int64(0)
		for {
			n, err := file.ReadAt(buf, offset) // not doing offset - overlapSize here cuz maybe offset is 0. no worries i do it later
			if n > 0 {
				out <- "<OVERLAP_START>" + string(buf[:n]) // flag that no overlap will be provided for this message
			} else {
				fmt.Println("0 bytes read.")
			}

			if err == io.EOF { // EOF = end of file XDDDD
				out <- "<END>"
				break
			}

			if err != nil {
				fmt.Printf("error reading chunk: %v\n", err)
				break
			}
			offset += int64(chunkSize)
			offset -= overlapSize // so it starts reading from overlapSize earlier.
		}
	}()

	return out, nil
}

// SavePDFAsTxt - Used primarily by setup.go to save our pdf books in zero-formatting .txt format. Require's poppler's pdftotext
func SavePDFAsTxt(pdfPath string, txtPath string) error {
	err := exec.Command("pdftotext", "-enc", "ASCII7", pdfPath, txtPath).Run()
	if err != nil {
		return err
	}

	// we can't just write to the file as it is being read CONCURRENTLY, may lead to race conditions so we use another temporary file
	tempPath := txtPath + ".tmp"
	tmpFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	writer := bufio.NewWriter(tmpFile)
	// for memory usage

	previousLineEmpty := false
	dataStream := ReadFileBuffered(txtPath)
	for line := range dataStream {
		// technically this ensures that we receive new content before appending but theres still issues OK IDK ALLAHU AALAM
		transformLine := CleanText(line) // remove our arabic from the line
		// this is used because we dont want 2 consecutive empty lines; only one for formatting purpose
		if transformLine == "" {
			if previousLineEmpty == true {
				continue
			}
			previousLineEmpty = true
		} else {
			previousLineEmpty = false
		}

		_, err := writer.Write([]byte(transformLine + "\n"))
		if err != nil {
			return fmt.Errorf("failed to write to temp file: %w", err)
		}
	}

	// we flush so to ensure that all data is appended to disk and not saved in buffer or smth cuz we're immediately saving
	if err = writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush temp file: %w", err)
	}

	// finally, rename the .tmp file to the actual file as we don't need to use it anymore
	return os.Rename(tempPath, txtPath)
}

// CleanText removes garbage lines from a given string. AI-GENERATED FUNCTION SO DONT EXPECT ME TO MAINTAIN IT >:(
func CleanText(text string) string {
	var cleanedLines []string
	// commonly appeared bad chars in the prompt i gave to gemini i guess
	badChars := map[rune]bool{
		'G': true, 'H': true, 'E': true, 'L': true, 'Z': true, '%': true,
		'_': true, '7': true, 'q': true, 'd': true, 'C': true, 'I': true,
		':': true, '$': true, '"': true, '(': true, ')': true, '{': true,
		'}': true, '~': true, 'Y': true, '#': true, '!': true,
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		// keep chapter lines i guess as theyre used by LLM for chunking
		if isPurelyNumeric(trimmedLine) {
			cleanedLines = append(cleanedLines, line)
			continue
		}

		// keep numbered paragraphs like "14. And..."
		parts := strings.SplitN(trimmedLine, ".", 2)
		if len(parts) > 1 && isPurelyNumeric(parts[0]) {
			cleanedLines = append(cleanedLines, line)
			continue
		}

		//i dont like runes.
		runes := []rune(trimmedLine)
		if len(runes) == 0 {
			continue
		}

		// stores how many characters are L mans so we can remove them
		badCount := 0
		alphaCount := 0
		hasSpace := false
		for _, r := range runes {
			if _, isBad := badChars[r]; isBad {
				badCount++
			}
			if unicode.IsLetter(r) {
				alphaCount++
			}
			if unicode.IsSpace(r) {
				hasSpace = true
			}
		}

		//rRemove very short lines with less numbers
		if len(runes) < 10 && hasSpace && alphaCount < 4 {
			continue
		}

		// high ratio of bad characters
		if float64(badCount)/float64(len(runes)) > 0.4 {
			continue
		}

		// low ratio of letters (idk why this exists ok i didnt tell it to add this)
		if len(runes) > 1 && float64(alphaCount)/float64(len(runes)) < 0.5 {
			continue
		}

		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, "\n")
}

// isPurelyNumeric Check if a string only has numbers (used by CleanText)
func isPurelyNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func ChunkString(input string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(input); i += chunkSize {
		end := i + chunkSize
		if end > len(input) {
			end = len(input)
		}
		chunks = append(chunks, input[i:end])
	}
	return chunks
}

func SaveDataToLogs(data string) {
	os.MkdirAll("assets/logs", 0755)
	os.WriteFile("assets/logs/data.txt", []byte(data), 0644)
}
