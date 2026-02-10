---
page_title: "splunkes_saved_search Resource - Splunk ES Provider"
description: |-
  Manages a saved search in Splunk.
---

# splunkes_saved_search (Resource)

Manages generic saved searches in Splunk. This resource handles standard saved searches, alerts, and reports. For ES correlation searches with MITRE ATT&CK and notable events, use [`splunkes_correlation_search`](correlation_search.md) instead.

This resource interacts with the Splunk REST API at `/servicesNS/{owner}/{app}/saved/searches`.

## Example Usage

```hcl
resource "splunkes_saved_search" "failed_logins" {
  name        = "Failed Login Report"
  description = "Daily report of failed login attempts"
  app         = "search"

  search = <<-SPL
    index=security sourcetype=WinEventLog:Security EventCode=4625
    | stats count by src_ip, user
    | sort -count
  SPL

  is_scheduled   = true
  cron_schedule  = "0 8 * * *"

  dispatch_earliest_time = "-24h@h"
  dispatch_latest_time   = "now"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) The name of the saved search. Changing this forces a new resource.
* `search` - (Required, String) The SPL search query.
* `description` - (Optional, String) Description. Default: `""`.
* `app` - (Optional, String) Splunk app context. Default: `"search"`.
* `owner` - (Optional, String) Splunk user owner. Default: `"nobody"`.
* `disabled` - (Optional, Bool) Whether the saved search is disabled. Default: `false`.
* `is_scheduled` - (Optional, Bool) Whether the saved search runs on a schedule. Default: `false`.
* `cron_schedule` - (Optional, String) Cron schedule for the saved search. Default: `""`.
* `dispatch_earliest_time` - (Optional, String) Earliest time for the search window. Default: `""`.
* `dispatch_latest_time` - (Optional, String) Latest time for the search window. Default: `""`.
* `actions` - (Optional, String) Comma-separated alert actions. Default: `""`.
* `action_email_to` - (Optional, String) Email recipients for email alert actions. Default: `""`.
* `action_email_subject` - (Optional, String) Email subject for email alert actions. Default: `""`.
* `alert_type` - (Optional, String) Alert type. Default: `""`.
* `alert_comparator` - (Optional, String) Alert comparator. Default: `""`.
* `alert_threshold` - (Optional, String) Alert threshold. Default: `""`.
* `alert_severity` - (Optional, Int64) Alert severity (1-5). Default: `3`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The saved search ID.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_saved_search.example "search_name"
```

The import ID is the name of the search.

~> **Note:** After import, `owner` defaults to `"nobody"` and `app` defaults to `"search"`.

## API Reference

This resource uses the **Splunk Enterprise REST API**:

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Create | `POST` | `/servicesNS/{owner}/{app}/saved/searches` |
| Read | `GET` | `/servicesNS/{owner}/{app}/saved/searches/{name}` |
| Update | `POST` | `/servicesNS/{owner}/{app}/saved/searches/{name}` |
| Delete | `DELETE` | `/servicesNS/{owner}/{app}/saved/searches/{name}` |

**Official Splunk Documentation:**

* [Search endpoint descriptions](https://help.splunk.com/en/splunk-enterprise/rest-api-reference/9.4/search-endpoints/search-endpoint-descriptions) - `saved/searches` endpoint
* [Creating searches using the REST API](https://docs.splunk.com/Documentation/Splunk/latest/RESTTUT/RESTsearches)
