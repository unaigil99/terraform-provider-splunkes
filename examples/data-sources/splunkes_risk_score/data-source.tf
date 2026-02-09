data "splunkes_risk_score" "cfo_risk" {
  entity      = "cfo@company.com"
  entity_type = "user"
}

output "cfo_risk_score" {
  value = data.splunkes_risk_score.cfo_risk.risk_score
}
