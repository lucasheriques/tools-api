package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/convert/html", convertHTMLHandler)

	fmt.Println("Starting tools-api server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!!!!")
}

// Assuming Gotenberg is accessible at 'http://gotenberg:3000' from within the cluster
const gotenbergURL = "http://gotenberg:3000/forms/chromium/convert/html"

func convertHTMLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the HTML content from the request body
	htmlContent, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read HTML content", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gotenberg")
	if err != nil {
		http.Error(w, "Failed to create a temporary directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir) // Clean up the directory afterwards

	// Create an index.html file inside the temporary directory
	tempFilePath := filepath.Join(tempDir, "index.html")
	err = os.WriteFile(tempFilePath, htmlContent, 0644)
	if err != nil {
		http.Error(w, "Failed to write HTML content to index.html", http.StatusInternalServerError)
		return
	}

	// Prepare the multipart request
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Add index.html file to the request
	file, err := os.Open(tempFilePath)
	if err != nil {
		http.Error(w, "Failed to open index.html", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileWriter, err := multipartWriter.CreateFormFile("files", tempFilePath)
	if err != nil {
		http.Error(w, "Failed to add index.html to the request", http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		http.Error(w, "Failed to copy index.html content", http.StatusInternalServerError)
		return
	}

	err = multipartWriter.Close()
	if err != nil {
		http.Error(w, "Failed to close multipart writer", http.StatusInternalServerError)
		return
	}

	// Send the request to Gotenberg
	response, err := http.Post(gotenbergURL, multipartWriter.FormDataContentType(), &requestBody)
	if err != nil {
		http.Error(w, "Failed to send request to Gotenberg", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf("Gotenberg returned error: %s", response.Status)
		http.Error(w, "Gotenberg failed to convert HTML to PDF", response.StatusCode)
		return
	}

	// Copy the Gotenberg response (PDF file) to the client
	w.Header().Set("Content-Type", "application/pdf")
	if _, err := io.Copy(w, response.Body); err != nil {
		http.Error(w, "Failed to send PDF content to client", http.StatusInternalServerError)
		return
	}
}
