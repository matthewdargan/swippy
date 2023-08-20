package ebay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// A FindingClient represents a client that interacts with the eBay Finding API.
type FindingClient struct {
	*http.Client
	AppID   string
	BaseURL string
}

const findingURL = "https://svcs.ebay.com/services/search/FindingService/v1?REST-PAYLOAD"

// NewFindingClient returns a new FindingClient given an HTTP client and a valid eBay application ID.
func NewFindingClient(client *http.Client, appID string) *FindingClient {
	return &FindingClient{Client: client, AppID: appID, BaseURL: findingURL}
}

// An APIError is returned to represent a custom error that includes an error message
// and an HTTP status code.
type APIError struct {
	Err        error
	StatusCode int
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("ebay: %v", e.Err)
	}
	return "ebay: API error occurred"
}

// FindItemsByCategories searches the eBay Finding API using the provided category, additional parameters,
// and a valid eBay application ID.
func (c *FindingClient) FindItemsByCategories(
	ctx context.Context, params map[string]string,
) (FindItemsByCategoriesResponse, error) {
	var findItems FindItemsByCategoriesResponse
	err := c.findItems(ctx, params, &findItemsByCategoryParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

// FindItemsByKeywords searches the eBay Finding API using the provided keywords, additional parameters,
// and a valid eBay application ID.
func (c *FindingClient) FindItemsByKeywords(
	ctx context.Context, params map[string]string,
) (FindItemsByKeywordsResponse, error) {
	var findItems FindItemsByKeywordsResponse
	err := c.findItems(ctx, params, &findItemsByKeywordsParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

// FindItemsAdvanced searches the eBay Finding API using the provided category and/or keywords, additional parameters,
// and a valid eBay application ID.
func (c *FindingClient) FindItemsAdvanced(
	ctx context.Context, params map[string]string,
) (FindItemsAdvancedResponse, error) {
	var findItems FindItemsAdvancedResponse
	err := c.findItems(ctx, params, &findItemsAdvancedParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

// FindItemsByProduct searches the eBay Finding API using the provided product, additional parameters,
// and a valid eBay application ID.
func (c *FindingClient) FindItemsByProduct(
	ctx context.Context, params map[string]string,
) (FindItemsByProductResponse, error) {
	var findItems FindItemsByProductResponse
	err := c.findItems(ctx, params, &findItemsByProductParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

// FindItemsInEBayStores searches the eBay Finding API using the provided category, keywords, and/or store name,
// additional parameters, and a valid eBay application ID.
func (c *FindingClient) FindItemsInEBayStores(
	ctx context.Context, params map[string]string,
) (FindItemsInEBayStoresResponse, error) {
	var findItems FindItemsInEBayStoresResponse
	err := c.findItems(ctx, params, &findItemsInEBayStoresParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

var (
	// ErrFailedRequest is returned when the eBay Finding API request fails.
	ErrFailedRequest = errors.New("failed to perform eBay Finding API request")

	// ErrInvalidStatus is returned when the eBay Finding API request returns an invalid status code.
	ErrInvalidStatus = errors.New("failed to perform eBay Finding API request with status code")

	// ErrDecodeAPIResponse is returned when there is an error decoding the eBay Finding API response body.
	ErrDecodeAPIResponse = errors.New("failed to decode eBay Finding API response body")
)

func (c *FindingClient) findItems(
	ctx context.Context, params map[string]string, fParams findItemsParams, items FindItems,
) error {
	err := fParams.validateParams(params)
	if err != nil {
		return &APIError{Err: err, StatusCode: http.StatusBadRequest}
	}
	req, err := fParams.newRequest(ctx, c.BaseURL)
	if err != nil {
		return &APIError{Err: err, StatusCode: http.StatusInternalServerError}
	}
	resp, err := c.Do(req)
	if err != nil {
		return &APIError{Err: fmt.Errorf("%w: %w", ErrFailedRequest, err), StatusCode: http.StatusInternalServerError}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &APIError{
			Err:        fmt.Errorf("%w %d", ErrInvalidStatus, resp.StatusCode),
			StatusCode: http.StatusInternalServerError,
		}
	}
	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		return &APIError{
			Err:        fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err),
			StatusCode: http.StatusInternalServerError,
		}
	}
	return nil
}
