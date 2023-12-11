variable "aws_region" {
  description = "AWS Region"
  default     = "us-east-1"
}

provider "aws" {
  region = var.aws_region
}

variable "ebay_app_id" {
  description = "eBay App ID"
}

resource "aws_ssm_parameter" "ebay_app_id" {
  name        = "ebay-app-id"
  description = "eBay App ID"
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

variable "aws_account_id" {
  description = "AWS Account ID"
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

resource "aws_cloudwatch_log_group" "find_by_category_logs" {
  name = "/aws/lambda/${aws_lambda_function.find_by_category.function_name}"
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

resource "aws_cloudwatch_log_group" "find_by_keyword_logs" {
  name = "/aws/lambda/${aws_lambda_function.find_by_keyword.function_name}"
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

resource "aws_cloudwatch_log_group" "find_advanced_logs" {
  name = "/aws/lambda/${aws_lambda_function.find_advanced.function_name}"
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

resource "aws_cloudwatch_log_group" "find_by_product_logs" {
  name = "/aws/lambda/${aws_lambda_function.find_by_product.function_name}"
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

resource "aws_cloudwatch_log_group" "find_in_ebay_stores_logs" {
  name = "/aws/lambda/${aws_lambda_function.find_in_ebay_stores.function_name}"
}

resource "aws_api_gateway_rest_api" "swippy_api" {
  name        = "swippy-api"
  description = "Swippy API Gateway"
}

resource "aws_api_gateway_resource" "root" {
  rest_api_id = aws_api_gateway_rest_api.swippy_api.id
  parent_id   = aws_api_gateway_rest_api.swippy_api.root_resource_id
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "proxy" {
  rest_api_id   = aws_api_gateway_rest_api.swippy_api.id
  resource_id   = aws_api_gateway_resource.root.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "find_by_category_lambda" {
  rest_api_id             = aws_api_gateway_rest_api.swippy_api.id
  resource_id             = aws_api_gateway_resource.root.id
  http_method             = aws_api_gateway_method.proxy.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.find_by_category.invoke_arn
}

resource "aws_api_gateway_integration" "find_by_keyword_lambda" {
  rest_api_id             = aws_api_gateway_rest_api.swippy_api.id
  resource_id             = aws_api_gateway_resource.root.id
  http_method             = aws_api_gateway_method.proxy.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.find_by_keyword.invoke_arn
}

resource "aws_api_gateway_integration" "find_advanced_lambda" {
  rest_api_id             = aws_api_gateway_rest_api.swippy_api.id
  resource_id             = aws_api_gateway_resource.root.id
  http_method             = aws_api_gateway_method.proxy.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.find_advanced.invoke_arn
}

resource "aws_api_gateway_integration" "find_by_product_lambda" {
  rest_api_id             = aws_api_gateway_rest_api.swippy_api.id
  resource_id             = aws_api_gateway_resource.root.id
  http_method             = aws_api_gateway_method.proxy.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.find_by_product.invoke_arn
}

resource "aws_api_gateway_integration" "find_in_ebay_stores_lambda" {
  rest_api_id             = aws_api_gateway_rest_api.swippy_api.id
  resource_id             = aws_api_gateway_resource.root.id
  http_method             = aws_api_gateway_method.proxy.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.find_in_ebay_stores.invoke_arn
}

resource "aws_lambda_permission" "allow_execution" {
  count = length([
    aws_lambda_function.find_by_category.function_name,
    aws_lambda_function.find_by_keyword.function_name,
    aws_lambda_function.find_advanced.function_name,
    aws_lambda_function.find_by_product.function_name,
    aws_lambda_function.find_in_ebay_stores.function_name,
  ])
  statement_id = "AllowExecutionFromAPIGateway"
  action       = "lambda:InvokeFunction"
  function_name = [
    aws_lambda_function.find_by_category.function_name,
    aws_lambda_function.find_by_keyword.function_name,
    aws_lambda_function.find_advanced.function_name,
    aws_lambda_function.find_by_product.function_name,
    aws_lambda_function.find_in_ebay_stores.function_name,
  ][count.index]
  principal = "apigateway.amazonaws.com"
}

resource "aws_api_gateway_deployment" "deploy" {
  depends_on  = [aws_lambda_permission.allow_execution]
  rest_api_id = aws_api_gateway_rest_api.swippy_api.id
}