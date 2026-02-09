# Correlation search with MITRE ATT&CK mapping and notable event creation
resource "splunkes_correlation_search" "brute_force" {
  name        = "Brute Force Login Attempts"
  description = "Detects multiple failed login attempts from a single source"
  app         = "SplunkEnterpriseSecuritySuite"

  search = <<-SPL
    index=security sourcetype=WinEventLog:Security EventCode=4625
    | stats count min(_time) as firstTime max(_time) as lastTime by src_ip, Account_Name, dest
    | where count > 10
  SPL

  cron_schedule          = "*/10 * * * *"
  dispatch_earliest_time = "-20m@m"
  dispatch_latest_time   = "now"

  correlation_search_enabled = true
  correlation_search_label   = "Brute Force Detection"

  notable_enabled          = true
  notable_rule_title       = "Brute Force: $count$ failed logins from $src_ip$"
  notable_rule_description = "Multiple failed login attempts detected from $src_ip$ targeting $Account_Name$ on $dest$"
  notable_security_domain  = "access"
  notable_severity         = "high"

  risk_enabled      = true
  risk_score        = 60
  risk_message      = "Brute force activity from $src_ip$ targeting $Account_Name$"
  risk_object_field = "src_ip"
  risk_object_type  = "system"

  mitre_attack_ids  = ["T1110.001", "T1110.003"]
  kill_chain_phases = ["Exploitation"]
  analytic_story    = ["Credential Access"]

  alert_severity = 4
}
