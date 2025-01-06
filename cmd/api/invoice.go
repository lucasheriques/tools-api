package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"tools.lucasfaria.dev/internal/convert"
	"tools.lucasfaria.dev/internal/generate"
	"tools.lucasfaria.dev/internal/validator"
)

var validCurrencies = []string{
	"usd", // United States Dollar
	"eur", // Euro
	"jpy", // Japanese Yen
	"gbp", // British Pound Sterling
	"aud", // Australian Dollar
	"cad", // Canadian Dollar
	"chf", // Swiss Franc
	"cny", // Chinese Yuan
	"sek", // Swedish Krona
	"nzd", // New Zealand Dollar
	"mxn", // Mexican Peso
	"sgd", // Singapore Dollar
	"hkd", // Hong Kong Dollar
	"nok", // Norwegian Krone
	"krw", // South Korean Won
	"try", // Turkish Lira
	"rub", // Russian Ruble
	"inr", // Indian Rupee
	"brl", // Brazilian Real
	"zar", // South African Rand
}

func (app *application) createFakeInvoice(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	now := time.Now()
	qs := r.URL.Query()
	paymentMethods := app.readCSV(qs, "paymentMethods", []string{"ach"})
	for i, method := range paymentMethods {
		paymentMethods[i] = strings.ToLower(method)
	}
	vendorName := app.readString(qs, "vendorName", "")
	accountNumber := app.readInt64(qs, "accountNumber", app.getRandomAccountNumber(), v)
	numberOfItems := app.readInt(qs, "numberOfItems", rand.Intn(8)+1, v)
	invoiceDate := app.readDate(qs, "createdAt", now, v)
	dueDate := app.readDate(qs, "dueAt", now.AddDate(0, 0, 30), v)
	currency := strings.ToLower(app.readString(qs, "currency", "usd"))

	v.Check(validator.PermittedValues(paymentMethods, []string{"ach", "check", "wire"}), "paymentMethods", "must be list of ['ach', 'check', 'wire']")
	v.Check(numberOfItems >= 1 && numberOfItems <= 20, "numberOfItems", "must be between 1 and 20")
	v.Check(invoiceDate.Before(dueDate), "invoiceDate", "must be before dueDate")
	v.Check(validator.PermittedValue(currency, validCurrencies...), "currency", fmt.Sprintf("must be one of %v", validCurrencies))

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	app.logger.Info("Creating invoice with the following parameters: " +
		fmt.Sprintf("paymentMethods=%v, vendorName=%v, accountNumber=%v, numberOfItems=%v, invoiceDate=%v, dueDate=%v, currency=%v",
			paymentMethods, vendorName, accountNumber, numberOfItems, invoiceDate, dueDate, currency))

	randomInvoice := generate.GenerateRandomInvoiceData(&generate.GenerateInvoiceOptions{
		PaymentMethods: paymentMethods,
		VendorName:     vendorName,
		AccountNumber:  accountNumber,
		NumberOfItems:  numberOfItems,
		InvoiceDate:    invoiceDate.Format("January 2, 2006"),
		DueDate:        dueDate.Format("January 2, 2006"),
		Currency:       currency,
	})

	invoiceHtml, err := generate.GenerateInvoiceHtml(&randomInvoice)

	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("failed to create index.html file: %v", err))
		return
	}

	app.logger.Info("Converting HTML to PDF v2...")
	pdfContent, err := convert.HtmlToPdfV2(invoiceHtml)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("failed to convert HTML to PDF: %v", err))
		return
	}

	app.logger.Info("Sending PDF content to client...")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfContent)))
	if _, err := io.Copy(w, bytes.NewReader(pdfContent)); err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("failed to send PDF content to client: %v", err))
		return
	}

	app.logger.Info("Successfully converted HTML to PDF and sent to client")
}

func (app *application) createInvoice(w http.ResponseWriter, r *http.Request) {
	var input *generate.InvoiceData
	app.logger.Info("Creating invoice with the JSON body")

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.logger.Error("failed to decode invoice data", "error", err.Error())
		app.badRequestResponse(w, r, err)
		return
	}

	app.logger.Info("Generating invoice HTML")
	invoiceHtml, err := generate.GenerateInvoiceHtml(input)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("failed to create invoice.html file: %v", err))
		return
	}

	app.logger.Info("Converting HTML to PDF")
	pdfContent, err := convert.HtmlToPdfV2(invoiceHtml)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("failed to convert HTML to PDF: %v", err))
		return
	}

	app.logger.Info("Sending PDF content to client")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfContent)))
	if _, err := io.Copy(w, bytes.NewReader(pdfContent)); err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("failed to send PDF content to client: %v", err))
		return
	}

	app.logger.Info("Successfully created invoice and sent to client")
}
