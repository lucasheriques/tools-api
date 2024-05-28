package invoices

import (
	"fmt"
	"html"
	"html/template"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jaswdr/faker/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"tools.lucasfaria.dev/internal/utils"
)

type GenerateInvoiceOptions struct {
	PaymentMethods []string
	VendorName     string
	AccountNumber  int64
}

type CompanyInfo struct {
	Name          string
	StreetAddress string
	CityStateZip  string
	Email         string
}
type InvoicePaymentDetails struct {
	Name  string
	Value string
}

type PaymentMethod struct {
	Rail    string
	Details []InvoicePaymentDetails
}

type InvoiceData struct {
	CompanyLogo    string
	InvoiceNumber  string
	InvoiceDate    string
	DueDate        string
	VendorInfo     CompanyInfo
	CustomerInfo   CompanyInfo
	PaymentMethods []PaymentMethod
	PaymentMethod  string
	PaymentDetails []InvoicePaymentDetails
	Items          []InvoiceItem
	Total          string
}

type InvoiceItem struct {
	Description string
	Price       string
}

const tmplFile = "invoice.tmpl"

func GenerateAccountNumber() int64 {
	min := int64(1e8)  // The smallest 9 digit number
	max := int64(1e12) // The smallest 13 digit number
	return min + rand.Int63n(max-min)
}

func getPaymentMethods(accountNumber int64, companyName, address string, ach, wire, check bool) []PaymentMethod {
	paymentMethods := []PaymentMethod{}

	if ach {
		paymentMethods = append(paymentMethods, PaymentMethod{
			Rail: "ACH",
			Details: []InvoicePaymentDetails{
				{Name: "Routing number", Value: "026001591"},
				{Name: "Account number", Value: strconv.FormatInt(accountNumber, 10)},
				{Name: "Beneficiary name", Value: companyName},
			}})
	}

	if wire {
		paymentMethods = append(paymentMethods, PaymentMethod{
			Rail: "Wire",
			Details: []InvoicePaymentDetails{
				{Name: "Bank name", Value: "Wells Fargo"},
				{Name: "Routing number", Value: "121000248"},
				{Name: "Account number", Value: strconv.FormatInt(accountNumber, 10)},
				{Name: "Beneficiary name", Value: companyName},
			}})
	}

	if check {
		paymentMethods = append(paymentMethods, PaymentMethod{
			Rail: "Check",
			Details: []InvoicePaymentDetails{
				{Name: "Payable to", Value: companyName},
				{Name: "Address", Value: address},
			}})
	}

	return paymentMethods
}

func generateInvoiceItems() ([]InvoiceItem, string) {
	fake := faker.New()
	items := []InvoiceItem{}
	total := 0.0
	// random number between 1 and 8
	for i := 0; i < rand.Intn(8)+1; i++ {
		price := fake.Float64(2, 100, 1000)
		total += price
		items = append(items, InvoiceItem{
			Description: cases.Title(language.English).String(fake.Company().BS()),
			Price:       fmt.Sprintf("$%.2f", price),
		})
	}
	return items, fmt.Sprintf("$%.2f", total)
}

func includePaymentRails(rail string, options []string) bool {
	return slices.Contains(options, rail)
}

func generateData(options GenerateInvoiceOptions) InvoiceData {
	fake := faker.New()
	vendorName := options.VendorName
	if vendorName == "" {
		vendorName = fake.Company().Name()
	}
	now := time.Now()
	vendorAddress := fake.Address()

	vendorStreetAddress := vendorAddress.StreetName() + " " + vendorAddress.StreetSuffix() + ", " + strconv.Itoa(fake.RandomNumber(3))
	vendorCityStateZip := vendorAddress.City() + ", " + vendorAddress.StateAbbr() + " " + strings.Split(vendorAddress.PostCode(), "-")[0]
	vendorFullAddress := vendorStreetAddress + ", " + vendorCityStateZip
	vendorEmail := "bills@" + utils.TransformIntoValidEmailName(vendorName) + ".com"

	accountNumber := options.AccountNumber

	invoiceItems, total := generateInvoiceItems()

	includeAchRail := includePaymentRails("ach", options.PaymentMethods)
	includeWireRail := includePaymentRails("wire", options.PaymentMethods)
	includeCheckRail := includePaymentRails("check", options.PaymentMethods)

	data := InvoiceData{
		CompanyLogo: fmt.Sprintf("https://ui-avatars.com/api/?background=0D8ABC&color=fff&name=%s&rounded=true&size=64", strings.ReplaceAll(vendorName, " ", "+")),
		// convert from int to string
		InvoiceNumber: strconv.Itoa(fake.RandomNumber(5)),
		// Invoice date should be today's date
		InvoiceDate: now.Format("January 2, 2006"),
		// Due date should be 30 days from today
		DueDate: now.AddDate(0, 0, 30).Format("January 2, 2006"),
		VendorInfo: CompanyInfo{
			Name:          vendorName,
			StreetAddress: vendorStreetAddress,
			CityStateZip:  vendorCityStateZip,
			Email:         vendorEmail,
		},
		CustomerInfo: CompanyInfo{
			Name:          "Acme Corp.",
			StreetAddress: "1234 Main St",
			CityStateZip:  "San Francisco, CA 94111",
			Email:         "john@acme.com",
		},
		PaymentMethods: getPaymentMethods(accountNumber, vendorName, vendorFullAddress, includeAchRail, includeWireRail, includeCheckRail),
		Items:          invoiceItems,
		Total:          total,
	}

	return data
}

func GenerateHtmlFile(options GenerateInvoiceOptions) (*os.File, error) {
	invoiceData := generateData(options)

	templ, err := template.New(tmplFile).Funcs(template.FuncMap{
		"nl2br": func(text string) template.HTML {
			return template.HTML(strings.Replace(html.EscapeString(text), "\n", "<br>", -1))
		},
		"spacesToPlus": func(text string) string {
			return strings.ReplaceAll(text, " ", "+")
		},
	}).ParseFiles(tmplFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing template template: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "invoice.html")
	if err != nil {
		return nil, fmt.Errorf("error creating index.html file: %v", err)
	}
	defer tmpFile.Close()

	err = templ.Execute(tmpFile, invoiceData)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %v", err)
	}

	return tmpFile, nil
}
