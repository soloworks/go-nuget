package nuget

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"time"
)

// PushNupkg PUTs a .nupkg binary to a NuGet Repository
func PushNupkg(fileContents []byte, apiKey string, host string) (int, int64, error) {

	// If no Source provided, exit
	if host == "" {
		return 0, 0, errors.New("Error: Please specify a Source/Host")
	}

	// Create MultiPart Writer
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	// Create new File part
	p, err := w.CreateFormFile("package", "package.nupkg")
	if err != nil {
		return 0, 0, err
	}
	// Write contents to part
	_, err = p.Write(fileContents)
	if err != nil {
		return 0, 0, err
	}
	// Close the writer
	err = w.Close()
	if err != nil {
		return 0, 0, err
	}

	// Create new PUT request
	request, err := http.NewRequest(http.MethodPut, host, body)
	if err != nil {
		return 0, 0, err
	}
	// Add the ApiKey if supplied
	if apiKey != "" {
		request.Header.Add("X-Nuget-Apikey", apiKey)
	}
	// Add the Content Type header from the reader - includes boundary
	request.Header.Add("Content-Type", w.FormDataContentType())

	// Push to the server
	startTime := time.Now()
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return 0, 0, err
	}
	duration := time.Now().Sub(startTime)

	// Return Results
	return resp.StatusCode, duration.Milliseconds(), nil
}
