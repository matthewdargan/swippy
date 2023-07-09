package ebay

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	findingURL                     = "https://svcs.ebay.com/services/search/FindingService/v1?REST-PAYLOAD"
	findingByKeywordsOperationName = "findItemsByKeywords"
	findingServiceVersion          = "1.0.0"
	findingResponseDataFormat      = "JSON"
)

var (
	// ErrCreateRequest is returned when there is a failure to create a new HTTP request with the provided URL.
	ErrCreateRequest = fmt.Errorf("failed to create new HTTP request with URL")

	// ErrFailedRequest is returned when the eBay Finding API request fails.
	ErrFailedRequest = fmt.Errorf("failed to perform eBay Finding API request")

	// ErrInvalidStatus is returned when the eBay Finding API request returns an invalid status code.
	ErrInvalidStatus = fmt.Errorf("failed to perform eBay Finding API request with status code")

	// ErrDecodeAPIResponse is returned when there is an error decoding the eBay Finding API response body.
	ErrDecodeAPIResponse = fmt.Errorf("failed to decode eBay Finding API response body")
)

// FindingParams contains the query parameters needed to refine Finding API requests.
type FindingParams struct {
	Keywords     string
	AspectFilter *AspectFilter
	ItemFilters  []ItemFilter
}

// AspectFilter is used to refine the number of results returned in a response
// by specifying an aspect value within a domain.
type AspectFilter struct {
	AspectName      string // Name of a standard item characteristic associated with a domain.
	AspectValueName string // Value name for a given aspect.
}

// ItemFilter is used to refine the number of results returned in a response by specifying filter criteria for items.
// Multiple item filters can be included in the same request.
type ItemFilter struct {
	Name       string  // Name of the item filter.
	Value      string  // Value associated with the item filter name.
	ParamName  *string // Additional parameter name for certain filters.
	ParamValue *string // Additional parameter value for certain filters.
}

// FindingClient is the interface that represents a client for performing requests to the eBay Finding API.
type FindingClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// A FindingServer represents a server that interacts with the eBay Finding API.
type FindingServer struct {
	client FindingClient
}

// NewFindingServer returns a new FindingServer given a client.
func NewFindingServer(client FindingClient) *FindingServer {
	return &FindingServer{client: client}
}

// FindItemsByKeywords searches the eBay Finding API using provided keywords.
func (svr *FindingServer) FindItemsByKeywords(params map[string]string, appID string) (*SearchResponse, error) {
	findingParams, err := svr.validateParams(params)
	if err != nil {
		return nil, err
	}

	req, err := svr.createRequest(findingParams, appID)
	if err != nil {
		return nil, err
	}

	resp, err := svr.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedRequest, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrInvalidStatus, resp.StatusCode)
	}

	return svr.parseResponse(resp)
}

func (svr *FindingServer) validateParams(params map[string]string) (*FindingParams, error) {
	keywords, ok := params["keywords"]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrValidationFailed, "keywords are required")
	}

	findingParams := &FindingParams{
		Keywords: keywords,
	}

	aspectName, anOk := params["aspectFilter.aspectName"]
	aspectValueName, avnOk := params["aspectFilter.aspectValueName"]
	if anOk != avnOk {
		return nil, fmt.Errorf("%w: %s", ErrValidationFailed, "aspectFilter requires both aspectName and aspectValueName")
	}
	if anOk && avnOk {
		findingParams.AspectFilter = &AspectFilter{
			AspectName:      aspectName,
			AspectValueName: aspectValueName,
		}
	}

	for idx := 0; ; idx++ {
		ifName, ok := params[fmt.Sprintf("itemFilter(%d).name", idx)]
		if !ok {
			break
		}

		ifValue, ok := params[fmt.Sprintf("itemFilter(%d).value", idx)]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrValidationFailed, "itemFilter requires both name and value")
		}

		itemFilter := ItemFilter{
			Name:  ifName,
			Value: ifValue,
		}

		ifParamName, pnOk := params[fmt.Sprintf("itemFilter(%d).paramName", idx)]
		ifParamValue, pvOk := params[fmt.Sprintf("itemFilter(%d).paramValue", idx)]
		if pnOk != pvOk {
			return nil, fmt.Errorf("%w: %s", ErrValidationFailed, "itemFilter requires both paramName and paramValue")
		}
		if pnOk && pvOk {
			itemFilter.ParamName = &ifParamName
			itemFilter.ParamValue = &ifParamValue
		}

		findingParams.ItemFilters = append(findingParams.ItemFilters, itemFilter)
	}

	return findingParams, nil
}

func (svr *FindingServer) createRequest(findingParams *FindingParams, appID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, findingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	qry := req.URL.Query()
	qry.Add("OPERATION-NAME", findingByKeywordsOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)
	qry.Add("keywords", findingParams.Keywords)

	if findingParams.AspectFilter != nil {
		qry.Add("aspectFilter.aspectName", findingParams.AspectFilter.AspectName)
		qry.Add("aspectFilter.aspectValueName", findingParams.AspectFilter.AspectValueName)
	}

	for idx, itemFilter := range findingParams.ItemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", idx), itemFilter.Name)
		qry.Add(fmt.Sprintf("itemFilter(%d).value", idx), itemFilter.Value)

		if itemFilter.ParamName != nil && itemFilter.ParamValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", idx), *itemFilter.ParamName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", idx), *itemFilter.ParamValue)
		}
	}

	req.URL.RawQuery = qry.Encode()

	return req, nil
}

func (svr *FindingServer) parseResponse(resp *http.Response) (*SearchResponse, error) {
	defer resp.Body.Close()

	var items SearchResponse
	err := json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}

	return &items, nil
}
