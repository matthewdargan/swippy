package finding

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	findingURL         = "https://svcs.ebay.com/services/search/FindingService/v1?REST-PAYLOAD"
	operationName      = "findItemsByKeywords"
	serviceVersion     = "1.0.0"
	responseDataFormat = "JSON"
	findingHTTPTimeout = 5
)

// FindItemsByKeywords searches the eBay Finding API using provided keywords.
func FindItemsByKeywords(keywords string, appID string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, findingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request with URL: %w", err)
	}
	qry := req.URL.Query()
	qry.Add("OPERATION-NAME", operationName)
	qry.Add("SERVICE-VERSION", serviceVersion)
	qry.Add("SECURITY-APPNAME", appID)
	qry.Add("RESPONSE-DATA-FORMAT", responseDataFormat)
	qry.Add("keywords", keywords)
	req.URL.RawQuery = qry.Encode()
	c := &http.Client{
		Timeout: time.Second * findingHTTPTimeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request to the eBay Finding API: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body from the eBay Finding API: %w", err)
	}

	return b, nil
}
