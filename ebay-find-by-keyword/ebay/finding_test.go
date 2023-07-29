package ebay_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
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
			Name: "returns error if params contains EndTimeFrom itemFilter with unparsable value",
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
			Name: "returns error if params contains EndTimeTo itemFilter with unparsable value",
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
		{
			Name: "can find items if params contains ExcludeCategory itemFilter with category ID 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExcludeCategory",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains ExcludeCategory itemFilter with category ID 5",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExcludeCategory",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains ExcludeCategory itemFilter with unparsable category ID",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExcludeCategory",
				"itemFilter.value": "not a category ID",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "not a category ID", 0),
		},
		{
			Name: "returns error if params contains ExcludeCategory itemFilter with category ID -1",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExcludeCategory",
				"itemFilter.value": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains ExcludeCategory itemFilter with category IDs 0 and 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ExcludeCategory",
				"itemFilter.value(0)": "0",
				"itemFilter.value(1)": "1",
			},
		},
		{
			Name: "returns error if params contains ExcludeCategory itemFilter with category IDs 0 and -1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ExcludeCategory",
				"itemFilter.value(0)": "0",
				"itemFilter.value(1)": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name:   "can find items if params contains ExcludeCategory itemFilter with 25 category IDs",
			Params: generateFilterParams("ExcludeCategory", 25),
		},
		{
			Name:          "returns error if params contains ExcludeCategory itemFilter with 26 category IDs",
			Params:        generateFilterParams("ExcludeCategory", 26),
			ExpectedError: ebay.ErrMaxExcludeCategories,
		},
		{
			Name: "can find items if params contains ExcludeSeller itemFilter with seller ID 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExcludeSeller",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains ExcludeSeller itemFilter with seller IDs 0 and 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ExcludeSeller",
				"itemFilter.value(0)": "0",
				"itemFilter.value(1)": "1",
			},
		},
		{
			Name: "returns error if params contains ExcludeSeller and Seller itemFilters",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "ExcludeSeller",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "Seller",
				"itemFilter(1).value": "0",
			},
			ExpectedError: ebay.ErrExcludeSellerCannotBeUsedWithSellers,
		},
		{
			Name: "returns error if params contains ExcludeSeller and TopRatedSellerOnly itemFilters",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "ExcludeSeller",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "TopRatedSellerOnly",
				"itemFilter(1).value": "true",
			},
			ExpectedError: ebay.ErrExcludeSellerCannotBeUsedWithSellers,
		},
		{
			Name:   "can find items if params contains ExcludeSeller itemFilter with 100 seller IDs",
			Params: generateFilterParams("ExcludeSeller", 100),
		},
		{
			Name:          "returns error if params contains ExcludeSeller itemFilter with 101 seller IDs",
			Params:        generateFilterParams("ExcludeSeller", 101),
			ExpectedError: ebay.ErrMaxExcludeSellers,
		},
		{
			Name: "can find items if params contains ExpeditedShippingType itemFilter.value=Expedited",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExpeditedShippingType",
				"itemFilter.value": "Expedited",
			},
		},
		{
			Name: "can find items if params contains ExpeditedShippingType itemFilter.value=OneDayShipping",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExpeditedShippingType",
				"itemFilter.value": "OneDayShipping",
			},
		},
		{
			Name: "returns error if params contains ExpeditedShippingType itemFilter with invalid shipping type",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ExpeditedShippingType",
				"itemFilter.value": "InvalidShippingType",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidExpeditedShippingType, "InvalidShippingType"),
		},
		{
			Name: "can find items if params contains FeedbackScoreMax itemFilter with max 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FeedbackScoreMax",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMax itemFilter with max 5",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FeedbackScoreMax",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains FeedbackScoreMax itemFilter with unparsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FeedbackScoreMax",
				"itemFilter.value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMax itemFilter with max -1",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FeedbackScoreMax",
				"itemFilter.value": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains FeedbackScoreMin itemFilter with max 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FeedbackScoreMin",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin itemFilter with max 5",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FeedbackScoreMin",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains FeedbackScoreMin itemFilter with unparsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FeedbackScoreMin",
				"itemFilter.value": "not a minimum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin itemFilter with max -1",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FeedbackScoreMin",
				"itemFilter.value": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 1 and min 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 0 and max 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max and min 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min and max 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 10 and min 5",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 5 and max 10",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "10",
			},
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 0 and unparsable min",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "not a minimum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with unparsable min and max 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "not a minimum",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 0 and min -1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min -1 and max 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "-1",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 0 and min 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "1",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 1 and max 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 5 and min 10",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "10",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 10 and max 5",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "5",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with unparsable max and min 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "not a maximum",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 0 and unparsable max",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max -1 and min 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "-1",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 0 and max -1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "can find items if params contains FreeShippingOnly itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FreeShippingOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains FreeShippingOnly itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FreeShippingOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains FreeShippingOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "FreeShippingOnly",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains HideDuplicateItems itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "HideDuplicateItems",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains HideDuplicateItems itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "HideDuplicateItems",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains HideDuplicateItems itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "HideDuplicateItems",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-AT",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-AT",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-AU",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-AU",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-CH",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-CH",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-DE",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-DE",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-ENCA",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-ENCA",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-ES",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-ES",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-FR",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-FR",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-FRBE",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-FRBE",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-FRCA",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-FRCA",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-GB",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-GB",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-HK",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-HK",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-IE",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-IE",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-IN",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-IN",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-IT",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-IT",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-MOTOR",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-MOTOR",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-MY",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-MY",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-NL",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-NL",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-NLBE",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-NLBE",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-PH",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-PH",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-PL",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-PL",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-SG",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-SG",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-US",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-US",
			},
		},
		{
			Name: "returns error if params contains ListedIn itemFilter with Global ID EBAY-ZZZ",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-ZZZ",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidGlobalID, "EBAY-ZZZ"),
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type Auction",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "Auction",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type AuctionWithBIN",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "AuctionWithBIN",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type Classified",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "Classified",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type FixedPrice",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "FixedPrice",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type StoreInventory",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "StoreInventory",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type All",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "All",
			},
		},
		{
			Name: "returns error if params contains ListingType itemFilter with invalid listing type",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "not a listing type",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidListingType, "not a listing type"),
		},
		{
			Name: "returns error if params contains ListingType itemFilters with All and Auction listing types",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "All",
				"itemFilter.value(1)": "Auction",
			},
			ExpectedError: ebay.ErrInvalidAllListingType,
		},
		{
			Name: "returns error if params contains ListingType itemFilters with StoreInventory and All listing types",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "StoreInventory",
				"itemFilter.value(1)": "All",
			},
			ExpectedError: ebay.ErrInvalidAllListingType,
		},
		{
			Name: "returns error if params contains ListingType itemFilters with 2 Auction listing types",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "Auction",
				"itemFilter.value(1)": "Auction",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrDuplicateListingType, "Auction"),
		},
		{
			Name: "returns error if params contains ListingType itemFilters with 2 StoreInventory listing types",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "StoreInventory",
				"itemFilter.value(1)": "StoreInventory",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrDuplicateListingType, "StoreInventory"),
		},
		{
			Name: "returns error if params contains ListingType itemFilters with Auction and AuctionWithBIN listing types",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "Auction",
				"itemFilter.value(1)": "AuctionWithBIN",
			},
			ExpectedError: ebay.ErrInvalidAuctionListingTypes,
		},
		{
			Name: "returns error if params contains ListingType itemFilters with AuctionWithBIN and Auction listing types",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "AuctionWithBIN",
				"itemFilter.value(1)": "Auction",
			},
			ExpectedError: ebay.ErrInvalidAuctionListingTypes,
		},
		{
			Name: "can find items if params contains LocalPickupOnly itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LocalPickupOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains LocalPickupOnly itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LocalPickupOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains LocalPickupOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LocalPickupOnly",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains LocalSearchOnly itemFilter.value=true, buyerPostalCode, and MaxDistance",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"buyerPostalCode":     "123",
				"itemFilter(0).name":  "LocalSearchOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "MaxDistance",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains LocalSearchOnly itemFilter.value=false, buyerPostalCode, and MaxDistance",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"buyerPostalCode":     "123",
				"itemFilter(0).name":  "LocalSearchOnly",
				"itemFilter(0).value": "false",
				"itemFilter(1).name":  "MaxDistance",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains LocalSearchOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"buyerPostalCode":     "123",
				"itemFilter(0).name":  "LocalSearchOnly",
				"itemFilter(0).value": "123",
				"itemFilter(1).name":  "MaxDistance",
				"itemFilter(1).value": "5",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "returns error if params contains LocalSearchOnly itemFilter but no buyerPostalCode",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "LocalSearchOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "MaxDistance",
				"itemFilter(1).value": "5",
			},
			ExpectedError: ebay.ErrBuyerPostalCodeMissing,
		},
		{
			Name: "returns error if params contains LocalSearchOnly itemFilter but no MaxDistance itemFilter",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"buyerPostalCode":  "123",
				"itemFilter.name":  "LocalSearchOnly",
				"itemFilter.value": "true",
			},
			ExpectedError: ebay.ErrMaxDistanceMissing,
		},
		{
			Name: "can find items if params contains valid LocatedIn itemFilter",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LocatedIn",
				"itemFilter.value": "US",
			},
		},
		{
			Name: "returns error if params contains LocatedIn itemFilter with lowercase characters",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LocatedIn",
				"itemFilter.value": "us",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCountryCode, "us"),
		},
		{
			Name: "returns error if params contains LocatedIn itemFilter with 1 uppercase character",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LocatedIn",
				"itemFilter.value": "U",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCountryCode, "U"),
		},
		{
			Name: "returns error if params contains LocatedIn itemFilter with 3 uppercase character",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LocatedIn",
				"itemFilter.value": "USA",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCountryCode, "USA"),
		},
		{
			Name: "can find items if params contains LocatedIn itemFilter with 25 country codes",
			Params: map[string]string{
				"keywords":             "marshmallows",
				"itemFilter.name":      "LocatedIn",
				"itemFilter.value(0)":  "AA",
				"itemFilter.value(1)":  "AB",
				"itemFilter.value(2)":  "AC",
				"itemFilter.value(3)":  "AD",
				"itemFilter.value(4)":  "AE",
				"itemFilter.value(5)":  "AF",
				"itemFilter.value(6)":  "AG",
				"itemFilter.value(7)":  "AH",
				"itemFilter.value(8)":  "AI",
				"itemFilter.value(9)":  "AJ",
				"itemFilter.value(10)": "AK",
				"itemFilter.value(11)": "AL",
				"itemFilter.value(12)": "AM",
				"itemFilter.value(13)": "AN",
				"itemFilter.value(14)": "AO",
				"itemFilter.value(15)": "AP",
				"itemFilter.value(16)": "AQ",
				"itemFilter.value(17)": "AR",
				"itemFilter.value(18)": "AS",
				"itemFilter.value(19)": "AT",
				"itemFilter.value(20)": "AU",
				"itemFilter.value(21)": "AV",
				"itemFilter.value(22)": "AW",
				"itemFilter.value(23)": "AX",
				"itemFilter.value(24)": "AY",
			},
		},
		{
			Name: "returns error if params contains LocatedIn itemFilter with 26 country codes",
			Params: map[string]string{
				"keywords":             "marshmallows",
				"itemFilter.name":      "LocatedIn",
				"itemFilter.value(0)":  "AA",
				"itemFilter.value(1)":  "AB",
				"itemFilter.value(2)":  "AC",
				"itemFilter.value(3)":  "AD",
				"itemFilter.value(4)":  "AE",
				"itemFilter.value(5)":  "AF",
				"itemFilter.value(6)":  "AG",
				"itemFilter.value(7)":  "AH",
				"itemFilter.value(8)":  "AI",
				"itemFilter.value(9)":  "AJ",
				"itemFilter.value(10)": "AK",
				"itemFilter.value(11)": "AL",
				"itemFilter.value(12)": "AM",
				"itemFilter.value(13)": "AN",
				"itemFilter.value(14)": "AO",
				"itemFilter.value(15)": "AP",
				"itemFilter.value(16)": "AQ",
				"itemFilter.value(17)": "AR",
				"itemFilter.value(18)": "AS",
				"itemFilter.value(19)": "AT",
				"itemFilter.value(20)": "AU",
				"itemFilter.value(21)": "AV",
				"itemFilter.value(22)": "AW",
				"itemFilter.value(23)": "AX",
				"itemFilter.value(24)": "AY",
				"itemFilter.value(25)": "AZ",
			},
			ExpectedError: ebay.ErrMaxLocatedIns,
		},
		{
			Name: "can find items if params contains LotsOnly itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LotsOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains LotsOnly itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LotsOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains LotsOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "LotsOnly",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains MaxBids itemFilter with max 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxBids",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains MaxBids itemFilter with max 5",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxBids",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MaxBids itemFilter with unparsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxBids",
				"itemFilter.value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MaxBids itemFilter with max -1",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxBids",
				"itemFilter.value": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains MinBids itemFilter with max 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinBids",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains MinBids itemFilter with max 5",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinBids",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MinBids itemFilter with unparsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinBids",
				"itemFilter.value": "not a minimum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids itemFilter with max -1",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinBids",
				"itemFilter.value": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with max 1 and min 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with min 0 and max 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with max and min 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with min and max 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with max 10 and min 5",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with min 5 and max 10",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "10",
			},
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max 0 and unparsable min",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "not a minimum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with unparsable min and max 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "not a minimum",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max 0 and min -1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min -1 and max 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "-1",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max 0 and min 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "1",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min 1 and max 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max 5 and min 10",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "10",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min 10 and max 5",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "5",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with unparsable max and min 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "not a maximum",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min 0 and unparsable max",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max -1 and min 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "-1",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min 0 and max -1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "-1",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "can find items if params contains MaxDistance itemFilter with max 5 and buyerPostalCode",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"buyerPostalCode":  "123",
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "can find items if params contains MaxDistance itemFilter with max 6 and buyerPostalCode",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"buyerPostalCode":  "123",
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "6",
			},
		},
		{
			Name: "returns error if params contains MaxDistance itemFilter with unparsable max",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"buyerPostalCode":  "123",
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "not a maximum", 5),
		},
		{
			Name: "returns error if params contains MaxDistance itemFilter with max 4 and buyerPostalCode",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"buyerPostalCode":  "123",
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "4",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "4", 5),
		},
		{
			Name: "returns error if params contains MaxDistance itemFilter with max 5 but no buyerPostalCode",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "5",
			},
			ExpectedError: ebay.ErrBuyerPostalCodeMissing,
		},
		{
			Name: "can find items if params contains MaxHandlingTime itemFilter with max 1",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxHandlingTime",
				"itemFilter.value": "1",
			},
		},
		{
			Name: "can find items if params contains MaxHandlingTime itemFilter with max 5",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxHandlingTime",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MaxHandlingTime itemFilter with max 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxHandlingTime",
				"itemFilter.value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "returns error if params contains MaxHandlingTime itemFilter with unparsable max",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxHandlingTime",
				"itemFilter.value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "not a maximum", 1),
		},
		{
			Name: "can find items if params contains MaxPrice itemFilter with max 0.0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxPrice",
				"itemFilter.value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MaxPrice itemFilter with max 5.0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxPrice",
				"itemFilter.value": "5.0",
			},
		},
		{
			Name: "can find items if params contains MaxPrice itemFilter with max 0.0, paramName Currency, and paramValue EUR",
			Params: map[string]string{
				"keywords":              "marshmallows",
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "EUR",
			},
		},
		{
			Name: "returns error if params contains MaxPrice itemFilter with unparsable max",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxPrice",
				"itemFilter.value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.ParseFloat: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MaxPrice itemFilter with max -1.0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxPrice",
				"itemFilter.value": "-1.0",
			},
			ExpectedError: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MaxPrice itemFilter with max 0.0, paramName NotCurrency, and paramValue EUR",
			Params: map[string]string{
				"keywords":              "marshmallows",
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "NotCurrency",
				"itemFilter.paramValue": "EUR",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidPriceParamName, "NotCurrency"),
		},
		{
			Name: "returns error if params contains MaxPrice itemFilter with max 0.0, paramName Currency, and paramValue ZZZ",
			Params: map[string]string{
				"keywords":              "marshmallows",
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "ZZZ",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "can find items if params contains MinPrice itemFilter with max 0.0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinPrice",
				"itemFilter.value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice itemFilter with max 5.0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinPrice",
				"itemFilter.value": "5.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice itemFilter with max 0.0, paramName Currency, and paramValue EUR",
			Params: map[string]string{
				"keywords":              "marshmallows",
				"itemFilter.name":       "MinPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "EUR",
			},
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with unparsable max",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinPrice",
				"itemFilter.value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.ParseFloat: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max -1.0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinPrice",
				"itemFilter.value": "-1.0",
			},
			ExpectedError: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max 0.0, paramName NotCurrency, and paramValue EUR",
			Params: map[string]string{
				"keywords":              "marshmallows",
				"itemFilter.name":       "MinPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "NotCurrency",
				"itemFilter.paramValue": "EUR",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidPriceParamName, "NotCurrency"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max 0.0, paramName Currency, and paramValue ZZZ",
			Params: map[string]string{
				"keywords":              "marshmallows",
				"itemFilter.name":       "MinPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "ZZZ",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with max 1.0 and min 0.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "1.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with min 0.0 and max 1.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "1.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with max and min 0.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with min and max 0.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with max 10.0 and min 5.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "10.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "5.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with min 5.0 and max 10.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "5.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "10.0",
			},
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 0.0 and unparsable min",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "not a minimum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.ParseFloat: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with unparsable min and max 0.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "not a minimum",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "0.0",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.ParseFloat: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 0.0 and min -1.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "-1.0",
			},
			ExpectedError: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min -1.0 and max 0.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "-1.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "0.0",
			},
			ExpectedError: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 0.0 and min 1.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "1.0",
			},
			ExpectedError: ebay.ErrInvalidMaxPrice,
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 1.0 and max 0.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "1.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "0.0",
			},
			ExpectedError: ebay.ErrInvalidMaxPrice,
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 5.0 and min 10.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "5.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "10.0",
			},
			ExpectedError: ebay.ErrInvalidMaxPrice,
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 10.0 and max 5.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "10.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "5.0",
			},
			ExpectedError: ebay.ErrInvalidMaxPrice,
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with unparsable max and min 0.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "not a maximum",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "0.0",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.ParseFloat: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 0.0 and unparsable max",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.ParseFloat: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max -1.0 and min 0.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "-1.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "0.0",
			},
			ExpectedError: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 0.0 and max -1.0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "-1.0",
			},
			ExpectedError: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 10.0 and min 5.0, paramName Invalid",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "10.0",
				"itemFilter(1).name":       "MinPrice",
				"itemFilter(1).value":      "5.0",
				"itemFilter(1).paramName":  "Invalid",
				"itemFilter(1).paramValue": "EUR",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidPriceParamName, "Invalid"),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 5.0, paramName Invalid and max 10.0",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramName":  "Invalid",
				"itemFilter(0).paramValue": "EUR",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "10.0",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidPriceParamName, "Invalid"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max 10.0 and min 5.0, paramValue ZZZ",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "10.0",
				"itemFilter(1).name":       "MinPrice",
				"itemFilter(1).value":      "5.0",
				"itemFilter(1).paramName":  "Currency",
				"itemFilter(1).paramValue": "ZZZ",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with min 5.0, paramValue ZZZ and max 10.0",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "ZZZ",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "10.0",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 10.0, paramName Invalid and min 5.0",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "10.0",
				"itemFilter(0).paramName":  "Invalid",
				"itemFilter(0).paramValue": "EUR",
				"itemFilter(1).name":       "MinPrice",
				"itemFilter(1).value":      "5.0",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidPriceParamName, "Invalid"),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 5.0 and max 10.0, paramName Invalid",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "10.0",
				"itemFilter(1).paramName":  "Invalid",
				"itemFilter(1).paramValue": "EUR",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidPriceParamName, "Invalid"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max 10.0, paramValue ZZZ and min 5.0",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "10.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "ZZZ",
				"itemFilter(1).name":       "MinPrice",
				"itemFilter(1).value":      "5.0",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with min 5.0 and max 10.0, paramValue ZZZ",
			Params: map[string]string{
				"keywords":                 "marshmallows",
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "10.0",
				"itemFilter(1).paramName":  "Currency",
				"itemFilter(1).paramValue": "ZZZ",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "can find items if params contains MaxQuantity itemFilter with max 1",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxQuantity",
				"itemFilter.value": "1",
			},
		},
		{
			Name: "can find items if params contains MaxQuantity itemFilter with max 5",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxQuantity",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MaxQuantity itemFilter with unparsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxQuantity",
				"itemFilter.value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MaxQuantity itemFilter with max 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MaxQuantity",
				"itemFilter.value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "can find items if params contains MinQuantity itemFilter with max 1",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinQuantity",
				"itemFilter.value": "1",
			},
		},
		{
			Name: "can find items if params contains MinQuantity itemFilter with max 5",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinQuantity",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MinQuantity itemFilter with unparsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinQuantity",
				"itemFilter.value": "not a minimum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity itemFilter with max 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "MinQuantity",
				"itemFilter.value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with max 2 and min 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "2",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with min 1 and max 2",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "2",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with max and min 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with min and max 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with max 10 and min 5",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with min 5 and max 10",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "10",
			},
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 1 and unparsable min",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "not a minimum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with unparsable min and max 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "not a minimum",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "1",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 1 and min 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 0 and max 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 1 and min 2",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "2",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 2 and max 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "2",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "1",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 5 and min 10",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "10",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 10 and max 5",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "5",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with unparsable max and min 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "not a maximum",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "1",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 1 and unparsable max",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "not a maximum",
			},
			ExpectedError: fmt.Errorf("ebay: %s: %w", `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 0 and min 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "1",
			},
			ExpectedError: fmt.Errorf("%w: %s (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 1 and max 0",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "0",
			},
			ExpectedError: fmt.Errorf("%w: %s must be greater than or equal to %s",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "can find items if params contains ModTimeFrom itemFilter with past timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ModTimeFrom",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains ModTimeFrom itemFilter with unparsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ModTimeFrom",
				"itemFilter.value": "not a timestamp",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains ModTimeFrom itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ModTimeFrom",
				"itemFilter.value": time.Now().Add(1 * time.Second).Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains ModTimeFrom itemFilter with future timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ModTimeFrom",
				"itemFilter.value": time.Now().Add(1 * time.Second).UTC().Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).UTC().Format(time.RFC3339)),
		},
		{
			Name: "can find items if params contains ReturnsAcceptedOnly itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ReturnsAcceptedOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains ReturnsAcceptedOnly itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ReturnsAcceptedOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains ReturnsAcceptedOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "ReturnsAcceptedOnly",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains Seller itemFilter with seller ID 0",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "Seller",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains Seller itemFilter with seller IDs 0 and 1",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "Seller",
				"itemFilter.value(0)": "0",
				"itemFilter.value(1)": "1",
			},
		},
		{
			Name: "returns error if params contains Seller and ExcludeSeller itemFilters",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "Seller",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "ExcludeSeller",
				"itemFilter(1).value": "0",
			},
			ExpectedError: ebay.ErrSellerCannotBeUsedWithOtherSellers,
		},
		{
			Name: "returns error if params contains Seller and TopRatedSellerOnly itemFilters",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter(0).name":  "Seller",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "TopRatedSellerOnly",
				"itemFilter(1).value": "true",
			},
			ExpectedError: ebay.ErrSellerCannotBeUsedWithOtherSellers,
		},
		{
			Name:   "can find items if params contains Seller itemFilter with 100 seller IDs",
			Params: generateFilterParams("Seller", 100),
		},
		{
			Name:          "returns error if params contains Seller itemFilter with 101 seller IDs",
			Params:        generateFilterParams("Seller", 101),
			ExpectedError: ebay.ErrMaxSellers,
		},
		{
			Name: "can find items if params contains SellerBusinessType itemFilter with Business type",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "SellerBusinessType",
				"itemFilter.value": "Business",
			},
		},
		{
			Name: "can find items if params contains SellerBusinessType itemFilter with Private type",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "SellerBusinessType",
				"itemFilter.value": "Private",
			},
		},
		{
			Name: "returns error if params contains SellerBusinessType itemFilter with NotBusiness type",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "SellerBusinessType",
				"itemFilter.value": "NotBusiness",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidSellerBusinessType, "NotBusiness"),
		},
		{
			Name: "returns error if params contains SellerBusinessType itemFilter with Business and Private types",
			Params: map[string]string{
				"keywords":            "marshmallows",
				"itemFilter.name":     "SellerBusinessType",
				"itemFilter.value(0)": "Business",
				"itemFilter.value(1)": "Private",
			},
			ExpectedError: ebay.ErrMultipleSellerBusinessTypes,
		},
		{
			Name: "can find items if params contains SoldItemsOnly itemFilter.value=true",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "SoldItemsOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains SoldItemsOnly itemFilter.value=false",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "SoldItemsOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains SoldItemsOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "SoldItemsOnly",
				"itemFilter.value": "123",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains StartTimeFrom itemFilter with future timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "StartTimeFrom",
				"itemFilter.value": time.Now().Add(1 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains StartTimeFrom itemFilter with unparsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "StartTimeFrom",
				"itemFilter.value": "not a timestamp",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains StartTimeFrom itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "StartTimeFrom",
				"itemFilter.value": time.Now().Add(1 * time.Second).Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains StartTimeFrom itemFilter with past timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "StartTimeFrom",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(-1*time.Second).UTC().Format(time.RFC3339)),
		},
		{
			Name: "can find items if params contains StartTimeTo itemFilter with future timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "StartTimeTo",
				"itemFilter.value": time.Now().Add(1 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains StartTimeTo itemFilter with unparsable value",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "StartTimeTo",
				"itemFilter.value": "not a timestamp",
			},
			ExpectedError: fmt.Errorf("%w: %s", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains StartTimeTo itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "StartTimeTo",
				"itemFilter.value": time.Now().Add(1 * time.Second).Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains StartTimeTo itemFilter with past timestamp",
			Params: map[string]string{
				"keywords":         "marshmallows",
				"itemFilter.name":  "StartTimeTo",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
			ExpectedError: fmt.Errorf("%w: %s",
				ebay.ErrInvalidDateTime, time.Now().Add(-1*time.Second).UTC().Format(time.RFC3339)),
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

func generateFilterParams(filterName string, count int) map[string]string {
	params := make(map[string]string)
	params["keywords"] = "marshmallows"
	params["itemFilter.name"] = filterName

	for i := 0; i < count; i++ {
		params[fmt.Sprintf("itemFilter.value(%d)", i)] = strconv.Itoa(i)
	}

	return params
}
