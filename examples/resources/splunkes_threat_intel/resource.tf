resource "splunkes_threat_intel" "phishing_domain" {
  collection  = "domain_intel"
  description = "Phishing domain targeting real estate closings"
  fields = {
    "domain"          = "secure-closing-docs.com"
    "threat_category" = "phishing"
    "weight"          = "4"
    "description"     = "Domain used in BEC attacks against title companies"
  }
}
