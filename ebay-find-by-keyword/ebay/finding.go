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

	// ErrInvalidItemFilterValue is returned when an item filter has an invalid 'value' parameter.
	ErrInvalidItemFilterValue = errors.New("ebay: invalid item filter value")

	// ErrUnsupportedItemFilterType is returned when an item filter 'name' parameter has an unsupported type.
	ErrUnsupportedItemFilterType = errors.New("ebay: unsupported item filter type")

	// ErrInvalidCountryCode is returned when an item filter 'value' parameter contains an invalid country code.
	ErrInvalidCountryCode = errors.New("ebay: invalid country code")

	// ErrInvalidConditionValue is returned when an item filter 'value' parameter contains an invalid condition ID or name.
	ErrInvalidConditionValue = errors.New("ebay: invalid condition value")

	// ErrInvalidCurrencyID is returned when an item filter 'value' parameter contains an invalid currency ID.
	ErrInvalidCurrencyID = errors.New("ebay: invalid currency ID")

	// ErrInvalidDateTime is returned when an item filter 'value' parameter contains an invalid date time value.
	ErrInvalidDateTime = errors.New("ebay: invalid date time value")

	// ErrInvalidNonNegativeInteger is returned when an item filter 'value' parameter contains
	// an invalid non-negative integer.
	ErrInvalidNonNegativeInteger = errors.New("ebay: invalid non-negative integer")

	// ErrInvalidPositiveInteger is returned when an item filter 'value' parameter contains an invalid positive integer.
	ErrInvalidPositiveInteger = errors.New("ebay: invalid positive integer")

	// ErrInvalidFilterRelationship is returned when an item filter relationship is invalid.
	ErrInvalidFilterRelationship = errors.New("ebay: invalid item filter relationship")

	// ErrInvalidExpeditedShippingValue is returned when an item filter 'value' parameter
	// contains an invalid expedited shipping value.
	ErrInvalidExpeditedShippingValue = errors.New("ebay: invalid expedited shipping value")

	// ErrInvalidGlobalID is returned when an item filter 'value' parameter contains an invalid global ID.
	ErrInvalidGlobalID = errors.New("ebay: invalid global ID")
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

	err := handleItemFilterType(&filter, nil)
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
)

func handleItemFilterType(filter *itemFilter, itemFilters []itemFilter) error {
	switch filter.name {
	case authorizedSellerOnly, bestOfferOnly, charityOnly, excludeAutoPay, featuredOnly, freeShippingOnly, getItFastOnly,
		hideDuplicateItems, localPickupOnly, localSearchOnly, lotsOnly, outletSellerOnly, returnsAcceptedOnly, soldItemsOnly,
		topRatedSellerOnly, worldOfGoodOnly:
		if filter.value != trueValue && filter.value != falseValue {
			return fmt.Errorf("%w: %s", ErrInvalidItemFilterValue, filter.value)
		}
	case availableTo:
		if !isValidCountryCode(filter.value) {
			return fmt.Errorf("%w: %s", ErrInvalidCountryCode, filter.value)
		}
	case condition:
		if !isValidConditionValue(filter.value) {
			return fmt.Errorf("%w: %s", ErrInvalidConditionValue, filter.value)
		}
	case currency:
		if !isValidCurrencyID(filter.value) {
			return fmt.Errorf("%w: %s", ErrInvalidCurrencyID, filter.value)
		}
	case endTimeFrom, endTimeTo, startTimeFrom, startTimeTo:
		if !isValidDateTime(filter.value, true) {
			return fmt.Errorf("%w: %s", ErrInvalidDateTime, filter.value)
		}
	case expeditedShippingType:
		if filter.value != "Expedited" && filter.value != "OneDayShipping" {
			return fmt.Errorf("%w: %s", ErrInvalidExpeditedShippingValue, filter.value)
		}
	case feedbackScoreMax, feedbackScoreMin:
		if !isValidIntegerInRange(filter.value, 0) {
			return fmt.Errorf("%w: %s", ErrInvalidNonNegativeInteger, filter.value)
		}

		relatedFilterName := feedbackScoreMin
		isCurrentMin := true
		if filter.name == feedbackScoreMax {
			relatedFilterName = feedbackScoreMin
			isCurrentMin = false
		}

		err := validateNumFilterRelationship(itemFilters, filter.name, filter.value, relatedFilterName, isCurrentMin)
		if err != nil {
			return err
		}
	case listedIn:
		if !isValidGlobalID(filter.value) {
			return fmt.Errorf("%w: %s", ErrInvalidGlobalID, filter.value)
		}
	case maxBids, minBids:
		if !isValidIntegerInRange(filter.value, 0) {
			return fmt.Errorf("%w: %s", ErrInvalidNonNegativeInteger, filter.value)
		}

		relatedFilterName := minBids
		isCurrentMin := true
		if filter.name == maxBids {
			relatedFilterName = minBids
			isCurrentMin = false
		}

		err := validateNumFilterRelationship(itemFilters, filter.name, filter.value, relatedFilterName, isCurrentMin)
		if err != nil {
			return err
		}
	case maxHandlingTime:
		if !isValidIntegerInRange(filter.value, 1) {
			return fmt.Errorf("%w: %s", ErrInvalidPositiveInteger, filter.value)
		}
	case maxQuantity, minQuantity:
		// TODO: Insert maxBids/minBids logic here
		if !isValidIntegerInRange(filter.value, 1) {
			return fmt.Errorf("%w: %s", ErrInvalidPositiveInteger, filter.value)
		}
	case modTimeFrom:
		if !isValidDateTime(filter.value, false) {
			return fmt.Errorf("%w: %s", ErrInvalidDateTime, filter.value)
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

func isValidConditionValue(value string) bool {
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

func isValidIntegerInRange(value string, min int) bool {
	num, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	return num >= min
}

// ErrInvalidInteger is a generic error for an invalid integer value in an item filter.
func ErrInvalidInteger(min int) error {
	// TODO: Use this instead of the NonNegative and Positive errors defined/used above.
	return fmt.Errorf("ebay: invalid integer (minimum value: %d)", min)
}

func validateNumFilterRelationship(
	itemFilters []itemFilter, currentName, currentValue, relatedName string, isCurrentMin bool,
) error {
	for _, f := range itemFilters {
		if f.name == relatedName {
			relatedValue, err := strconv.Atoi(f.value)
			if err != nil {
				return fmt.Errorf("ebay: %w: %s", err, f.value)
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
