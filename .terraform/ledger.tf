



resource "aws_iam_user" "ledger-api-read-only" {
  name = "ledger-api-read-only"
}

resource "aws_iam_access_key" "ledger-api-read-only-access-key" {
  user = aws_iam_user.ledger-api-read-only.name
}

resource "aws_iam_user" "ledger-api-admin" {
  name = "ledger-api-admin"
}

resource "aws_iam_user_policy_attachment" "ledger-api-read-only-ssm-read-policy-attachment" {
  user       = aws_iam_user.ledger-api-read-only.name
  policy_arn = aws_iam_policy.ledger-api-ssm-read-policy.arn
}

resource "aws_iam_user_policy_attachment" "ledger-api-read-only-kms-policy-attachment" {
  user       = aws_iam_user.ledger-api-admin.name
  policy_arn = aws_iam_policy.ledger-api-kms-policy.arn
}

resource "aws_iam_access_key" "ledger-api-admin-access-key" {
  user = aws_iam_user.ledger-api-admin.name
}

resource "aws_iam_user_policy_attachment" "ledger-api-admin-ssm-write-policy-attachment" {
  user       = aws_iam_user.ledger-api-admin.name
  policy_arn = aws_iam_policy.ledger-api-ssm-write-policy.arn
}

resource "aws_iam_user_policy_attachment" "ledger-api-admin-ssm-read-policy-attachment" {
  user       = aws_iam_user.ledger-api-admin.name
  policy_arn = aws_iam_policy.ledger-api-ssm-read-policy.arn
}

resource "aws_iam_user_policy_attachment" "ledger-api-admin-kms-policy-attachment" {
  user       = aws_iam_user.ledger-api-admin.name
  policy_arn = aws_iam_policy.ledger-api-kms-policy.arn
}

resource "aws_iam_user_policy_attachment" "ledger-api-admin-ssm-describe-policy-attachment" {
  user       = aws_iam_user.ledger-api-admin.name
  policy_arn = aws_iam_policy.ledger-api-ssm-describe-policy.arn
}
