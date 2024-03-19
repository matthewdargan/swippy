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
	"time"

	"github.com/matthewdargan/ebay"
)

const jsonContextType = "application/json"

func main() {
	c := &http.Client{Timeout: time.Second * 10}
	findingClient := ebay.NewFindingClient(c, os.Getenv("EBAY_APP_ID"))
	ctx := context.Background()
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
		// TODO: Write to database here.
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
		// TODO: Write to database here.
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
		// TODO: Write to database here.
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
		// TODO: Write to database here.
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
		// TODO: Write to database here.
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
