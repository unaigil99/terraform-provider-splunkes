---
page_title: "splunkes_risk_score Data Source - Splunk ES Provider"
description: |-
  Reads a risk score for an entity from Splunk Enterprise Security.
---

# splunkes_risk_score (Data Source)

Use this data source to read the current risk score for an entity (user or system) from Splunk Enterprise Security.

## Example Usage

```hcl
data "splunkes_risk_score" "cfo_risk" {
  entity      = "cfo@company.com"
  entity_type = "user"
}

output "cfo_risk_score" {
  value = data.splunkes_risk_score.cfo_risk.risk_score
}

output "cfo_risk_level" {
  value = data.splunkes_risk_score.cfo_risk.risk_level
}
```

## Argument Reference

* `entity` - (Required) The entity name to look up the risk score for (username, hostname, etc.).
* `entity_type` - (Optional) The entity type. Valid values: `user`, `system`. Defaults to `user`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `risk_score` - The computed risk score for the entity (integer).
* `risk_level` - The risk level classification: `low`, `medium`, `high`, or `critical`.

## API Reference

* [Get risk scores](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/risks_1/public_v2_risk_entity_risk_scores_retrieve)
* [ES API Overview](https://help.splunk.com/en/splunk-enterprise-security-8/rest-api-reference/8.0/overview/the-splunk-enterprise-security-api)
