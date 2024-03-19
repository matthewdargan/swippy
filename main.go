// Copyright 2024 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

// TODO: write package-level comment here.
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/matthewdargan/ebay"
)

func main() {
	c := &http.Client{Timeout: time.Second * 10}
	findingClient := ebay.NewFindingClient(c, os.Getenv("EBAY_APP_ID"))
	mux := http.NewServeMux()
	mux.HandleFunc("/find/advanced", func(w http.ResponseWriter, r *http.Request) {
		resp, err := findingClient.FindItemsAdvanced(ctx, req.QueryStringParameters)
	})
	mux.HandleFunc("/find/category", func(w http.ResponseWriter, r *http.Request) {
		resp, err := findingClient.FindItemsByCategory(ctx, req.QueryStringParameters)
	})
	mux.HandleFunc("/find/keywords", func(w http.ResponseWriter, r *http.Request) {
		resp, err := findingClient.FindItemsByKeywords(ctx, req.QueryStringParameters)
	})
	mux.HandleFunc("/find/product", func(w http.ResponseWriter, r *http.Request) {
		resp, err := findingClient.FindItemsByProduct(ctx, req.QueryStringParameters)
	})
	mux.HandleFunc("/find/ebay-stores", func(w http.ResponseWriter, r *http.Request) {
		resp, err := findingClient.FindItemsInEBayStores(ctx, req.QueryStringParameters)
	})
}
