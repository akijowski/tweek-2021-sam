resource "aws_iam_role" "this" {
  name = "${var.lambda_name}-role"

  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "basic_execution" {
  count      = var.enable_basic_execution ? 1 : 0
  role       = aws_iam_role.this.id
  policy_arn = data.aws_iam_policy.basic_execution.arn
}

resource "aws_iam_role_policy_attachment" "dynamo_access" {
  count      = var.enable_dynamo_access ? 1 : 0
  policy_arn = aws_iam_policy.dynamo_access[count.index].arn
  role       = aws_iam_role.this.id
}

resource "aws_iam_policy" "dynamo_access" {
  count       = var.enable_dynamo_access ? 1 : 0
  name        = "dynamodb-access"
  description = "Provides access to the ${var.dynamo_table_name} dynamodb table"
  policy      = data.aws_iam_policy_document.dynamo_access[count.index].json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }
    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "dynamo_access" {
  count = var.enable_dynamo_access ? 1 : 0
  statement {
    effect    = "Allow"
    actions   = [
      "dynamodb:BatchGet*",
      "dynamodb:DescribeStream",
      "dynamodb:DescribeTable",
      "dynamodb:Get*",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:BatchWrite*",
      "dynamodb:Delete*",
      "dynamodb:Update*",
      "dynamodb:PutItem"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${var.dynamo_table_name}"
    ]
  }
}

# https://docs.aws.amazon.com/lambda/latest/dg/lambda-intro-execution-role.html#permissions-executionrole-features
data "aws_iam_policy" "basic_execution" {
  name = "AWSLambdaBasicExecutionRole"
}