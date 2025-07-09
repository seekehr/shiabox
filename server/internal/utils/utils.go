package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
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
		transformLine := CleanNoise(line) // remove our arabic from the line
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

// CleanNoise - removes garbage characters thanks to the pdf conversion process. Sinful ChatGPT generated but idc about file encodings ;-;
func CleanNoise(input string) string {
	// Remove repeating glyph patterns
	garbagePattern := regexp.MustCompile(`(?:[GHELZ%_7qdCI:$"(){}]{2,}\s*){2,}`)
	input = garbagePattern.ReplaceAllString(input, "")

	// empty string if the line is mostly garbage
	if isMostlyGarbage(input) {
		return ""
	}

	// Remove non-printable characters
	var builder strings.Builder
	for _, r := range input {
		if unicode.IsPrint(r) && r != '\uFFFD' {
			builder.WriteRune(r)
		}
	}

	// Normalize spacing
	cleaned := strings.Join(strings.Fields(builder.String()), " ")

	// Final short + junk line check
	if len([]rune(cleaned)) < 15 && isMostlyGarbage(cleaned) {
		return ""
	}

	return cleaned
}

// isMostlyGarbage checks ratio of junk characters
func isMostlyGarbage(s string) bool {
	runes := []rune(s)
	if len(runes) == 0 {
		return true
	}

	badChars := "GHELZ%_7}{~\"$()Y#qdCI:"
	bad := 0
	for _, r := range runes {
		if strings.ContainsRune(badChars, r) {
			bad++
		}
	}

	return float64(bad)/float64(len(runes)) > 0.4
}

func SaveDataToLogs(data string) {
	os.MkdirAll("assets/logs", 0755)
	os.WriteFile("assets/logs/data.txt", []byte(data), 0644)
}
