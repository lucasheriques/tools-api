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

const gotenbergURL = "http://gotenberg.default.svc.cluster.local:80/forms/chromium/convert/html"

func ConvertHtmlStringToPdf(htmlContent []byte) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "gotenberg")
	if err != nil {
		return nil, fmt.Errorf("failed to create a temporary directory: %v", err)
	}

	tempFilePath := filepath.Join(tempDir, "index.html")
	err = os.WriteFile(tempFilePath, htmlContent, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write HTML content to index.html: %v", err)
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fw, err := w.CreateFormFile("files[index.html]", "index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}
	htmlFile, err := os.Open(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open index.html file: %v", err)
	}
	defer htmlFile.Close()

	if _, err = io.Copy(fw, htmlFile); err != nil {
		return nil, fmt.Errorf("failed to copy index.html file to form file: %v", err)
	}
	w.Close()

	req, err := http.NewRequest("POST", gotenbergURL, &b)
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg responded with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	pdfContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return pdfContent, nil
}
