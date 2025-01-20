package services

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type RequestService struct{}

var requestServerInstance *RequestService = nil

// Singleton instance of RequestService
func GetRequestService() (*RequestService, error) {
	if requestServerInstance == nil {
		requestServerInstance = &RequestService{}
	}

	return requestServerInstance, nil
}

func (rs *RequestService) Post(url string, payload string, headers map[string]string) (string, map[string]string, error) {
	// Convert payload to JSON
	// jsonData, err := json.Marshal(payload)
	// if err != nil {
	// 	fmt.Println("Error marshalling JSON:", err)
	// 	return "", err
	// }

	jsonData := payload

	// Create a new POST request
	req, err := http.NewRequest(headers["_ METHOD"], url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", nil, err
	}

	// Add headers
	for key, value := range headers {
		if strings.HasPrefix(key, "_ ") {
			continue
		}
		req.Header.Set(key, value)
	}

	if _, ok := headers["Accept-Encoding"]; !ok {
		req.Header.Set("Accept-Encoding", "gzip")
	}

	// Send the request
	client := &http.Client{}
	client.Timeout = time.Duration(5) * time.Second
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", nil, err
	}
	defer resp.Body.Close()

	// Check if the response is compressed
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Println("Error creating gzip reader:", err)
			return "", nil, err
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	// Read the response body
	body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", nil, err
	}

	respHeaders := make(map[string]string)
	for key, value := range resp.Header {
		respHeaders[key] = strings.Join(value, ",")
	}

	return string(body), respHeaders, nil
}
