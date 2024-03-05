package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"tools-api/convert"
	"tools-api/utils"

	"github.com/jaswdr/faker/v2"
)

const tmplFile = "invoice.tmpl"

type InvoiceData struct {
	CompanyLogo    string
	InvoiceNumber  string
	InvoiceDate    string
	DueDate        string
	VendorInfo     string
	CustomerInfo   string
	PaymentMethod  string
	PaymentDetails []InvoicePaymentDetails
	Items          []InvoiceItem
	Total          string
}

type InvoiceItem struct {
	Description string
	Price       string
}

type InvoicePaymentDetails struct {
	Name  string
	Value string
}

func generateData() InvoiceData {
	fake := faker.New()
	companyName := fake.Company().Name()
	now := time.Now()
	companyAddress := fake.Address()

	companyStreetAddress := companyAddress.StreetName() + " " + companyAddress.StreetSuffix() + ", " + strconv.Itoa(fake.RandomNumber(3))

	companyEmail := "bills@" + utils.TransformIntoValidEmailName(companyName) + "." + fake.Internet().Domain()

	data := InvoiceData{
		CompanyLogo: "https://example.com/logo.png",
		// convert from int to string
		InvoiceNumber: strconv.Itoa(fake.RandomNumber(5)),
		// Invoice date should be today's date
		InvoiceDate: now.Format("January 2, 2006"),
		// Due date should be 30 days from today
		DueDate: now.AddDate(0, 0, 30).Format("January 2, 2006"),
		VendorInfo: fmt.Sprintf(`%s
		%s
		%s %s
		%s`, companyName, companyStreetAddress, companyAddress.City(), companyAddress.StateAbbr(), companyEmail),
		CustomerInfo: `Acme Corp.
		John Doe
		john@example.com`,
		PaymentMethod: "ACH",
		PaymentDetails: []InvoicePaymentDetails{
			{Name: "Routing number", Value: "026001591"},
			{Name: "Account number", Value: "7534028150001"},
			{Name: "Beneficiary name", Value: "TechWave Solutions"},
		},
		Items: []InvoiceItem{
			{Description: "Website design", Price: "$300.00"},
			{Description: "Hosting (3 months)", Price: "$75.00"},
			{Description: "Domain name (1 year)", Price: "$10.00"},
		},
		Total: "$385.00",
	}

	return data
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", rootHandler)
	mux.HandleFunc("POST /fake-invoice", invoiceHandler)
	mux.HandleFunc("POST /convert/html", convertHTMLHandler)

	fmt.Println("Starting tools-api server at port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hey! This is my new Tools API.")
}

func generateHtmlFile(invoiceData InvoiceData) ([]byte, error) {
	templ, err := template.New(tmplFile).Funcs(template.FuncMap{
		"nl2br": func(text string) template.HTML {
			return template.HTML(strings.Replace(html.EscapeString(text), "\n", " <br/> ", -1))
		},
	}).ParseFiles(tmplFile)
	if err != nil {
		return nil, fmt.Errorf("g template: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "invoice-*.html")
	if err != nil {
		return nil, fmt.Errorf("error creating index.html file: %v", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	err = templ.Execute(tmpFile, invoiceData)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %v", err)
	}

	htmlContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("error reading HTML file: %v", err)
	}

	return htmlContent, nil
}

// Invoice should receive JSON payload with the InvoiceData
// Then it should create a index.html file with the invoice data inside a temporary folder
// The template is located at tmpl/invoice.tmpl
// After creating the template, it should call the convertHTMLHandler to convert the HTML to PDF
// And finally, return the PDF to the client
func invoiceHandler(w http.ResponseWriter, r *http.Request) {
	// var invoiceData InvoiceData
	// if err := json.NewDecoder(r.Body).Decode(&invoiceData); err != nil {
	// 	http.Error(w, "Failed to decode JSON payload", http.StatusBadRequest)
	// 	log.Printf("Failed to decode JSON payload: %v", err)
	// 	return
	// }
	// defer r.Body.Close()

	// Create the index.html file with the invoice data
	htmlContent, err := generateHtmlFile(generateData())
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
