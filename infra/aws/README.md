# AWS Infrastructure

This directory contains AWS infrastructure configuration and guidelines.

---

## Conventions

### Environments

| Environment | Purpose                        | Branch          |
| ----------- | ------------------------------ | --------------- |
| `local`     | Local dev via docker-compose   | any             |
| `dev`       | Shared development environment | `main`          |
| `staging`   | Pre-production validation      | `release/*`     |
| `prod`      | Production                     | `main` (tagged) |

### Naming Convention

All AWS resources follow: `{project}-{service}-{environment}-{resource-type}`

Examples:

- `myproject-auth-service-prod-rds`
- `myproject-api-gateway-dev-lambda`
- `myproject-events-staging-sqs`

### Tagging

All resources must be tagged:

```
Project:     [project-name]
Service:     [service-name]
Environment: [local|dev|staging|prod]
ManagedBy:   [terraform|cdk|manual]
Owner:       [team-name]
```

---

## Infrastructure as Code

Prefer IaC over manual console changes:

- **AWS CDK** (TypeScript) — recommended for new projects
- **Terraform** — acceptable for existing projects or multi-cloud

IaC definitions live in `infra/aws/[service-name]/` or `infra/cdk/`.

Any manual console change must be:

1. Documented in an ADR
2. Followed up with an IaC equivalent
3. Applied to all environments

---

## Secrets Management

- **Local**: `.env` files (gitignored)
- **Dev/Staging/Prod**: AWS Secrets Manager or SSM Parameter Store
- Never store secrets in environment variables injected via CI plain text
- Rotate secrets on team member offboarding

CI/CD reads secrets via:

```yaml
- uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
    aws-region: us-east-1
```

---

## Common Services Used

| AWS Service         | Purpose             |
| ------------------- | ------------------- |
| ECS Fargate         | Container hosting   |
| RDS (Postgres)      | Relational database |
| ElastiCache (Redis) | Cache / sessions    |
| MSK (Kafka)         | Event streaming     |
| SQS                 | Simple queuing      |
| S3                  | Object storage      |
| CloudFront          | CDN                 |
| API Gateway         | HTTP API proxy      |
| Secrets Manager     | Secret storage      |
| CloudWatch          | Logs and metrics    |
| ECR                 | Container registry  |

---

## GitHub Actions AWS Auth

Use OIDC (not long-lived access keys) for CI/CD authentication:

```yaml
permissions:
  id-token: write
  contents: read

steps:
  - uses: aws-actions/configure-aws-credentials@v4
    with:
      role-to-assume: arn:aws:iam::ACCOUNT_ID:role/github-actions-role
      aws-region: us-east-1
```

Set up the OIDC trust relationship once per AWS account. See AWS docs for the IAM role configuration.
