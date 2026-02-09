resource "splunkes_finding" "manual_finding" {
  rule_title       = "Suspicious Wire Transfer Request"
  rule_description = "Manual finding: Wire transfer request from unverified email domain"
  security_domain  = "threat"
  risk_score       = 85
  risk_object      = "finance-director@company.com"
  risk_object_type = "user"
  severity         = "critical"
}
