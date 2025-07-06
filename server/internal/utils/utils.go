package utils

import (
	"bytes"
	"net/http"
	"os"
)

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

func SaveDataToDisk(data string) {
	os.MkdirAll("assets/logs", 0755)
	os.WriteFile("assets/logs/data.txt", []byte(data), 0644)
}
