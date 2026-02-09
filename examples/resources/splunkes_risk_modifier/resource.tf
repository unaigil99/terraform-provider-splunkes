resource "splunkes_risk_modifier" "high_value_user" {
  entity        = "cfo@company.com"
  entity_type   = "user"
  risk_modifier = 25
  description   = "C-suite executive with wire transfer authority"
}
