package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"tools-api/convert"
	"tools-api/invoices"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/fake-invoice", invoiceHandler)
	mux.HandleFunc("POST /api/convert/html", convertHTMLHandler)

	fmt.Println("Starting tools-api server at port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func invoiceHandler(w http.ResponseWriter, r *http.Request) {
	// Read query parameters
	paymentMethod := r.FormValue("paymentMethod")

	htmlContent, err := invoices.GenerateHtmlFile(invoices.GenerateInvoiceOptions{
		PaymentMethod: paymentMethod,
	})

	if err != nil {
		log.Printf("Failed to create index.html file: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create index.html file: %v", err), http.StatusInternalServerError)
		return
	}

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

func convertHTMLHandler(w http.ResponseWriter, r *http.Request) {
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
