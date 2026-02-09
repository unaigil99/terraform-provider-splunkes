# =============================================================================
# Full Detection Pipeline Example
# =============================================================================
# This example demonstrates the complete dependency chain for a Splunk
# Enterprise Security detection pipeline managed as code.
#
# Dependency graph:
#
#   splunkes_lookup_table
#          │
#          ▼
#   splunkes_lookup_definition ──┐
#                                │
#   splunkes_kvstore_collection ─┤
#          │                     │
#          ▼                     │
#   splunkes_lookup_definition   │
#     (kvstore-backed)           │
#                                │
#   splunkes_macro ──────────────┤
#                                │
#   splunkes_analytic_story ─────┤
#                                │
#   splunkes_threat_intel ───────┤
#                                │
#                                ▼
#               splunkes_correlation_search
#                                │
#                                ▼
#                     splunkes_investigation
#                                │
#                                ▼
#                   splunkes_investigation_note
#
# =============================================================================

terraform {
  required_providers {
    splunkes = {
      source  = "Agentric/splunkes"
      version = "~> 1.0"
    }
  }
}

provider "splunkes" {
  url                  = var.splunk_url
  auth_token           = var.splunk_auth_token
  insecure_skip_verify = var.insecure_skip_verify
}

# -----------------------------------------------------------------------------
# Variables
# -----------------------------------------------------------------------------

variable "splunk_url" {
  description = "Splunk management endpoint"
  type        = string
  default     = "https://localhost:8089"
}

variable "splunk_auth_token" {
  description = "Splunk bearer auth token"
  type        = string
  sensitive   = true
}

variable "insecure_skip_verify" {
  description = "Skip TLS verification"
  type        = bool
  default     = true
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "dev"
}

locals {
  app   = "SplunkEnterpriseSecuritySuite"
  owner = "nobody"
  tags  = ["managed-by-terraform", var.environment]
}

# =============================================================================
# Layer 1: Enrichment Data (Lookups + KV Store)
# =============================================================================

# CSV lookup table with known-bad IP addresses
resource "splunkes_lookup_table" "threat_ips" {
  name  = "threat_ips_${var.environment}.csv"
  app   = local.app
  owner = local.owner
  content = <<-CSV
    ip,threat_category,confidence,description
    10.0.0.100,c2,high,Known C2 server
    192.168.1.200,scanner,medium,Port scanner
    172.16.0.50,exfil,high,Data exfiltration endpoint
  CSV
}

# Lookup definition pointing to the CSV file
resource "splunkes_lookup_definition" "threat_ips" {
  name                 = "threat_ips_lookup"
  filename             = splunkes_lookup_table.threat_ips.name
  app                  = local.app
  owner                = local.owner
  case_sensitive_match = false
  fields_list          = "ip,threat_category,confidence,description"
}

# KV store collection for tracking investigation metadata
resource "splunkes_kvstore_collection" "investigation_tracking" {
  name  = "investigation_tracking"
  app   = local.app
  owner = local.owner
  fields = {
    "case_id"     = "string"
    "analyst"     = "string"
    "status"      = "string"
    "risk_score"  = "number"
    "created_at"  = "time"
    "is_resolved" = "bool"
  }
  accelerated_fields = {
    "idx_case" = "{\"case_id\": 1}"
  }
}

# Lookup definition backed by KV store
resource "splunkes_lookup_definition" "investigation_tracking" {
  name          = "investigation_tracking_lookup"
  external_type = "kvstore"
  collection    = splunkes_kvstore_collection.investigation_tracking.name
  app           = local.app
  owner         = local.owner
  fields_list   = "case_id,analyst,status,risk_score,created_at,is_resolved"
}

# =============================================================================
# Layer 2: Reusable Search Components (Macros)
# =============================================================================

# Macro for common sysmon filter
resource "splunkes_macro" "sysmon_filter" {
  name       = "sysmon_process_creation"
  definition = "index=main sourcetype=XmlWinEventLog:Microsoft-Windows-Sysmon/Operational EventCode=1"
  app        = local.app
  owner      = local.owner
  description = "Sysmon process creation events (EventCode 1)"
}

# Macro for security event log filter
resource "splunkes_macro" "security_log_filter" {
  name       = "windows_security_events"
  definition = "index=main sourcetype=WinEventLog:Security"
  app        = local.app
  owner      = local.owner
  description = "Windows Security Event Log events"
}

# Macro with arguments for time-based thresholds
resource "splunkes_macro" "threshold_check" {
  name       = "threshold_check(2)"
  definition = "where count > $threshold$ | eval severity=if(count > $critical$, \"critical\", \"high\")"
  args       = "threshold,critical"
  app        = local.app
  owner      = local.owner
  description = "Threshold check with dynamic severity assignment"
}

# =============================================================================
# Layer 3: Threat Intelligence
# =============================================================================

resource "splunkes_threat_intel" "malicious_domain" {
  collection  = "domain_intel"
  description = "Known malicious domain for wire fraud campaign"
  fields = {
    "domain"          = "evil-title-company.com"
    "threat_category" = "phishing"
    "weight"          = "3"
    "description"     = "Wire fraud phishing domain targeting real estate"
  }
}

resource "splunkes_threat_intel" "c2_ip" {
  collection  = "ip_intel"
  description = "C2 server for real estate BEC campaign"
  fields = {
    "ip"              = "198.51.100.42"
    "threat_category" = "c2"
    "weight"          = "5"
    "description"     = "BEC campaign C2 infrastructure"
  }
}

# =============================================================================
# Layer 4: Analytic Story (groups detections)
# =============================================================================

resource "splunkes_analytic_story" "wire_fraud_detection" {
  name        = "Real Estate Wire Fraud Detection"
  app         = local.app
  owner       = local.owner
  description = "Detects wire fraud attempts targeting real estate transactions including BEC, domain spoofing, and unauthorized wire transfer modifications."

  category               = ["Adversary Tactics", "Financial Fraud"]
  data_models             = ["Email", "Network_Traffic", "Web"]
  providing_technologies  = ["Microsoft 365", "Palo Alto Networks", "Proofpoint"]

  detection_searches = [
    splunkes_correlation_search.bec_wire_fraud.name,
    splunkes_correlation_search.credential_dumping.name,
  ]
}

# =============================================================================
# Layer 5: Correlation Searches (Detection Rules)
# =============================================================================

# Detection: Business Email Compromise - Wire Fraud
resource "splunkes_correlation_search" "bec_wire_fraud" {
  name        = "BEC Wire Fraud - Suspicious Wire Transfer Instructions"
  description = "Detects email patterns indicative of business email compromise targeting wire transfer instructions in real estate transactions."
  app         = local.app
  owner       = local.owner

  search = <<-SPL
    index=email sourcetype=ms365:messageTrace
    (subject="*wire*" OR subject="*transfer*" OR subject="*closing*" OR subject="*escrow*")
    | eval sender_domain=mvindex(split(sender, "@"), 1)
    | lookup ${splunkes_lookup_definition.threat_ips.name} domain AS sender_domain OUTPUT threat_category
    | where isnotnull(threat_category) OR match(sender_domain, ".*typo.*|.*-.*\\.com")
    | stats count min(_time) as firstTime max(_time) as lastTime
      values(subject) as subjects values(recipient) as recipients
      by sender sender_domain threat_category
    | where count > 0
    | `${splunkes_macro.threshold_check.name}`
  SPL

  cron_schedule          = "*/5 * * * *"
  dispatch_earliest_time = "-10m@m"
  dispatch_latest_time   = "now"

  # ES Correlation Search
  correlation_search_enabled = true
  correlation_search_label   = "BEC Wire Fraud Detection"

  # Notable Event
  notable_enabled          = true
  notable_rule_title       = "BEC Wire Fraud: Suspicious wire transfer email from $sender$"
  notable_rule_description = "A suspicious email related to wire transfers was detected from $sender_domain$. Threat category: $threat_category$"
  notable_security_domain  = "threat"
  notable_severity         = "critical"
  notable_drilldown_name   = "View sender email history"
  notable_drilldown_search = "index=email sender=$sender$ | table _time subject recipient sender"
  notable_recommended_actions = "email,phantom"

  # Risk Scoring
  risk_enabled      = true
  risk_score        = 80
  risk_message      = "BEC wire fraud attempt detected from $sender_domain$ targeting $recipients$"
  risk_object_field = "sender"
  risk_object_type  = "user"

  # MITRE ATT&CK Mapping
  mitre_attack_ids = ["T1566.001", "T1534"]
  kill_chain_phases = ["Delivery", "Actions on Objectives"]
  analytic_story   = ["Real Estate Wire Fraud Detection"]

  alert_severity = 5  # Critical
}

# Detection: Credential Dumping via LSASS
resource "splunkes_correlation_search" "credential_dumping" {
  name        = "Credential Dumping - LSASS Memory Access"
  description = "Detects attempts to access LSASS process memory, indicative of credential dumping attacks."
  app         = local.app
  owner       = local.owner

  search = <<-SPL
    `${splunkes_macro.sysmon_filter.name}`
    EventCode=10 TargetImage=*lsass.exe
    (CallTrace=*dbgcore.dll* OR CallTrace=*dbghelp.dll*)
    | stats count min(_time) as firstTime max(_time) as lastTime
      by dest SourceImage TargetImage CallTrace
  SPL

  cron_schedule          = "*/5 * * * *"
  dispatch_earliest_time = "-15m@m"
  dispatch_latest_time   = "now"

  correlation_search_enabled = true
  correlation_search_label   = "LSASS Credential Dumping"

  notable_enabled          = true
  notable_rule_title       = "Credential Dumping: LSASS access on $dest$"
  notable_rule_description = "Process $SourceImage$ accessed LSASS memory on $dest$"
  notable_security_domain  = "endpoint"
  notable_severity         = "high"

  risk_enabled      = true
  risk_score        = 70
  risk_message      = "LSASS credential dumping detected on $dest$ by $SourceImage$"
  risk_object_field = "dest"
  risk_object_type  = "system"

  mitre_attack_ids  = ["T1003.001"]
  kill_chain_phases = ["Exploitation"]
  analytic_story    = ["Credential Dumping"]
  cis20             = ["CIS 8"]
  nist              = ["DE.CM"]

  alert_severity = 4  # Error/High

  # This detection depends on macros being created first
  depends_on = [
    splunkes_macro.sysmon_filter,
  ]
}

# =============================================================================
# Layer 6: Investigations & Notes
# =============================================================================

# Standing investigation for wire fraud monitoring
resource "splunkes_investigation" "wire_fraud_standing" {
  name        = "Wire Fraud Monitoring - ${var.environment}"
  description = "Standing investigation for monitoring and tracking wire fraud attempts across real estate transactions."
  status      = "in_progress"
  priority    = "high"
  tags        = local.tags
}

# Runbook note attached to the investigation
resource "splunkes_investigation_note" "wire_fraud_runbook" {
  investigation_id = splunkes_investigation.wire_fraud_standing.id

  content = <<-NOTE
    ## Wire Fraud Investigation Runbook

    ### Triage Steps
    1. Verify sender domain against known-good closing agent domains
    2. Check for domain typosquatting (Levenshtein distance analysis)
    3. Confirm wire instructions match previous communications
    4. Contact closing agent via phone (number on file, NOT from email)

    ### Escalation
    - If wire sent: IMMEDIATELY contact bank for wire recall
    - If credentials compromised: Reset all affected accounts
    - If ongoing campaign: Engage threat intelligence team

    ### Evidence Collection
    - Preserve email headers (X-Originating-IP, SPF/DKIM/DMARC results)
    - Screenshot any spoofed domains or login pages
    - Capture network logs for C2 communication
  NOTE
}

# =============================================================================
# Layer 7: Risk Modifiers (for high-value targets)
# =============================================================================

resource "splunkes_risk_modifier" "closing_coordinator" {
  entity        = "closing-coordinator@company.com"
  entity_type   = "user"
  risk_modifier = 20
  description   = "Elevated risk: handles wire transfers for real estate closings"
}

resource "splunkes_risk_modifier" "transaction_server" {
  entity        = "txn-server-01"
  entity_type   = "system"
  risk_modifier = 15
  description   = "Critical asset: transaction management server"
}

# =============================================================================
# Data Sources (read existing ES data)
# =============================================================================

# Read current risk score for a monitored entity
data "splunkes_risk_score" "closing_coordinator_risk" {
  entity      = "closing-coordinator@company.com"
  entity_type = "user"

  depends_on = [splunkes_risk_modifier.closing_coordinator]
}

# Read an existing macro
data "splunkes_macro" "existing_cim" {
  name = "cim_Endpoint_Processes_indexes"
  app  = local.app
}

# =============================================================================
# Outputs
# =============================================================================

output "bec_detection_name" {
  description = "Name of the BEC wire fraud correlation search"
  value       = splunkes_correlation_search.bec_wire_fraud.name
}

output "credential_dumping_detection_name" {
  description = "Name of the credential dumping correlation search"
  value       = splunkes_correlation_search.credential_dumping.name
}

output "investigation_id" {
  description = "ID of the standing wire fraud investigation"
  value       = splunkes_investigation.wire_fraud_standing.id
}

output "closing_coordinator_risk" {
  description = "Current risk score for closing coordinator"
  value       = data.splunkes_risk_score.closing_coordinator_risk.risk_score
}
