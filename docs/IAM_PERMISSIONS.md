# 🔒 IAM Permissions

This document outlines the necessary IAM permissions for ZTiAWS tools.

## For SSM Session Manager (`ssm` tool):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:StartSession",
        "ssm:TerminateSession",
        "ssm:ResumeSession",
        "ec2:DescribeInstances"
      ],
      "Resource": "*"
    }
  ]
}
```

## For AWS SSO Authentication (`authaws` tool):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sso:GetRoleCredentials",
        "sso:ListAccountRoles",
        "sso:ListAccounts"
      ],
      "Resource": "*"
    }
  ]
}
```
