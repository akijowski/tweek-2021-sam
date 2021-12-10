variable "dynamo_billing_mode" {
  type        = string
  description = "The billing mode for the DynamoDB table.  Must be one of PROVISIONED or PAY_PER_REQUEST"
  default     = "PAY_PER_REQUEST"
}

variable "dynamo_hash_key" {
  type        = string
  description = "The attribute name to be used as the Hash (Partition) key, must be present in the attributes"
}

variable "dynamo_range_key" {
  type        = string
  description = "The attribute name to be used as the Range (Sort) key, must be present in the attributes"
}

variable "dynamo_table_name" {
  type        = string
  description = "The name of the DynamoDB table that needs to be created"
}

variable "lambda_name" {
  type        = string
  description = "Required: the name of the Lambda Function"
}
