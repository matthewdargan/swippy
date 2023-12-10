provider "aws" {
  region = "us-east-1"
}

resource "aws_iam_role" "ebay_find_role" {
  name = "ebay-find-function-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow",
      Principal = { Service = "lambda.amazonaws.com" },
      Action    = "sts:AssumeRole",
    }]
  })
}

resource "aws_iam_role_policy" "ebay_find_policy" {
  name = "swippy-api-ebay-find-function-policy"
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
        Resource = "arn:aws:ssm::*:parameter/ebay-app-id",
      },
    ],
  })
}

resource "aws_lambda_function" "ebay_find_by_category" {
  function_name    = "ebay-find-by-category"
  handler          = "bin/ebay-find-by-category/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/ebay-find-by-category.zip"
  source_code_hash = filebase64("bin/ebay-find-by-category.zip")
  role             = aws_iam_role.ebay_find_role.arn
}

resource "aws_lambda_function" "ebay_find_by_keyword" {
  function_name    = "ebay-find-by-keyword"
  handler          = "bin/ebay-find-by-keyword/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/ebay-find-by-keyword.zip"
  source_code_hash = filebase64("bin/ebay-find-by-keyword.zip")
  role             = aws_iam_role.ebay_find_role.arn
}

resource "aws_lambda_function" "ebay_find_advanced" {
  function_name    = "ebay-find-advanced"
  handler          = "bin/ebay-find-advanced/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/ebay-find-advanced.zip"
  source_code_hash = filebase64("bin/ebay-find-advanced.zip")
  role             = aws_iam_role.ebay_find_role.arn
}

resource "aws_lambda_function" "ebay_find_by_product" {
  function_name    = "ebay-find-by-product"
  handler          = "bin/ebay-find-by-product/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/ebay-find-by-product.zip"
  source_code_hash = filebase64("bin/ebay-find-by-product.zip")
  role             = aws_iam_role.ebay_find_role.arn
}

resource "aws_lambda_function" "ebay_find_in_ebay_stores" {
  function_name    = "ebay-find-in-ebay-stores"
  handler          = "bin/ebay-find-in-ebay-stores/bootstrap"
  runtime          = "provided.al2"
  filename         = "bin/ebay-find-in-ebay-stores.zip"
  source_code_hash = filebase64("bin/ebay-find-in-ebay-stores.zip")
  role             = aws_iam_role.ebay_find_role.arn
}