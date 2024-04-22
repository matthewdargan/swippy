# Swippy

[![GoDoc](https://godoc.org/github.com/matthewdargan/swippy-api?status.svg)](https://godoc.org/github.com/matthewdargan/swippy-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/matthewdargan/swippy-api)](https://goreportcard.com/report/github.com/matthewdargan/swippy-api)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Swippy retrieves from the
[eBay Finding API](https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide-landing.html)
and stores results in a PostgreSQL database.

Usage:

    swippy -m method -p params

The `-m` flag indicates the eBay Finding API method to call.

The `-p` flag specifies the query parameters for the eBay Finding API call.

The `EBAY_APP_ID` and `DB_URL` environment variables are required.

## Examples

Retrieve phones by keyword:

```sh
swippy -m keyword -p 'keywords=phone'
```

Retrieve phones by category:

```sh
swippy -m category -p 'categoryId=9355'
```
