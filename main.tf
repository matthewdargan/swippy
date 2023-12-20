variable "aws_region" {
  default = "us-east-1"
}

provider "aws" {
  region = var.aws_region
}

resource "aws_s3_bucket" "tfstate" {
  bucket = "swippy-api-terraform-state"
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_public_access_block" "block_public_access" {
  bucket                  = aws_s3_bucket.tfstate.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "versioning_enabled" {
  bucket = aws_s3_bucket.tfstate.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_dynamodb_table" "tflock" {
  name         = "terraform-lock"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"
  attribute {
    name = "LockID"
    type = "S"
  }
}

terraform {
  backend "s3" {
    bucket         = "swippy-api-terraform-state"
    key            = "terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-lock"
    encrypt        = true
  }
}

variable "ebay_app_id" {}

resource "aws_ssm_parameter" "ebay_app_id" {
  name   = "ebay-app-id"
  type   = "SecureString"
  value  = var.ebay_app_id
  key_id = "alias/aws/ssm"
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

variable "aws_account_id" {}

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

resource "aws_lambda_function" "find_advanced" {
  function_name    = "find-advanced"
  handler          = "bin/find-advanced/bootstrap"
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  filename         = "bin/find-advanced.zip"
  source_code_hash = filebase64("bin/find-advanced.zip")
  role             = aws_iam_role.ebay_find_role.arn
  tags             = { Project = "swippy-api" }
}

resource "aws_cloudwatch_log_group" "find_advanced_logs" {
  name = "/aws/lambda/${aws_lambda_function.find_advanced.function_name}"
}

resource "aws_lambda_function" "find_by_category" {
  function_name    = "find-by-category"
  handler          = "bin/find-by-category/bootstrap"
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  filename         = "bin/find-by-category.zip"
  source_code_hash = filebase64("bin/find-by-category.zip")
  role             = aws_iam_role.ebay_find_role.arn
  tags             = { Project = "swippy-api" }
}

resource "aws_cloudwatch_log_group" "find_by_category_logs" {
  name = "/aws/lambda/${aws_lambda_function.find_by_category.function_name}"
}

resource "aws_lambda_function" "find_by_keywords" {
  function_name    = "find-by-keywords"
  handler          = "bin/find-by-keywords/bootstrap"
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  filename         = "bin/find-by-keywords.zip"
  source_code_hash = filebase64("bin/find-by-keywords.zip")
  role             = aws_iam_role.ebay_find_role.arn
  tags             = { Project = "swippy-api" }
}

resource "aws_cloudwatch_log_group" "find_by_keywords_logs" {
  name = "/aws/lambda/${aws_lambda_function.find_by_keywords.function_name}"
}

resource "aws_lambda_function" "find_by_product" {
  function_name    = "find-by-product"
  handler          = "bin/find-by-product/bootstrap"
  runtime          = "provided.al2"
  architectures    = ["arm64"]
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
  architectures    = ["arm64"]
  filename         = "bin/find-in-ebay-stores.zip"
  source_code_hash = filebase64("bin/find-in-ebay-stores.zip")
  role             = aws_iam_role.ebay_find_role.arn
  tags             = { Project = "swippy-api" }
}

resource "aws_cloudwatch_log_group" "find_in_ebay_stores_logs" {
  name = "/aws/lambda/${aws_lambda_function.find_in_ebay_stores.function_name}"
}

resource "aws_apigatewayv2_api" "swippy_api_gw" {
  name          = "swippy-api-gw"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "dev" {
  api_id      = aws_apigatewayv2_api.swippy_api_gw.id
  name        = "dev"
  auto_deploy = true
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.swippy_api_gw_logs.arn
    format = jsonencode({
      requestId               = "$context.requestId"
      sourceIp                = "$context.identity.sourceIp"
      requestTime             = "$context.requestTime"
      protocol                = "$context.protocol"
      httpMethod              = "$context.httpMethod"
      resourcePath            = "$context.resourcePath"
      routeKey                = "$context.routeKey"
      status                  = "$context.status"
      responseLength          = "$context.responseLength"
      integrationErrorMessage = "$context.integrationErrorMessage"
      }
    )
  }
}

resource "aws_apigatewayv2_integration" "find_advanced_lambda" {
  api_id             = aws_apigatewayv2_api.swippy_api_gw.id
  integration_uri    = aws_lambda_function.find_advanced.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "find_advanced_lambda" {
  api_id    = aws_apigatewayv2_api.swippy_api_gw.id
  route_key = "GET /find_advanced"
  target    = "integrations/${aws_apigatewayv2_integration.find_advanced_lambda.id}"
}

resource "aws_apigatewayv2_integration" "find_by_category_lambda" {
  api_id             = aws_apigatewayv2_api.swippy_api_gw.id
  integration_uri    = aws_lambda_function.find_by_category.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "find_by_category_lambda" {
  api_id    = aws_apigatewayv2_api.swippy_api_gw.id
  route_key = "GET /find_by_category"
  target    = "integrations/${aws_apigatewayv2_integration.find_by_category_lambda.id}"
}

resource "aws_apigatewayv2_integration" "find_by_keywords_lambda" {
  api_id             = aws_apigatewayv2_api.swippy_api_gw.id
  integration_uri    = aws_lambda_function.find_by_keywords.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "find_by_keywords_lambda" {
  api_id    = aws_apigatewayv2_api.swippy_api_gw.id
  route_key = "GET /find_by_keywords"
  target    = "integrations/${aws_apigatewayv2_integration.find_by_keywords_lambda.id}"
}

resource "aws_apigatewayv2_integration" "find_by_product_lambda" {
  api_id             = aws_apigatewayv2_api.swippy_api_gw.id
  integration_uri    = aws_lambda_function.find_by_product.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "find_by_product_lambda" {
  api_id    = aws_apigatewayv2_api.swippy_api_gw.id
  route_key = "GET /find_by_product"
  target    = "integrations/${aws_apigatewayv2_integration.find_by_product_lambda.id}"
}

resource "aws_apigatewayv2_integration" "find_in_ebay_stores_lambda" {
  api_id             = aws_apigatewayv2_api.swippy_api_gw.id
  integration_uri    = aws_lambda_function.find_in_ebay_stores.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "find_in_ebay_stores_lambda" {
  api_id    = aws_apigatewayv2_api.swippy_api_gw.id
  route_key = "GET /find_in_ebay_stores"
  target    = "integrations/${aws_apigatewayv2_integration.find_in_ebay_stores_lambda.id}"
}

resource "aws_cloudwatch_log_group" "swippy_api_gw_logs" {
  name              = "/aws/api_gw/${aws_apigatewayv2_api.swippy_api_gw.name}"
  retention_in_days = 30
}

resource "aws_lambda_permission" "allow_execution" {
  count = length([
    aws_lambda_function.find_advanced.function_name,
    aws_lambda_function.find_by_category.function_name,
    aws_lambda_function.find_by_keywords.function_name,
    aws_lambda_function.find_by_product.function_name,
    aws_lambda_function.find_in_ebay_stores.function_name,
  ])
  statement_id = "AllowExecutionFromAPIGateway"
  action       = "lambda:InvokeFunction"
  function_name = [
    aws_lambda_function.find_advanced.function_name,
    aws_lambda_function.find_by_category.function_name,
    aws_lambda_function.find_by_keywords.function_name,
    aws_lambda_function.find_by_product.function_name,
    aws_lambda_function.find_in_ebay_stores.function_name,
  ][count.index]
  principal = "apigateway.amazonaws.com"
}