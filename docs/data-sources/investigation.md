---
page_title: "splunkes_investigation Data Source - Splunk ES Provider"
description: |-
  Reads an existing investigation from Splunk Enterprise Security.
---

# splunkes_investigation (Data Source)

Use this data source to read an existing investigation from Splunk Enterprise Security.

## Example Usage

```hcl
data "splunkes_investigation" "active" {
  id = "abc123-def456"
}

output "investigation_status" {
  value = data.splunkes_investigation.active.status
}

output "assigned_to" {
  value = data.splunkes_investigation.active.assignee
}
```

## Argument Reference

* `id` - (Required) The ID of the investigation to look up.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `name` - The investigation title.
* `status` - The current status (new, in_progress, pending, on_hold, resolved, closed).
* `assignee` - The assigned analyst username.
* `priority` - The priority level (informational, low, medium, high, critical).

## API Reference

* [Retrieve investigations](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/investigation_1/public_v2_list_investigations)
* [ES API Overview](https://help.splunk.com/en/splunk-enterprise-security-8/rest-api-reference/8.0/overview/the-splunk-enterprise-security-api)
