variable "dynamo_enable_ttl" {
  type        = bool
  description = "Setting to true will enable TTL on the table"
  default     = true
}

variable "dynamo_hash_key" {
  type        = string
  description = "The attribute name to be used as the Hash (Partition) key, must be present in the attributes"
}

variable "dynamo_range_key" {
  type        = string
  description = "The attribute name to be used as the Range (Sort) key, must be present in the attributes"
  default     = null
}

variable "dynamo_table_name" {
  type        = string
  description = "The name of the DynamoDB table that needs to be created"
}

variable "dynamo_ttl_attribute" {
  type        = string
  description = "The name of the item attribute to run against the TTL expression"
  default     = "timestamp"
}