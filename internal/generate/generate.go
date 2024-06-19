package generate

import (
	"fmt"
	"html"
	"html/template"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/jaswdr/faker/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"tools.lucasfaria.dev/internal/utils"
)

type GenerateInvoiceOptions struct {
	PaymentMethods []string
	VendorName     string
	AccountNumber  int64
	NumberOfItems  int
	InvoiceDate    string
	DueDate        string
	Currency       string
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

func getCurrenctSymbol(currency string) string {
	switch currency {
	case "usd":
		return "$"
	case "eur":
		return "€"
	case "jpy":
		return "¥"
	case "gbp":
		return "£"
	case "aud":
		return "A$"
	case "cad":
		return "C$"
	case "chf":
		return "CHF"
	case "cny":
		return "¥"
	case "sek":
		return "kr"
	case "nzd":
		return "NZ$"
	case "mxn":
		return "Mex$"
	case "sgd":
		return "S$"
	case "hkd":
		return "HK$"
	case "nok":
		return "kr"
	case "krw":
		return "₩"
	case "try":
		return "₺"
	case "rub":
		return "₽"
	case "inr":
		return "₹"
	case "brl":
		return "R$"
	case "zar":
		return "R"
	default:
		return "$"
	}
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

func generateInvoiceItems(numOfItems int, currency string) ([]InvoiceItem, string) {
	fake := faker.New()
	items := []InvoiceItem{}
	total := 0.0
	for i := 0; i < numOfItems; i++ {
		price := fake.Float64(2, 100, 1000)
		total += price
		items = append(items, InvoiceItem{
			Description: cases.Title(language.English).String(fake.Company().BS()),
			Price:       fmt.Sprintf("%s%.2f", currency, price),
		})
	}
	return items, fmt.Sprintf("%s%.2f", currency, total)
}

func includePaymentRails(rail string, options []string) bool {
	return slices.Contains(options, rail)
}

func generateData(options *GenerateInvoiceOptions) InvoiceData {
	fake := faker.New()
	vendorName := options.VendorName
	if vendorName == "" {
		vendorName = fake.Company().Name()
	}
	vendorAddress := fake.Address()
	currency := getCurrenctSymbol(options.Currency)

	vendorStreetAddress := vendorAddress.StreetName() + " " + vendorAddress.StreetSuffix() + ", " + strconv.Itoa(fake.RandomNumber(3))
	vendorCityStateZip := vendorAddress.City() + ", " + vendorAddress.StateAbbr() + " " + strings.Split(vendorAddress.PostCode(), "-")[0]
	vendorFullAddress := vendorStreetAddress + ", " + vendorCityStateZip
	vendorEmail := "bills@" + utils.TransformIntoValidEmailName(vendorName) + ".com"

	accountNumber := options.AccountNumber

	invoiceItems, total := generateInvoiceItems(options.NumberOfItems, currency)

	includeAchRail := includePaymentRails("ach", options.PaymentMethods)
	includeWireRail := includePaymentRails("wire", options.PaymentMethods)
	includeCheckRail := includePaymentRails("check", options.PaymentMethods)

	data := InvoiceData{
		CompanyLogo: fmt.Sprintf("https://ui-avatars.com/api/?background=0D8ABC&color=fff&name=%s&rounded=true&size=64", strings.ReplaceAll(vendorName, " ", "+")),
		// convert from int to string
		InvoiceNumber: strconv.Itoa(fake.RandomNumber(5)),
		// Invoice date should be today's date
		InvoiceDate: options.InvoiceDate,
		// Due date should be 30 days from today
		DueDate: options.DueDate,
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
			Email:         "mary@acme.com",
		},
		PaymentMethods: getPaymentMethods(accountNumber, vendorName, vendorFullAddress, includeAchRail, includeWireRail, includeCheckRail),
		Items:          invoiceItems,
		Total:          total,
	}

	return data
}

func GenerateHtmlFile(options *GenerateInvoiceOptions) (*os.File, error) {
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
