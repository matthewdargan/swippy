package finding

import (
	"io"
	"net/http"
	"os"
	"time"
)

const (
	findingURL         = "https://svcs.ebay.com/services/search/FindingService/v1?REST-PAYLOAD"
	operationName      = "findItemsByKeywords"
	serviceVersion     = "1.0.0"
	responseDataFormat = "JSON"
)

func FindItemsByKeywords(keywords string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, findingURL, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("OPERATION-NAME", operationName)
	q.Add("SERVICE-VERSION", serviceVersion)
	q.Add("SECURITY-APPNAME", os.Getenv("EBAY_APP_ID"))
	q.Add("RESPONSE-DATA-FORMAT", responseDataFormat)
	q.Add("keywords", keywords)
	req.URL.RawQuery = q.Encode()

	c := &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}
