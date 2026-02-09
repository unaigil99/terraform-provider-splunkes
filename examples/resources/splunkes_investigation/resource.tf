resource "splunkes_investigation" "incident_tracking" {
  name        = "Phishing Campaign Investigation"
  description = "Investigation tracking a targeted phishing campaign against the finance team"
  status      = "in_progress"
  priority    = "high"
  assignee    = "soc_analyst_1"
  tags        = ["phishing", "finance", "q1-2026"]
}
