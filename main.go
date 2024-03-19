// Copyright 2024 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

// Package swippy-api provides an HTTP server for interacting with the eBay
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

const jsonContextType = "application/json"

func main() {
	log.SetPrefix("swippy-api: ")
	c := &http.Client{Timeout: time.Second * 10}
	findingClient := ebay.NewFindingClient(c, os.Getenv("EBAY_APP_ID"))
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("SWIPPY_DB_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/find/advanced", func(w http.ResponseWriter, r *http.Request) {
		params, err := urlParams(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := findingClient.FindItemsAdvanced(ctx, params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "swippy-api: failed to marshal eBay API response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(resp.ItemsResponse) > 0 && len(resp.ItemsResponse[0].ErrorMessage) > 0 {
			w.Header().Set("Content-Type", jsonContextType)
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write(data); err != nil {
				log.Printf("error writing response: %v", err)
			}
			return
		}
		insertItemsResponse(ctx, conn, resp.ItemsResponse)
		w.Header().Set("Content-Type", jsonContextType)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			log.Printf("error writing response: %v", err)
		}
	})
	mux.HandleFunc("/find/category", func(w http.ResponseWriter, r *http.Request) {
		params, err := urlParams(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := findingClient.FindItemsByCategory(ctx, params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "swippy-api: failed to marshal eBay API response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(resp.ItemsResponse) > 0 && len(resp.ItemsResponse[0].ErrorMessage) > 0 {
			w.Header().Set("Content-Type", jsonContextType)
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write(data); err != nil {
				log.Printf("error writing response: %v", err)
			}
			return
		}
		insertItemsResponse(ctx, conn, resp.ItemsResponse)
		w.Header().Set("Content-Type", jsonContextType)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			log.Printf("error writing response: %v", err)
		}
	})
	mux.HandleFunc("/find/keywords", func(w http.ResponseWriter, r *http.Request) {
		params, err := urlParams(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := findingClient.FindItemsByKeywords(ctx, params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "swippy-api: failed to marshal eBay API response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(resp.ItemsResponse) > 0 && len(resp.ItemsResponse[0].ErrorMessage) > 0 {
			w.Header().Set("Content-Type", jsonContextType)
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write(data); err != nil {
				log.Printf("error writing response: %v", err)
			}
			return
		}
		insertItemsResponse(ctx, conn, resp.ItemsResponse)
		w.Header().Set("Content-Type", jsonContextType)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			log.Printf("error writing response: %v", err)
		}
	})
	mux.HandleFunc("/find/product", func(w http.ResponseWriter, r *http.Request) {
		params, err := urlParams(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := findingClient.FindItemsByProduct(ctx, params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "swippy-api: failed to marshal eBay API response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(resp.ItemsResponse) > 0 && len(resp.ItemsResponse[0].ErrorMessage) > 0 {
			w.Header().Set("Content-Type", jsonContextType)
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write(data); err != nil {
				log.Printf("error writing response: %v", err)
			}
			return
		}
		insertItemsResponse(ctx, conn, resp.ItemsResponse)
		w.Header().Set("Content-Type", jsonContextType)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			log.Printf("error writing response: %v", err)
		}
	})
	mux.HandleFunc("/find/ebay-stores", func(w http.ResponseWriter, r *http.Request) {
		params, err := urlParams(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := findingClient.FindItemsInEBayStores(ctx, params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "swippy-api: failed to marshal eBay API response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(resp.ItemsResponse) > 0 && len(resp.ItemsResponse[0].ErrorMessage) > 0 {
			w.Header().Set("Content-Type", jsonContextType)
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write(data); err != nil {
				log.Printf("error writing response: %v", err)
			}
			return
		}
		insertItemsResponse(ctx, conn, resp.ItemsResponse)
		w.Header().Set("Content-Type", jsonContextType)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			log.Printf("error writing response: %v", err)
		}
	})
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}

func urlParams(vs url.Values) (map[string]string, error) {
	m := make(map[string]string, len(vs))
	for k, v := range vs {
		if len(v) > 1 {
			return nil, fmt.Errorf("swippy-api: query string parameter %s contains more than one value", k)
		}
		m[k] = v[0]
	}
	return m, nil
}

type eBayEntry struct {
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

func insertItemsResponse(ctx context.Context, conn *pgx.Conn, rs []ebay.FindItemsResponse) {
	for _, r := range rs {
		eBayEntries, err := responseToEBayEntry(r)
		if err != nil {
			log.Printf("failed to convert eBay API response to entries: %v", err)
			continue
		}
		_, err = conn.CopyFrom(
			ctx,
			pgx.Identifier{"ebay_responses"},
			[]string{
				"timestamp", "version", "condition_display_name", "condition_id", "country", "gallery_url", "global_id", "is_multi_variation_listing", "item_id",
				"listing_info_best_offer_enabled", "listing_info_buy_it_now_available", "listing_info_end_time", "listing_info_listing_type",
				"listing_info_start_time", "listing_info_watch_count", "location", "postal_code", "primary_category_id", "primary_category_name",
				"product_id_type", "product_id_value", "selling_status_converted_current_price_currency", "selling_status_converted_current_price_value",
				"selling_status_current_price_currency", "selling_status_current_price_value", "selling_status_selling_state", "selling_status_time_left",
				"shipping_service_cost_currency", "shipping_service_cost_value", "shipping_type", "ship_to_locations", "subtitle", "title",
				"top_rated_listing", "view_item_url",
			},
			pgx.CopyFromSlice(len(eBayEntries), func(i int) ([]any, error) {
				return []any{
					eBayEntries[i].timestamp, eBayEntries[i].version, eBayEntries[i].conditionDisplayName, eBayEntries[i].conditionID, eBayEntries[i].country,
					eBayEntries[i].galleryURL, eBayEntries[i].globalID, eBayEntries[i].isMultiVariationListing, eBayEntries[i].itemID,
					eBayEntries[i].listingInfoBestOfferEnabled, eBayEntries[i].listingInfoBuyItNowAvailable, eBayEntries[i].listingInfoEndTime,
					eBayEntries[i].listingInfoListingType, eBayEntries[i].listingInfoStartTime, eBayEntries[i].listingInfoWatchCount, eBayEntries[i].location,
					eBayEntries[i].postalCode, eBayEntries[i].primaryCategoryID, eBayEntries[i].primaryCategoryName, eBayEntries[i].productIDType,
					eBayEntries[i].productIDValue, eBayEntries[i].sellingStatusConvertedCurrentPriceCurrency, eBayEntries[i].sellingStatusConvertedCurrentPriceValue,
					eBayEntries[i].sellingStatusCurrentPriceCurrency, eBayEntries[i].sellingStatusCurrentPriceValue, eBayEntries[i].sellingStatusSellingState,
					eBayEntries[i].sellingStatusTimeLeft, eBayEntries[i].shippingServiceCostCurrency, eBayEntries[i].shippingServiceCostValue,
					eBayEntries[i].shippingType, eBayEntries[i].shipToLocations, eBayEntries[i].subtitle, eBayEntries[i].title,
					eBayEntries[i].topRatedListing, eBayEntries[i].viewItemURL,
				}, nil
			}),
		)
		if err != nil {
			log.Printf("failed to insert data: %v", err)
		}
	}
}

func responseToEBayEntry(resp ebay.FindItemsResponse) ([]eBayEntry, error) {
	eBayEntries := make([]eBayEntry, len(resp.SearchResult[0].Item))
	for i := range eBayEntries {
		entry, err := newEBayEntry(resp.SearchResult[0].Item[i])
		if err != nil {
			return nil, err
		}
		entry.timestamp = resp.Timestamp[0]
		entry.version = resp.Version[0]
		eBayEntries[i] = *entry
	}
	return eBayEntries, nil
}

func newEBayEntry(item ebay.SearchItem) (*eBayEntry, error) {
	conditionID, err := strconv.Atoi(item.Condition[0].ConditionID[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert conditionID to int: %w", err)
	}
	isMultiVariationListing, err := strconv.ParseBool(item.IsMultiVariationListing[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert isMultiVariationListing to bool: %w", err)
	}
	itemID, err := strconv.ParseInt(item.ItemID[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot convert itemID to int64: %w", err)
	}
	bestOfferEnabled, err := strconv.ParseBool(item.ListingInfo[0].BestOfferEnabled[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert bestOfferEnabled to bool: %w", err)
	}
	buyItNowAvailable, err := strconv.ParseBool(item.ListingInfo[0].BuyItNowAvailable[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert buyItNowAvailable to bool: %w", err)
	}
	var watchCount *int
	if len(item.ListingInfo[0].WatchCount) > 0 {
		v, cErr := strconv.Atoi(item.ListingInfo[0].WatchCount[0])
		if cErr != nil {
			return nil, fmt.Errorf("cannot convert watchCount to int: %w", cErr)
		}
		watchCount = &v
	}
	primaryCategoryID, err := strconv.ParseInt(item.PrimaryCategory[0].CategoryID[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot convert primaryCategoryID to int64: %w", err)
	}
	var productIDType *string
	var productIDValue *int64
	if len(item.ProductID) > 0 {
		productIDType = &item.ProductID[0].Type
		v, cErr := strconv.ParseInt(item.ProductID[0].Value, 10, 64)
		if cErr != nil {
			return nil, fmt.Errorf("cannot convert productID value to int64: %w", cErr)
		}
		productIDValue = &v
	}
	var sellingStatusSellingState, sellingStatusTimeLeft *string
	if len(item.SellingStatus[0].SellingState) > 0 {
		sellingStatusSellingState = &item.SellingStatus[0].SellingState[0]
		sellingStatusTimeLeft = &item.SellingStatus[0].TimeLeft[0]
	}
	var sellingStatusPriceCurrency, sellingStatusConvertedPriceCurrency *string
	var sellingStatusPriceValue, sellingStatusConvertedPriceValue *float64
	if len(item.SellingStatus[0].CurrentPrice) > 0 {
		sellingStatusPriceCurrency = &item.SellingStatus[0].CurrentPrice[0].CurrencyID
		v, cErr := strconv.ParseFloat(item.SellingStatus[0].CurrentPrice[0].Value, 64)
		if cErr != nil {
			return nil, fmt.Errorf("cannot convert selling status current price value to float64: %w", cErr)
		}
		sellingStatusPriceValue = &v
		sellingStatusConvertedPriceCurrency = &item.SellingStatus[0].ConvertedCurrentPrice[0].CurrencyID
		v, cErr = strconv.ParseFloat(item.SellingStatus[0].ConvertedCurrentPrice[0].Value, 64)
		if cErr != nil {
			return nil, fmt.Errorf("cannot convert selling status converted current price value to float64: %w", cErr)
		}
		sellingStatusConvertedPriceValue = &v
	}
	var shippingServiceCurrency, shippingType, shipToLocations *string
	var shippingServiceValue *float64
	if len(item.ShippingInfo[0].ShippingServiceCost) > 0 {
		shippingServiceCurrency = &item.ShippingInfo[0].ShippingServiceCost[0].CurrencyID
		v, cErr := strconv.ParseFloat(item.ShippingInfo[0].ShippingServiceCost[0].Value, 64)
		if cErr != nil {
			return nil, fmt.Errorf("cannot convert shipping service cost value to float64: %w", cErr)
		}
		shippingServiceValue = &v
		shippingType = &item.ShippingInfo[0].ShippingType[0]
		shipToLocations = &item.ShippingInfo[0].ShipToLocations[0]
	}
	topRatedListing, err := strconv.ParseBool(item.TopRatedListing[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert topRatedListing to bool: %w", err)
	}
	return &eBayEntry{
		conditionDisplayName:         item.Condition[0].ConditionDisplayName[0],
		conditionID:                  conditionID,
		country:                      item.Country[0],
		galleryURL:                   firstElem(item.GalleryURL),
		globalID:                     item.GlobalID[0],
		isMultiVariationListing:      isMultiVariationListing,
		itemID:                       itemID,
		listingInfoBestOfferEnabled:  bestOfferEnabled,
		listingInfoBuyItNowAvailable: buyItNowAvailable,
		listingInfoEndTime:           item.ListingInfo[0].EndTime[0],
		listingInfoListingType:       item.ListingInfo[0].ListingType[0],
		listingInfoStartTime:         item.ListingInfo[0].StartTime[0],
		listingInfoWatchCount:        watchCount,
		location:                     firstElem(item.Location),
		postalCode:                   firstElem(item.PostalCode),
		primaryCategoryID:            primaryCategoryID,
		primaryCategoryName:          item.PrimaryCategory[0].CategoryName[0],
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
		subtitle:                                   firstElem(item.Subtitle),
		title:                                      item.Title[0],
		topRatedListing:                            topRatedListing,
		viewItemURL:                                firstElem(item.ViewItemURL),
	}, nil
}

func firstElem(ss []string) *string {
	if len(ss) > 0 {
		return &ss[0]
	}
	return nil
}
