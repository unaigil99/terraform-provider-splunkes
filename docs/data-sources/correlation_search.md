---
page_title: "splunkes_correlation_search Data Source - Splunk ES Provider"
description: |-
  Reads an existing correlation search from Splunk Enterprise Security.
---

# splunkes_correlation_search (Data Source)

Use this data source to read an existing correlation search from Splunk Enterprise Security. This is useful for referencing existing detections that are not managed by Terraform.

## Example Usage

```hcl
data "splunkes_correlation_search" "existing" {
  name = "Access - Brute Force Access Behavior Detected - Rule"
  app  = "SplunkEnterpriseSecuritySuite"
}

output "search_query" {
  value = data.splunkes_correlation_search.existing.search
}

output "is_enabled" {
  value = !data.splunkes_correlation_search.existing.disabled
}
```

## Argument Reference

* `name` - (Required) The name of the correlation search to look up.
* `app` - (Optional) The Splunk app context. Defaults to `SplunkEnterpriseSecuritySuite`.
* `owner` - (Optional) The Splunk user owner. Defaults to `nobody`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `search` - The SPL search query.
* `description` - The description of the correlation search.
* `disabled` - Whether the correlation search is disabled.
* `cron_schedule` - The cron schedule.
* `notable_enabled` - Whether notable event creation is enabled.
* `security_domain` - The security domain.
* `severity` - The severity level.
* `risk_score` - The default risk score.
