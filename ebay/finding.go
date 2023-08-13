package ebay

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	findingURL                       = "https://svcs.ebay.com/services/search/FindingService/v1?REST-PAYLOAD"
	findItemsByKeywordsOperationName = "findItemsByKeywords"
	findItemsByCategoryOperationName = "findItemsByCategory"
	findItemsAdvancedOperationName   = "findItemsAdvanced"
	findingServiceVersion            = "1.0.0"
	findingResponseDataFormat        = "JSON"
)

var (
	// ErrCategoryIDMissing is returned when the 'categoryId' parameter is missing in a findItemsByCategory request.
	ErrCategoryIDMissing = errors.New("ebay: category ID parameter is missing")

	// ErrCategoryIDKeywordsMissing is returned when the 'categoryId' and 'keywords' parameters
	// are missing in a findItemsAdvanced request.
	ErrCategoryIDKeywordsMissing = errors.New("ebay: both category ID and keywords parameters are missing")

	// ErrProductIDMissing is returned when the 'productId' or 'productId.@type' parameters
	// are missing in a findItemsByProduct request.
	ErrProductIDMissing = errors.New("ebay: product ID parameter or product ID type are missing")

	maxCategoryIDs = 3

	// ErrMaxCategoryIDs is returned when the 'categoryId' parameter contains more category IDs than the maximum allowed.
	ErrMaxCategoryIDs = fmt.Errorf("ebay: maximum category IDs to specify is %d", maxCategoryIDs)

	maxCategoryIDLen = 10

	// ErrInvalidCategoryIDLength is returned when an individual category ID in the 'categoryId' parameter
	// exceed the maximum length of 10 characters or is empty.
	ErrInvalidCategoryIDLength = fmt.Errorf(
		"ebay: invalid category ID length: must be between 1 and %d characters", maxCategoryIDLen)

	// ErrKeywordsMissing is returned when the 'keywords' parameter is missing.
	ErrKeywordsMissing = errors.New("ebay: keywords parameter is missing")

	minKeywordsLen, maxKeywordsLen = 2, 350

	// ErrInvalidKeywordsLength is returned when the 'keywords' parameter as a whole
	// exceeds the maximum length of 350 characters or has a length less than 2 characters.
	ErrInvalidKeywordsLength = fmt.Errorf(
		"ebay: invalid keywords length: must be between %d and %d characters", minKeywordsLen, maxKeywordsLen)

	maxKeywordLen = 98

	// ErrInvalidKeywordLength is returned when an individual keyword in the 'keywords' parameter
	// exceeds the maximum length of 98 characters.
	ErrInvalidKeywordLength = fmt.Errorf("ebay: invalid keyword length: must be no more than %d characters", maxKeywordLen)

	// ErrInvalidProductIDLength is returned when the 'productId' parameter is empty.
	ErrInvalidProductIDLength = errors.New("ebay: invalid product ID length")

	isbnShortLen, isbnLongLen = 10, 13

	// ErrInvalidISBNLength is returned when the 'productId.type' parameter is an ISBN (International Standard Book Number)
	// and the 'productId' parameter is not exactly 10 or 13 characters.
	ErrInvalidISBNLength = fmt.Errorf("ebay: invalid ISBN length: must be either %d or %d characters",
		isbnShortLen, isbnLongLen)

	// ErrInvalidISBN is returned when the 'productId.type' parameter is an ISBN (International Standard Book Number)
	// and the 'productId' parameter contains an invalid ISBN.
	ErrInvalidISBN = errors.New("ebay: invalid ISBN")

	upcLen = 12

	// ErrInvalidUPCLength is returned when the 'productId.type' parameter is a UPC (Universal Product Code)
	// and the 'productId' parameter is not 12 digits.
	ErrInvalidUPCLength = fmt.Errorf("ebay: invalid UPC length: must be %d digits", upcLen)

	// ErrInvalidUPC is returned when the 'productId.type' parameter is a UPC (Universal Product Code)
	// and the 'productId' parameter contains an invalid UPC.
	ErrInvalidUPC = errors.New("ebay: invalid UPC")

	eanShortLen, eanLongLen = 8, 13

	// ErrInvalidEANLength is returned when the 'productId.type' parameter is an EAN (European Article Number)
	// and the 'productId' parameter is not exactly 8 or 13 characters.
	ErrInvalidEANLength = fmt.Errorf("ebay: invalid EAN length: must be either %d or %d characters",
		eanShortLen, eanLongLen)

	// ErrInvalidEAN is returned when the 'productId.type' parameter is an EAN (European Article Number)
	// and the 'productId' parameter contains an invalid EAN.
	ErrInvalidEAN = errors.New("ebay: invalid EAN")

	// ErrUnsupportedProductIDType is returned when the 'productId.type' parameter has an unsupported type.
	ErrUnsupportedProductIDType = errors.New("ebay: unsupported product ID type")

	// ErrInvalidFilterSyntax is returned when both syntax types for filters are used in the params.
	ErrInvalidFilterSyntax = errors.New("ebay: invalid filter syntax: both syntax types are present")

	// ErrIncompleteFilterNameOnly is returned when a filter is missing the 'value' parameter.
	ErrIncompleteFilterNameOnly = errors.New("ebay: incomplete item filter: missing")

	// ErrIncompleteItemFilterParam is returned when an item filter is missing
	// either the 'paramName' or 'paramValue' parameter, as both 'paramName' and 'paramValue'
	// are required when either one is specified.
	ErrIncompleteItemFilterParam = errors.New(
		"ebay: incomplete item filter: both paramName and paramValue must be specified together")

	// ErrFailedRequest is returned when the eBay Finding API request fails.
	ErrFailedRequest = errors.New("ebay: failed to perform eBay Finding API request")

	// ErrInvalidStatus is returned when the eBay Finding API request returns an invalid status code.
	ErrInvalidStatus = errors.New("ebay: failed to perform eBay Finding API request with status code")

	// ErrDecodeAPIResponse is returned when there is an error decoding the eBay Finding API response body.
	ErrDecodeAPIResponse = errors.New("ebay: failed to decode eBay Finding API response body")

	// ErrInvalidBooleanValue is returned when a parameter has an invalid boolean value.
	ErrInvalidBooleanValue = errors.New("ebay: invalid boolean value, allowed values are true and false")

	// ErrUnsupportedItemFilterType is returned when an item filter 'name' parameter has an unsupported type.
	ErrUnsupportedItemFilterType = errors.New("ebay: unsupported item filter type")

	// ErrInvalidCountryCode is returned when an item filter 'values' parameter contains an invalid country code.
	ErrInvalidCountryCode = errors.New("ebay: invalid country code")

	// ErrInvalidCondition is returned when an item filter 'values' parameter contains an invalid condition ID or name.
	ErrInvalidCondition = errors.New("ebay: invalid condition")

	// ErrInvalidCurrencyID is returned when an item filter 'values' parameter contains an invalid currency ID.
	ErrInvalidCurrencyID = errors.New("ebay: invalid currency ID")

	// ErrInvalidDateTime is returned when an item filter 'values' parameter contains an invalid date time.
	ErrInvalidDateTime = errors.New("ebay: invalid date time value")

	maxExcludeCategories = 25

	// ErrMaxExcludeCategories is returned when an item filter 'values' parameter
	// contains more categories to exclude than the maximum allowed.
	ErrMaxExcludeCategories = fmt.Errorf("ebay: maximum categories to exclude is %d", maxExcludeCategories)

	maxExcludeSellers = 100

	// ErrMaxExcludeSellers is returned when an item filter 'values' parameter
	// contains more categories to exclude than the maximum allowed.
	ErrMaxExcludeSellers = fmt.Errorf("ebay: maximum sellers to exclude is %d", maxExcludeSellers)

	// ErrExcludeSellerCannotBeUsedWithSellers is returned when there is an attempt to use
	// the ExcludeSeller item filter together with either the Seller or TopRatedSellerOnly item filters.
	ErrExcludeSellerCannotBeUsedWithSellers = errors.New(
		"ebay: ExcludeSeller item filter cannot be used together with either the Seller or TopRatedSellerOnly item filters")

	// ErrInvalidInteger is returned when an item filter 'values' parameter contains an invalid integer.
	ErrInvalidInteger = errors.New("ebay: invalid integer")

	// ErrInvalidNumericFilter is returned when a numeric item filter is invalid.
	ErrInvalidNumericFilter = errors.New("ebay: invalid numeric item filter")

	// ErrInvalidExpeditedShippingType is returned when an item filter 'values' parameter
	// contains an invalid expedited shipping type.
	ErrInvalidExpeditedShippingType = errors.New("ebay: invalid expedited shipping type")

	// ErrInvalidGlobalID is returned when an item filter 'values' parameter contains an invalid global ID.
	ErrInvalidGlobalID = errors.New("ebay: invalid global ID")

	// ErrInvalidAllListingType is returned when an item filter 'values' parameter
	// contains the 'All' listing type and other listing types.
	ErrInvalidAllListingType = errors.New("ebay: 'All' listing type cannot be combined with other listing types")

	// ErrInvalidListingType is returned when an item filter 'values' parameter contains an invalid listing type.
	ErrInvalidListingType = errors.New("ebay: invalid listing type")

	// ErrDuplicateListingType is returned when an item filter 'values' parameter contains duplicate listing types.
	ErrDuplicateListingType = errors.New("ebay: duplicate listing type")

	// ErrInvalidAuctionListingTypes is returned when an item filter 'values' parameter
	// contains both 'Auction' and 'AuctionWithBIN' listing types.
	ErrInvalidAuctionListingTypes = errors.New("ebay: 'Auction' and 'AuctionWithBIN' listing types cannot be combined")

	// ErrBuyerPostalCodeMissing is returned when the LocalSearchOnly, MaxDistance item filter,
	// or DistanceNearest sortOrder is used, but the buyerPostalCode parameter is missing in the request.
	ErrBuyerPostalCodeMissing = errors.New("ebay: buyerPostalCode is missing")

	// ErrMaxDistanceMissing is returned when the LocalSearchOnly item filter is used,
	// but the MaxDistance item filter is missing in the request.
	ErrMaxDistanceMissing = errors.New("ebay: MaxDistance item filter is missing when using LocalSearchOnly item filter")

	maxLocatedIns = 25

	// ErrMaxLocatedIns is returned when an item filter 'values' parameter
	// contains more countries to locate items in than the maximum allowed.
	ErrMaxLocatedIns = fmt.Errorf("ebay: maximum countries to locate items in is %d", maxLocatedIns)

	// ErrInvalidPrice is returned when an item filter 'values' parameter contains an invalid price.
	ErrInvalidPrice = errors.New("ebay: invalid price")

	// ErrInvalidPriceParamName is returned when an item filter 'paramName' parameter
	// contains anything other than "Currency".
	ErrInvalidPriceParamName = errors.New(`ebay: invalid price parameter name, must be "Currency"`)

	// ErrInvalidMaxPrice is returned when an item filter 'values' parameter
	// contains a maximum price less than a minimum price.
	ErrInvalidMaxPrice = errors.New("ebay: maximum price must be greater than or equal to minimum price")

	maxSellers = 100

	// ErrMaxSellers is returned when an item filter 'values' parameter
	// contains more categories to include than the maximum allowed.
	ErrMaxSellers = fmt.Errorf("ebay: maximum sellers to include is %d", maxExcludeSellers)

	// ErrSellerCannotBeUsedWithOtherSellers is returned when there is an attempt to use
	// the Seller item filter together with either the ExcludeSeller or TopRatedSellerOnly item filters.
	ErrSellerCannotBeUsedWithOtherSellers = errors.New(
		"ebay: Seller item filter cannot be used together with either the ExcludeSeller or TopRatedSellerOnly item filters")

	// ErrMultipleSellerBusinessTypes is returned when an item filter 'values' parameter
	// contains multiple seller business types.
	ErrMultipleSellerBusinessTypes = errors.New("ebay: multiple seller business types found")

	// ErrInvalidSellerBusinessType is returned when an item filter 'values' parameter
	// contains an invalid seller business type.
	ErrInvalidSellerBusinessType = errors.New("ebay: invalid seller business type")

	// ErrTopRatedSellerCannotBeUsedWithSellers is returned when there is an attempt to use
	// the TopRatedSellerOnly item filter together with either the Seller or ExcludeSeller item filters.
	ErrTopRatedSellerCannotBeUsedWithSellers = errors.New(
		"ebay: TopRatedSellerOnly item filter cannot be used together with either the Seller or ExcludeSeller item filters")

	// ErrInvalidValueBoxInventory is returned when an item filter 'values' parameter
	// contains an invalid value box inventory.
	ErrInvalidValueBoxInventory = errors.New("ebay: invalid value box inventory")

	// ErrInvalidOutputSelector is returned when the 'outputSelector' parameter contains an invalid output selector.
	ErrInvalidOutputSelector = errors.New("ebay: invalid output selector")

	maxCustomIDLen = 256

	// ErrInvalidCustomIDLength is returned when the 'affiliate.customId' parameter
	// exceeds the maximum length of 256 characters.
	ErrInvalidCustomIDLength = fmt.Errorf(
		"ebay: invalid affiliate custom ID length: must be no more than %d characters", maxCustomIDLen)

	// ErrIncompleteAffiliateParams is returned when an affiliate is missing
	// either the 'networkId' or 'trackingId' parameter, as both 'networkId' and 'trackingId'
	// are required when either one is specified.
	ErrIncompleteAffiliateParams = errors.New(
		"ebay: incomplete affiliate: both network and tracking IDs must be specified together")

	beFreeID, ebayPartnerNetworkID = 2, 9

	// ErrInvalidNetworkID is returned when the 'affiliate.networkId' parameter
	// is outside the valid range of 2 (Be Free) and 9 (eBay Partner Network).
	ErrInvalidNetworkID = fmt.Errorf("ebay: invalid affiliate network ID: must be between %d and %d",
		beFreeID, ebayPartnerNetworkID)

	// ErrInvalidCampaignID is returned when the 'affiliate.networkId' parameter is 9 (eBay Partner Network)
	// and the 'affiliate.trackingId' parameter is not a 10-digit number (eBay Partner Network's Campaign ID).
	ErrInvalidCampaignID = errors.New("ebay: invalid affiliate Campaign ID length: must be a 10-digit number")

	// ErrInvalidEntriesPerPage is returned when the 'buyerPostalCode' parameter contains an invalid postal code.
	ErrInvalidPostalCode = errors.New("ebay: invalid postal code")

	minPaginationValue, maxPaginationValue = 1, 100

	// ErrInvalidEntriesPerPage is returned when the 'paginationInput.entriesPerPage' parameter
	// is outside the valid range of 1 to 100.
	ErrInvalidEntriesPerPage = fmt.Errorf("ebay: invalid pagination entries per page, must be between %d and %d",
		minPaginationValue, maxPaginationValue)

	// ErrInvalidPageNumber is returned when the 'paginationInput.pageNumber' parameter
	// is outside the valid range of 1 to 100.
	ErrInvalidPageNumber = fmt.Errorf("ebay: invalid pagination page number, must be between %d and %d",
		minPaginationValue, maxPaginationValue)

	// ErrAuctionListingMissing is returned when the 'sortOrder' parameter BidCountFewest or BidCountMost,
	// but a 'Auction' listing type is not specified in the item filters.
	ErrAuctionListingMissing = errors.New("ebay: 'Auction' listing type required for sorting by bid count")

	// ErrUnsupportedSortOrderType is returned when the 'sortOrder' parameter has an unsupported type.
	ErrUnsupportedSortOrderType = errors.New("ebay: invalid sort order type")
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

// FindItemsByCategories searches the eBay Finding API using the provided category, additional parameters,
// and a valid eBay application ID.
func (svr *FindingServer) FindItemsByCategories(
	params map[string]string, appID string,
) (FindItemsByCategoriesResponse, error) {
	resp, err := svr.requestItems(params, &findItemsByCategoryParams{}, appID)
	if err != nil {
		return FindItemsByCategoriesResponse{}, err
	}

	itemsResp, err := svr.parseFindItemsByCategoryResponse(resp)
	if err != nil {
		return FindItemsByCategoriesResponse{}, &APIError{Err: err.Error(), StatusCode: http.StatusInternalServerError}
	}

	return itemsResp, nil
}

// FindItemsByKeywords searches the eBay Finding API using the provided keywords, additional parameters,
// and a valid eBay application ID.
func (svr *FindingServer) FindItemsByKeywords(
	params map[string]string, appID string,
) (FindItemsByKeywordsResponse, error) {
	resp, err := svr.requestItems(params, &findItemsByKeywordsParams{}, appID)
	if err != nil {
		return FindItemsByKeywordsResponse{}, err
	}

	itemsResp, err := svr.parseFindItemsByKeywordsResponse(resp)
	if err != nil {
		return FindItemsByKeywordsResponse{}, &APIError{Err: err.Error(), StatusCode: http.StatusInternalServerError}
	}

	return itemsResp, nil
}

// FindItemsAdvanced searches the eBay Finding API using the provided category and/or keywords, additional parameters,
// and a valid eBay application ID.
func (svr *FindingServer) FindItemsAdvanced(
	params map[string]string, appID string,
) (FindItemsAdvancedResponse, error) {
	resp, err := svr.requestItems(params, &findItemsAdvancedParams{}, appID)
	if err != nil {
		return FindItemsAdvancedResponse{}, err
	}

	itemsResp, err := svr.parseFindItemsAdvancedResponse(resp)
	if err != nil {
		return FindItemsAdvancedResponse{}, &APIError{Err: err.Error(), StatusCode: http.StatusInternalServerError}
	}

	return itemsResp, nil
}

// FindItemsByProduct searches the eBay Finding API using the provided product, additional parameters,
// and a valid eBay application ID.
func (svr *FindingServer) FindItemsByProduct(
	params map[string]string, appID string,
) (FindItemsByProductResponse, error) {
	resp, err := svr.requestItems(params, &findItemsByProductParams{}, appID)
	if err != nil {
		return FindItemsByProductResponse{}, err
	}

	itemsResp, err := svr.parseFindItemsByProductResponse(resp)
	if err != nil {
		return FindItemsByProductResponse{}, &APIError{Err: err.Error(), StatusCode: http.StatusInternalServerError}
	}

	return itemsResp, nil
}

func (svr *FindingServer) requestItems(
	params map[string]string, fParams findItemsParams, appID string,
) (*http.Response, error) {
	err := fParams.validateParams(params)
	if err != nil {
		return nil, &APIError{Err: err.Error(), StatusCode: http.StatusBadRequest}
	}

	req, err := fParams.createRequest(appID)
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

	return resp, nil
}

type findItemsParams interface {
	validateParams(params map[string]string) error
	createRequest(appID string) (*http.Request, error)
}

type findItemsByCategoryParams struct {
	aspectFilters   []aspectFilter
	categoryIDs     string
	itemFilters     []itemFilter
	outputSelectors []string
	affiliate       *affiliate
	buyerPostalCode *string
	paginationInput *paginationInput
	sortOrder       *string
}

type aspectFilter struct {
	aspectName       string
	aspectValueNames []string
}

type itemFilter struct {
	name       string
	values     []string
	paramName  *string
	paramValue *string
}

type affiliate struct {
	customID     *string
	geoTargeting *string
	networkID    *string
	trackingID   *string
}

type paginationInput struct {
	entriesPerPage *string
	pageNumber     *string
}

func (fp *findItemsByCategoryParams) validateParams(params map[string]string) error {
	categoryID, ok := params["categoryId"]
	if !ok {
		return ErrCategoryIDMissing
	}
	err := processCategoryIDs(categoryID)
	if err != nil {
		return err
	}
	fp.categoryIDs = categoryID

	aspectFilters, err := processAspectFilters(params)
	if err != nil {
		return err
	}
	fp.aspectFilters = aspectFilters

	itemFilters, err := processItemFilters(params)
	if err != nil {
		return err
	}
	fp.itemFilters = itemFilters

	outputSelectors, err := processOutputSelectors(params)
	if err != nil {
		return err
	}
	fp.outputSelectors = outputSelectors

	affiliate, err := processAffiliate(params)
	if err != nil {
		return err
	}
	fp.affiliate = affiliate

	buyerPostalCode, ok := params["buyerPostalCode"]
	if ok {
		if !isValidPostalCode(buyerPostalCode) {
			return ErrInvalidPostalCode
		}
		fp.buyerPostalCode = &buyerPostalCode
	}

	paginationInput, err := processPaginationInput(params)
	if err != nil {
		return err
	}
	fp.paginationInput = paginationInput

	sortOrder, ok := params["sortOrder"]
	if ok {
		err := validateSortOrder(sortOrder, fp.itemFilters, fp.buyerPostalCode != nil)
		if err != nil {
			return err
		}
		fp.sortOrder = &sortOrder
	}

	return nil
}

func (fp *findItemsByCategoryParams) createRequest(appID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, findingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ebay: %w", err)
	}

	qry := req.URL.Query()
	qry.Add("OPERATION-NAME", findItemsByCategoryOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)

	for i, f := range fp.aspectFilters {
		qry.Add(fmt.Sprintf("aspectFilter(%d).aspectName", i), f.aspectName)
		for j, v := range f.aspectValueNames {
			qry.Add(fmt.Sprintf("aspectFilter(%d).aspectValueName(%d)", i, j), v)
		}
	}

	qry.Add("categoryId", fp.categoryIDs)
	for i, f := range fp.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", i), f.name)
		for j, v := range f.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", i, j), v)
		}

		if f.paramName != nil && f.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", i), *f.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", i), *f.paramValue)
		}
	}

	for i := range fp.outputSelectors {
		qry.Add(fmt.Sprintf("outputSelector(%d)", i), fp.outputSelectors[i])
	}

	if fp.affiliate != nil {
		if fp.affiliate.customID != nil {
			qry.Add("affiliate.customId", *fp.affiliate.customID)
		}
		if fp.affiliate.geoTargeting != nil {
			qry.Add("affiliate.geoTargeting", *fp.affiliate.geoTargeting)
		}
		if fp.affiliate.networkID != nil {
			qry.Add("affiliate.networkId", *fp.affiliate.networkID)
		}
		if fp.affiliate.trackingID != nil {
			qry.Add("affiliate.trackingId", *fp.affiliate.trackingID)
		}
	}
	if fp.buyerPostalCode != nil {
		qry.Add("buyerPostalCode", *fp.buyerPostalCode)
	}
	if fp.paginationInput != nil {
		if fp.paginationInput.entriesPerPage != nil {
			qry.Add("paginationInput.entriesPerPage", *fp.paginationInput.entriesPerPage)
		}
		if fp.paginationInput.pageNumber != nil {
			qry.Add("paginationInput.pageNumber", *fp.paginationInput.pageNumber)
		}
	}
	if fp.sortOrder != nil {
		qry.Add("sortOrder", *fp.sortOrder)
	}

	req.URL.RawQuery = qry.Encode()

	return req, nil
}

type findItemsByKeywordsParams struct {
	aspectFilters   []aspectFilter
	itemFilters     []itemFilter
	keywords        string
	outputSelectors []string
	affiliate       *affiliate
	buyerPostalCode *string
	paginationInput *paginationInput
	sortOrder       *string
}

func (fp *findItemsByKeywordsParams) validateParams(params map[string]string) error {
	keywords, err := processKeywords(params)
	if err != nil {
		return err
	}
	fp.keywords = keywords

	aspectFilters, err := processAspectFilters(params)
	if err != nil {
		return err
	}
	fp.aspectFilters = aspectFilters

	itemFilters, err := processItemFilters(params)
	if err != nil {
		return err
	}
	fp.itemFilters = itemFilters

	outputSelectors, err := processOutputSelectors(params)
	if err != nil {
		return err
	}
	fp.outputSelectors = outputSelectors

	affiliate, err := processAffiliate(params)
	if err != nil {
		return err
	}
	fp.affiliate = affiliate

	buyerPostalCode, ok := params["buyerPostalCode"]
	if ok {
		if !isValidPostalCode(buyerPostalCode) {
			return ErrInvalidPostalCode
		}
		fp.buyerPostalCode = &buyerPostalCode
	}

	paginationInput, err := processPaginationInput(params)
	if err != nil {
		return err
	}
	fp.paginationInput = paginationInput

	sortOrder, ok := params["sortOrder"]
	if ok {
		err := validateSortOrder(sortOrder, fp.itemFilters, fp.buyerPostalCode != nil)
		if err != nil {
			return err
		}
		fp.sortOrder = &sortOrder
	}

	return nil
}

func (fp *findItemsByKeywordsParams) createRequest(appID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, findingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ebay: %w", err)
	}

	qry := req.URL.Query()
	qry.Add("OPERATION-NAME", findItemsByKeywordsOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)

	for i, f := range fp.aspectFilters {
		qry.Add(fmt.Sprintf("aspectFilter(%d).aspectName", i), f.aspectName)
		for j, v := range f.aspectValueNames {
			qry.Add(fmt.Sprintf("aspectFilter(%d).aspectValueName(%d)", i, j), v)
		}
	}

	for i, f := range fp.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", i), f.name)
		for j, v := range f.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", i, j), v)
		}

		if f.paramName != nil && f.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", i), *f.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", i), *f.paramValue)
		}
	}

	qry.Add("keywords", fp.keywords)
	for i := range fp.outputSelectors {
		qry.Add(fmt.Sprintf("outputSelector(%d)", i), fp.outputSelectors[i])
	}

	if fp.affiliate != nil {
		if fp.affiliate.customID != nil {
			qry.Add("affiliate.customId", *fp.affiliate.customID)
		}
		if fp.affiliate.geoTargeting != nil {
			qry.Add("affiliate.geoTargeting", *fp.affiliate.geoTargeting)
		}
		if fp.affiliate.networkID != nil {
			qry.Add("affiliate.networkId", *fp.affiliate.networkID)
		}
		if fp.affiliate.trackingID != nil {
			qry.Add("affiliate.trackingId", *fp.affiliate.trackingID)
		}
	}
	if fp.buyerPostalCode != nil {
		qry.Add("buyerPostalCode", *fp.buyerPostalCode)
	}
	if fp.paginationInput != nil {
		if fp.paginationInput.entriesPerPage != nil {
			qry.Add("paginationInput.entriesPerPage", *fp.paginationInput.entriesPerPage)
		}
		if fp.paginationInput.pageNumber != nil {
			qry.Add("paginationInput.pageNumber", *fp.paginationInput.pageNumber)
		}
	}
	if fp.sortOrder != nil {
		qry.Add("sortOrder", *fp.sortOrder)
	}

	req.URL.RawQuery = qry.Encode()

	return req, nil
}

type findItemsAdvancedParams struct {
	aspectFilters     []aspectFilter
	categoryIDs       *string
	descriptionSearch *string
	itemFilters       []itemFilter
	keywords          *string
	outputSelectors   []string
	affiliate         *affiliate
	buyerPostalCode   *string
	paginationInput   *paginationInput
	sortOrder         *string
}

func (fp *findItemsAdvancedParams) validateParams(params map[string]string) error {
	categoryID, cOk := params["categoryId"]
	_, ok := params["keywords"]
	if !cOk && !ok {
		return ErrCategoryIDKeywordsMissing
	}

	if cOk {
		err := processCategoryIDs(categoryID)
		if err != nil {
			return err
		}
		fp.categoryIDs = &categoryID
	}
	if ok {
		keywords, err := processKeywords(params)
		if err != nil {
			return err
		}
		fp.keywords = &keywords
	}

	aspectFilters, err := processAspectFilters(params)
	if err != nil {
		return err
	}
	fp.aspectFilters = aspectFilters

	ds, ok := params["descriptionSearch"]
	if ok {
		if ds != trueValue && ds != falseValue {
			return fmt.Errorf("%w: %q", ErrInvalidBooleanValue, ds)
		}
		fp.descriptionSearch = &ds
	}

	itemFilters, err := processItemFilters(params)
	if err != nil {
		return err
	}
	fp.itemFilters = itemFilters

	outputSelectors, err := processOutputSelectors(params)
	if err != nil {
		return err
	}
	fp.outputSelectors = outputSelectors

	affiliate, err := processAffiliate(params)
	if err != nil {
		return err
	}
	fp.affiliate = affiliate

	buyerPostalCode, ok := params["buyerPostalCode"]
	if ok {
		if !isValidPostalCode(buyerPostalCode) {
			return ErrInvalidPostalCode
		}
		fp.buyerPostalCode = &buyerPostalCode
	}

	paginationInput, err := processPaginationInput(params)
	if err != nil {
		return err
	}
	fp.paginationInput = paginationInput

	sortOrder, ok := params["sortOrder"]
	if ok {
		err := validateSortOrder(sortOrder, fp.itemFilters, fp.buyerPostalCode != nil)
		if err != nil {
			return err
		}
		fp.sortOrder = &sortOrder
	}

	return nil
}

func (fp *findItemsAdvancedParams) createRequest(appID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, findingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ebay: %w", err)
	}

	qry := req.URL.Query()
	qry.Add("OPERATION-NAME", findItemsAdvancedOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)

	for i, f := range fp.aspectFilters {
		qry.Add(fmt.Sprintf("aspectFilter(%d).aspectName", i), f.aspectName)
		for j, v := range f.aspectValueNames {
			qry.Add(fmt.Sprintf("aspectFilter(%d).aspectValueName(%d)", i, j), v)
		}
	}

	if fp.categoryIDs != nil {
		qry.Add("categoryId", *fp.categoryIDs)
	}
	if fp.descriptionSearch != nil {
		qry.Add("descriptionSearch", *fp.descriptionSearch)
	}

	for i, f := range fp.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", i), f.name)
		for j, v := range f.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", i, j), v)
		}

		if f.paramName != nil && f.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", i), *f.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", i), *f.paramValue)
		}
	}

	if fp.keywords != nil {
		qry.Add("keywords", *fp.keywords)
	}

	for i := range fp.outputSelectors {
		qry.Add(fmt.Sprintf("outputSelector(%d)", i), fp.outputSelectors[i])
	}

	if fp.affiliate != nil {
		if fp.affiliate.customID != nil {
			qry.Add("affiliate.customId", *fp.affiliate.customID)
		}
		if fp.affiliate.geoTargeting != nil {
			qry.Add("affiliate.geoTargeting", *fp.affiliate.geoTargeting)
		}
		if fp.affiliate.networkID != nil {
			qry.Add("affiliate.networkId", *fp.affiliate.networkID)
		}
		if fp.affiliate.trackingID != nil {
			qry.Add("affiliate.trackingId", *fp.affiliate.trackingID)
		}
	}
	if fp.buyerPostalCode != nil {
		qry.Add("buyerPostalCode", *fp.buyerPostalCode)
	}
	if fp.paginationInput != nil {
		if fp.paginationInput.entriesPerPage != nil {
			qry.Add("paginationInput.entriesPerPage", *fp.paginationInput.entriesPerPage)
		}
		if fp.paginationInput.pageNumber != nil {
			qry.Add("paginationInput.pageNumber", *fp.paginationInput.pageNumber)
		}
	}
	if fp.sortOrder != nil {
		qry.Add("sortOrder", *fp.sortOrder)
	}

	req.URL.RawQuery = qry.Encode()

	return req, nil
}

type findItemsByProductParams struct {
	itemFilters     []itemFilter
	outputSelectors []string
	product         productID
	affiliate       *affiliate
	buyerPostalCode *string
	paginationInput *paginationInput
	sortOrder       *string
}

type productID struct {
	idType string
	value  string
}

func (fp *findItemsByProductParams) validateParams(params map[string]string) error {
	productIDType, ptOk := params["productId.@type"]
	productValue, pOk := params["productId"]
	if !ptOk || !pOk {
		return ErrProductIDMissing
	}
	p := productID{idType: productIDType, value: productValue}
	err := p.processProductID()
	if err != nil {
		return err
	}
	fp.product = p

	itemFilters, err := processItemFilters(params)
	if err != nil {
		return err
	}
	fp.itemFilters = itemFilters

	outputSelectors, err := processOutputSelectors(params)
	if err != nil {
		return err
	}
	fp.outputSelectors = outputSelectors

	affiliate, err := processAffiliate(params)
	if err != nil {
		return err
	}
	fp.affiliate = affiliate

	buyerPostalCode, ok := params["buyerPostalCode"]
	if ok {
		if !isValidPostalCode(buyerPostalCode) {
			return ErrInvalidPostalCode
		}
		fp.buyerPostalCode = &buyerPostalCode
	}

	paginationInput, err := processPaginationInput(params)
	if err != nil {
		return err
	}
	fp.paginationInput = paginationInput

	sortOrder, ok := params["sortOrder"]
	if ok {
		err := validateSortOrder(sortOrder, fp.itemFilters, fp.buyerPostalCode != nil)
		if err != nil {
			return err
		}
		fp.sortOrder = &sortOrder
	}

	return nil
}

func (fp *findItemsByProductParams) createRequest(appID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, findingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ebay: %w", err)
	}

	qry := req.URL.Query()
	qry.Add("OPERATION-NAME", findItemsByKeywordsOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)

	for i, f := range fp.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", i), f.name)
		for j, v := range f.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", i, j), v)
		}

		if f.paramName != nil && f.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", i), *f.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", i), *f.paramValue)
		}
	}

	for i := range fp.outputSelectors {
		qry.Add(fmt.Sprintf("outputSelector(%d)", i), fp.outputSelectors[i])
	}
	qry.Add("productId.@type", fp.product.idType)
	qry.Add("productId", fp.product.value)

	if fp.affiliate != nil {
		if fp.affiliate.customID != nil {
			qry.Add("affiliate.customId", *fp.affiliate.customID)
		}
		if fp.affiliate.geoTargeting != nil {
			qry.Add("affiliate.geoTargeting", *fp.affiliate.geoTargeting)
		}
		if fp.affiliate.networkID != nil {
			qry.Add("affiliate.networkId", *fp.affiliate.networkID)
		}
		if fp.affiliate.trackingID != nil {
			qry.Add("affiliate.trackingId", *fp.affiliate.trackingID)
		}
	}
	if fp.buyerPostalCode != nil {
		qry.Add("buyerPostalCode", *fp.buyerPostalCode)
	}
	if fp.paginationInput != nil {
		if fp.paginationInput.entriesPerPage != nil {
			qry.Add("paginationInput.entriesPerPage", *fp.paginationInput.entriesPerPage)
		}
		if fp.paginationInput.pageNumber != nil {
			qry.Add("paginationInput.pageNumber", *fp.paginationInput.pageNumber)
		}
	}
	if fp.sortOrder != nil {
		qry.Add("sortOrder", *fp.sortOrder)
	}

	req.URL.RawQuery = qry.Encode()

	return req, nil
}

func processCategoryIDs(categoryID string) error {
	if categoryID == "" {
		return ErrInvalidCategoryIDLength
	}

	categoryIDs := strings.Split(categoryID, ",")
	if len(categoryIDs) > maxCategoryIDs {
		return ErrMaxCategoryIDs
	}

	for i := range categoryIDs {
		categoryIDs[i] = strings.TrimSpace(categoryIDs[i])
		if len(categoryIDs[i]) > maxCategoryIDLen {
			return ErrInvalidCategoryIDLength
		}
	}

	return nil
}

func processKeywords(params map[string]string) (string, error) {
	keywords, ok := params["keywords"]
	if !ok {
		return "", ErrKeywordsMissing
	}

	if len(keywords) < minKeywordsLen || len(keywords) > maxKeywordsLen {
		return "", ErrInvalidKeywordsLength
	}

	individualKeywords := splitKeywords(keywords)
	for _, k := range individualKeywords {
		if len(k) > maxKeywordLen {
			return "", ErrInvalidKeywordLength
		}
	}

	return keywords, nil
}

// Split keywords based on special characters acting as search operators.
// See https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-keywords.html
func splitKeywords(keywords string) []string {
	const specialChars = ` ,()"-*@+`

	return strings.FieldsFunc(keywords, func(r rune) bool {
		return strings.ContainsRune(specialChars, r)
	})
}

const (
	// Product ID type enumeration values from the eBay documentation.
	// See https://developer.ebay.com/Devzone/finding/CallRef/types/ProductId.html
	referenceID = "ReferenceID"
	isbn        = "ISBN"
	upc         = "UPC"
	ean         = "EAN"
)

func (p *productID) processProductID() error {
	switch p.idType {
	case referenceID:
		if len(p.value) < 1 {
			return ErrInvalidProductIDLength
		}
	case isbn:
		if len(p.value) != isbnShortLen && len(p.value) != isbnLongLen {
			return ErrInvalidISBNLength
		}
		if !isValidISBN(p.value) {
			return ErrInvalidISBN
		}
	case upc:
		if len(p.value) != upcLen {
			return ErrInvalidUPCLength
		}
		if !isValidEAN(p.value) {
			return ErrInvalidUPC
		}
	case ean:
		if len(p.value) != eanShortLen && len(p.value) != eanLongLen {
			return ErrInvalidEANLength
		}
		if !isValidEAN(p.value) {
			return ErrInvalidEAN
		}
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedProductIDType, p.idType)
	}

	return nil
}

func isValidISBN(isbn string) bool {
	if len(isbn) == isbnShortLen {
		var sum, acc int
		for i, r := range isbn {
			digit := int(r - '0')
			if !isDigit(digit) {
				if i == 9 && r == 'X' {
					digit = 10
				} else {
					return false
				}
			}

			acc += digit
			sum += acc
		}

		return sum%11 == 0
	}

	const altMultiplier = 3
	var sum int
	for i, r := range isbn {
		digit := int(r - '0')
		if !isDigit(digit) {
			return false
		}

		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * altMultiplier
		}
	}

	return sum%10 == 0
}

func isDigit(digit int) bool {
	return digit >= 0 && digit <= 9
}

func isValidEAN(ean string) bool {
	const altMultiplier = 3
	var sum int
	for i, r := range ean[:len(ean)-1] {
		digit := int(r - '0')
		if !isDigit(digit) {
			return false
		}

		switch {
		case len(ean) == eanShortLen && i%2 == 0,
			len(ean) == eanLongLen && i%2 != 0,
			len(ean) == upcLen && i%2 == 0:
			sum += digit * altMultiplier
		default:
			sum += digit
		}
	}

	checkDigit := int(ean[len(ean)-1] - '0')
	if !isDigit(checkDigit) {
		return false
	}

	return (sum+checkDigit)%10 == 0
}

func processAspectFilters(params map[string]string) ([]aspectFilter, error) {
	// Check if both "aspectFilter.aspectName" and "aspectFilter(0).aspectName" syntax types occur in the params.
	_, nonNumberedExists := params["aspectFilter.aspectName"]
	_, numberedExists := params["aspectFilter(0).aspectName"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidFilterSyntax
	}

	if nonNumberedExists {
		return processNonNumberedAspectFilter(params)
	}

	return processNumberedAspectFilters(params)
}

func processNonNumberedAspectFilter(params map[string]string) ([]aspectFilter, error) {
	filterValues, err := parseFilterValues(params, "aspectFilter.aspectValueName")
	if err != nil {
		return nil, err
	}

	filter := aspectFilter{
		// No need to check aspectFilter.aspectName exists due to how this function is invoked from processAspectFilters.
		aspectName:       params["aspectFilter.aspectName"],
		aspectValueNames: filterValues,
	}

	return []aspectFilter{filter}, nil
}

func processNumberedAspectFilters(params map[string]string) ([]aspectFilter, error) {
	var aspectFilters []aspectFilter
	for i := 0; ; i++ {
		name, ok := params[fmt.Sprintf("aspectFilter(%d).aspectName", i)]
		if !ok {
			break
		}

		filterValues, err := parseFilterValues(params, fmt.Sprintf("aspectFilter(%d).aspectValueName", i))
		if err != nil {
			return nil, err
		}

		aspectFilter := aspectFilter{
			aspectName:       name,
			aspectValueNames: filterValues,
		}

		aspectFilters = append(aspectFilters, aspectFilter)
	}

	return aspectFilters, nil
}

func processItemFilters(params map[string]string) ([]itemFilter, error) {
	// Check if both "itemFilter.name" and "itemFilter(0).name" syntax types occur in the params.
	_, nonNumberedExists := params["itemFilter.name"]
	_, numberedExists := params["itemFilter(0).name"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidFilterSyntax
	}

	if nonNumberedExists {
		return processNonNumberedItemFilter(params)
	}

	return processNumberedItemFilters(params)
}

func processNonNumberedItemFilter(params map[string]string) ([]itemFilter, error) {
	filterValues, err := parseFilterValues(params, "itemFilter.value")
	if err != nil {
		return nil, err
	}

	filter := itemFilter{
		// No need to check itemFilter.name exists due to how this function is invoked from processItemFilters.
		name:   params["itemFilter.name"],
		values: filterValues,
	}

	pn, pnOk := params["itemFilter.paramName"]
	pv, pvOk := params["itemFilter.paramValue"]
	if pnOk != pvOk {
		return nil, ErrIncompleteItemFilterParam
	}
	if pnOk && pvOk {
		filter.paramName = &pn
		filter.paramValue = &pv
	}

	err = handleItemFilterType(&filter, nil, params)
	if err != nil {
		return nil, err
	}

	return []itemFilter{filter}, nil
}

func processNumberedItemFilters(params map[string]string) ([]itemFilter, error) {
	var itemFilters []itemFilter
	for i := 0; ; i++ {
		name, ok := params[fmt.Sprintf("itemFilter(%d).name", i)]
		if !ok {
			break
		}

		filterValues, err := parseFilterValues(params, fmt.Sprintf("itemFilter(%d).value", i))
		if err != nil {
			return nil, err
		}

		itemFilter := itemFilter{
			name:   name,
			values: filterValues,
		}

		pn, pnOk := params[fmt.Sprintf("itemFilter(%d).paramName", i)]
		pv, pvOk := params[fmt.Sprintf("itemFilter(%d).paramValue", i)]
		if pnOk != pvOk {
			return nil, ErrIncompleteItemFilterParam
		}
		if pnOk && pvOk {
			itemFilter.paramName = &pn
			itemFilter.paramValue = &pv
		}

		itemFilters = append(itemFilters, itemFilter)
	}

	for i := range itemFilters {
		err := handleItemFilterType(&itemFilters[i], itemFilters, params)
		if err != nil {
			return nil, err
		}
	}

	return itemFilters, nil
}

func parseFilterValues(params map[string]string, filterAttr string) ([]string, error) {
	var filterValues []string
	for i := 0; ; i++ {
		k := fmt.Sprintf("%s(%d)", filterAttr, i)
		if v, ok := params[k]; ok {
			filterValues = append(filterValues, v)
		} else {
			break
		}
	}

	if v, ok := params[filterAttr]; ok {
		filterValues = append(filterValues, v)
	}

	if len(filterValues) == 0 {
		return nil, fmt.Errorf("%w %q", ErrIncompleteFilterNameOnly, filterAttr)
	}

	// Check if both "filterAttr" and "filterAttr(0)" syntax types occur in the params.
	_, nonNumberedExists := params[filterAttr]
	_, numberedExists := params[filterAttr+"(0)"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidFilterSyntax
	}

	return filterValues, nil
}

const (
	// ItemFilterType enumeration values from the eBay documentation.
	// See https://developer.ebay.com/devzone/finding/CallRef/types/ItemFilterType.html
	authorizedSellerOnly  = "AuthorizedSellerOnly"
	availableTo           = "AvailableTo"
	bestOfferOnly         = "BestOfferOnly"
	charityOnly           = "CharityOnly"
	condition             = "Condition"
	currency              = "Currency"
	endTimeFrom           = "EndTimeFrom"
	endTimeTo             = "EndTimeTo"
	excludeAutoPay        = "ExcludeAutoPay"
	excludeCategory       = "ExcludeCategory"
	excludeSeller         = "ExcludeSeller"
	expeditedShippingType = "ExpeditedShippingType"
	feedbackScoreMax      = "FeedbackScoreMax"
	feedbackScoreMin      = "FeedbackScoreMin"
	freeShippingOnly      = "FreeShippingOnly"
	hideDuplicateItems    = "HideDuplicateItems"
	listedIn              = "ListedIn"
	listingType           = "ListingType"
	localPickupOnly       = "LocalPickupOnly"
	localSearchOnly       = "LocalSearchOnly"
	locatedIn             = "LocatedIn"
	lotsOnly              = "LotsOnly"
	maxBids               = "MaxBids"
	maxDistance           = "MaxDistance"
	maxHandlingTime       = "MaxHandlingTime"
	maxPrice              = "MaxPrice"
	maxQuantity           = "MaxQuantity"
	minBids               = "MinBids"
	minPrice              = "MinPrice"
	minQuantity           = "MinQuantity"
	modTimeFrom           = "ModTimeFrom"
	returnsAcceptedOnly   = "ReturnsAcceptedOnly"
	seller                = "Seller"
	sellerBusinessType    = "SellerBusinessType"
	soldItemsOnly         = "SoldItemsOnly"
	startTimeFrom         = "StartTimeFrom"
	startTimeTo           = "StartTimeTo"
	topRatedSellerOnly    = "TopRatedSellerOnly"
	valueBoxInventory     = "ValueBoxInventory"

	trueValue           = "true"
	falseValue          = "false"
	trueNum             = "1"
	falseNum            = "0"
	smallestMaxDistance = 5
)

// Valid Currency ID values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/CallRef/Enums/currencyIdList.html
var validCurrencyIDs = []string{
	"AUD", "CAD", "CHF", "CNY", "EUR", "GBP", "HKD", "INR", "MYR", "PHP", "PLN", "SEK", "SGD", "TWD", "USD",
}

// Valid Global ID values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/CallRef/Enums/GlobalIdList.html
var validGlobalIDs = []string{
	"EBAY-AT",
	"EBAY-AU",
	"EBAY-CH",
	"EBAY-DE",
	"EBAY-ENCA",
	"EBAY-ES",
	"EBAY-FR",
	"EBAY-FRBE",
	"EBAY-FRCA",
	"EBAY-GB",
	"EBAY-HK",
	"EBAY-IE",
	"EBAY-IN",
	"EBAY-IT",
	"EBAY-MOTOR",
	"EBAY-MY",
	"EBAY-NL",
	"EBAY-NLBE",
	"EBAY-PH",
	"EBAY-PL",
	"EBAY-SG",
	"EBAY-US",
}

func handleItemFilterType(filter *itemFilter, itemFilters []itemFilter, params map[string]string) error {
	switch filter.name {
	case authorizedSellerOnly, bestOfferOnly, charityOnly, excludeAutoPay, freeShippingOnly, hideDuplicateItems,
		localPickupOnly, lotsOnly, returnsAcceptedOnly, soldItemsOnly:
		if filter.values[0] != trueValue && filter.values[0] != falseValue {
			return fmt.Errorf("%w: %q", ErrInvalidBooleanValue, filter.values[0])
		}
	case availableTo:
		if !isValidCountryCode(filter.values[0]) {
			return fmt.Errorf("%w: %q", ErrInvalidCountryCode, filter.values[0])
		}
	case condition:
		if !isValidCondition(filter.values[0]) {
			return fmt.Errorf("%w: %q", ErrInvalidCondition, filter.values[0])
		}
	case currency:
		if !slices.Contains(validCurrencyIDs, filter.values[0]) {
			return fmt.Errorf("%w: %q", ErrInvalidCurrencyID, filter.values[0])
		}
	case endTimeFrom, endTimeTo, startTimeFrom, startTimeTo:
		if !isValidDateTime(filter.values[0], true) {
			return fmt.Errorf("%w: %q", ErrInvalidDateTime, filter.values[0])
		}
	case excludeCategory:
		err := validateExcludeCategories(filter.values)
		if err != nil {
			return err
		}
	case excludeSeller:
		err := validateExcludeSellers(filter.values, itemFilters)
		if err != nil {
			return err
		}
	case expeditedShippingType:
		if filter.values[0] != "Expedited" && filter.values[0] != "OneDayShipping" {
			return fmt.Errorf("%w: %q", ErrInvalidExpeditedShippingType, filter.values[0])
		}
	case feedbackScoreMax, feedbackScoreMin:
		err := validateNumericFilter(filter, itemFilters, 0, feedbackScoreMax, feedbackScoreMin)
		if err != nil {
			return err
		}
	case listedIn:
		if !slices.Contains(validGlobalIDs, filter.values[0]) {
			return fmt.Errorf("%w: %q", ErrInvalidGlobalID, filter.values[0])
		}
	case listingType:
		err := validateListingTypes(filter.values)
		if err != nil {
			return err
		}
	case localSearchOnly:
		err := validateLocalSearchOnly(filter.values, itemFilters, params)
		if err != nil {
			return err
		}
	case locatedIn:
		err := validateLocatedIns(filter.values)
		if err != nil {
			return err
		}
	case maxBids, minBids:
		err := validateNumericFilter(filter, itemFilters, 0, maxBids, minBids)
		if err != nil {
			return err
		}
	case maxDistance:
		if _, ok := params["buyerPostalCode"]; !ok {
			return ErrBuyerPostalCodeMissing
		}
		if !isValidIntegerInRange(filter.values[0], smallestMaxDistance) {
			return invalidIntegerError(filter.values[0], smallestMaxDistance)
		}
	case maxHandlingTime:
		if !isValidIntegerInRange(filter.values[0], 1) {
			return invalidIntegerError(filter.values[0], 1)
		}
	case maxPrice, minPrice:
		err := validatePriceRange(filter, itemFilters)
		if err != nil {
			return err
		}
	case maxQuantity, minQuantity:
		err := validateNumericFilter(filter, itemFilters, 1, maxQuantity, minQuantity)
		if err != nil {
			return err
		}
	case modTimeFrom:
		if !isValidDateTime(filter.values[0], false) {
			return fmt.Errorf("%w: %q", ErrInvalidDateTime, filter.values[0])
		}
	case seller:
		err := validateSellers(filter.values, itemFilters)
		if err != nil {
			return err
		}
	case sellerBusinessType:
		err := validateSellerBusinessType(filter.values)
		if err != nil {
			return err
		}
	case topRatedSellerOnly:
		err := validateTopRatedSellerOnly(filter.values[0], itemFilters)
		if err != nil {
			return err
		}
	case valueBoxInventory:
		if filter.values[0] != trueNum && filter.values[0] != falseNum {
			return fmt.Errorf("%w: %q", ErrInvalidValueBoxInventory, filter.values[0])
		}
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedItemFilterType, filter.name)
	}

	return nil
}

func isValidCountryCode(value string) bool {
	const countryCodeLen = 2
	if len(value) != countryCodeLen {
		return false
	}

	for _, r := range value {
		if !unicode.IsUpper(r) {
			return false
		}
	}

	return true
}

// Valid Condition IDs from the eBay documentation.
// See https://developer.ebay.com/Devzone/finding/CallRef/Enums/conditionIdList.html#ConditionDefinitions
var validConditionIDs = []int{1000, 1500, 1750, 2000, 2010, 2020, 2030, 2500, 2750, 3000, 4000, 5000, 6000, 7000}

func isValidCondition(value string) bool {
	cID, err := strconv.Atoi(value)
	if err == nil {
		return slices.Contains(validConditionIDs, cID)
	}

	// Value is a condition name, refer to the eBay documentation for condition name definitions.
	// See https://developer.ebay.com/Devzone/finding/CallRef/Enums/conditionIdList.html
	return true
}

func isValidDateTime(value string, future bool) bool {
	dateTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return false
	}

	if dateTime.Location() != time.UTC {
		return false
	}

	now := time.Now().UTC()
	if future && dateTime.Before(now) {
		return false
	}
	if !future && dateTime.After(now) {
		return false
	}

	return true
}

func validateExcludeCategories(values []string) error {
	if len(values) > maxExcludeCategories {
		return ErrMaxExcludeCategories
	}

	for _, v := range values {
		if !isValidIntegerInRange(v, 0) {
			return invalidIntegerError(v, 0)
		}
	}

	return nil
}

func validateExcludeSellers(values []string, itemFilters []itemFilter) error {
	if len(values) > maxExcludeSellers {
		return ErrMaxExcludeSellers
	}

	for _, f := range itemFilters {
		if f.name == seller || f.name == topRatedSellerOnly {
			return ErrExcludeSellerCannotBeUsedWithSellers
		}
	}

	return nil
}

func validateNumericFilter(
	filter *itemFilter, itemFilters []itemFilter, minAllowedValue int, filterA, filterB string,
) error {
	v, err := strconv.Atoi(filter.values[0])
	if err != nil {
		return fmt.Errorf("ebay: %w", err)
	}
	if minAllowedValue > v {
		return invalidIntegerError(filter.values[0], minAllowedValue)
	}

	var filterAValue, filterBValue *int
	for _, f := range itemFilters {
		if f.name == filterA {
			val, err := strconv.Atoi(f.values[0])
			if err != nil {
				return fmt.Errorf("ebay: %w", err)
			}
			filterAValue = &val
		} else if f.name == filterB {
			val, err := strconv.Atoi(f.values[0])
			if err != nil {
				return fmt.Errorf("ebay: %w", err)
			}
			filterBValue = &val
		}
	}

	if filterAValue != nil && filterBValue != nil && *filterBValue > *filterAValue {
		return fmt.Errorf("%w: %q must be greater than or equal to %q", ErrInvalidNumericFilter, filterA, filterB)
	}

	return nil
}

func invalidIntegerError(value string, min int) error {
	return fmt.Errorf("%w: %q (minimum value: %d)", ErrInvalidInteger, value, min)
}

func isValidIntegerInRange(value string, min int) bool {
	n, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	return n >= min
}

// Valid Listing Type values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/CallRef/types/ItemFilterType.html#ListingType
var validListingTypes = []string{"Auction", "AuctionWithBIN", "Classified", "FixedPrice", "StoreInventory", "All"}

func validateListingTypes(values []string) error {
	seenTypes := make(map[string]bool)
	hasAuction, hasAuctionWithBIN := false, false
	for _, v := range values {
		if v == "All" && len(values) > 1 {
			return ErrInvalidAllListingType
		}

		found := false
		for _, lt := range validListingTypes {
			if v == lt {
				found = true
				if v == "Auction" {
					hasAuction = true
				} else if v == "AuctionWithBIN" {
					hasAuctionWithBIN = true
				}

				break
			}
		}

		if !found {
			return fmt.Errorf("%w: %q", ErrInvalidListingType, v)
		}
		if seenTypes[v] {
			return fmt.Errorf("%w: %q", ErrDuplicateListingType, v)
		}
		if hasAuction && hasAuctionWithBIN {
			return ErrInvalidAuctionListingTypes
		}

		seenTypes[v] = true
	}

	return nil
}

func validateLocalSearchOnly(values []string, itemFilters []itemFilter, params map[string]string) error {
	if _, ok := params["buyerPostalCode"]; !ok {
		return ErrBuyerPostalCodeMissing
	}

	foundMaxDistance := slices.ContainsFunc(itemFilters, func(f itemFilter) bool {
		return f.name == maxDistance
	})
	if !foundMaxDistance {
		return ErrMaxDistanceMissing
	}

	if values[0] != trueValue && values[0] != falseValue {
		return fmt.Errorf("%w: %q", ErrInvalidBooleanValue, values[0])
	}

	return nil
}

func validateLocatedIns(values []string) error {
	if len(values) > maxLocatedIns {
		return ErrMaxLocatedIns
	}

	for _, v := range values {
		if !isValidCountryCode(v) {
			return fmt.Errorf("%w: %q", ErrInvalidCountryCode, v)
		}
	}

	return nil
}

func validatePriceRange(filter *itemFilter, itemFilters []itemFilter) error {
	price, err := parsePrice(filter)
	if err != nil {
		return err
	}

	var relatedFilterName string
	if filter.name == maxPrice {
		relatedFilterName = minPrice
	} else if filter.name == minPrice {
		relatedFilterName = maxPrice
	}

	for i := range itemFilters {
		if itemFilters[i].name == relatedFilterName {
			relatedPrice, err := parsePrice(&itemFilters[i])
			if err != nil {
				return err
			}

			if (filter.name == maxPrice && price < relatedPrice) ||
				(filter.name == minPrice && price > relatedPrice) {
				return ErrInvalidMaxPrice
			}
		}
	}

	return nil
}

func parsePrice(filter *itemFilter) (float64, error) {
	const minAllowedPrice float64 = 0.0
	price, err := strconv.ParseFloat(filter.values[0], 64)
	if err != nil {
		return 0, fmt.Errorf("ebay: %w", err)
	}
	if minAllowedPrice > price {
		return 0, fmt.Errorf("%w: %f (minimum value: %f)", ErrInvalidPrice, price, minAllowedPrice)
	}

	if filter.paramName != nil && *filter.paramName != currency {
		return 0, fmt.Errorf("%w: %q", ErrInvalidPriceParamName, *filter.paramName)
	}

	if filter.paramValue != nil && !slices.Contains(validCurrencyIDs, *filter.paramValue) {
		return 0, fmt.Errorf("%w: %q", ErrInvalidCurrencyID, *filter.paramValue)
	}

	return price, nil
}

func validateSellers(values []string, itemFilters []itemFilter) error {
	if len(values) > maxSellers {
		return ErrMaxSellers
	}

	for _, f := range itemFilters {
		if f.name == excludeSeller || f.name == topRatedSellerOnly {
			return ErrSellerCannotBeUsedWithOtherSellers
		}
	}

	return nil
}

func validateSellerBusinessType(values []string) error {
	if len(values) > 1 {
		return fmt.Errorf("%w", ErrMultipleSellerBusinessTypes)
	}

	if values[0] != "Business" && values[0] != "Private" {
		return fmt.Errorf("%w: %q", ErrInvalidSellerBusinessType, values[0])
	}

	return nil
}

func validateTopRatedSellerOnly(value string, itemFilters []itemFilter) error {
	if value != trueValue && value != falseValue {
		return fmt.Errorf("%w: %q", ErrInvalidBooleanValue, value)
	}

	for _, f := range itemFilters {
		if f.name == seller || f.name == excludeSeller {
			return ErrTopRatedSellerCannotBeUsedWithSellers
		}
	}

	return nil
}

// Valid OutputSelectorType values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/callref/types/OutputSelectorType.html
var validOutputSelectors = []string{
	"AspectHistogram",
	"CategoryHistogram",
	"ConditionHistogram",
	"GalleryInfo",
	"PictureURLLarge",
	"PictureURLSuperSize",
	"SellerInfo",
	"StoreInfo",
	"UnitPriceInfo",
}

func processOutputSelectors(params map[string]string) ([]string, error) {
	// Check if both "outputSelector" and "outputSelector(0)" syntax types occur in the params.
	outputSelector, nonNumberedExists := params["outputSelector"]
	_, numberedExists := params["outputSelector(0)"]

	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidFilterSyntax
	}

	if nonNumberedExists {
		if !slices.Contains(validOutputSelectors, outputSelector) {
			return nil, ErrInvalidOutputSelector
		}

		return []string{outputSelector}, nil
	}

	var os []string
	for i := 0; ; i++ {
		s, ok := params[fmt.Sprintf("outputSelector(%d)", i)]
		if !ok {
			break
		}

		if !slices.Contains(validOutputSelectors, s) {
			return nil, ErrInvalidOutputSelector
		}

		os = append(os, s)
	}

	return os, nil
}

func processAffiliate(params map[string]string) (*affiliate, error) {
	var aff affiliate
	customID, ok := params["affiliate.customId"]
	if ok {
		if len(customID) > maxCustomIDLen {
			return nil, ErrInvalidCustomIDLength
		}
		aff.customID = &customID
	}

	geoTargeting, ok := params["affiliate.geoTargeting"]
	if ok {
		if geoTargeting != trueValue && geoTargeting != falseValue {
			return nil, fmt.Errorf("%w: %q", ErrInvalidBooleanValue, geoTargeting)
		}
		aff.geoTargeting = &geoTargeting
	}

	networkID, nOk := params["affiliate.networkId"]
	trackingID, tOk := params["affiliate.trackingId"]
	if nOk != tOk {
		return nil, ErrIncompleteAffiliateParams
	}
	if !nOk {
		return &aff, nil
	}

	nID, err := strconv.Atoi(networkID)
	if err != nil {
		return nil, fmt.Errorf("ebay: %w", err)
	}

	if nID < beFreeID || nID > ebayPartnerNetworkID {
		return nil, ErrInvalidNetworkID
	}
	if nID == ebayPartnerNetworkID {
		err := validateTrackingID(trackingID)
		if err != nil {
			return nil, err
		}
	}

	aff.networkID = &networkID
	aff.trackingID = &trackingID

	return &aff, nil
}

func validateTrackingID(trackingID string) error {
	_, err := strconv.Atoi(trackingID)
	if err != nil {
		return fmt.Errorf("ebay: %w", err)
	}

	const maxCampIDLen = 10
	if len(trackingID) != maxCampIDLen {
		return ErrInvalidCampaignID
	}

	return nil
}

func isValidPostalCode(postalCode string) bool {
	const minPostalCodeLen = 3

	return len(postalCode) >= minPostalCodeLen
}

func processPaginationInput(params map[string]string) (*paginationInput, error) {
	entriesPerPage, eOk := params["paginationInput.entriesPerPage"]
	pageNumber, pOk := params["paginationInput.pageNumber"]
	if !eOk && !pOk {
		return &paginationInput{}, nil
	}

	var pInput paginationInput
	if eOk {
		v, err := strconv.Atoi(entriesPerPage)
		if err != nil {
			return nil, fmt.Errorf("ebay: %w", err)
		}

		if v < minPaginationValue || v > maxPaginationValue {
			return nil, ErrInvalidEntriesPerPage
		}
		pInput.entriesPerPage = &entriesPerPage
	}
	if pOk {
		v, err := strconv.Atoi(pageNumber)
		if err != nil {
			return nil, fmt.Errorf("ebay: %w", err)
		}

		if v < minPaginationValue || v > maxPaginationValue {
			return nil, ErrInvalidPageNumber
		}
		pInput.pageNumber = &pageNumber
	}

	return &pInput, nil
}

const (
	// SortOrderType enumeration values from the eBay documentation.
	// See https://developer.ebay.com/devzone/finding/CallRef/types/SortOrderType.html
	bestMatch                = "BestMatch"
	bidCountFewest           = "BidCountFewest"
	bidCountMost             = "BidCountMost"
	countryAscending         = "CountryAscending"
	countryDescending        = "CountryDescending"
	currentPriceHighest      = "CurrentPriceHighest"
	distanceNearest          = "DistanceNearest"
	endTimeSoonest           = "EndTimeSoonest"
	pricePlusShippingHighest = "PricePlusShippingHighest"
	pricePlusShippingLowest  = "PricePlusShippingLowest"
	startTimeNewest          = "StartTimeNewest"
	watchCountDecreaseSort   = "WatchCountDecreaseSort"
)

func validateSortOrder(sortOrder string, itemFilters []itemFilter, hasBuyerPostalCode bool) error {
	switch sortOrder {
	case bestMatch, countryAscending, countryDescending, currentPriceHighest, endTimeSoonest,
		pricePlusShippingHighest, pricePlusShippingLowest, startTimeNewest, watchCountDecreaseSort:
		return nil
	case bidCountFewest, bidCountMost:
		hasAuctionListing := slices.ContainsFunc(itemFilters, func(f itemFilter) bool {
			return f.name == listingType && slices.Contains(f.values, "Auction")
		})
		if !hasAuctionListing {
			return ErrAuctionListingMissing
		}
	case distanceNearest:
		if !hasBuyerPostalCode {
			return ErrBuyerPostalCodeMissing
		}
	default:
		return ErrUnsupportedSortOrderType
	}

	return nil
}

func (svr *FindingServer) parseFindItemsByKeywordsResponse(resp *http.Response) (FindItemsByKeywordsResponse, error) {
	defer resp.Body.Close()

	var itemsResp FindItemsByKeywordsResponse
	err := json.NewDecoder(resp.Body).Decode(&itemsResp)
	if err != nil {
		return FindItemsByKeywordsResponse{}, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}

	return itemsResp, nil
}

func (svr *FindingServer) parseFindItemsByCategoryResponse(resp *http.Response) (FindItemsByCategoriesResponse, error) {
	defer resp.Body.Close()

	var itemsResp FindItemsByCategoriesResponse
	err := json.NewDecoder(resp.Body).Decode(&itemsResp)
	if err != nil {
		return FindItemsByCategoriesResponse{}, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}

	return itemsResp, nil
}

func (svr *FindingServer) parseFindItemsAdvancedResponse(resp *http.Response) (FindItemsAdvancedResponse, error) {
	defer resp.Body.Close()

	var itemsResp FindItemsAdvancedResponse
	err := json.NewDecoder(resp.Body).Decode(&itemsResp)
	if err != nil {
		return FindItemsAdvancedResponse{}, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}

	return itemsResp, nil
}

func (svr *FindingServer) parseFindItemsByProductResponse(resp *http.Response) (FindItemsByProductResponse, error) {
	defer resp.Body.Close()

	var itemsResp FindItemsByProductResponse
	err := json.NewDecoder(resp.Body).Decode(&itemsResp)
	if err != nil {
		return FindItemsByProductResponse{}, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}

	return itemsResp, nil
}
