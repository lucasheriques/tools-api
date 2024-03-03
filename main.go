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

func convertHTMLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		log.Println("Method not allowed")
		return
	}

	htmlContent, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read HTML content", http.StatusInternalServerError)
		log.Printf("Failed to read HTML content: %v", err)
		return
	}
	defer r.Body.Close()

	log.Println("Converting HTML to PDF...")
	pdfContent, err := convert.ConvertHtmlStringToPdf(htmlContent)
	if err != nil {
		http.Error(w, "Failed to convert HTML to PDF", http.StatusInternalServerError)
		log.Printf("Error converting HTML to PDF: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	if _, err := io.Copy(w, bytes.NewReader(pdfContent)); err != nil {
		log.Printf("Error sending PDF content to client: %v", err)
		http.Error(w, "Failed to send PDF content to client", http.StatusInternalServerError)
		return
	}

	log.Println("Successfully converted HTML to PDF and sent to client")
}
