package ebay

import "time"

type SearchResponse struct {
	FindItemsByKeywordsResponse []FindItemsByKeywordsResponse `json:"findItemsByKeywordsResponse"`
}

type FindItemsByKeywordsResponse struct {
	Ack              []string       `json:"ack"`
	Version          []string       `json:"version"`
	Timestamp        []time.Time    `json:"timestamp"`
	SearchResult     []SearchResult `json:"searchResult"`
	PaginationOutput []Pagination   `json:"paginationOutput"`
	ItemSearchURL    []string       `json:"itemSearchURL"`
}

type SearchResult struct {
	Count string `json:"@count"`
	Item  []Item `json:"item"`
}

type Item struct {
	ItemID                  []string            `json:"itemId"`
	Title                   []string            `json:"title"`
	GlobalID                []string            `json:"globalId"`
	Subtitle                []string            `json:"subtitle"`
	PrimaryCategory         []PrimaryCategory   `json:"primaryCategory"`
	GalleryURL              []string            `json:"galleryURL"`
	ViewItemURL             []string            `json:"viewItemURL"`
	ProductID               []ProductID         `json:"productId"`
	AutoPay                 []string            `json:"autoPay"`
	PostalCode              []string            `json:"postalCode"`
	Location                []string            `json:"location"`
	Country                 []string            `json:"country"`
	ShippingInfo            []ShippingInfo      `json:"shippingInfo"`
	SellingStatus           []SellingStatus     `json:"sellingStatus"`
	ListingInfo             []ListingInfo       `json:"listingInfo"`
	ReturnsAccepted         []string            `json:"returnsAccepted"`
	Condition               []Condition         `json:"condition"`
	IsMultiVariationListing []string            `json:"isMultiVariationListing"`
	TopRatedListing         []string            `json:"topRatedListing"`
	DiscountPriceInfo       []DiscountPriceInfo `json:"discountPriceInfo"`
}

type PrimaryCategory struct {
	CategoryID   []string `json:"categoryId"`
	CategoryName []string `json:"categoryName"`
}

type ProductID struct {
	Type  string `json:"@type"`
	Value string `json:"__value__"`
}

type ShippingInfo struct {
	ShippingServiceCost     []Price  `json:"shippingServiceCost"`
	ShippingType            []string `json:"shippingType"`
	ShipToLocations         []string `json:"shipToLocations"`
	ExpeditedShipping       []string `json:"expeditedShipping"`
	OneDayShippingAvailable []string `json:"oneDayShippingAvailable"`
	HandlingTime            []string `json:"handlingTime"`
}

type Price struct {
	CurrencyID string `json:"@currencyId"`
	Value      string `json:"__value__"`
}

type SellingStatus struct {
	CurrentPrice          []Price  `json:"currentPrice"`
	ConvertedCurrentPrice []Price  `json:"convertedCurrentPrice"`
	SellingState          []string `json:"sellingState"`
	TimeLeft              []string `json:"timeLeft"`
}

type ListingInfo struct {
	BestOfferEnabled  []string    `json:"bestOfferEnabled"`
	BuyItNowAvailable []string    `json:"buyItNowAvailable"`
	StartTime         []time.Time `json:"startTime"`
	EndTime           []time.Time `json:"endTime"`
	ListingType       []string    `json:"listingType"`
	Gift              []string    `json:"gift"`
	WatchCount        []string    `json:"watchCount"`
}

type Condition struct {
	ConditionID          []string `json:"conditionId"`
	ConditionDisplayName []string `json:"conditionDisplayName"`
}

type DiscountPriceInfo struct {
	OriginalRetailPrice []Price  `json:"originalRetailPrice"`
	PricingTreatment    []string `json:"pricingTreatment"`
	SoldOnEbay          []string `json:"soldOnEbay"`
	SoldOffEbay         []string `json:"soldOffEbay"`
}

type Pagination struct {
	PageNumber     []string `json:"pageNumber"`
	EntriesPerPage []string `json:"entriesPerPage"`
	TotalPages     []string `json:"totalPages"`
	TotalEntries   []string `json:"totalEntries"`
}
