variable "aws_region" {
  description = "The AWS region where the resources will be created."
  default     = "us-east-1"
}

variable "aws_account_id" {
  description = "The AWS account ID that the resources will reside."
}

variable "ebay_app_id" {
  description = "The eBay App ID."
}

provider "aws" {
  region = var.aws_region
}

resource "aws_ssm_parameter" "ebay_app_id" {
  name        = "ebay-app-id"
  description = "Ebay App ID"
  type        = "SecureString"
  value       = var.ebay_app_id
  key_id      = "alias/aws/ssm"
}

resource "aws_iam_role" "ebay_find_role" {
  name = "swippy-api-lambda-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Effect    = "Allow",
      Principal = { Service = "lambda.amazonaws.com" },
      Action    = "sts:AssumeRole",
    }]
  })
}

resource "aws_iam_role_policy" "ebay_find_policy" {
  name = "swippy-api-lambda-policy"
  role = aws_iam_role.ebay_find_role.id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:TagResource",
        ],
        Resource = "*",
      },
      {
        Effect = "Allow",
        Action = [
          "ssm:GetParameter",
        ],
        Resource = "arn:aws:ssm:${var.aws_region}:${var.aws_account_id}:parameter/ebay-app-id",
      },
    ],
  })
}

resource "aws_lambda_function" "find_by_category" {
  function_name    = "find-by-category"
  handler          = "bin/find-by-category/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/find-by-category.zip"
  source_code_hash = filebase64("bin/find-by-category.zip")
  role             = aws_iam_role.ebay_find_role.arn
  tags             = { Project = "swippy-api" }
}

resource "aws_lambda_function" "find_by_keyword" {
  function_name    = "find-by-keyword"
  handler          = "bin/find-by-keyword/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/find-by-keyword.zip"
  source_code_hash = filebase64("bin/find-by-keyword.zip")
  role             = aws_iam_role.ebay_find_role.arn
  tags             = { Project = "swippy-api" }
}

resource "aws_lambda_function" "find_advanced" {
  function_name    = "find-advanced"
  handler          = "bin/find-advanced/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/find-advanced.zip"
  source_code_hash = filebase64("bin/find-advanced.zip")
  role             = aws_iam_role.ebay_find_role.arn
  tags             = { Project = "swippy-api" }
}

resource "aws_lambda_function" "find_by_product" {
  function_name    = "find-by-product"
  handler          = "bin/find-by-product/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/find-by-product.zip"
  source_code_hash = filebase64("bin/find-by-product.zip")
  role             = aws_iam_role.ebay_find_role.arn
  tags             = { Project = "swippy-api" }
}

resource "aws_lambda_function" "find_in_ebay_stores" {
  function_name    = "find-in-ebay-stores"
  handler          = "bin/find-in-ebay-stores/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/find-in-ebay-stores.zip"
  source_code_hash = filebase64("bin/find-in-ebay-stores.zip")
  role             = aws_iam_role.ebay_find_role.arn
  tags             = { Project = "swippy-api" }
}