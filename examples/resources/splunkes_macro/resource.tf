# Simple macro
resource "splunkes_macro" "sysmon_events" {
  name        = "sysmon_events"
  definition  = "index=main sourcetype=XmlWinEventLog:Microsoft-Windows-Sysmon/Operational"
  description = "Base filter for all Sysmon events"
  app         = "SplunkEnterpriseSecuritySuite"
}

# Macro with arguments
resource "splunkes_macro" "threshold_alert" {
  name        = "threshold_alert(2)"
  definition  = "where count > $min_count$ | eval alert_level=if(count > $high_count$, \"critical\", \"warning\")"
  args        = "min_count,high_count"
  description = "Parameterized threshold alert with dynamic severity"
  app         = "SplunkEnterpriseSecuritySuite"
}
