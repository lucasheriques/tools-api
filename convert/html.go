package convert

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const gotenbergURL = "http://gotenberg:3000/forms/chromium/convert/html"

func ConvertHtmlStringToPdf(htmlContent []byte) ([]byte, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gotenberg")
	if err != nil {
		return nil, fmt.Errorf("failed to create a temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up the directory afterwards

	// Create an index.html file inside the temporary directory
	tempFilePath := filepath.Join(tempDir, "index.html")
	err = os.WriteFile(tempFilePath, []byte(htmlContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write HTML content to index.html: %w", err)
	}

	// Create a new multipart writer
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add the index.html file to the multipart writer
	fw, err := w.CreateFormFile("files[index.html]", "index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	htmlFile, err := os.Open(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open index.html file: %w", err)
	}
	defer htmlFile.Close()
	_, err = io.Copy(fw, htmlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to copy index.html file to form file: %w", err)
	}

	// Close the multipart writer
	w.Close()

	// Create a new HTTP request
	req, err := http.NewRequest("POST", gotenbergURL, &b)
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request responded with not ok: %s", resp.Status)
	}

	// Read the response body
	pdfContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return pdfContent, nil
}
