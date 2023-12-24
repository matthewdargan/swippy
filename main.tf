variable "aws_region" {
  default  = "us-east-1"
  type     = string
  nullable = false
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

variable "ebay_app_id" {
  type      = string
  sensitive = true
  nullable  = false
}

resource "aws_ssm_parameter" "ebay_app_id" {
  name   = "ebay-app-id"
  type   = "SecureString"
  value  = var.ebay_app_id
  key_id = "alias/aws/ssm"
}

variable "swippy_db_url" {
  type      = string
  sensitive = true
  nullable  = false
}

resource "aws_ssm_parameter" "swippy_db_url" {
  name   = "swippy-db-url"
  type   = "SecureString"
  value  = var.swippy_db_url
  key_id = "alias/aws/ssm"
}

resource "aws_sqs_queue" "swippy_api_queue" {
  name = "swippy-api-queue"
  tags = { Project = "swippy-api" }
}

resource "aws_iam_role" "db_insert_role" {
  name = "swippy-api-lambda-role-db-insert"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Effect    = "Allow",
      Principal = { Service = "lambda.amazonaws.com" },
      Action    = "sts:AssumeRole",
    }]
  })
}

resource "aws_iam_role_policy" "db_insert_policy" {
  name = "swippy-api-lambda-policy-db-insert"
  role = aws_iam_role.db_insert_role.name
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ],
        Resource = "${aws_cloudwatch_log_group.db_insert_lambda_log_group.arn}:*",
      },
      {
        Effect   = "Allow",
        Action   = "ssm:GetParameter",
        Resource = aws_ssm_parameter.swippy_db_url.arn,
      },
      {
        Effect = "Allow",
        Action = [
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ReceiveMessage",
        ],
        Resource = aws_sqs_queue.swippy_api_queue.arn,
      },
    ],
  })
}

resource "aws_lambda_function" "db_insert_lambda_function" {
  function_name    = "db-insert"
  handler          = "bin/db-insert/bootstrap"
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  filename         = "bin/db-insert.zip"
  source_code_hash = filebase64("bin/db-insert.zip")
  role             = aws_iam_role.db_insert_role.arn
  tags             = { Project = "swippy-api" }
}

resource "aws_cloudwatch_log_group" "db_insert_lambda_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.db_insert_lambda_function.function_name}"
  retention_in_days = 30
}

resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = aws_sqs_queue.swippy_api_queue.arn
  function_name    = aws_lambda_function.db_insert_lambda_function.arn
}

variable "lambda_functions" {
  type = list(string)
  default = [
    "find-advanced",
    "find-by-category",
    "find-by-keywords",
    "find-by-product",
    "find-in-ebay-stores",
  ]
}

resource "aws_iam_role" "lambda_roles" {
  count = length(var.lambda_functions)
  name  = "swippy-api-lambda-role-${var.lambda_functions[count.index]}"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Effect    = "Allow",
      Principal = { Service = "lambda.amazonaws.com" },
      Action    = "sts:AssumeRole",
    }]
  })
}

resource "aws_lambda_function" "lambda_functions" {
  count            = length(var.lambda_functions)
  function_name    = var.lambda_functions[count.index]
  handler          = "bin/${var.lambda_functions[count.index]}/bootstrap"
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  filename         = "bin/${var.lambda_functions[count.index]}.zip"
  source_code_hash = filebase64("bin/${var.lambda_functions[count.index]}.zip")
  role             = aws_iam_role.lambda_roles[count.index].arn
  tags             = { Project = "swippy-api" }
}

resource "aws_cloudwatch_log_group" "lambda_log_groups" {
  count             = length(var.lambda_functions)
  name              = "/aws/lambda/${aws_lambda_function.lambda_functions[count.index].function_name}"
  retention_in_days = 30
}

resource "aws_iam_role_policy" "lambda_policies" {
  count = length(var.lambda_functions)
  name  = "swippy-api-lambda-policy-${var.lambda_functions[count.index]}"
  role  = aws_iam_role.lambda_roles[count.index].name
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ],
        Resource = "${aws_cloudwatch_log_group.lambda_log_groups[count.index].arn}:*",
      },
      {
        Effect   = "Allow",
        Action   = "ssm:GetParameter",
        Resource = aws_ssm_parameter.ebay_app_id.arn,
      },
      {
        Effect = "Allow",
        Action = [
          "sqs:GetQueueUrl",
          "sqs:SendMessage",
        ]
        Resource = aws_sqs_queue.swippy_api_queue.arn,
      },
    ],
  })
}

resource "aws_apigatewayv2_api" "swippy_api_gw" {
  name          = "swippy-api-gw"
  protocol_type = "HTTP"
}

resource "aws_cloudwatch_log_group" "swippy_api_gw_logs" {
  name              = "/aws/api_gw/${aws_apigatewayv2_api.swippy_api_gw.name}"
  retention_in_days = 30
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

resource "aws_apigatewayv2_integration" "lambda_integrations" {
  count              = length(var.lambda_functions)
  api_id             = aws_apigatewayv2_api.swippy_api_gw.id
  integration_uri    = aws_lambda_function.lambda_functions[count.index].invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "lambda_routes" {
  count     = length(var.lambda_functions)
  api_id    = aws_apigatewayv2_api.swippy_api_gw.id
  route_key = "GET /${var.lambda_functions[count.index]}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integrations[count.index].id}"
}

resource "aws_lambda_permission" "allow_execution" {
  count         = length(var.lambda_functions)
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda_functions[count.index].function_name
  principal     = "apigateway.amazonaws.com"
}