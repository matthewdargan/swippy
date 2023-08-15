package ebay

import "time"

type FindItems interface {
	Items() []FindItemsResponse
}

type FindItemsByCategoriesResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsByCategoryResponse"`
}

func (r FindItemsByCategoriesResponse) Items() []FindItemsResponse {
	return r.ItemsResponse
}

type FindItemsByKeywordsResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsByKeywordsResponse"`
}

func (r FindItemsByKeywordsResponse) Items() []FindItemsResponse {
	return r.ItemsResponse
}

type FindItemsAdvancedResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsAdvancedResponse"`
}

func (r FindItemsAdvancedResponse) Items() []FindItemsResponse {
	return r.ItemsResponse
}

type FindItemsByProductResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsByProductResponse"`
}

func (r FindItemsByProductResponse) Items() []FindItemsResponse {
	return r.ItemsResponse
}

type FindItemsInEBayStoresResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsIneBayStoresResponse"`
}

func (r FindItemsInEBayStoresResponse) Items() []FindItemsResponse {
	return r.ItemsResponse
}

type FindItemsResponse struct {
	Ack              []string           `json:"ack"`
	ErrorMessage     []ErrorMessage     `json:"errorMessage"`
	ItemSearchURL    []string           `json:"itemSearchURL"`
	PaginationOutput []PaginationOutput `json:"paginationOutput"`
	SearchResult     []SearchResult     `json:"searchResult"`
	Timestamp        []time.Time        `json:"timestamp"`
	Version          []string           `json:"version"`
}

type ErrorMessage struct {
	Error []ErrorData `json:"error"`
}

type ErrorData struct {
	Category    []string         `json:"category"`
	Domain      []string         `json:"domain"`
	ErrorID     []string         `json:"errorId"`
	ExceptionID []string         `json:"exceptionId"`
	Message     []string         `json:"message"`
	Parameter   []ErrorParameter `json:"parameter"`
	Severity    []string         `json:"severity"`
	Subdomain   []string         `json:"subdomain"`
}

type ErrorParameter struct {
	Name  string `json:"@name"`
	Value string `json:"__value__"`
}

type PaginationOutput struct {
	EntriesPerPage []string `json:"entriesPerPage"`
	PageNumber     []string `json:"pageNumber"`
	TotalEntries   []string `json:"totalEntries"`
	TotalPages     []string `json:"totalPages"`
}

type SearchResult struct {
	Count string       `json:"@count"`
	Item  []SearchItem `json:"item"`
}

type SearchItem struct {
	AutoPay                 []string            `json:"autoPay"`
	CharityID               []string            `json:"charityId"`
	Compatibility           []string            `json:"compatibility"`
	Condition               []Condition         `json:"condition"`
	Country                 []string            `json:"country"`
	DiscountPriceInfo       []DiscountPriceInfo `json:"discountPriceInfo"`
	Distance                []Distance          `json:"distance"`
	EBayPlusEnabled         []string            `json:"eBayPlusEnabled"`
	EekStatus               []string            `json:"eekStatus"`
	GalleryInfoContainer    []GalleryURL        `json:"galleryInfoContainer"`
	GalleryPlusPictureURL   []string            `json:"galleryPlusPictureURL"`
	GalleryURL              []string            `json:"galleryURL"`
	GlobalID                []string            `json:"globalId"`
	IsMultiVariationListing []string            `json:"isMultiVariationListing"`
	ItemID                  []string            `json:"itemId"`
	ListingInfo             []ListingInfo       `json:"listingInfo"`
	Location                []string            `json:"location"`
	PaymentMethod           []string            `json:"paymentMethod"`
	PictureURLLarge         []string            `json:"pictureURLLarge"`
	PictureURLSuperSize     []string            `json:"pictureURLSuperSize"`
	PostalCode              []string            `json:"postalCode"`
	PrimaryCategory         []Category          `json:"primaryCategory"`
	ProductID               []ProductID         `json:"productId"`
	ReturnsAccepted         []string            `json:"returnsAccepted"`
	SecondaryCategory       []Category          `json:"secondaryCategory"`
	SellerInfo              []SellerInfo        `json:"sellerInfo"`
	SellingStatus           []SellingStatus     `json:"sellingStatus"`
	ShippingInfo            []ShippingInfo      `json:"shippingInfo"`
	StoreInfo               []Storefront        `json:"storeInfo"`
	Subtitle                []string            `json:"subtitle"`
	Title                   []string            `json:"title"`
	TopRatedListing         []string            `json:"topRatedListing"`
	UnitPrice               []UnitPriceInfo     `json:"unitPrice"`
	ViewItemURL             []string            `json:"viewItemURL"`
}

type Condition struct {
	ConditionDisplayName []string `json:"conditionDisplayName"`
	ConditionID          []string `json:"conditionId"`
}

type DiscountPriceInfo struct {
	MinimumAdvertisedPriceExposure []string `json:"minimumAdvertisedPriceExposure"`
	OriginalRetailPrice            []Price  `json:"originalRetailPrice"`
	PricingTreatment               []string `json:"pricingTreatment"`
	SoldOffEbay                    []string `json:"soldOffEbay"`
	SoldOnEbay                     []string `json:"soldOnEbay"`
}

type Price struct {
	CurrencyID string `json:"@currencyId"`
	Value      string `json:"__value__"`
}

type Distance struct {
	Unit  string `json:"@unit"`
	Value string `json:"__value__"`
}

type GalleryURL struct {
	GallerySize string `json:"@gallerySize"`
	Value       string `json:"__value__"`
}

type ListingInfo struct {
	BestOfferEnabled       []string    `json:"bestOfferEnabled"`
	BuyItNowAvailable      []string    `json:"buyItNowAvailable"`
	BuyItNowPrice          []Price     `json:"buyItNowPrice"`
	ConvertedBuyItNowPrice []Price     `json:"convertedBuyItNowPrice"`
	EndTime                []time.Time `json:"endTime"`
	Gift                   []string    `json:"gift"`
	ListingType            []string    `json:"listingType"`
	StartTime              []time.Time `json:"startTime"`
	WatchCount             []string    `json:"watchCount"`
}

type Category struct {
	CategoryID   []string `json:"categoryId"`
	CategoryName []string `json:"categoryName"`
}

type ProductID struct {
	Type  string `json:"@type"`
	Value string `json:"__value__"`
}

type SellerInfo struct {
	FeedbackRatingStar      []string `json:"feedbackRatingStar"`
	FeedbackScore           []string `json:"feedbackScore"`
	PositiveFeedbackPercent []string `json:"positiveFeedbackPercent"`
	SellerUserName          []string `json:"sellerUserName"`
	TopRatedSeller          []string `json:"topRatedSeller"`
}

type SellingStatus struct {
	BidCount              []string `json:"bidCount"`
	ConvertedCurrentPrice []Price  `json:"convertedCurrentPrice"`
	CurrentPrice          []Price  `json:"currentPrice"`
	SellingState          []string `json:"sellingState"`
	TimeLeft              []string `json:"timeLeft"`
}

type ShippingInfo struct {
	ExpeditedShipping       []string `json:"expeditedShipping"`
	HandlingTime            []string `json:"handlingTime"`
	IntermediatedShipping   []string `json:"intermediatedShipping"`
	OneDayShippingAvailable []string `json:"oneDayShippingAvailable"`
	ShippingServiceCost     []Price  `json:"shippingServiceCost"`
	ShippingType            []string `json:"shippingType"`
	ShipToLocations         []string `json:"shipToLocations"`
}

type Storefront struct {
	StoreName []string `json:"storeName"`
	StoreURL  []string `json:"storeURL"`
}

type UnitPriceInfo struct {
	Quantity []string `json:"quantity"`
	Type     []string `json:"type"`
}
