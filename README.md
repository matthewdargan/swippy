# Swippy API

[![GoDoc](https://godoc.org/github.com/matthewdargan/swippy-api?status.svg)](https://godoc.org/github.com/matthewdargan/swippy-api)
[![Build Status](https://github.com/matthewdargan/swippy-api/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/matthewdargan/swippy-api/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/matthewdargan/swippy-api)](https://goreportcard.com/report/github.com/matthewdargan/swippy-api)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Swippy API is a serverless app designed to interact with the [eBay Finding API](https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide-landing.html) to perform various searches, retrieve information about items, and save item data to a database.

## Architecture

![Architecture Diagram](docs/swippy_architecture.drawio.svg)

The API architecture includes AWS API Gateway, Lambda, SQS, and PostgreSQL. The [OpenTofu](https://opentofu.org) configuration sets up the infrastructure for the Swippy API.

### Lambda functions

Each Lambda function uses [Go](https://go.dev) for its implementation.

#### Find-advanced

This Lambda function handles requests to the [findItemsAdvanced](https://developer.ebay.com/Devzone/finding/CallRef/findItemsAdvanced.html) eBay Finding API endpoint.

#### Find-by-category

This Lambda function handles requests to the [findItemsByCategory](https://developer.ebay.com/Devzone/finding/CallRef/findItemsByCategory.html) eBay Finding API endpoint.

#### Find-by-keywords

This Lambda function handles requests to the [findItemsByKeywords](https://developer.ebay.com/Devzone/finding/CallRef/findItemsByKeywords.html) eBay Finding API endpoint.

#### Find-by-product

This Lambda function handles requests to the [findItemsByProduct](https://developer.ebay.com/Devzone/finding/CallRef/findItemsByProduct.html) eBay Finding API endpoint.

#### Find-in-ebay-stores

This Lambda function handles requests to the [findItemsIneBayStores](https://developer.ebay.com/Devzone/finding/CallRef/findItemsIneBayStores.html) eBay Finding API endpoint.

#### Db-insert

`swippy-api-queue` triggers this Lambda function to insert eBay item data into the Swippy Database.

## Installation

### Prerequisites

- [OpenTofu](https://opentofu.org/docs/intro/install/)
- AWS credentials configured locally

### Usage

Clone this repository:

```sh
git@github.com:matthewdargan/swippy-api.git
```

Create a `terraform.tfvars` file at the root project directory with your specific values, including `ebay_app_id`, `swippy_db_url`, and optionally `aws_region` (defaults to us-east-1).

Build and zip the Lambda functions, and initialize and apply the OpenTofu configuration:

```sh
make apply
```

Follow the prompts to confirm the changes and provision the Swippy API infrastructure.

### Cleanup

To remove all provisioned resources, run the following command:

```sh
tofu destroy
```

Follow the prompts to confirm the destruction of the resources.
