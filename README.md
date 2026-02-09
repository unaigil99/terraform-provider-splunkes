# Terraform Provider for Splunk Enterprise Security

A Terraform provider for managing Splunk Enterprise Security configurations as code. Built with the modern `terraform-plugin-framework` (protocol v6).

## Features

### Resources (12)

| Resource | API | Description |
|----------|-----|-------------|
| `splunkes_correlation_search` | Saved Searches + ES params | Full correlation search with MITRE ATT&CK, notable events, risk scoring |
| `splunkes_saved_search` | Saved Searches | Generic saved searches, alerts, and reports |
| `splunkes_macro` | configs/conf-macros | Reusable SPL search macros |
| `splunkes_lookup_definition` | data/transforms/lookups | Lookup definitions (CSV and KV store backed) |
| `splunkes_lookup_table` | data/lookup-table-files | CSV lookup table file management |
| `splunkes_kvstore_collection` | storage/collections/config | KV store collection schemas |
| `splunkes_investigation` | ES v2 Investigations | ES investigations lifecycle |
| `splunkes_investigation_note` | ES v2 Investigation Notes | Notes on investigations |
| `splunkes_finding` | ES v2 Findings | Manual security findings |
| `splunkes_risk_modifier` | ES v2 Risks | Risk score modifiers for entities |
| `splunkes_threat_intel` | data/threat_intel/item | Threat intelligence indicators |
| `splunkes_analytic_story` | configs/conf-analyticstories | Detection grouping / analytic stories |

### Data Sources (7)

| Data Source | Description |
|-------------|-------------|
| `splunkes_correlation_search` | Read existing correlation search |
| `splunkes_investigation` | Read existing investigation |
| `splunkes_finding` | Read existing finding |
| `splunkes_risk_score` | Read risk score for an entity |
| `splunkes_identity` | Read ES identity (read-only API) |
| `splunkes_asset` | Read ES asset (read-only API) |
| `splunkes_macro` | Read existing search macro |

## Resource Dependency Graph

```
splunkes_lookup_table ─────────────────────┐
                                           ▼
splunkes_kvstore_collection ──► splunkes_lookup_definition ──┐
                                                             │
splunkes_macro ──────────────────────────────────────────────┤
                                                             │
splunkes_threat_intel ──────────────────────────────────────┤
                                                             │
splunkes_analytic_story ────────────────────────────────────┤
                                                             ▼
                                        splunkes_correlation_search
                                                             │
                                                             ▼
                                        splunkes_finding ─► splunkes_investigation
                                                             │
                                        splunkes_risk_modifier                    ▼
                                                        splunkes_investigation_note
```

Dependencies are handled **implicitly** via Terraform references (e.g., referencing
`splunkes_macro.my_macro.name` in a correlation search SPL query) or **explicitly**
via `depends_on` blocks.

## Quick Start

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
  url                  = "https://splunk.example.com:8089"
  auth_token           = var.splunk_token
  insecure_skip_verify = true
}

# Create a macro for reuse
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
  notable_enabled            = true
  notable_severity           = "critical"
  notable_security_domain    = "endpoint"
  notable_rule_title         = "Credential Dumping on $dest$"

  risk_enabled      = true
  risk_score        = 80
  risk_object_field = "dest"
  risk_object_type  = "system"

  mitre_attack_ids = ["T1003.001"]
}
```

## Authentication

Three authentication methods (in priority order):

1. **Bearer Token** (recommended for automation):
   ```hcl
   provider "splunkes" {
     url        = "https://splunk:8089"
     auth_token = var.token
   }
   ```

2. **Username/Password** (session-based):
   ```hcl
   provider "splunkes" {
     url      = "https://splunk:8089"
     username = var.username
     password = var.password
   }
   ```

3. **Environment Variables**:
   ```bash
   export SPLUNK_URL="https://splunk:8089"
   export SPLUNK_AUTH_TOKEN="your-token"
   # or
   export SPLUNK_USERNAME="admin"
   export SPLUNK_PASSWORD="changeme"
   ```

## Building

```bash
go install
```

## Testing

```bash
# Unit tests
make test

# Acceptance tests (requires running Splunk instance)
SPLUNK_URL=https://localhost:8089 \
SPLUNK_USERNAME=admin \
SPLUNK_PASSWORD=changeme \
make testacc
```

## License

MPL-2.0
