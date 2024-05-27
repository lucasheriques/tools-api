package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"tools.lucasfaria.dev/internal/convert"
	"tools.lucasfaria.dev/internal/invoices"
)

func (app *application) createFakeInvoice(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	paymentMethod := app.readString(qs, "paymentMethod", "ach")
	vendorName := app.readString(qs, "vendorName", "Acme Corp.")
	accountNumber := app.readString(qs, "accountNumber", invoices.GenerateAccountNumber())

	htmlContent, err := invoices.GenerateHtmlFile(invoices.GenerateInvoiceOptions{
		PaymentMethod: paymentMethod,
		VendorName:    vendorName,
		AccountNumber: accountNumber,
	})
	if err != nil {
		app.logger.Error(fmt.Sprintf("Failed to create index.html file: %v", err))
		http.Error(w, fmt.Sprintf("Failed to create index.html file: %v", err), http.StatusInternalServerError)
		return
	}

	app.logger.Info("Converting HTML to PDF...")
	pdfContent, err := convert.ConvertHtmlStringToPdf(htmlContent)
	if err != nil {
		http.Error(w, "Failed to convert HTML to PDF", http.StatusInternalServerError)
		app.logger.Error(fmt.Sprintf("Error converting HTML to PDF: %v", err))
		return
	}

	app.logger.Info("Sending PDF content to client...")
	w.Header().Set("Content-Type", "application/pdf")
	if _, err := io.Copy(w, bytes.NewReader(pdfContent)); err != nil {
		app.logger.Error(fmt.Sprintf("Error sending PDF content to client: %v", err))
		http.Error(w, "Failed to send PDF content to client", http.StatusInternalServerError)
		return
	}

	app.logger.Info("Successfully converted HTML to PDF and sent to client")

}
