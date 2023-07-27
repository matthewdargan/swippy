package ebay_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/matthewdargan/swippy-api/ebay-find-by-keyword/ebay"
)

var (
	ErrClientFailure = errors.New("http: client failed")
	appID            = "super secret ID"
	searchResp       = ebay.SearchResponse{
		FindItemsByKeywordsResponse: []ebay.FindItemsByKeywordsResponse{
			{
				Ack:       []string{"Success"},
				Version:   []string{"1.0"},
				Timestamp: []time.Time{time.Date(2023, 6, 24, 0, 0, 0, 0, time.UTC)},
				SearchResult: []ebay.SearchResult{
					{
						Count: "1",
						Item: []ebay.Item{
							{
								ItemID:   []string{"1234567890"},
								Title:    []string{"Sample Item"},
								GlobalID: []string{"global-id-123"},
								Subtitle: []string{"Sample Item Subtitle"},
								PrimaryCategory: []ebay.PrimaryCategory{
									{
										CategoryID:   []string{"category-id-123"},
										CategoryName: []string{"Sample Category"},
									},
								},
								GalleryURL:  []string{"https://example.com/sample-item.jpg"},
								ViewItemURL: []string{"https://example.com/sample-item"},
								ProductID: []ebay.ProductID{
									{
										Type:  "product-type-123",
										Value: "product-value-123",
									},
								},
								AutoPay:    []string{"true"},
								PostalCode: []string{"12345"},
								Location:   []string{"Sample Location"},
								Country:    []string{"Sample Country"},
								ShippingInfo: []ebay.ShippingInfo{
									{
										ShippingServiceCost: []ebay.Price{
											{
												CurrencyID: "USD",
												Value:      "5.99",
											},
										},
										ShippingType:            []string{"Standard"},
										ShipToLocations:         []string{"US"},
										ExpeditedShipping:       []string{"false"},
										OneDayShippingAvailable: []string{"false"},
										HandlingTime:            []string{"1"},
									},
								},
								SellingStatus: []ebay.SellingStatus{
									{
										CurrentPrice: []ebay.Price{
											{
												CurrencyID: "USD",
												Value:      "19.99",
											},
										},
										ConvertedCurrentPrice: []ebay.Price{
											{
												CurrencyID: "USD",
												Value:      "19.99",
											},
										},
										SellingState: []string{"Active"},
										TimeLeft:     []string{"P1D"},
									},
								},
								ListingInfo: []ebay.ListingInfo{
									{
										BestOfferEnabled:  []string{"true"},
										BuyItNowAvailable: []string{"false"},
										StartTime:         []time.Time{time.Date(2023, 6, 24, 0, 0, 0, 0, time.UTC)},
										EndTime:           []time.Time{time.Date(2023, 7, 24, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)},
										ListingType:       []string{"Auction"},
										Gift:              []string{"false"},
										WatchCount:        []string{"10"},
									},
								},
								ReturnsAccepted: []string{"true"},
								Condition: []ebay.Condition{
									{
										ConditionID:          []string{"1000"},
										ConditionDisplayName: []string{"New"},
									},
								},
								IsMultiVariationListing: []string{"false"},
								TopRatedListing:         []string{"true"},
								DiscountPriceInfo: []ebay.DiscountPriceInfo{
									{
										OriginalRetailPrice: []ebay.Price{
											{
												CurrencyID: "USD",
												Value:      "29.99",
											},
										},
										PricingTreatment: []string{"STP"},
										SoldOnEbay:       []string{"true"},
										SoldOffEbay:      []string{"false"},
									},
								},
							},
						},
					},
				},
				PaginationOutput: []ebay.Pagination{
					{
						PageNumber:     []string{"1"},
						EntriesPerPage: []string{"10"},
						TotalPages:     []string{"1"},
						TotalEntries:   []string{"1"},
					},
				},
				ItemSearchURL: []string{"https://example.com/search?q=sample"},
			},
		},
	}
)

type MockFindingClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockFindingClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestFindItemsByKeywords(t *testing.T) {
	t.Parallel()
	params := map[string]string{
		"keywords": "marshmallows",
	}

	t.Run("can find items by keywords", func(t *testing.T) {
		t.Parallel()
		client := &MockFindingClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				body, err := json.Marshal(searchResp)
				assertNoError(t, err)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(body)),
				}, nil
			},
		}
		svr := ebay.NewFindingServer(client)
		resp, err := svr.FindItemsByKeywords(params, appID)
		assertNoError(t, err)
		assertSearchResponse(t, resp, &searchResp)
	})

	t.Run("returns error if the client returns an error", func(t *testing.T) {
		t.Parallel()
		client := &MockFindingClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, ErrClientFailure
			},
		}
		svr := ebay.NewFindingServer(client)
		_, err := svr.FindItemsByKeywords(params, appID)
		assertError(t, err)

		expected := fmt.Sprintf("%v: %v", ebay.ErrFailedRequest, ErrClientFailure)
		got := err.Error()
		assertErrorEquals(t, got, expected)
		assertStatusCodeEquals(t, err, http.StatusInternalServerError)
	})

	t.Run("returns error if the client request was not successful", func(t *testing.T) {
		t.Parallel()
		badStatusCodes := []int{
			http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusPaymentRequired,
			http.StatusForbidden,
			http.StatusNotFound,
			http.StatusMethodNotAllowed,
			http.StatusNotAcceptable,
			http.StatusProxyAuthRequired,
			http.StatusRequestTimeout,
			http.StatusConflict,
			http.StatusGone,
			http.StatusLengthRequired,
			http.StatusPreconditionFailed,
			http.StatusRequestEntityTooLarge,
			http.StatusRequestURITooLong,
			http.StatusUnsupportedMediaType,
			http.StatusRequestedRangeNotSatisfiable,
			http.StatusExpectationFailed,
			http.StatusTeapot,
			http.StatusMisdirectedRequest,
			http.StatusUnprocessableEntity,
			http.StatusLocked,
			http.StatusFailedDependency,
			http.StatusTooEarly,
			http.StatusUpgradeRequired,
			http.StatusPreconditionRequired,
			http.StatusTooManyRequests,
			http.StatusRequestHeaderFieldsTooLarge,
			http.StatusUnavailableForLegalReasons,
			http.StatusInternalServerError,
			http.StatusNotImplemented,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
			http.StatusHTTPVersionNotSupported,
			http.StatusVariantAlsoNegotiates,
			http.StatusInsufficientStorage,
			http.StatusLoopDetected,
			http.StatusNotExtended,
			http.StatusNetworkAuthenticationRequired,
		}

		for _, statusCode := range badStatusCodes {
			client := &MockFindingClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: statusCode}, nil
				},
			}
			svr := ebay.NewFindingServer(client)
			_, err := svr.FindItemsByKeywords(params, appID)
			assertError(t, err)

			expected := fmt.Sprintf("%v: %d", ebay.ErrInvalidStatus, statusCode)
			got := err.Error()
			assertErrorEquals(t, got, expected)
			assertStatusCodeEquals(t, err, http.StatusInternalServerError)
		}
	})

	t.Run("returns error if the response cannot be parsed into SearchResponse", func(t *testing.T) {
		t.Parallel()
		badData := `[123.1, 234.2]`
		client := &MockFindingClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				body := []byte(badData)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(body)),
				}, nil
			},
		}
		svr := ebay.NewFindingServer(client)
		_, err := svr.FindItemsByKeywords(params, appID)
		assertError(t, err)

		expected := fmt.Sprintf("%v: %v", ebay.ErrDecodeAPIResponse,
			"json: cannot unmarshal array into Go value of type ebay.SearchResponse")
		got := err.Error()
		assertErrorEquals(t, got, expected)
		assertStatusCodeEquals(t, err, http.StatusInternalServerError)
	})
}

type findItemsTestCase struct {
	Name          string
	Params        map[string]string
	ExpectedError error
}

func TestValidateParams(t *testing.T) {
	t.Parallel()
	testCases := []findItemsTestCase{
		{
			Name:          "returns error if params does not contain keywords",
			Params:        map[string]string{},
			ExpectedError: ebay.ErrKeywordsMissing,
		},
		{
			Name: "can find items by aspectFilter",
			Params: map[string]string{
				"keywords":                     "marshmallows",
				"aspectFilter.aspectName":      "squish level",
				"aspectFilter.aspectValueName": "very squishy",
			},
		},
		{
			Name: "returns error if params contains aspectFilter but not keywords",
			Params: map[string]string{
				"aspectFilter.aspectName":      "squish level",
				"aspectFilter.aspectValueName": "very squishy",
			},
			ExpectedError: ebay.ErrKeywordsMissing,
		},
		{
			Name: "returns error if params contains aspectName but not aspectValueName",
			Params: map[string]string{
				"keywords":                "marshmallows",
				"aspectFilter.aspectName": "squish level",
			},
			ExpectedError: ebay.ErrIncompleteAspectFilter,
		},
		{
			Name: "returns error if params contains aspectValueName but not aspectName",
			Params: map[string]string{
				"keywords":                     "marshmallows",
				"aspectFilter.aspectValueName": "very squishy",
			},
			ExpectedError: ebay.ErrIncompleteAspectFilter,
		},
		{
			Name: "can find items by basic, non-numbered itemFilter with non-numbered value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items by basic, non-numbered itemFilter with numbered values",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ExcludeCategory",
				"itemFilter.value(0)": "1",
				"itemFilter.value(1)": "2",
			},
		},
		{
			Name: "can find items by non-numbered itemFilter with name, value, paramName, and paramValue",
			Params: map[string]string{
				"keywords":              "marshmallows",
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "5.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "EUR",
			},
		},
		{
			Name: "returns error if params contains non-numbered itemFilter but not keywords",
			Params: map[string]string{
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "true",
			},
			ExpectedError: ebay.ErrKeywordsMissing,
		},
		{
			Name: "returns error if params contains non-numbered itemFilter name but not value",
			Params: map[string]string{
				"keywords":        "marshmallows",
				"itemFilter.name": "BestOfferOnly",
			},
			ExpectedError: ebay.ErrIncompleteItemFilterNameOnly,
		},
		{
			// The numbered itemFilter.value(1) will be ignored because indexing does not start at 0.
			Name: "returns error if params contains non-numbered itemFilter name and numbered value greater than 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "BestOfferOnly",
				"itemFilter.value(1)": "true",
			},
			ExpectedError: ebay.ErrIncompleteItemFilterNameOnly,
		},
		{
			// The numbered itemFilter.value(1) will be ignored because indexing does not start at 0.
			// Therefore, only the itemFilter.value is considered and this becomes a basic, non-numbered itemFilter.
			Name: "can find items by basic, non-numbered itemFilter with non-numbered value and numbered value greater than 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "BestOfferOnly",
				"itemFilter.value":    "true",
				"itemFilter.value(1)": "true",
			},
		},
		{
			// The itemFilter will be ignored if no itemFilter.name param is found before other itemFilter params.
			Name: "can find items if params contains non-numbered itemFilter value only",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.value": "true",
			},
		},
		{
			// The itemFilter will be ignored if no itemFilter.name param is found before other itemFilter params.
			Name: "can find items if params contains non-numbered itemFilter paramName only",
			Params: map[string]string{
				"keywords":             "marshmallows",
				"itemFilter.paramName": "Currency",
			},
		},
		{
			// The itemFilter will be ignored if no itemFilter.name param is found before other itemFilter params.
			Name: "can find items if params contains non-numbered itemFilter paramValue only",
			Params: map[string]string{
				"keywords":              "marshmallows",
				"itemFilter.paramValue": "EUR",
			},
		},
		{
			Name: "returns error if params contains non-numbered itemFilter paramName but not paramValue",
			Params: map[string]string{
				"keywords":             "marshmallows",
				"itemFilter.name":      "MaxPrice",
				"itemFilter.value":     "5.0",
				"itemFilter.paramName": "Currency",
			},
			ExpectedError: ebay.ErrIncompleteItemFilterParam,
		},
		{
			Name: "returns error if params contains non-numbered itemFilter paramValue but not paramName",
			Params: map[string]string{
				"keywords":              "marshmallows",
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "5.0",
				"itemFilter.paramValue": "EUR",
			},
			ExpectedError: ebay.ErrIncompleteItemFilterParam,
		},
		{
			Name: "returns error if params contain numbered and non-numbered itemFilter syntax types",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "BestOfferOnly",
				"itemFilter.value":    "true",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "5.0",
			},
			ExpectedError: ebay.ErrInvalidItemFilterSyntax,
		},
		{
			Name: "returns error if params contain non-numbered itemFilter with numbered and non-numbered value syntax types",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ExcludeCategory",
				"itemFilter.value":    "1",
				"itemFilter.value(0)": "2",
			},
			ExpectedError: ebay.ErrInvalidItemFilterSyntax,
		},
		{
			Name: "returns error if params contain numbered itemFilter with numbered and non-numbered value syntax types",
			Params: map[string]string{
				"keywords":               "marshmallows",
				"itemFilter(0).name":     "ExcludeCategory",
				"itemFilter(0).value":    "1",
				"itemFilter(0).value(0)": "2",
			},
			ExpectedError: ebay.ErrInvalidItemFilterSyntax,
		},
		{
			Name: "can find items by basic, numbered itemFilter with non-numbered value",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "BestOfferOnly",
				"itemFilter(0).value": "true",
			},
		},
		{
			Name: "can find items by basic, numbered itemFilter with numbered values",
			Params: map[string]string{
				"keywords":               "marshmallows",
				"itemFilter(0).name":     "ExcludeCategory",
				"itemFilter(0).value(0)": "1",
				"itemFilter(0).value(1)": "2",
			},
		},
		{
			Name: "can find items by numbered itemFilter with name, value, paramName, and paramValue",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "EUR",
			},
		},
		{
			Name: "can find items by 2 basic, numbered itemFilters",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "BestOfferOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "5.0",
			},
		},
		{
			Name: "can find items by 1st advanced, numbered and 2nd basic, numbered itemFilters",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "EUR",
				"itemFilter(1).name":       "BestOfferOnly",
				"itemFilter(1).value":      "true",
			},
		},
		{
			Name: "can find items by 1st basic, numbered and 2nd advanced, numbered itemFilters",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "BestOfferOnly",
				"itemFilter(0).value":      "true",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "5.0",
				"itemFilter(1).paramName":  "Currency",
				"itemFilter(1).paramValue": "EUR",
			},
		},
		{
			Name: "can find items by 2 advanced, numbered itemFilters",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "1.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "EUR",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "5.0",
				"itemFilter(1).paramName":  "Currency",
				"itemFilter(1).paramValue": "EUR",
			},
		},
		{
			Name: "returns error if params contains numbered itemFilter but not keywords",
			Params: map[string]string{
				"itemFilter(0).name":  "BestOfferOnly",
				"itemFilter(0).value": "true",
			},
			ExpectedError: ebay.ErrKeywordsMissing,
		},
		{
			Name: "returns error if params contains numbered itemFilter name but not value",
			Params: map[string]string{
				"keywords":           "marshmallows",
				"itemFilter(0).name": "BestOfferOnly",
			},
			ExpectedError: ebay.ErrIncompleteItemFilterNameOnly,
		},
		{
			// The numbered itemFilter(0).value(1) will be ignored because indexing does not start at 0.
			Name: "returns error if params contains numbered itemFilter name and numbered value greater than 0",
			Params: map[string]string{
				"keywords":               "marshmallows",
				"itemFilter(0).name":     "BestOfferOnly",
				"itemFilter(0).value(1)": "true",
			},
			ExpectedError: ebay.ErrIncompleteItemFilterNameOnly,
		},
		{
			// The numbered itemFilter(0).value(1) will be ignored because indexing does not start at 0.
			// Therefore, only the itemFilter(0).value is considered and this becomes a basic, numbered itemFilter.
			Name: "can find items by basic, numbered itemFilter with non-numbered value and numbered value greater than 0",
			Params: map[string]string{
				"keywords":               "marshmallows",
				"itemFilter(0).name":     "BestOfferOnly",
				"itemFilter(0).value":    "true",
				"itemFilter(0).value(1)": "true",
			},
		},
		{
			// The itemFilter will be ignored if no itemFilter(0).name param is found before other itemFilter params.
			Name: "can find items if params contains numbered itemFilter value only",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).value": "true",
			},
		},
		{
			// The itemFilter will be ignored if no itemFilter(0).name param is found before other itemFilter params.
			Name: "can find items if params contains numbered itemFilter paramName only",
			Params: map[string]string{
				"keywords":                "marshmallows",
				"itemFilter(0).paramName": "Currency",
			},
		},
		{
			// The itemFilter will be ignored if no itemFilter(0).name param is found before other itemFilter params.
			Name: "can find items if params contains numbered itemFilter paramValue only",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).paramValue": "EUR",
			},
		},
		{
			Name: "returns error if params contains numbered itemFilter paramName but not paramValue",
			Params: map[string]string{
				"keywords":                "marshmallows",
				"itemFilter(0).name":      "MaxPrice",
				"itemFilter(0).value":     "5.0",
				"itemFilter(0).paramName": "Currency",
			},
			ExpectedError: ebay.ErrIncompleteItemFilterParam,
		},
		{
			Name: "returns error if params contains numbered itemFilter paramValue but not paramName",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramValue": "EUR",
			},
			ExpectedError: ebay.ErrIncompleteItemFilterParam,
		},
		{
			Name: "returns error if params contains non-numbered, unsupported itemFilter name",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "UnsupportedFilter",
				"itemFilter.value": "true",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrUnsupportedItemFilterType, "UnsupportedFilter"),
		},
		{
			Name: "returns error if params contains numbered, unsupported itemFilter name",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "UnsupportedFilter",
				"itemFilter(0).value": "true",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrUnsupportedItemFilterType, "UnsupportedFilter"),
		},
		{
			Name: "returns error if params contains numbered supported and unsupported itemFilter names",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "BestOfferOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "UnsupportedFilter",
				"itemFilter(1).value": "true",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrUnsupportedItemFilterType, "UnsupportedFilter"),
		},
		{
			Name: "can find items if params contains AuthorizedSellerOnly itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "AuthorizedSellerOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains AuthorizedSellerOnly itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "AuthorizedSellerOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains AuthorizedSellerOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "AuthorizedSellerOnly",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains valid AvailableTo itemFilter",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "AvailableTo",
				"itemFilter.value": "US",
			},
		},
		{
			Name: "returns error if params contains AvailableTo itemFilter with lowercase characters",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "AvailableTo",
				"itemFilter.value": "us",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCountryCode, "us"),
		},
		{
			Name: "returns error if params contains AvailableTo itemFilter with 1 uppercase character",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "AvailableTo",
				"itemFilter.value": "U",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCountryCode, "U"),
		},
		{
			Name: "returns error if params contains AvailableTo itemFilter with 3 uppercase character",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "AvailableTo",
				"itemFilter.value": "USA",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCountryCode, "USA"),
		},
		{
			Name: "can find items if params contains BestOfferOnly itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains BestOfferOnly itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains BestOfferOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains CharityOnly itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "CharityOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains CharityOnly itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "CharityOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains CharityOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "CharityOnly",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition name",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "dirty",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 1000",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "1000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 1500",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "1500",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 1750",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "1750",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2000",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2010",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2010",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2020",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2020",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2030",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2030",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2500",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2500",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2750",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2750",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 3000",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "3000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 4000",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "4000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 5000",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "5000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 6000",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "6000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 7000",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "7000",
			},
		},
		{
			Name: "returns error if params contains Condition itemFilter with condition ID 1",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Condition",
				"itemFilter.value": "1",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCondition, "1"),
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID AUD",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "AUD",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID CAD",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "CAD",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID CHF",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "CHF",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID CNY",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "CNY",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID EUR",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "EUR",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID GBP",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "GBP",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID HKD",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "HKD",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID INR",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "INR",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID MYR",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "MYR",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID PHP",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "PHP",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID PLN",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "PLN",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID SEK",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "SEK",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID SGD",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "SGD",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID TWD",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "TWD",
			},
		},
		{
			Name: "returns error if params contains Currency itemFilter with currency ID ZZZ",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Currency",
				"itemFilter.value": "ZZZ",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "can find items if params contains EndTimeFrom itemFilter with future timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "EndTimeFrom",
				"itemFilter.value": time.Now().Add(1 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains EndTimeFrom itemFilter with non-parsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "EndTimeFrom",
				"itemFilter.value": "not a timestamp",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains EndTimeFrom itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "EndTimeFrom",
				"itemFilter.value": time.Now().Add(1 * time.Second).Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains EndTimeFrom itemFilter with past timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "EndTimeFrom",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(-1*time.Second).UTC().Format(time.RFC3339)),
		},
		{
			Name: "can find items if params contains EndTimeTo itemFilter with future timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "EndTimeTo",
				"itemFilter.value": time.Now().Add(1 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains EndTimeTo itemFilter with non-parsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "EndTimeTo",
				"itemFilter.value": "not a timestamp",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains EndTimeTo itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "EndTimeTo",
				"itemFilter.value": time.Now().Add(1 * time.Second).Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains EndTimeTo itemFilter with past timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "EndTimeTo",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(-1*time.Second).UTC().Format(time.RFC3339)),
		},
		{
			Name: "can find items if params contains ExcludeAutoPay itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExcludeAutoPay",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains ExcludeAutoPay itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExcludeAutoPay",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains ExcludeAutoPay itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExcludeAutoPay",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
	}

	for _, tc := range testCases {
		testCase := tc
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			client := &MockFindingClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					body, err := json.Marshal(searchResp)
					assertNoError(t, err)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(body)),
					}, nil
				},
			}
			svr := ebay.NewFindingServer(client)
			resp, err := svr.FindItemsByKeywords(testCase.Params, appID)

			if testCase.ExpectedError != nil {
				assertError(t, err)
				assertErrorEquals(t, err.Error(), testCase.ExpectedError.Error())
				assertStatusCodeEquals(t, err, http.StatusBadRequest)
			} else {
				assertNoError(t, err)
				assertSearchResponse(t, resp, &searchResp)
			}
		})
	}
}

func assertError(tb testing.TB, err error) {
	tb.Helper()
	if err == nil {
		tb.Fatal("expected an error but did not get one")
	}
}

func assertNoError(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("did not expect error but got one, %v", err)
	}
}

func assertErrorEquals(tb testing.TB, got, expected string) {
	tb.Helper()
	if got != expected {
		tb.Errorf("got %v, expected %v", got, expected)
	}
}

func assertStatusCodeEquals(tb testing.TB, err error, expectedStatusCode int) {
	tb.Helper()
	var apiError *ebay.APIError
	if !errors.As(err, &apiError) {
		tb.Error("expected APIError")
	} else if apiError.StatusCode != expectedStatusCode {
		tb.Errorf("got status code %d, expected %d", apiError.StatusCode, expectedStatusCode)
	}
}

func assertSearchResponse(tb testing.TB, got, expected *ebay.SearchResponse) {
	tb.Helper()
	if !reflect.DeepEqual(*got, *expected) {
		tb.Errorf("got %v, expected %v", got, expected)
	}
}
