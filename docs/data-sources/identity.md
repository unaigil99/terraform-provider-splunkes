---
page_title: "splunkes_identity Data Source - Splunk ES Provider"
description: |-
  Reads identity information from Splunk Enterprise Security.
---

# splunkes_identity (Data Source)

Use this data source to read identity (user) information from the Splunk Enterprise Security identity framework. This is a read-only API.

## Example Usage

```hcl
data "splunkes_identity" "analyst" {
  id = "jsmith"
}

output "analyst_email" {
  value = data.splunkes_identity.analyst.email
}

output "analyst_department" {
  value = data.splunkes_identity.analyst.bunit
}
```

## Argument Reference

* `id` - (Required) The identity ID to look up.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `first_name` - First name of the identity.
* `last_name` - Last name of the identity.
* `email` - Email address.
* `bunit` - Business unit.
* `category` - Category classification.
* `priority` - Priority level.
* `watchlist` - Whether the identity is on a watchlist (boolean).
