package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
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

// SavePDFAsTxt - Used primarily by setup.go to save our pdf books in zero-formatting .txt format. Require's poppler's pdftotext
func SavePDFAsTxt(pdfPath string, txtPath string) error {
	return exec.Command("pdftotext", pdfPath, txtPath).Run()
}

// RemoveArabic - Use regex to remove arabic unicode representations from text
func RemoveArabic(text string) string {
	re := regexp.MustCompile(`[\x{0600}-\x{06FF}\x{0750}-\x{077F}\x{08A0}-\x{08FF}\x{FB50}-\x{FDFF}\x{FE70}-\x{FEFF}]+`)
	return re.ReplaceAllString(text, "")
}

func SaveDataToLogs(data string) {
	os.MkdirAll("assets/logs", 0755)
	os.WriteFile("assets/logs/data.txt", []byte(data), 0644)
}
