package ebay

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"unicode"
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

	// ErrInvalidItemFilterSyntax is returned when both syntax types for item filters are used in the params.
	ErrInvalidItemFilterSyntax = errors.New("ebay: invalid item filter syntax: both syntax types are present")

	// ErrIncompleteItemFilterNameOnly is returned when an item filter is missing the 'values' parameter.
	ErrIncompleteItemFilterNameOnly = errors.New("ebay: incomplete item filter: missing value")

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

	// ErrInvalidBooleanValue is returned when an item filter has an invalid boolean 'values' parameter.
	ErrInvalidBooleanValue = errors.New("ebay: invalid boolean item filter value, allowed values are true and false")

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

	// ErrBuyerPostalCodeMissing is returned when the LocalSearchOnly or MaxDistance item filter is used,
	// but the buyerPostalCode parameter is missing in the request.
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

// FindItemsByKeywords searches the eBay Finding API using the provided parameters and a valid eBay application ID.
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
	name       string
	values     []string
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
	// Check if both "itemFilter.name" and "itemFilter(0).name" syntax types occur in the params.
	_, nonNumberedExists := params["itemFilter.name"]
	_, numberedExists := params["itemFilter(0).name"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidItemFilterSyntax
	}

	if nonNumberedExists {
		return processNonNumberedItemFilter(params)
	}

	return processNumberedItemFilters(params)
}

func processNonNumberedItemFilter(params map[string]string) ([]itemFilter, error) {
	filterValues, err := parseItemFilterValues(params, "itemFilter")
	if err != nil {
		return nil, err
	}

	filter := itemFilter{
		// No need to check itemFilter.name exists due to how this function is invoked from processItemFilters.
		name:   params["itemFilter.name"],
		values: filterValues,
	}

	pName, pnOk := params["itemFilter.paramName"]
	pValue, pvOk := params["itemFilter.paramValue"]
	if pnOk != pvOk {
		return nil, ErrIncompleteItemFilterParam
	}
	if pnOk && pvOk {
		filter.paramName = &pName
		filter.paramValue = &pValue
	}

	err = handleItemFilterType(&filter, nil, params)
	if err != nil {
		return nil, err
	}

	return []itemFilter{filter}, nil
}

func processNumberedItemFilters(params map[string]string) ([]itemFilter, error) {
	var itemFilters []itemFilter
	for idx := 0; ; idx++ {
		name, ok := params[fmt.Sprintf("itemFilter(%d).name", idx)]
		if !ok {
			break
		}

		filterValues, err := parseItemFilterValues(params, fmt.Sprintf("itemFilter(%d)", idx))
		if err != nil {
			return nil, err
		}

		itemFilter := itemFilter{
			name:   name,
			values: filterValues,
		}

		pName, pnOk := params[fmt.Sprintf("itemFilter(%d).paramName", idx)]
		pValue, pvOk := params[fmt.Sprintf("itemFilter(%d).paramValue", idx)]
		if pnOk != pvOk {
			return nil, ErrIncompleteItemFilterParam
		}
		if pnOk && pvOk {
			itemFilter.paramName = &pName
			itemFilter.paramValue = &pValue
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

func parseItemFilterValues(params map[string]string, prefix string) ([]string, error) {
	var filterValues []string
	for i := 0; ; i++ {
		key := fmt.Sprintf(prefix+".value(%d)", i)
		if v, ok := params[key]; ok {
			filterValues = append(filterValues, v)
		} else {
			break
		}
	}

	if v, ok := params[prefix+".value"]; ok {
		filterValues = append(filterValues, v)
	}

	if len(filterValues) == 0 {
		return nil, ErrIncompleteItemFilterNameOnly
	}

	// Check if both "itemFilter.value" and "itemFilter.value(0)" syntax types occur in the params.
	_, nonNumberedExists := params[prefix+".value"]
	_, numberedExists := params[prefix+".value(0)"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidItemFilterSyntax
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

func handleItemFilterType(filter *itemFilter, itemFilters []itemFilter, params map[string]string) error {
	switch filter.name {
	case authorizedSellerOnly, bestOfferOnly, charityOnly, excludeAutoPay, freeShippingOnly, hideDuplicateItems,
		localPickupOnly, lotsOnly, returnsAcceptedOnly, soldItemsOnly:
		if filter.values[0] != trueValue && filter.values[0] != falseValue {
			return fmt.Errorf("%w: %s", ErrInvalidBooleanValue, filter.values[0])
		}
	case availableTo:
		if !isValidCountryCode(filter.values[0]) {
			return fmt.Errorf("%w: %s", ErrInvalidCountryCode, filter.values[0])
		}
	case condition:
		if !isValidCondition(filter.values[0]) {
			return fmt.Errorf("%w: %s", ErrInvalidCondition, filter.values[0])
		}
	case currency:
		if !isValidCurrencyID(filter.values[0]) {
			return fmt.Errorf("%w: %s", ErrInvalidCurrencyID, filter.values[0])
		}
	case endTimeFrom, endTimeTo, startTimeFrom, startTimeTo:
		if !isValidDateTime(filter.values[0], true) {
			return fmt.Errorf("%w: %s", ErrInvalidDateTime, filter.values[0])
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
			return fmt.Errorf("%w: %s", ErrInvalidExpeditedShippingType, filter.values[0])
		}
	case feedbackScoreMax, feedbackScoreMin:
		err := validateNumericFilter(filter, itemFilters, 0, feedbackScoreMax, feedbackScoreMin)
		if err != nil {
			return err
		}
	case listedIn:
		if !isValidGlobalID(filter.values[0]) {
			return fmt.Errorf("%w: %s", ErrInvalidGlobalID, filter.values[0])
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
			return fmt.Errorf("%w: %s", ErrInvalidDateTime, filter.values[0])
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
			return fmt.Errorf("%w: %s", ErrInvalidValueBoxInventory, filter.values[0])
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedItemFilterType, filter.name)
	}

	return nil
}

const countryCodeLength = 2

func isValidCountryCode(value string) bool {
	if len(value) != countryCodeLength {
		return false
	}

	for _, ch := range value {
		if !unicode.IsUpper(ch) {
			return false
		}
	}

	return true
}

// Valid Condition IDs from the eBay documentation.
// See https://developer.ebay.com/Devzone/finding/CallRef/Enums/conditionIdList.html#ConditionDefinitions
var validConditionIDs = []int{1000, 1500, 1750, 2000, 2010, 2020, 2030, 2500, 2750, 3000, 4000, 5000, 6000, 7000}

func isValidCondition(value string) bool {
	conditionID, err := strconv.Atoi(value)
	if err == nil {
		for _, id := range validConditionIDs {
			if conditionID == id {
				return true
			}
		}
	} else {
		// Value is a condition name, refer to the eBay documentation for condition name definitions.
		// See https://developer.ebay.com/Devzone/finding/CallRef/Enums/conditionIdList.html
		return true
	}

	return false
}

// Valid Currency ID values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/CallRef/Enums/currencyIdList.html
var validCurrencyIDs = []string{
	"AUD", "CAD", "CHF", "CNY", "EUR", "GBP", "HKD", "INR", "MYR", "PHP", "PLN", "SEK", "SGD", "TWD", "USD",
}

func isValidCurrencyID(value string) bool {
	for _, currency := range validCurrencyIDs {
		if value == currency {
			return true
		}
	}

	return false
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

	for _, flt := range itemFilters {
		if flt.name == seller || flt.name == topRatedSellerOnly {
			return ErrExcludeSellerCannotBeUsedWithSellers
		}
	}

	return nil
}

func validateNumericFilter(
	filter *itemFilter, itemFilters []itemFilter, minAllowedValue int, filterA, filterB string,
) error {
	value, err := strconv.Atoi(filter.values[0])
	if err != nil {
		return fmt.Errorf("ebay: %w", err)
	}
	if minAllowedValue > value {
		return invalidIntegerError(filter.values[0], minAllowedValue)
	}

	var filterAValue, filterBValue *int
	for _, flt := range itemFilters {
		if flt.name == filterA {
			val, err := strconv.Atoi(flt.values[0])
			if err != nil {
				return fmt.Errorf("ebay: %w", err)
			}
			filterAValue = &val
		} else if flt.name == filterB {
			val, err := strconv.Atoi(flt.values[0])
			if err != nil {
				return fmt.Errorf("ebay: %w", err)
			}
			filterBValue = &val
		}
	}

	if filterAValue != nil && filterBValue != nil && *filterBValue > *filterAValue {
		return fmt.Errorf("%w: %s must be greater than or equal to %s", ErrInvalidNumericFilter, filterA, filterB)
	}

	return nil
}

func invalidIntegerError(value string, min int) error {
	return fmt.Errorf("%w: %s (minimum value: %d)", ErrInvalidInteger, value, min)
}

func isValidIntegerInRange(value string, min int) bool {
	num, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	return num >= min
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

func isValidGlobalID(value string) bool {
	for _, id := range validGlobalIDs {
		if value == id {
			return true
		}
	}

	return false
}

// Valid Listing Type values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/CallRef/types/ItemFilterType.html#ListingType
var validListingTypes = []string{"Auction", "AuctionWithBIN", "Classified", "FixedPrice", "StoreInventory", "All"}

func validateListingTypes(values []string) error {
	seenTypes := make(map[string]bool)
	hasAuction, hasAuctionWithBIN := false, false
	for _, val := range values {
		if val == "All" && len(values) > 1 {
			return ErrInvalidAllListingType
		}

		found := false
		for _, lt := range validListingTypes {
			if val == lt {
				found = true
				if val == "Auction" {
					hasAuction = true
				} else if val == "AuctionWithBIN" {
					hasAuctionWithBIN = true
				}

				break
			}
		}

		if !found {
			return fmt.Errorf("%w: %s", ErrInvalidListingType, val)
		}
		if seenTypes[val] {
			return fmt.Errorf("%w: %s", ErrDuplicateListingType, val)
		}
		if hasAuction && hasAuctionWithBIN {
			return ErrInvalidAuctionListingTypes
		}

		seenTypes[val] = true
	}

	return nil
}

func validateLocalSearchOnly(values []string, itemFilters []itemFilter, params map[string]string) error {
	if _, ok := params["buyerPostalCode"]; !ok {
		return ErrBuyerPostalCodeMissing
	}

	foundMaxDistance := false
	for _, flt := range itemFilters {
		if flt.name == maxDistance {
			foundMaxDistance = true

			break
		}
	}

	if !foundMaxDistance {
		return ErrMaxDistanceMissing
	}

	if values[0] != trueValue && values[0] != falseValue {
		return fmt.Errorf("%w: %s", ErrInvalidBooleanValue, values[0])
	}

	return nil
}

func validateLocatedIns(values []string) error {
	if len(values) > maxLocatedIns {
		return ErrMaxLocatedIns
	}

	for _, v := range values {
		if !isValidCountryCode(v) {
			return fmt.Errorf("%w: %s", ErrInvalidCountryCode, v)
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

const minAllowedPrice float64 = 0.0

func parsePrice(filter *itemFilter) (float64, error) {
	price, err := strconv.ParseFloat(filter.values[0], 64)
	if err != nil {
		return 0, fmt.Errorf("ebay: %w", err)
	}
	if minAllowedPrice > price {
		return 0, fmt.Errorf("%w: %f (minimum value: %f)", ErrInvalidPrice, price, minAllowedPrice)
	}

	if filter.paramName != nil && *filter.paramName != currency {
		return 0, fmt.Errorf("%w: %s", ErrInvalidPriceParamName, *filter.paramName)
	}

	if filter.paramValue != nil && !isValidCurrencyID(*filter.paramValue) {
		return 0, fmt.Errorf("%w: %s", ErrInvalidCurrencyID, *filter.paramValue)
	}

	return price, nil
}

func validateSellers(values []string, itemFilters []itemFilter) error {
	if len(values) > maxSellers {
		return ErrMaxSellers
	}

	for _, flt := range itemFilters {
		if flt.name == excludeSeller || flt.name == topRatedSellerOnly {
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
		return fmt.Errorf("%w: %s", ErrInvalidSellerBusinessType, values[0])
	}

	return nil
}

func validateTopRatedSellerOnly(value string, itemFilters []itemFilter) error {
	if value != trueValue && value != falseValue {
		return fmt.Errorf("%w: %s", ErrInvalidBooleanValue, value)
	}

	for _, flt := range itemFilters {
		if flt.name == seller || flt.name == excludeSeller {
			return ErrTopRatedSellerCannotBeUsedWithSellers
		}
	}

	return nil
}

func (svr *FindingServer) createRequest(fParams *findingParams, appID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, findingURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ebay: %w", err)
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
		for j, v := range itemFilter.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", idx, j), v)
		}

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
