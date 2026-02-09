---
page_title: "Installation Guide - Splunk ES Provider"
description: |-
  Step-by-step guide for installing and configuring the Splunk ES Terraform provider.
---

# Installation Guide

This guide walks you through installing the `terraform-provider-splunkes` provider and integrating it into your Terraform project.

---

## Prerequisites

| Tool | Minimum Version | Check Command |
|------|----------------|---------------|
| Terraform | 1.0+ | `terraform version` |
| Go | 1.22+ (only for building from source) | `go version` |
| Splunk Enterprise | 9.x+ | Check in Splunk Settings |
| Splunk Enterprise Security | 7.x+ | Check in ES About page |

---

## Step 1: Build the Provider

```bash
# Clone the repository
git clone https://github.com/Agentric/terraform-provider-splunkes.git
cd terraform-provider-splunkes

# Build the binary
go build -o terraform-provider-splunkes

# Verify the binary
./terraform-provider-splunkes --version
```

The build produces a single binary (~23MB) with no external dependencies.

---

## Step 2: Install the Provider

You have two options for making the provider available to Terraform:

### Option A: `dev_overrides` (Recommended for Development)

This is the simplest approach. Add to `~/.terraformrc` (Linux/macOS) or `%APPDATA%\terraform.rc` (Windows):

```hcl
provider_installation {
  dev_overrides {
    "Agentric/splunkes" = "/absolute/path/to/directory/containing/binary"
  }
  direct {}
}
```

With `dev_overrides`, you do **not** need to run `terraform init`. Terraform will use the binary directly.

**Important:** The path must point to the **directory** containing the binary, not the binary itself.

### Option B: Local Plugin Directory

Install the binary into Terraform's plugin directory:

```bash
# Determine your platform
OS=$(go env GOOS)    # linux, darwin, windows
ARCH=$(go env GOARCH) # amd64, arm64

# Create the plugin directory
PLUGIN_DIR="$HOME/.terraform.d/plugins/registry.terraform.io/Agentric/splunkes/1.0.0/${OS}_${ARCH}"
mkdir -p "$PLUGIN_DIR"

# Copy the binary
cp terraform-provider-splunkes "$PLUGIN_DIR/"
```

With this approach, you need to run `terraform init` in your project directory.

---

## Step 3: Configure Your Terraform Project

Create a new Terraform project or add the provider to an existing one:

```bash
mkdir my-splunk-security
cd my-splunk-security
```

Create `main.tf`:

```hcl
terraform {
  required_providers {
    splunkes = {
      source  = "Agentric/splunkes"
      version = "~> 1.0"
    }
  }
}

provider "splunkes" {
  url                  = "https://your-splunk-server:8089"
  auth_token           = var.splunk_token
  insecure_skip_verify = true  # Set to false in production with valid TLS
}

variable "splunk_token" {
  description = "Splunk bearer auth token"
  type        = string
  sensitive   = true
}
```

---

## Step 4: Set Up Authentication

### Option 1: Terraform Variables (Recommended)

Create a `terraform.tfvars` file (add to `.gitignore`!):

```hcl
splunk_token = "your-bearer-token-here"
```

### Option 2: Environment Variables

```bash
export SPLUNK_URL="https://your-splunk-server:8089"
export SPLUNK_AUTH_TOKEN="your-bearer-token-here"
export SPLUNK_INSECURE_SKIP_VERIFY="true"
```

With environment variables, the provider block can be empty:

```hcl
provider "splunkes" {}
```

### Option 3: Username/Password

```hcl
provider "splunkes" {
  url      = "https://your-splunk-server:8089"
  username = var.splunk_username
  password = var.splunk_password
}
```

### Generating a Splunk Auth Token

In Splunk:
1. Go to **Settings > Tokens**
2. Click **New Token**
3. Set the audience and expiration
4. Copy the generated token

Or via CLI:
```bash
curl -k -u admin:changeme \
  https://splunk:8089/services/authorization/tokens \
  -d name=admin \
  -d audience=terraform
```

---

## Step 5: Verify the Setup

Create a simple test resource:

```hcl
# test.tf
resource "splunkes_macro" "test" {
  name       = "terraform_test_macro"
  definition = "index=_internal | head 1"
  app        = "search"
}
```

Run Terraform:

```bash
# Initialize (skip if using dev_overrides)
terraform init

# Preview changes
terraform plan

# Apply
terraform apply

# Verify in Splunk
# Go to Settings > Advanced Search > Search Macros

# Clean up
terraform destroy
```

---

## Step 6: Production Setup

For production use, consider:

### State Backend

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "splunk-es/terraform.tfstate"
    region = "us-east-1"
  }
}
```

### Variable Management

```hcl
# variables.tf
variable "splunk_token" {
  type      = string
  sensitive = true
}

variable "environment" {
  type    = string
  default = "dev"
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}
```

### CI/CD Integration

```yaml
# .github/workflows/terraform.yml
name: Terraform
on:
  push:
    branches: [main]
  pull_request:

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build Provider
        run: |
          cd terraform-provider-splunkes
          go build -o terraform-provider-splunkes

      - uses: hashicorp/setup-terraform@v3

      - name: Terraform Init
        run: terraform init

      - name: Terraform Plan
        run: terraform plan -no-color
        env:
          SPLUNK_URL: ${{ secrets.SPLUNK_URL }}
          SPLUNK_AUTH_TOKEN: ${{ secrets.SPLUNK_AUTH_TOKEN }}
```

---

## Troubleshooting

### "provider not found" Error

Ensure the binary path in `dev_overrides` points to the directory, not the file:

```hcl
# Correct
"Agentric/splunkes" = "/home/user/terraform-provider-splunkes"

# Wrong
"Agentric/splunkes" = "/home/user/terraform-provider-splunkes/terraform-provider-splunkes"
```

### "TLS handshake error"

Set `insecure_skip_verify = true` or install the Splunk CA certificate:

```hcl
provider "splunkes" {
  url                  = "https://splunk:8089"
  auth_token           = var.token
  insecure_skip_verify = true
}
```

### "401 Unauthorized"

Verify your token is valid:

```bash
curl -k -H "Authorization: Bearer YOUR_TOKEN" \
  https://splunk:8089/services/authentication/current-context
```

### "Permission denied" During Build

Use custom cache directories:

```bash
GOMODCACHE=/tmp/gomodcache GOCACHE=/tmp/gobuildcache go build -o terraform-provider-splunkes
```
