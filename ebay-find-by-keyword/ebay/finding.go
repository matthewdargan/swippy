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

	// ErrInvalidItemFilterSyntax is returned when both syntax types for itemFilters are used in the params.
	ErrInvalidItemFilterSyntax = errors.New(
		"ebay: invalid item filter syntax: both itemFilter.name and itemFilter(0).name are present")

	// ErrIncompleteItemFilterNameOnly is returned when an item filter is missing the 'values' parameter.
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

	// ErrInvalidInteger is returned when an item filter 'values' parameter contains an invalid integer.
	ErrInvalidInteger = errors.New("ebay: invalid integer")

	// ErrInvalidFilterRelationship is returned when an item filter relationship is invalid.
	ErrInvalidFilterRelationship = errors.New("ebay: invalid item filter relationship")

	// ErrInvalidExpeditedShippingType is returned when an item filter 'values' parameter
	// contains an invalid expedited shipping type.
	ErrInvalidExpeditedShippingType = errors.New("ebay: invalid expedited shipping type")

	// ErrInvalidGlobalID is returned when an item filter 'values' parameter contains an invalid global ID.
	ErrInvalidGlobalID = errors.New("ebay: invalid global ID")

	// ErrInvalidMaxPrice is returned when an item filter 'values' parameter contains an invalid maximum price.
	ErrInvalidMaxPrice = errors.New("ebay: invalid maximum price")

	// ErrInvalidMinPrice is returned when an item filter 'values' parameter contains an invalid minimum price.
	ErrInvalidMinPrice = errors.New("ebay: invalid minimum price")

	// ErrInvalidPaymentMethod is returned when an item filter 'values' parameter contains an invalid payment method.
	ErrInvalidPaymentMethod = errors.New("ebay: invalid payment method")

	// ErrInvalidSellerBusinessType is returned when an item filter 'values' parameter
	// contains an invalid seller business type.
	ErrInvalidSellerBusinessType = errors.New("ebay: invalid seller business type")

	// ErrMultipleSellerBusinessType is returned when multiple item filters contain seller business types.
	ErrMultipleSellerBusinessType = errors.New("ebay: multiple sellerBusinessType types found")

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

	err = handleItemFilterType(&filter, nil)
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

		filterValues, err := parseItemFilterValues(params, "itemFilter")
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
		err := handleItemFilterType(&itemFilters[i], itemFilters)
		if err != nil {
			return nil, err
		}
	}

	return itemFilters, nil
}

func parseItemFilterValues(params map[string]string, prefix string) ([]string, error) {
	var filterValues []string
	valueExists := false
	if v, ok := params[prefix+".value"]; ok {
		filterValues = []string{v}
		valueExists = true
	} else {
		for i := 0; ; i++ {
			key := fmt.Sprintf(prefix+".value(%d)", i)
			if v, ok := params[key]; ok {
				filterValues = append(filterValues, v)
				valueExists = true
			} else {
				break
			}
		}
	}

	if !valueExists {
		return nil, ErrIncompleteItemFilterNameOnly
	}

	return filterValues, nil
}

const (
	// ItemFilterType values from the eBay documentation.
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
	featuredOnly          = "FeaturedOnly"
	feedbackScoreMax      = "FeedbackScoreMax"
	feedbackScoreMin      = "FeedbackScoreMin"
	freeShippingOnly      = "FreeShippingOnly"
	getItFastOnly         = "GetItFastOnly"
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
	outletSellerOnly      = "OutletSellerOnly"
	paymentMethod         = "PaymentMethod"
	returnsAcceptedOnly   = "ReturnsAcceptedOnly"
	seller                = "Seller"
	sellerBusinessType    = "SellerBusinessType"
	soldItemsOnly         = "SoldItemsOnly"
	startTimeFrom         = "StartTimeFrom"
	startTimeTo           = "StartTimeTo"
	topRatedSellerOnly    = "TopRatedSellerOnly"
	valueBoxInventory     = "ValueBoxInventory"
	worldOfGoodOnly       = "WorldOfGoodOnly"

	trueValue  = "true"
	falseValue = "false"
	trueNum    = "1"
	falseNum   = "0"
)

func handleItemFilterType(filter *itemFilter, itemFilters []itemFilter) error {
	switch filter.name {
	// TODO: Refactor case localSearchOnly - add logic for only can be used with MaxDistance item filter,
	// and the request must also specify buyerPostalCode.
	case authorizedSellerOnly, bestOfferOnly, charityOnly, excludeAutoPay, featuredOnly, freeShippingOnly, getItFastOnly,
		hideDuplicateItems, localPickupOnly, localSearchOnly, lotsOnly, outletSellerOnly, returnsAcceptedOnly, soldItemsOnly,
		worldOfGoodOnly:
		if filter.values[0] != trueValue && filter.values[0] != falseValue {
			return fmt.Errorf("%w: %s", ErrInvalidBooleanValue, filter.values)
		}
	case availableTo:
		if !isValidCountryCode(filter.values[0]) {
			return fmt.Errorf("%w: %s", ErrInvalidCountryCode, filter.values)
		}
	case condition:
		if !isValidCondition(filter.values[0]) {
			return fmt.Errorf("%w: %s", ErrInvalidCondition, filter.values)
		}
	case currency:
		if !isValidCurrencyID(filter.values[0]) {
			return fmt.Errorf("%w: %s", ErrInvalidCurrencyID, filter.values)
		}
	case endTimeFrom, endTimeTo, startTimeFrom, startTimeTo:
		if !isValidDateTime(filter.values[0], true) {
			return fmt.Errorf("%w: %s", ErrInvalidDateTime, filter.values)
		}
	case excludeCategory:
		err := validateExcludeCategory(filter.values)
		if err != nil {
			return err
		}
	// TODO: Implement case excludeSeller
	case expeditedShippingType:
		if filter.values[0] != "Expedited" && filter.values[0] != "OneDayShipping" {
			return fmt.Errorf("%w: %s", ErrInvalidExpeditedShippingType, filter.values)
		}
	case feedbackScoreMax, feedbackScoreMin:
		if !isValidIntegerInRange(filter.values[0], 0) {
			return invalidIntegerError(filter.values[0], 0)
		}

		relatedFilterName := feedbackScoreMax
		isCurrentMin := true
		if filter.name == feedbackScoreMax {
			relatedFilterName = feedbackScoreMin
			isCurrentMin = false
		}

		err := validateNumFilterRelationship(itemFilters, filter.name, filter.values[0], relatedFilterName, isCurrentMin)
		if err != nil {
			return err
		}
	case listedIn:
		if !isValidGlobalID(filter.values[0]) {
			return fmt.Errorf("%w: %s", ErrInvalidGlobalID, filter.values)
		}
	// TODO: Implement case listingType, case locatedIn
	// case locatedIn:
	// 	err := validateLocatedInValues(itemFilters, filter.value)
	// 	if err != nil {
	// 		return err
	// 	}
	case maxBids, minBids:
		if !isValidIntegerInRange(filter.values[0], 0) {
			return invalidIntegerError(filter.values[0], 0)
		}

		relatedFilterName := maxBids
		isCurrentMin := true
		if filter.name == maxBids {
			relatedFilterName = minBids
			isCurrentMin = false
		}

		err := validateNumFilterRelationship(itemFilters, filter.name, filter.values[0], relatedFilterName, isCurrentMin)
		if err != nil {
			return err
		}
	// TODO: Implement case maxDistance
	case maxHandlingTime:
		if !isValidIntegerInRange(filter.values[0], 1) {
			return invalidIntegerError(filter.values[0], 1)
		}
	case maxPrice, minPrice:
		err := validateMinMaxPriceType(itemFilters)
		if err != nil {
			return err
		}
	case maxQuantity, minQuantity:
		if !isValidIntegerInRange(filter.values[0], 1) {
			return invalidIntegerError(filter.values[0], 1)
		}

		relatedFilterName := maxQuantity
		isCurrentMin := true
		if filter.name == maxQuantity {
			relatedFilterName = minQuantity
			isCurrentMin = false
		}

		err := validateNumFilterRelationship(itemFilters, filter.name, filter.values[0], relatedFilterName, isCurrentMin)
		if err != nil {
			return err
		}
	case modTimeFrom:
		if !isValidDateTime(filter.values[0], false) {
			return fmt.Errorf("%w: %s", ErrInvalidDateTime, filter.values)
		}
	case paymentMethod:
		if !isValidPaymentMethod(filter.values[0]) {
			return fmt.Errorf("%w: %s", ErrInvalidPaymentMethod, filter.values)
		}
	// TODO: Implement case seller
	case sellerBusinessType:
		err := validateSellerBusinessType(itemFilters, filter.values[0])
		if err != nil {
			return err
		}
	case topRatedSellerOnly:
		err := validateTopRatedSellerOnly(itemFilters, filter.values[0])
		if err != nil {
			return err
		}
	case valueBoxInventory:
		if filter.values[0] != trueNum && filter.values[0] != falseNum {
			return fmt.Errorf("%w: %s", ErrInvalidValueBoxInventory, filter.values)
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

func validateExcludeCategory(values []string) error {
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

func isValidIntegerInRange(value string, min int) bool {
	num, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	return num >= min
}

func invalidIntegerError(value string, min int) error {
	return fmt.Errorf("%w: %s (minimum value: %d)", ErrInvalidInteger, value, min)
}

func validateNumFilterRelationship(
	itemFilters []itemFilter, currentName, currentValue, relatedName string, isCurrentMin bool,
) error {
	for _, f := range itemFilters {
		if f.name == relatedName {
			relatedValue, err := strconv.Atoi(f.values[0])
			if err != nil {
				return fmt.Errorf("ebay: %w: %s", err, f.values)
			}

			value, err := strconv.Atoi(currentValue)
			if err != nil {
				return fmt.Errorf("ebay: %w: %s", err, currentValue)
			}

			if isCurrentMin && value >= relatedValue {
				return fmt.Errorf("%w: %s must be less than or equal to %s", ErrInvalidFilterRelationship, currentName, relatedName)
			}

			if !isCurrentMin && value <= relatedValue {
				return fmt.Errorf("%w: %s must be greater than or equal to %s",
					ErrInvalidFilterRelationship, currentName, relatedName)
			}

			break
		}
	}

	return nil
}

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

func validateMinMaxPriceType(itemFilters []itemFilter) error {
	var minPriceValue, maxPriceValue float64
	var minPriceParam, maxPriceParam *string
	for _, filter := range itemFilters {
		if filter.name == maxPrice {
			value, err := strconv.ParseFloat(filter.values[0], 64)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidMaxPrice, filter.values)
			}
			maxPriceValue = value
			maxPriceParam = filter.paramName
		} else if filter.name == minPrice {
			value, err := strconv.ParseFloat(filter.values[0], 64)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidMinPrice, filter.values)
			}
			minPriceValue = value
			minPriceParam = filter.paramName
		}
	}

	if maxPriceValue != 0.0 && minPriceValue != 0.0 {
		if maxPriceValue < minPriceValue {
			return fmt.Errorf("%w: maximum price must be greater than or equal to minimum price", ErrInvalidMaxPrice)
		}
	}

	if maxPriceParam != nil && minPriceParam != nil {
		err := validateCurrencyParams(itemFilters, maxPriceParam, minPriceParam)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateCurrencyParams(itemFilters []itemFilter, maxPriceParam, minPriceParam *string) error {
	var maxCurrencyID, minCurrencyID string
	for _, f := range itemFilters {
		if f.paramName != nil && f.paramValue != nil {
			if f.paramName == maxPriceParam {
				maxCurrencyID = *f.paramValue
			} else if f.paramName == minPriceParam {
				minCurrencyID = *f.paramValue
			}
		}
	}

	if maxCurrencyID != "" {
		if !isValidCurrencyID(maxCurrencyID) {
			return fmt.Errorf("%w: %s", ErrInvalidCurrencyID, maxCurrencyID)
		}
	}

	if minCurrencyID != "" {
		if !isValidCurrencyID(minCurrencyID) {
			return fmt.Errorf("%w: %s", ErrInvalidCurrencyID, minCurrencyID)
		}
	}

	return nil
}

var supportedPaymentMethods = []string{
	"AmEx",
	"CashOnPickup",
	"CCAccepted",
	"COD",
	"CreditCard",
	"CustomCode",
	"DirectDebit",
	"Discover",
	"ELV",
	"LoanCheck",
	"MOCC",
	"MoneyXferAccepted",
	"MoneyXferAcceptedInCheckout",
	"None",
	"Other",
	"OtherOnlinePayments",
	"PaymentSeeDescription",
	"PayPal",
	"PersonalCheck",
	"VisaMC",
}

func isValidPaymentMethod(value string) bool {
	for _, method := range supportedPaymentMethods {
		if value == method {
			return true
		}
	}

	return false
}

func validateSellerBusinessType(itemFilters []itemFilter, value string) error {
	if value != "Business" && value != "Private" {
		return fmt.Errorf("%w: %s", ErrInvalidSellerBusinessType, value)
	}

	cnt := 0
	for _, f := range itemFilters {
		if f.name == sellerBusinessType {
			cnt++
		}
	}

	if cnt > 1 {
		return fmt.Errorf("%w: %s", ErrMultipleSellerBusinessType, sellerBusinessType)
	}

	return nil
}

func validateTopRatedSellerOnly(itemFilters []itemFilter, value string) error {
	if value != trueValue && value != falseValue {
		return fmt.Errorf("%w: %s", ErrInvalidBooleanValue, value)
	}

	for _, f := range itemFilters {
		if f.name == seller || f.name == excludeSeller {
			return ErrTopRatedSellerCannotBeUsedWithSellers
		}
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
