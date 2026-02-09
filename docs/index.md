---
page_title: "Provider: Splunk Enterprise Security"
description: |-
  The Splunk ES provider manages security detection rules, correlation searches, investigations, and more in Splunk Enterprise Security.
---

# Splunk Enterprise Security Provider

> **Agentric/splunkes** | Version ~> 1.0 | Protocol v6

The Splunk ES provider enables **Security-as-Code** management of Splunk Enterprise Security configurations. Manage correlation searches, detection rules, investigations, threat intelligence, and risk scoring entirely through Terraform.

## Quick Navigation

| Section | Description |
|---------|-------------|
| [Installation](#installation) | How to install and configure the provider |
| [Authentication](#authentication) | Auth methods (token, username/password, env vars) |
| [Resources](#resources-12) | 12 managed resources |
| [Data Sources](#data-sources-7) | 7 read-only data sources |
| [Full Example](#full-detection-pipeline) | End-to-end detection pipeline |

---

## Installation

### Option A: Local Development Build

```bash
# 1. Clone the provider
git clone https://github.com/Agentric/terraform-provider-splunkes.git
cd terraform-provider-splunkes

# 2. Build the binary
go build -o terraform-provider-splunkes

# 3. Install locally for Terraform
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/Agentric/splunkes/1.0.0/$(go env GOOS)_$(go env GOARCH)
cp terraform-provider-splunkes ~/.terraform.d/plugins/registry.terraform.io/Agentric/splunkes/1.0.0/$(go env GOOS)_$(go env GOARCH)/
```

### Option B: Using `dev_overrides` (Recommended for Development)

Add to your `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "Agentric/splunkes" = "/path/to/terraform-provider-splunkes"  # Directory containing the binary
  }
  direct {}
}
```

Then no `terraform init` is needed - Terraform will use the local binary directly.

### Option C: Terraform Registry (Future)

```hcl
terraform {
  required_providers {
    splunkes = {
      source  = "Agentric/splunkes"
      version = "~> 1.0"
    }
  }
}
```

---

## Provider Configuration

```hcl
provider "splunkes" {
  url                  = "https://splunk.example.com:8089"
  auth_token           = var.splunk_token
  insecure_skip_verify = true
  timeout              = 60
}
```

### Schema

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `url` | String | No | `https://localhost:8089` | Splunk management REST API URL |
| `username` | String | No | - | Admin username for basic auth |
| `password` | String (Sensitive) | No | - | Admin password for basic auth |
| `auth_token` | String (Sensitive) | No | - | Bearer token for token auth |
| `insecure_skip_verify` | Bool | No | `false` | Skip TLS certificate verification |
| `timeout` | Number | No | `60` | HTTP request timeout in seconds |

---

## Authentication

The provider supports three authentication methods (in priority order):

### 1. Bearer Token (Recommended)

```hcl
provider "splunkes" {
  url        = "https://splunk.example.com:8089"
  auth_token = var.splunk_token
}
```

### 2. Username / Password

```hcl
provider "splunkes" {
  url      = "https://splunk.example.com:8089"
  username = var.splunk_username
  password = var.splunk_password
}
```

### 3. Environment Variables

```bash
export SPLUNK_URL="https://splunk.example.com:8089"
export SPLUNK_AUTH_TOKEN="your-bearer-token"
# OR
export SPLUNK_USERNAME="admin"
export SPLUNK_PASSWORD="changeme"
export SPLUNK_INSECURE_SKIP_VERIFY="true"
export SPLUNK_TIMEOUT="60"
```

```hcl
# No explicit config needed when using env vars
provider "splunkes" {}
```

---

## Resources (12)

### Detection & Search

| Resource | Description | Docs |
|----------|-------------|------|
| [splunkes_correlation_search](resources/correlation_search.md) | ES correlation searches with MITRE ATT&CK, notable events, risk scoring | [View](resources/correlation_search.md) |
| [splunkes_saved_search](resources/saved_search.md) | Generic saved searches, alerts, and reports | [View](resources/saved_search.md) |
| [splunkes_macro](resources/macro.md) | Reusable SPL search macros | [View](resources/macro.md) |
| [splunkes_analytic_story](resources/analytic_story.md) | Group detections into analytic stories | [View](resources/analytic_story.md) |

### Enrichment Data

| Resource | Description | Docs |
|----------|-------------|------|
| [splunkes_lookup_definition](resources/lookup_definition.md) | Lookup definitions (CSV or KV store backed) | [View](resources/lookup_definition.md) |
| [splunkes_lookup_table](resources/lookup_table.md) | CSV lookup table file management | [View](resources/lookup_table.md) |
| [splunkes_kvstore_collection](resources/kvstore_collection.md) | KV store collection schemas | [View](resources/kvstore_collection.md) |
| [splunkes_threat_intel](resources/threat_intel.md) | Threat intelligence indicators | [View](resources/threat_intel.md) |

### Investigation & Response

| Resource | Description | Docs |
|----------|-------------|------|
| [splunkes_investigation](resources/investigation.md) | ES investigation lifecycle management | [View](resources/investigation.md) |
| [splunkes_investigation_note](resources/investigation_note.md) | Notes attached to investigations | [View](resources/investigation_note.md) |
| [splunkes_finding](resources/finding.md) | Manual security findings | [View](resources/finding.md) |
| [splunkes_risk_modifier](resources/risk_modifier.md) | Risk score modifiers for entities | [View](resources/risk_modifier.md) |

---

## Data Sources (7)

| Data Source | Description | Docs |
|-------------|-------------|------|
| [splunkes_correlation_search](data-sources/correlation_search.md) | Read an existing correlation search | [View](data-sources/correlation_search.md) |
| [splunkes_macro](data-sources/macro.md) | Read an existing search macro | [View](data-sources/macro.md) |
| [splunkes_investigation](data-sources/investigation.md) | Read an existing investigation | [View](data-sources/investigation.md) |
| [splunkes_finding](data-sources/finding.md) | Read an existing finding | [View](data-sources/finding.md) |
| [splunkes_risk_score](data-sources/risk_score.md) | Read risk score for an entity | [View](data-sources/risk_score.md) |
| [splunkes_identity](data-sources/identity.md) | Read ES identity information | [View](data-sources/identity.md) |
| [splunkes_asset](data-sources/asset.md) | Read ES asset information | [View](data-sources/asset.md) |

---

## Resource Dependency Graph

```
 Layer 1: Enrichment Data
 ========================
 splunkes_lookup_table ─────────────────────────┐
                                                ▼
 splunkes_kvstore_collection ──► splunkes_lookup_definition ──┐
                                                              │
 Layer 2: Search Components                                   │
 ==========================                                   │
 splunkes_macro ──────────────────────────────────────────────┤
                                                              │
 Layer 3: Threat Intelligence                                 │
 ============================                                 │
 splunkes_threat_intel ──────────────────────────────────────┤
                                                              │
 Layer 4: Detection Grouping                                  │
 ===========================                                  │
 splunkes_analytic_story ───────────────────────────────────┤
                                                              ▼
 Layer 5: Detection Rules       splunkes_correlation_search
 ========================                    │
                                             ▼
 Layer 6: Response           splunkes_investigation
 =================                    │
                                      ├──► splunkes_investigation_note
                                      │
                             splunkes_finding
                             splunkes_risk_modifier
```

Dependencies are handled **implicitly** via Terraform attribute references or **explicitly** via `depends_on` blocks.

---

## Full Detection Pipeline

See [`examples/full_detection_pipeline/main.tf`](../examples/full_detection_pipeline/main.tf) for a complete end-to-end example that demonstrates:

1. CSV lookup tables and KV store collections (enrichment data)
2. Lookup definitions (CSV and KV store backed)
3. Search macros (reusable SPL components)
4. Threat intelligence indicators
5. Analytic stories (detection grouping)
6. Correlation searches with MITRE ATT&CK mapping
7. Investigations with runbook notes
8. Risk modifiers for high-value targets
9. Data sources for reading existing ES data

### Minimal Example

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
  url        = "https://splunk.example.com:8089"
  auth_token = var.splunk_token
}

# Create a search macro
resource "splunkes_macro" "sysmon" {
  name       = "sysmon_events"
  definition = "index=main sourcetype=XmlWinEventLog:Microsoft-Windows-Sysmon/Operational"
}

# Create a correlation search that uses the macro
resource "splunkes_correlation_search" "lsass_dump" {
  name   = "LSASS Memory Access Detection"
  search = "`${splunkes_macro.sysmon.name}` EventCode=10 TargetImage=*lsass.exe"

  cron_schedule              = "*/5 * * * *"
  correlation_search_enabled = true

  notable_enabled         = true
  notable_severity        = "critical"
  notable_security_domain = "endpoint"
  notable_rule_title      = "Credential Dumping on $dest$"

  risk_enabled      = true
  risk_score        = 80
  risk_object_field = "dest"
  risk_object_type  = "system"

  mitre_attack_ids = ["T1003.001"]
}
```

---

## Supported Splunk APIs

| API | Base Path | Used By |
|-----|-----------|---------|
| Splunk REST API | `/servicesNS/{owner}/{app}/...` | Saved searches, macros, lookups, KV store, analytic stories |
| ES v2 API | `/servicesNS/nobody/missioncontrol/public/v2/...` | Investigations, findings, risks, assets, identity |
| Threat Intel API | `/services/data/threat_intel/item/...` | Threat intelligence indicators |

---

## Requirements

| Requirement | Version |
|-------------|---------|
| Terraform | >= 1.0 |
| Go (for building) | >= 1.22 |
| Splunk Enterprise Security | >= 7.x |
| Splunk Enterprise | >= 9.x |
