output "dynamodb_table_arn" {
  value = module.dynamodb.dynamodb_table_arn
}

output "lambda_iam_role_arn" {
  value = module.iam_role.lambda_execution_role_arn
}