resource "aws_dynamodb_table" "this" {
  name         = var.dynamo_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = var.dynamo_hash_key
  range_key    = var.dynamo_range_key
  attribute {
    name = var.dynamo_hash_key
    type = "S"
  }
  attribute {
    name = var.dynamo_range_key
    type = "S"
  }
  ttl {
    attribute_name = var.dynamo_ttl_attribute
    enabled        = var.dynamo_enable_ttl
  }
}