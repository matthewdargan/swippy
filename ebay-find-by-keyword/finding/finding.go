package finding

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const findingURL = "https://svcs.ebay.com/services/search/FindingService/v1?OPERATION-NAME=findItemsByKeywords"

func FindItemsByKeywords(keywords string) ([]byte, error) {
	appID := os.Getenv("EBAY_APP_ID")
	apiURL := fmt.Sprintf("%s&SERVICE-VERSION=1.0.0&SECURITY-APPNAME=%s&RESPONSE-DATA-FORMAT=JSON&keywords=%s", findingURL, appID, keywords)
	resp, err := http.Get(apiURL)
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
