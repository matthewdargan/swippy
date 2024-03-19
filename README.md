# Swippy API

[![GoDoc](https://godoc.org/github.com/matthewdargan/swippy-api?status.svg)](https://godoc.org/github.com/matthewdargan/swippy-api)
[![Build Status](https://github.com/matthewdargan/swippy-api/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/matthewdargan/swippy-api/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/matthewdargan/swippy-api)](https://goreportcard.com/report/github.com/matthewdargan/swippy-api)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Swippy API is a RESTful API designed to interact with the
[eBay Finding API](https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide-landing.html)
to perform various searches, retrieve information about items, and save item
data to a database.

## Endpoints

### `GET /find/advanced`

Handles requests to the
[findItemsAdvanced](https://developer.ebay.com/Devzone/finding/CallRef/findItemsAdvanced.html)
eBay Finding API endpoint.

### `GET /find/category`

Handles requests to the
[findItemsByCategory](https://developer.ebay.com/Devzone/finding/CallRef/findItemsByCategory.html)
eBay Finding API endpoint.

### `GET /find/keywords`

Handles requests to the
[findItemsByKeywords](https://developer.ebay.com/Devzone/finding/CallRef/findItemsByKeywords.html)
eBay Finding API endpoint.

### `GET /find/product`

Handles requests to the
[findItemsByProduct](https://developer.ebay.com/Devzone/finding/CallRef/findItemsByProduct.html)
eBay Finding API endpoint.

### `GET /find/ebay-stores`

Handles requests to the
[findItemsIneBayStores](https://developer.ebay.com/Devzone/finding/CallRef/findItemsIneBayStores.html)
eBay Finding API endpoint.
