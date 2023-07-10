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

var ErrClientFailure = errors.New("http: client failed")

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
	appID := "super secret ID"
	searchResp := ebay.SearchResponse{
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

	t.Run("returns error if params does not contain keywords", func(t *testing.T) {
		t.Parallel()
		client := &MockFindingClient{}
		svr := ebay.NewFindingServer(client)
		emptyParams := make(map[string]string)
		_, err := svr.FindItemsByKeywords(emptyParams, appID)
		assertError(t, err)

		expected := ebay.ErrKeywordsMissing.Error()
		got := err.Error()
		assertErrorEquals(t, got, expected)
		assertStatusCodeEquals(t, err, http.StatusBadRequest)
	})

	t.Run("can find items by keywords and aspectFilter", func(t *testing.T) {
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
		aspectFilterParams := map[string]string{
			"keywords":                     "marshmallows",
			"aspectFilter.aspectName":      "squish level",
			"aspectFilter.aspectValueName": "very squishy",
		}
		resp, err := svr.FindItemsByKeywords(aspectFilterParams, appID)
		assertNoError(t, err)
		assertSearchResponse(t, resp, &searchResp)
	})

	t.Run("returns error if params contains aspectName but not aspectValueName", func(t *testing.T) {
		t.Parallel()
		client := &MockFindingClient{}
		svr := ebay.NewFindingServer(client)
		aspectFilterParams := map[string]string{
			"keywords":                "marshmallows",
			"aspectFilter.aspectName": "squish level",
		}
		_, err := svr.FindItemsByKeywords(aspectFilterParams, appID)
		assertError(t, err)

		expected := ebay.ErrIncompleteAspectFilter.Error()
		got := err.Error()
		assertErrorEquals(t, got, expected)
		assertStatusCodeEquals(t, err, http.StatusBadRequest)
	})

	t.Run("can find items by keywords and basic itemFilter", func(t *testing.T) {
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
		itemFilterParams := map[string]string{
			"keywords":            "marshmallows",
			"itemFilter(0).name":  "color",
			"itemFilter(0).value": "purple",
		}
		resp, err := svr.FindItemsByKeywords(itemFilterParams, appID)
		assertNoError(t, err)
		assertSearchResponse(t, resp, &searchResp)
	})

	t.Run("can find items by keywords and itemFilter with name, value, paramName, and paramValue", func(t *testing.T) {
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
		itemFilterParams := map[string]string{
			"keywords":                 "marshmallows",
			"itemFilter(0).name":       "color",
			"itemFilter(0).value":      "purple",
			"itemFilter(0).paramName":  "brightness",
			"itemFilter(0).paramValue": "super bright",
		}
		resp, err := svr.FindItemsByKeywords(itemFilterParams, appID)
		assertNoError(t, err)
		assertSearchResponse(t, resp, &searchResp)
	})

	t.Run("can find items by keywords and 2 basic itemFilters", func(t *testing.T) {
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
		itemFilterParams := map[string]string{
			"keywords":            "marshmallows",
			"itemFilter(0).name":  "color",
			"itemFilter(0).value": "purple",
			"itemFilter(1).name":  "size",
			"itemFilter(1).value": "small",
		}
		resp, err := svr.FindItemsByKeywords(itemFilterParams, appID)
		assertNoError(t, err)
		assertSearchResponse(t, resp, &searchResp)
	})

	t.Run("can find items by keywords, 1st advanced itemFilter, and 2nd basic itemFilter", func(t *testing.T) {
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
		itemFilterParams := map[string]string{
			"keywords":                 "marshmallows",
			"itemFilter(0).name":       "color",
			"itemFilter(0).value":      "purple",
			"itemFilter(0).paramName":  "brightness",
			"itemFilter(0).paramValue": "super bright",
			"itemFilter(1).name":       "size",
			"itemFilter(1).value":      "small",
		}
		resp, err := svr.FindItemsByKeywords(itemFilterParams, appID)
		assertNoError(t, err)
		assertSearchResponse(t, resp, &searchResp)
	})

	t.Run("can find items by keywords, 1st basic itemFilter, and 2nd advanced itemFilter", func(t *testing.T) {
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
		itemFilterParams := map[string]string{
			"keywords":                 "marshmallows",
			"itemFilter(0).name":       "size",
			"itemFilter(0).value":      "small",
			"itemFilter(1).name":       "color",
			"itemFilter(1).value":      "purple",
			"itemFilter(1).paramName":  "brightness",
			"itemFilter(1).paramValue": "super bright",
		}
		resp, err := svr.FindItemsByKeywords(itemFilterParams, appID)
		assertNoError(t, err)
		assertSearchResponse(t, resp, &searchResp)
	})

	t.Run("can find items by keywords and 2 advanced itemFilters", func(t *testing.T) {
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
		itemFilterParams := map[string]string{
			"keywords":                 "marshmallows",
			"itemFilter(0).name":       "color",
			"itemFilter(0).value":      "purple",
			"itemFilter(0).paramName":  "brightness",
			"itemFilter(0).paramValue": "super bright",
			"itemFilter(1).name":       "size",
			"itemFilter(1).value":      "small",
			"itemFilter(1).paramName":  "width",
			"itemFilter(1).paramValue": "wide",
		}
		resp, err := svr.FindItemsByKeywords(itemFilterParams, appID)
		assertNoError(t, err)
		assertSearchResponse(t, resp, &searchResp)
	})

	t.Run("returns error if params contains itemFilter name but not value", func(t *testing.T) {
		t.Parallel()
		client := &MockFindingClient{}
		svr := ebay.NewFindingServer(client)
		itemFilterParams := map[string]string{
			"keywords":           "marshmallows",
			"itemFilter(0).name": "color",
		}
		_, err := svr.FindItemsByKeywords(itemFilterParams, appID)
		assertError(t, err)

		expected := ebay.ErrIncompleteItemFilterNameOnly.Error()
		got := err.Error()
		assertErrorEquals(t, got, expected)
		assertStatusCodeEquals(t, err, http.StatusBadRequest)
	})

	t.Run("returns error if params contains itemFilter paramName but not paramValue", func(t *testing.T) {
		t.Parallel()
		client := &MockFindingClient{}
		svr := ebay.NewFindingServer(client)
		itemFilterParams := map[string]string{
			"keywords":                "marshmallows",
			"itemFilter(0).name":      "color",
			"itemFilter(0).value":     "purple",
			"itemFilter(0).paramName": "brightness",
		}
		_, err := svr.FindItemsByKeywords(itemFilterParams, appID)
		assertError(t, err)

		expected := ebay.ErrIncompleteItemFilterParam.Error()
		got := err.Error()
		assertErrorEquals(t, got, expected)
		assertStatusCodeEquals(t, err, http.StatusBadRequest)
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
