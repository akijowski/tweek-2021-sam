variable "enable_basic_execution" {
  type = bool
  description = "Setting this to true will add the AWSLambdaBasicExecutionRole managed IAM role"
  default = false
}

variable "enable_dynamo_access" {
  type = bool
  description = "Setting this to true will add IAM permissions to read and write from the table specified in in `var.dynamo_table_name`"
  default = false
}

variable "enable_xray_write_access" {
  type = bool
  description = "Setting this to true will add IAM permissions to write XRay traces to the XRay daemon process by adding the AWSXRayDaemonWriteAccess managed IAM role"
  default = false
}

variable "dynamo_table_name" {
  type = string
  description = "Required if var.enable_dynamo_access is true.  This is the dynamodb table needed for access"
}

variable "lambda_name" {
  type        = string
  description = "Required: the name of the Lambda Function"
}