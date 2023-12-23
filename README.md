# Swippy API

Swippy API is a serverless application designed to interact with the [eBay Finding API](https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide-landing.html) to perform various searches, retrieve information about items, and save item data to a database.

## Architecture

![Architecture Diagram](docs/swippy_architecture.drawio.svg)

The API is architected with AWS API Gateway, Lambda, SQS, and PostgreSQL.

### Lambda Functions

Each Lambda function uses [Go](https://go.dev) for its implementation.

#### find-advanced

This Lambda function handles requests to the [findItemsAdvanced](https://developer.ebay.com/Devzone/finding/CallRef/findItemsAdvanced.html) eBay Finding API endpoint.

#### find-by-category

This Lambda function handles requests to the [findItemsByCategory](https://developer.ebay.com/Devzone/finding/CallRef/findItemsByCategory.html) eBay Finding API endpoint.

#### find-by-keywords

This Lambda function handles requests to the [findItemsByKeywords](https://developer.ebay.com/Devzone/finding/CallRef/findItemsByKeywords.html) eBay Finding API endpoint.

#### find-by-product

This Lambda function handles requests to the [findItemsByProduct](https://developer.ebay.com/Devzone/finding/CallRef/findItemsByProduct.html) eBay Finding API endpoint.

#### find-in-ebay-stores

This Lambda function handles requests to the [findItemsIneBayStores](https://developer.ebay.com/Devzone/finding/CallRef/findItemsIneBayStores.html) eBay Finding API endpoint.

### OpenTofu Configuration

The [OpenTofu](https://opentofu.org) configuration sets up the infrastructure for the Swippy API. It includes the following AWS resources:

- S3 Bucket for OpenTofu State (aws_s3_bucket.tfstate)
- DynamoDB Table for OpenTofu State Lock (aws_dynamodb_table.tflock)
- Lambda Functions (aws_lambda_function)
- IAM Lambda Execution Roles (aws_iam_role)
- IAM Role Policies for Lambda Functions (aws_iam_role_policy)
- CloudWatch Log Groups (aws_cloudwatch_log_group)
- API Gateway (aws_apigatewayv2_api)
- API Gateway Stage (aws_apigatewayv2_stage)
- API Gateway Integrations and Routes for each Lambda function

## Installation

### Prerequisites

- [OpenTofu](https://opentofu.org/docs/intro/install/)
- AWS credentials configured locally

### Usage

Clone this repository:

```sh
git@github.com:matthewdargan/swippy-api.git
```

Create a `terraform.tfvars` file at the root project directory with your specific values, including `aws_account_id`, `ebay_app_id`, and optionally `aws_region` (defaults to us-east-1).

Build and zip the Lambda functions, and initialize and apply the OpenTofu configuration:

```sh
make apply
```

Follow the prompts to confirm the changes. Once you finish this process, AWS will provision the Swippy API infrastructure.

### Cleanup

To remove all provisioned resources, run the following command:

```sh
tofu destroy
```

Follow the prompts to confirm the destruction of the resources.
