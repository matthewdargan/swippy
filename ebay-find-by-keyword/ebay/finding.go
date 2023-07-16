package ebay

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

const (
	findingURL                     = "https://svcs.ebay.com/services/search/FindingService/v1?REST-PAYLOAD"
	findingByKeywordsOperationName = "findItemsByKeywords"
	findingServiceVersion          = "1.0.0"
	findingResponseDataFormat      = "JSON"
)

var (
	// ErrKeywordsMissing is returned when the 'keywords' parameter is missing.
	ErrKeywordsMissing = errors.New("ebay: keywords parameter is required")

	// ErrIncompleteAspectFilter is returned when the aspect filter is missing
	// either the 'aspectName' or 'aspectValueName' parameter.
	ErrIncompleteAspectFilter = errors.New("ebay: incomplete aspect filter: aspectName and aspectValueName are required")

	// ErrInvalidItemFilterSyntax is returned when both syntax types for itemFilters are used in the params.
	ErrInvalidItemFilterSyntax = errors.New(
		"ebay: invalid item filter syntax: both itemFilter.name and itemFilter(0).name are present")

	// ErrIncompleteItemFilterNameOnly is returned when an item filter is missing the 'value' parameter.
	ErrIncompleteItemFilterNameOnly = errors.New("ebay: incomplete item filter: missing value")

	// ErrIncompleteItemFilterParam is returned when an item filter is missing
	// either the 'paramName' or 'paramValue' parameter, as both 'paramName' and 'paramValue'
	// are required when either one is specified.
	ErrIncompleteItemFilterParam = errors.New(
		"ebay: incomplete item filter: both paramName and paramValue must be specified together")

	// ErrCreateRequest is returned when there is a failure to create a new HTTP request with the provided URL.
	ErrCreateRequest = errors.New("ebay: failed to create new HTTP request with URL")

	// ErrFailedRequest is returned when the eBay Finding API request fails.
	ErrFailedRequest = errors.New("ebay: failed to perform eBay Finding API request")

	// ErrInvalidStatus is returned when the eBay Finding API request returns an invalid status code.
	ErrInvalidStatus = errors.New("ebay: failed to perform eBay Finding API request with status code")

	// ErrDecodeAPIResponse is returned when there is an error decoding the eBay Finding API response body.
	ErrDecodeAPIResponse = errors.New("ebay: failed to decode eBay Finding API response body")
)

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

// An APIError is returned to represent a custom error that includes an error message
// and an HTTP status code.
type APIError struct {
	Err        string
	StatusCode int
}

func (e *APIError) Error() string {
	return e.Err
}

// FindItemsByKeywords searches the eBay Finding API using provided keywords.
func (svr *FindingServer) FindItemsByKeywords(params map[string]string, appID string) (*SearchResponse, error) {
	fParams, err := svr.validateParams(params)
	if err != nil {
		return nil, &APIError{Err: err.Error(), StatusCode: http.StatusBadRequest}
	}

	req, err := svr.createRequest(fParams, appID)
	if err != nil {
		return nil, &APIError{Err: err.Error(), StatusCode: http.StatusInternalServerError}
	}

	resp, err := svr.client.Do(req)
	if err != nil {
		return nil, &APIError{Err: ErrFailedRequest.Error() + ": " + err.Error(), StatusCode: http.StatusInternalServerError}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			Err:        ErrInvalidStatus.Error() + ": " + strconv.Itoa(resp.StatusCode),
			StatusCode: http.StatusInternalServerError,
		}
	}

	searchResp, err := svr.parseResponse(resp)
	if err != nil {
		return nil, &APIError{Err: err.Error(), StatusCode: http.StatusInternalServerError}
	}

	return searchResp, nil
}

type findingParams struct {
	keywords     string
	aspectFilter *aspectFilter
	itemFilters  []itemFilter
}

type aspectFilter struct {
	aspectName      string
	aspectValueName string
}

type itemFilter struct {
	// TODO: Change name to more specifically support only the ItemFilterType: https://developer.ebay.com/devzone/finding/CallRef/types/ItemFilterType.html.
	name       string
	value      string
	paramName  *string
	paramValue *string
}

func (svr *FindingServer) validateParams(params map[string]string) (*findingParams, error) {
	keywords, ok := params["keywords"]
	if !ok {
		return nil, ErrKeywordsMissing
	}

	fParams := &findingParams{
		keywords: keywords,
	}

	aspectName, anOk := params["aspectFilter.aspectName"]
	aspectValueName, avnOk := params["aspectFilter.aspectValueName"]
	if anOk != avnOk {
		return nil, ErrIncompleteAspectFilter
	}
	if anOk && avnOk {
		fParams.aspectFilter = &aspectFilter{
			aspectName:      aspectName,
			aspectValueName: aspectValueName,
		}
	}

	itemFilters, err := svr.processItemFilters(params)
	if err != nil {
		return nil, err
	}
	fParams.itemFilters = itemFilters

	return fParams, nil
}

func (svr *FindingServer) processItemFilters(params map[string]string) ([]itemFilter, error) {
	_, nonNumberedExists := params["itemFilter.name"]
	_, numberedExists := params["itemFilter(0).name"]

	// Check if both syntax types occur in the params.
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidItemFilterSyntax
	}

	if nonNumberedExists {
		return processNonNumberedItemFilter(params)
	}

	return processNumberedItemFilters(params)
}

func processNonNumberedItemFilter(params map[string]string) ([]itemFilter, error) {
	ifValue, vOk := params["itemFilter.value"]
	if !vOk {
		return nil, ErrIncompleteItemFilterNameOnly
	}

	filter := itemFilter{
		// No need to check itemFilter.name exists due to how this function is invoked from processItemFilters.
		name:  params["itemFilter.name"],
		value: ifValue,
	}

	ifParamName, pnOk := params["itemFilter.paramName"]
	ifParamValue, pvOk := params["itemFilter.paramValue"]
	if pnOk != pvOk {
		return nil, ErrIncompleteItemFilterParam
	}
	if pnOk && pvOk {
		filter.paramName = &ifParamName
		filter.paramValue = &ifParamValue
	}

	err := handleItemFilterType(&filter)
	if err != nil {
		return nil, err
	}

	return []itemFilter{filter}, nil
}

func processNumberedItemFilters(params map[string]string) ([]itemFilter, error) {
	var itemFilters []itemFilter
	for idx := 0; ; idx++ {
		ifName, nOk := params[fmt.Sprintf("itemFilter(%d).name", idx)]
		if !nOk {
			break
		}

		ifValue, vOk := params[fmt.Sprintf("itemFilter(%d).value", idx)]
		if !vOk {
			return nil, ErrIncompleteItemFilterNameOnly
		}

		itemFilter := itemFilter{
			name:  ifName,
			value: ifValue,
		}

		ifParamName, pnOk := params[fmt.Sprintf("itemFilter(%d).paramName", idx)]
		ifParamValue, pvOk := params[fmt.Sprintf("itemFilter(%d).paramValue", idx)]
		if pnOk != pvOk {
			return nil, ErrIncompleteItemFilterParam
		}
		if pnOk && pvOk {
			itemFilter.paramName = &ifParamName
			itemFilter.paramValue = &ifParamValue
		}

		err := handleItemFilterType(&itemFilter)
		if err != nil {
			return nil, err
		}

		itemFilters = append(itemFilters, itemFilter)
	}

	return itemFilters, nil
}

const (
	bestOfferOnly        = "BestOfferOnly"
	charityOnly          = "CharityOnly"
	authorizedSellerOnly = "AuthorizedSellerOnly"
	trueValue            = "true"
	falseValue           = "false"
)

func handleItemFilterType(filter *itemFilter) error {
	switch filter.name {
	case bestOfferOnly, charityOnly, authorizedSellerOnly:
		if filter.value != trueValue && filter.value != falseValue {
			return fmt.Errorf("Invalid value for %s itemFilter: %s", filter.name, filter.value)
		}
	default:
		return fmt.Errorf("Unsupported itemFilter type: %s", filter.name)
	}

	return nil
}

func (svr *FindingServer) createRequest(fParams *findingParams, appID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, findingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	qry := req.URL.Query()
	qry.Add("OPERATION-NAME", findingByKeywordsOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)
	qry.Add("keywords", fParams.keywords)

	if fParams.aspectFilter != nil {
		qry.Add("aspectFilter.aspectName", fParams.aspectFilter.aspectName)
		qry.Add("aspectFilter.aspectValueName", fParams.aspectFilter.aspectValueName)
	}

	for idx, itemFilter := range fParams.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", idx), itemFilter.name)
		qry.Add(fmt.Sprintf("itemFilter(%d).value", idx), itemFilter.value)

		if itemFilter.paramName != nil && itemFilter.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", idx), *itemFilter.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", idx), *itemFilter.paramValue)
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
