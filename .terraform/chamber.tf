resource "aws_kms_key" "chamber_parameter_store_key" {
  description             = "Parameter store kms master key"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

resource "aws_kms_alias" "chamber_parameter_store_alias" {
  name          = "alias/parameter_store_key"
  target_key_id = aws_kms_key.chamber_parameter_store_key.id
}

