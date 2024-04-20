// Copyright 2024 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

// Package swippy-api provides a RESTful API for interacting with the eBay
// Finding API.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/matthewdargan/ebay"
)

func main() {
	log.SetPrefix("swippy-api: ")
	c := ebay.NewFindingClient(&http.Client{Timeout: time.Second * 10}, os.Getenv("EBAY_APP_ID"))
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	http.HandleFunc("GET /find/advanced", makeHandler(conn, func(w http.ResponseWriter, ps map[string]string) []ebay.FindItemsResponse {
		resp, err := c.FindItemsAdvanced(context.Background(), ps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		return resp.ItemsResponse
	}))
	http.HandleFunc("GET /find/category", makeHandler(conn, func(w http.ResponseWriter, ps map[string]string) []ebay.FindItemsResponse {
		resp, err := c.FindItemsByCategory(context.Background(), ps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		return resp.ItemsResponse
	}))
	http.HandleFunc("GET /find/keywords", makeHandler(conn, func(w http.ResponseWriter, ps map[string]string) []ebay.FindItemsResponse {
		resp, err := c.FindItemsByKeywords(context.Background(), ps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		return resp.ItemsResponse
	}))
	http.HandleFunc("GET /find/product", makeHandler(conn, func(w http.ResponseWriter, ps map[string]string) []ebay.FindItemsResponse {
		resp, err := c.FindItemsByProduct(context.Background(), ps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		return resp.ItemsResponse
	}))
	http.HandleFunc("GET /find/ebay-stores", makeHandler(conn, func(w http.ResponseWriter, ps map[string]string) []ebay.FindItemsResponse {
		resp, err := c.FindItemsAdvanced(context.Background(), ps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		return resp.ItemsResponse
	}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type itemHandler func(http.ResponseWriter, map[string]string) []ebay.FindItemsResponse

func makeHandler(conn *pgx.Conn, f itemHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		getItems(w, r, conn, f)
	}
}

const jsonContextType = "application/json"

func getItems(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, f itemHandler) {
	ps, err := params(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	is := f(w, ps)
	data, err := json.Marshal(is)
	if err != nil {
		http.Error(w, fmt.Sprintf("swippy-api: failed to marshal eBay API response: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if len(is) > 0 && len(is[0].ErrorMessage) > 0 {
		w.Header().Set("Content-Type", jsonContextType)
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write(data); err != nil {
			log.Printf("error writing response: %v", err)
		}
		return
	}
	w.Header().Set("Content-Type", jsonContextType)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Printf("error writing response: %v", err)
	}
	if len(is) > 0 {
		insertItems(conn, is)
	}
}

func params(vs url.Values) (map[string]string, error) {
	m := make(map[string]string, len(vs))
	for k, v := range vs {
		if len(v) > 1 {
			return nil, fmt.Errorf("swippy-api: parameter %q contains more than one value", k)
		}
		m[k] = v[0]
	}
	return m, nil
}

type eBayItem struct {
	timestamp                                  time.Time
	version                                    string
	conditionDisplayName                       string
	conditionID                                int
	country                                    string
	galleryURL                                 *string
	globalID                                   string
	isMultiVariationListing                    bool
	itemID                                     int64
	listingInfoBestOfferEnabled                bool
	listingInfoBuyItNowAvailable               bool
	listingInfoEndTime                         time.Time
	listingInfoListingType                     string
	listingInfoStartTime                       time.Time
	listingInfoWatchCount                      *int
	location                                   *string
	postalCode                                 *string
	primaryCategoryID                          int64
	primaryCategoryName                        string
	productIDType                              *string
	productIDValue                             *int64
	sellingStatusConvertedCurrentPriceCurrency *string
	sellingStatusConvertedCurrentPriceValue    *float64
	sellingStatusCurrentPriceCurrency          *string
	sellingStatusCurrentPriceValue             *float64
	sellingStatusSellingState                  *string
	sellingStatusTimeLeft                      *string
	shippingServiceCostCurrency                *string
	shippingServiceCostValue                   *float64
	shippingType                               *string
	shipToLocations                            *string
	subtitle                                   *string
	title                                      string
	topRatedListing                            bool
	viewItemURL                                *string
}

func insertItems(conn *pgx.Conn, rs []ebay.FindItemsResponse) {
	var eBayItems []eBayItem
	for _, r := range rs {
		items, err := responseToItems(r)
		if err != nil {
			log.Printf("failed to convert eBay API response to items: %v", err)
			continue
		}
		eBayItems = append(eBayItems, items...)
	}
	_, err := conn.CopyFrom(
		context.Background(), pgx.Identifier{"item"},
		[]string{
			"timestamp", "version", "condition_display_name", "condition_id",
			"country", "gallery_url", "global_id",
			"is_multi_variation_listing", "item_id",
			"listing_info_best_offer_enabled",
			"listing_info_buy_it_now_available", "listing_info_end_time",
			"listing_info_listing_type",
			"listing_info_start_time", "listing_info_watch_count", "location",
			"postal_code", "primary_category_id", "primary_category_name",
			"product_id_type", "product_id_value",
			"selling_status_converted_current_price_currency",
			"selling_status_converted_current_price_value",
			"selling_status_current_price_currency",
			"selling_status_current_price_value",
			"selling_status_selling_state", "selling_status_time_left",
			"shipping_service_cost_currency", "shipping_service_cost_value",
			"shipping_type", "ship_to_locations", "subtitle", "title",
			"top_rated_listing", "view_item_url",
		},
		pgx.CopyFromSlice(len(eBayItems), func(i int) ([]any, error) {
			return []any{
				eBayItems[i].timestamp, eBayItems[i].version,
				eBayItems[i].conditionDisplayName, eBayItems[i].conditionID,
				eBayItems[i].country, eBayItems[i].galleryURL,
				eBayItems[i].globalID, eBayItems[i].isMultiVariationListing,
				eBayItems[i].itemID,
				eBayItems[i].listingInfoBestOfferEnabled,
				eBayItems[i].listingInfoBuyItNowAvailable,
				eBayItems[i].listingInfoEndTime,
				eBayItems[i].listingInfoListingType,
				eBayItems[i].listingInfoStartTime,
				eBayItems[i].listingInfoWatchCount, eBayItems[i].location,
				eBayItems[i].postalCode, eBayItems[i].primaryCategoryID,
				eBayItems[i].primaryCategoryName, eBayItems[i].productIDType,
				eBayItems[i].productIDValue,
				eBayItems[i].sellingStatusConvertedCurrentPriceCurrency,
				eBayItems[i].sellingStatusConvertedCurrentPriceValue,
				eBayItems[i].sellingStatusCurrentPriceCurrency,
				eBayItems[i].sellingStatusCurrentPriceValue,
				eBayItems[i].sellingStatusSellingState,
				eBayItems[i].sellingStatusTimeLeft,
				eBayItems[i].shippingServiceCostCurrency,
				eBayItems[i].shippingServiceCostValue,
				eBayItems[i].shippingType, eBayItems[i].shipToLocations,
				eBayItems[i].subtitle, eBayItems[i].title,
				eBayItems[i].topRatedListing, eBayItems[i].viewItemURL,
			}, nil
		}),
	)
	if err != nil {
		log.Printf("failed to insert data: %v", err)
	}
}

func responseToItems(resp ebay.FindItemsResponse) ([]eBayItem, error) {
	items := make([]eBayItem, len(resp.SearchResult[0].Item))
	for i := range items {
		it, err := item(resp.SearchResult[0].Item[i])
		if err != nil {
			return nil, err
		}
		it.timestamp = resp.Timestamp[0]
		it.version = resp.Version[0]
		items[i] = *it
	}
	return items, nil
}

func item(it ebay.SearchItem) (*eBayItem, error) {
	conditionID, err := strconv.Atoi(it.Condition[0].ConditionID[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert conditionID to int: %w", err)
	}
	isMultiVariationListing, err := strconv.ParseBool(it.IsMultiVariationListing[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert isMultiVariationListing to bool: %w", err)
	}
	itemID, err := strconv.ParseInt(it.ItemID[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot convert itemID to int64: %w", err)
	}
	bestOfferEnabled, err := strconv.ParseBool(it.ListingInfo[0].BestOfferEnabled[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert bestOfferEnabled to bool: %w", err)
	}
	buyItNowAvailable, err := strconv.ParseBool(it.ListingInfo[0].BuyItNowAvailable[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert buyItNowAvailable to bool: %w", err)
	}
	var watchCount *int
	if len(it.ListingInfo[0].WatchCount) > 0 {
		var v int
		v, err = strconv.Atoi(it.ListingInfo[0].WatchCount[0])
		if err != nil {
			return nil, fmt.Errorf("cannot convert watchCount to int: %w", err)
		}
		watchCount = &v
	}
	primaryCategoryID, err := strconv.ParseInt(it.PrimaryCategory[0].CategoryID[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot convert primaryCategoryID to int64: %w", err)
	}
	var productIDType *string
	var productIDValue *int64
	if len(it.ProductID) > 0 {
		productIDType = &it.ProductID[0].Type
		var v int64
		v, err = strconv.ParseInt(it.ProductID[0].Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert productID value to int64: %w", err)
		}
		productIDValue = &v
	}
	var sellingStatusSellingState, sellingStatusTimeLeft *string
	if len(it.SellingStatus[0].SellingState) > 0 {
		sellingStatusSellingState = &it.SellingStatus[0].SellingState[0]
		sellingStatusTimeLeft = &it.SellingStatus[0].TimeLeft[0]
	}
	var sellingStatusPriceCurrency, sellingStatusConvertedPriceCurrency *string
	var sellingStatusPriceValue, sellingStatusConvertedPriceValue *float64
	if len(it.SellingStatus[0].CurrentPrice) > 0 {
		sellingStatusPriceCurrency = &it.SellingStatus[0].CurrentPrice[0].CurrencyID
		var v float64
		v, err = strconv.ParseFloat(it.SellingStatus[0].CurrentPrice[0].Value, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert selling status current price value to float64: %w", err)
		}
		sellingStatusPriceValue = &v
		sellingStatusConvertedPriceCurrency = &it.SellingStatus[0].ConvertedCurrentPrice[0].CurrencyID
		v, err = strconv.ParseFloat(it.SellingStatus[0].ConvertedCurrentPrice[0].Value, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert selling status converted current price value to float64: %w", err)
		}
		sellingStatusConvertedPriceValue = &v
	}
	var shippingServiceCurrency, shippingType, shipToLocations *string
	var shippingServiceValue *float64
	if len(it.ShippingInfo[0].ShippingServiceCost) > 0 {
		shippingServiceCurrency = &it.ShippingInfo[0].ShippingServiceCost[0].CurrencyID
		var v float64
		v, err = strconv.ParseFloat(it.ShippingInfo[0].ShippingServiceCost[0].Value, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert shipping service cost value to float64: %w", err)
		}
		shippingServiceValue = &v
		shippingType = &it.ShippingInfo[0].ShippingType[0]
		shipToLocations = &it.ShippingInfo[0].ShipToLocations[0]
	}
	topRatedListing, err := strconv.ParseBool(it.TopRatedListing[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert topRatedListing to bool: %w", err)
	}
	return &eBayItem{
		conditionDisplayName:         it.Condition[0].ConditionDisplayName[0],
		conditionID:                  conditionID,
		country:                      it.Country[0],
		galleryURL:                   firstElem(it.GalleryURL),
		globalID:                     it.GlobalID[0],
		isMultiVariationListing:      isMultiVariationListing,
		itemID:                       itemID,
		listingInfoBestOfferEnabled:  bestOfferEnabled,
		listingInfoBuyItNowAvailable: buyItNowAvailable,
		listingInfoEndTime:           it.ListingInfo[0].EndTime[0],
		listingInfoListingType:       it.ListingInfo[0].ListingType[0],
		listingInfoStartTime:         it.ListingInfo[0].StartTime[0],
		listingInfoWatchCount:        watchCount,
		location:                     firstElem(it.Location),
		postalCode:                   firstElem(it.PostalCode),
		primaryCategoryID:            primaryCategoryID,
		primaryCategoryName:          it.PrimaryCategory[0].CategoryName[0],
		productIDType:                productIDType,
		productIDValue:               productIDValue,
		sellingStatusConvertedCurrentPriceCurrency: sellingStatusConvertedPriceCurrency,
		sellingStatusConvertedCurrentPriceValue:    sellingStatusConvertedPriceValue,
		sellingStatusCurrentPriceCurrency:          sellingStatusPriceCurrency,
		sellingStatusCurrentPriceValue:             sellingStatusPriceValue,
		sellingStatusSellingState:                  sellingStatusSellingState,
		sellingStatusTimeLeft:                      sellingStatusTimeLeft,
		shippingServiceCostCurrency:                shippingServiceCurrency,
		shippingServiceCostValue:                   shippingServiceValue,
		shippingType:                               shippingType,
		shipToLocations:                            shipToLocations,
		subtitle:                                   firstElem(it.Subtitle),
		title:                                      it.Title[0],
		topRatedListing:                            topRatedListing,
		viewItemURL:                                firstElem(it.ViewItemURL),
	}, nil
}

func firstElem(ss []string) *string {
	if len(ss) > 0 {
		return &ss[0]
	}
	return nil
}
