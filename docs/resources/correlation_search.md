---
page_title: "splunkes_correlation_search Resource - Splunk ES Provider"
description: |-
  Manages a correlation search in Splunk Enterprise Security.
---

# splunkes_correlation_search (Resource)

Manages correlation searches in Splunk Enterprise Security. Correlation searches are scheduled saved searches that detect security threats and generate notable events, risk scores, and MITRE ATT&CK annotations.

This resource interacts with the Splunk REST API at `/servicesNS/{owner}/{app}/saved/searches`.

## Example Usage

```hcl
resource "splunkes_correlation_search" "brute_force" {
  name        = "Brute Force Login Attempts"
  description = "Detects multiple failed login attempts from a single source"
  app         = "SplunkEnterpriseSecuritySuite"

  search = <<-SPL
    index=security sourcetype=WinEventLog:Security EventCode=4625
    | stats count min(_time) as firstTime max(_time) as lastTime by src_ip, Account_Name, dest
    | where count > 10
  SPL

  cron_schedule          = "*/10 * * * *"
  dispatch_earliest_time = "-20m@m"
  dispatch_latest_time   = "now"

  correlation_search_enabled = true
  correlation_search_label   = "Brute Force Detection"

  notable_enabled          = true
  notable_rule_title       = "Brute Force: $count$ failed logins from $src_ip$"
  notable_rule_description = "Multiple failed login attempts detected from $src_ip$ targeting $Account_Name$ on $dest$"
  notable_security_domain  = "access"
  notable_severity         = "high"

  risk_enabled      = true
  risk_score        = 60
  risk_message      = "Brute force activity from $src_ip$ targeting $Account_Name$"
  risk_object_field = "src_ip"
  risk_object_type  = "system"

  mitre_attack_ids  = ["T1110.001", "T1110.003"]
  kill_chain_phases = ["Exploitation"]
  analytic_story    = ["Credential Access"]

  alert_severity = 4
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) The name of the correlation search. Changing this forces a new resource.
* `search` - (Required, String) The SPL search query.
* `cron_schedule` - (Required, String) The cron schedule for the search. Example: `"*/5 * * * *"`.
* `description` - (Optional, String) Description of the correlation search. Default: `""`.
* `app` - (Optional, String) Splunk app context. Default: `"SplunkEnterpriseSecuritySuite"`.
* `owner` - (Optional, String) Splunk user owner. Default: `"nobody"`.
* `disabled` - (Optional, Bool) Whether the search is disabled. Default: `false`.
* `is_scheduled` - (Optional, Bool) Whether the search runs on a schedule. Default: `true`.
* `dispatch_earliest_time` - (Optional, String) Earliest time for the search window. Default: `"-60m@m"`.
* `dispatch_latest_time` - (Optional, String) Latest time for the search window. Default: `"now"`.
* `schedule_priority` - (Optional, String) Scheduling priority. Valid values are `default`, `higher`, `highest`. Default: `"default"`.
* `correlation_search_enabled` - (Optional, Bool) Enable ES correlation search mode. Default: `true`.
* `correlation_search_label` - (Optional, String) Label for the correlation search. Default: `""`.
* `security_domain` - (Optional, String) Security domain. Valid values are `access`, `endpoint`, `network`, `threat`, `identity`, `audit`. Default: `"threat"`.
* `notable_enabled` - (Optional, Bool) Enable notable event creation. Default: `false`.
* `notable_rule_title` - (Optional, String) Title for notable events. Supports `$field$` substitution. Default: `""`.
* `notable_rule_description` - (Optional, String) Description for notable events. Default: `""`.
* `notable_security_domain` - (Optional, String) Security domain for notable events. Default: `"threat"`.
* `notable_severity` - (Optional, String) Severity for notable events. Valid values are `informational`, `low`, `medium`, `high`, `critical`. Default: `"medium"`.
* `notable_drilldown_name` - (Optional, String) Drilldown search name. Default: `""`.
* `notable_drilldown_search` - (Optional, String) Drilldown SPL query. Default: `""`.
* `notable_default_owner` - (Optional, String) Default owner for notable events. Default: `"unassigned"`.
* `notable_recommended_actions` - (Optional, String) Comma-separated recommended actions. Default: `""`.
* `risk_enabled` - (Optional, Bool) Enable risk scoring. Default: `false`.
* `risk_score` - (Optional, Int64) Default risk score (0-100). Default: `0`.
* `risk_message` - (Optional, String) Risk event message. Default: `""`.
* `risk_object_field` - (Optional, String) Field containing the risk object. Default: `""`.
* `risk_object_type` - (Optional, String) Risk object type. Valid values are `system`, `user`, `other`. Default: `"system"`.
* `mitre_attack_ids` - (Optional, List of String) MITRE ATT&CK technique IDs. Example: `["T1003.001", "T1110"]`.
* `kill_chain_phases` - (Optional, List of String) Kill chain phases.
* `cis20` - (Optional, List of String) CIS 20 controls.
* `nist` - (Optional, List of String) NIST framework mappings.
* `analytic_story` - (Optional, List of String) Associated analytic stories.
* `alert_type` - (Optional, String) Alert type. Valid values are `always`, `custom`, `number of events`. Default: `"number of events"`.
* `alert_comparator` - (Optional, String) Alert comparator (e.g., `greater than`). Default: `"greater than"`.
* `alert_threshold` - (Optional, String) Alert threshold. Default: `"0"`.
* `alert_severity` - (Optional, Int64) Alert severity integer. Valid values are `1` (debug), `2` (info), `3` (warn), `4` (error), `5` (critical). Default: `3`.
* `alert_suppress` - (Optional, Bool) Enable alert suppression. Default: `false`.
* `alert_suppress_period` - (Optional, String) Suppression period. Default: `""`.
* `alert_suppress_fields` - (Optional, String) Comma-separated suppression fields. Default: `""`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The saved search ID.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_correlation_search.example "search_name"
```

The import ID is the name of the search.

~> **Note:** After import, `owner` defaults to `"nobody"` and `app` defaults to `"SplunkEnterpriseSecuritySuite"`.

## API Reference

This resource uses the **Splunk Enterprise REST API** (`saved/searches` endpoint with ES correlation search attributes):

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Create | `POST` | `/servicesNS/{owner}/{app}/saved/searches` |
| Read | `GET` | `/servicesNS/{owner}/{app}/saved/searches/{name}` |
| Update | `POST` | `/servicesNS/{owner}/{app}/saved/searches/{name}` |
| Delete | `DELETE` | `/servicesNS/{owner}/{app}/saved/searches/{name}` |

Correlation searches are saved searches with `action.correlationsearch.enabled=1` and additional ES-specific attributes stored in `savedsearches.conf`.

**Official Splunk Documentation:**

* [Search endpoint descriptions](https://help.splunk.com/en/splunk-enterprise/rest-api-reference/9.4/search-endpoints/search-endpoint-descriptions) - `saved/searches` endpoint
* [Configure correlation searches in ES](https://help.splunk.com/en/splunk-enterprise-security-7/administer/7.2/correlation-searches/configure-correlation-searches-in-splunk-enterprise-security)
* [List correlation searches in ES](https://help.splunk.com/en/splunk-enterprise-security-7/administer/7.3/correlation-searches/list-correlation-searches-in-splunk-enterprise-security)
* [Correlation Search creation via REST API](https://splunk.my.site.com/customer/s/article/Correlation-Search-creation-in-Enterprise-Security-through-REST-API)
