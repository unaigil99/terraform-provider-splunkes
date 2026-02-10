---
page_title: "splunkes_finding Data Source - Splunk ES Provider"
description: |-
  Reads an existing finding from Splunk Enterprise Security.
---

# splunkes_finding (Data Source)

Use this data source to read an existing security finding from Splunk Enterprise Security.

## Example Usage

```hcl
data "splunkes_finding" "recent" {
  id = "finding-abc123"
}

output "finding_title" {
  value = data.splunkes_finding.recent.rule_title
}

output "finding_severity" {
  value = data.splunkes_finding.recent.severity
}
```

## Argument Reference

* `id` - (Required) The ID of the finding to look up.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `rule_title` - The title of the finding.
* `security_domain` - The security domain (access, endpoint, network, identity, threat, audit).
* `risk_score` - The risk score.
* `risk_object` - The entity affected.
* `risk_object_type` - The entity type (user, system, other).
* `severity` - The severity level.

## API Reference

* [Retrieve findings](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/findings_1/public_v2_get_findings)
* [Retrieve finding by ID](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/findings_1/public_v2_get_finding_by_id)
* [ES API Overview](https://help.splunk.com/en/splunk-enterprise-security-8/rest-api-reference/8.0/overview/the-splunk-enterprise-security-api)
