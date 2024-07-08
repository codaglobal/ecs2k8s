# Example - Importing secrets from AWS Secrets manager to Kubernetes


An sample task definition is given here which can be used with this utility to migrate an sample nginx task running on ECS with secrets referenced from AWS Secrets Manager into Kubernetes.
This feature is opt-in using the `--include-secrets` flag when running the `migrate-task` or `generate-k8s-spec` ecs subcommands. The secret values are read using the AWS SDK and mounted as native kubernetes secrets.

## Requirements

1. AWS ECS Task definition with Secrets parameters set
2. Secrets created in AWS Secret Manager
3. Required IAM roles for AWS account running the utility to fetch values from Secrets Manager


Note: Replace `AWS_ACCOUNT_ID`, `SECRET_NAME`, `SECRET_JSON_KEY` with appropriate values before creating this task definition.