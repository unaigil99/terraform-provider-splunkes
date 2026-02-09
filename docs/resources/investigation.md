---
page_title: "splunkes_investigation Resource - Splunk ES Provider"
description: |-
  Manages an investigation in Splunk Enterprise Security.
---

# splunkes_investigation (Resource)

Manages investigations in Splunk Enterprise Security. Investigations track security incidents through their lifecycle, allowing analysts to coordinate response activities, assign ownership, and document progress.

This resource interacts with the ES v2 API at `/servicesNS/nobody/missioncontrol/public/v2/investigations`.

~> **Note:** When this resource is destroyed, the investigation is **closed** (set to status `"closed"`) rather than deleted, because the ES v2 API does not support deleting investigations.

## Example Usage

```hcl
resource "splunkes_investigation" "phishing_campaign" {
  name        = "Phishing Campaign Investigation"
  description = "Investigation tracking a targeted phishing campaign against the finance team"
  status      = "in_progress"
  priority    = "high"
  assignee    = "soc_analyst_1"
  tags        = ["phishing", "finance", "q1-2026"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The investigation title.
* `description` - (Optional, String) Investigation description. Default: `""`.
* `status` - (Optional, String) Status of the investigation. Valid values are `new`, `in_progress`, `pending`, `on_hold`, `resolved`, `closed`. Default: `"new"`.
* `assignee` - (Optional, String) Assigned analyst username. Default: `""`.
* `priority` - (Optional, String) Priority level. Valid values are `informational`, `low`, `medium`, `high`, `critical`, `unknown`. Default: `"unknown"`.
* `tags` - (Optional, List of String) Tags for categorization.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The investigation ID.
* `created_time` - The timestamp when the investigation was created.
* `modified_time` - The timestamp when the investigation was last modified.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_investigation.example "investigation_id"
```

The import ID is the investigation ID from Splunk ES.

~> **Note:** On destroy, the investigation is closed (status set to `"closed"`) rather than deleted, as the ES v2 API does not support deletion of investigations.
