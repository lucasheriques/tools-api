package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"tools-api/convert"
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

	pdfFile, err := convert.ConvertHtmlStringToPdf(htmlContent)
	if err != nil {
		http.Error(w, "Failed to convert HTML to PDF", http.StatusInternalServerError)
		return
	}

	// Copy the Gotenberg response (PDF file) to the client
	w.Header().Set("Content-Type", "application/pdf")
	if _, err := io.Copy(w, bytes.NewReader(pdfFile)); err != nil {
		http.Error(w, "Failed to send PDF content to client", http.StatusInternalServerError)
		return
	}

	// Log the request
	log.Printf("Converted HTML to PDF")

}
