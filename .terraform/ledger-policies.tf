resource "aws_iam_policy" "ledger-api-ssm-write-policy" {
  name   = "ledger-api-ssm-write-policy"
  policy = data.aws_iam_policy_document.ledger-api-ssm-write.json
}

data "aws_iam_policy_document" "ledger-api-ssm-write" {
  statement {
    actions = [
      "ssm:PutParameter",
      "ssm:DeleteParameter",
      "ssm:DeleteParameters"
    ]
    resources = [
      "arn:aws:ssm:us-east-1:847870459364:parameter/ledger-api/*"
    ]
  }
}

resource "aws_iam_policy" "ledger-api-ssm-read-policy" {
  name   = "ledger-api-ssm-read-policy"
  policy = data.aws_iam_policy_document.ledger-api-ssm-read.json
}

data "aws_iam_policy_document" "ledger-api-ssm-read" {
  statement {
    actions = [
      "ssm:GetParameterHistory",
      "ssm:GetParametersByPath",
      "ssm:GetParameters",
      "ssm:GetParameter",
    ]
    resources = [
      "arn:aws:ssm:us-east-1:847870459364:parameter/ledger-api/*"
    ]
  }
}

resource "aws_iam_policy" "ledger-api-kms-policy" {
  name   = "ledger-api-kms-policy"
  policy = data.aws_iam_policy_document.ledger-api-kms.json
}

data "aws_iam_policy_document" "ledger-api-kms" {
  statement {
    actions = [
      "kms:Encrypt",
      "kms:Decrypt"
    ]
    resources = [
      aws_kms_key.chamber_parameter_store_key.arn
    ]
  }
}

resource "aws_iam_policy" "ledger-api-ssm-describe-policy" {
  name   = "ledger-api-ssm-describe-policy"
  policy = data.aws_iam_policy_document.ledger-api-ssm-describe.json
}

data "aws_iam_policy_document" "ledger-api-ssm-describe" {
  statement {
    actions = [
      "ssm:DescribeParameters",
    ]
    resources = [
      "*"
    ]
  }
}
