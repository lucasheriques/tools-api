package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"tools.lucasfaria.dev/internal/convert"
	"tools.lucasfaria.dev/internal/generate"
	"tools.lucasfaria.dev/internal/validator"
)

func (app *application) createFakeInvoice(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	now := time.Now()
	qs := r.URL.Query()
	paymentMethods := app.readCSV(qs, "paymentMethods", []string{"ach"})
	vendorName := app.readString(qs, "vendorName", "")
	accountNumber := app.readInt64(qs, "accountNumber", app.getRandomAccountNumber(), v)
	numberOfItems := app.readInt(qs, "numberOfItems", rand.Intn(8)+1, v)
	invoiceDate := app.readDate(qs, "createdAt", now, v)
	dueDate := app.readDate(qs, "dueAt", now.AddDate(0, 0, 30), v)

	v.Check(validator.PermittedValues(paymentMethods, []string{"ach", "check", "wire"}), "paymentMethods", "must be list of ['ach', 'check', 'wire']")
	v.Check(numberOfItems >= 1 && numberOfItems <= 20, "numberOfItems", "must be between 1 and 20")
	v.Check(invoiceDate.Before(dueDate), "invoiceDate", "must be before dueDate")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	htmlContent, err := generate.GenerateHtmlFile(&generate.GenerateInvoiceOptions{
		PaymentMethods: paymentMethods,
		VendorName:     vendorName,
		AccountNumber:  accountNumber,
		NumberOfItems:  numberOfItems,
		InvoiceDate:    invoiceDate.Format("January 2, 2006"),
		DueDate:        dueDate.Format("January 2, 2006"),
	})
	if err != nil {
		app.logger.Error(fmt.Sprintf("Failed to create index.html file: %v", err))
		http.Error(w, fmt.Sprintf("Failed to create index.html file: %v", err), http.StatusInternalServerError)
		return
	}

	app.logger.Info("Converting HTML to PDF v2...")
	pdfContent, err := convert.HtmlToPdfV2(htmlContent)
	if err != nil {
		http.Error(w, "Failed to convert HTML to PDF", http.StatusInternalServerError)
		app.logger.Error(fmt.Sprintf("Error converting HTML to PDF: %v", err))
		return
	}

	app.logger.Info("Sending PDF content to client...")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Size", fmt.Sprintf("%d", len(pdfContent)))
	if _, err := io.Copy(w, bytes.NewReader(pdfContent)); err != nil {
		app.logger.Error(fmt.Sprintf("Error sending PDF content to client: %v", err))
		http.Error(w, "Failed to send PDF content to client", http.StatusInternalServerError)
		return
	}

	app.logger.Info("Successfully converted HTML to PDF and sent to client")
}
