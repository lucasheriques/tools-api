package convert

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dcaraxes/gotenberg-go-client/v8"
)

const gotenbergURL = "http://gotenberg.default.svc.cluster.local:80"

func HtmlToPdfV2(htmlFile *os.File) ([]byte, error) {
	client := &gotenberg.Client{
		Hostname: gotenbergURL,
	}

	index, err := gotenberg.NewDocumentFromPath("index.html", htmlFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to create new document: %v", err)
	}

	req := gotenberg.NewHTMLRequest(index)
	// req.SkipNetworkIdleEvent()

	client.Store(req, "/gotenberg/test.pdf")

	resp, _ := client.Post(req)
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg responded with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg responded with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	pdfContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return pdfContent, nil
}
