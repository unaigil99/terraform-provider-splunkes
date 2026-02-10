---
page_title: "splunkes_macro Resource - Splunk ES Provider"
description: |-
  Manages a search macro in Splunk.
---

# splunkes_macro (Resource)

Manages search macros in Splunk. Macros are reusable SPL snippets that can be referenced in searches using backtick syntax.

This resource interacts with the Splunk REST API at `/servicesNS/{owner}/{app}/configs/conf-macros`.

## Example Usage

```hcl
# Simple macro
resource "splunkes_macro" "sysmon_events" {
  name        = "sysmon_events"
  definition  = "index=main sourcetype=XmlWinEventLog:Microsoft-Windows-Sysmon/Operational"
  description = "Base filter for all Sysmon events"
  app         = "SplunkEnterpriseSecuritySuite"
}

# Macro with arguments
resource "splunkes_macro" "threshold_alert" {
  name        = "threshold_alert(2)"
  definition  = "where count > $min_count$ | eval alert_level=if(count > $high_count$, \"critical\", \"warning\")"
  args        = "min_count,high_count"
  description = "Parameterized threshold alert with dynamic severity"
  app         = "SplunkEnterpriseSecuritySuite"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) Name of the macro. Include the argument count in parentheses for macros with arguments (e.g., `"my_macro(2)"`). Changing this forces a new resource.
* `definition` - (Required, String) The SPL definition. Use `$arg_name$` syntax for arguments.
* `description` - (Optional, String) Description of the macro. Default: `""`.
* `app` - (Optional, String) Splunk app context. Default: `"search"`.
* `owner` - (Optional, String) Splunk user owner. Default: `"nobody"`.
* `args` - (Optional, String) Comma-separated argument names. Default: `""`.
* `validation` - (Optional, String) Validation expression for arguments. Default: `""`.
* `errormsg` - (Optional, String) Error message displayed when validation fails. Default: `""`.
* `iseval` - (Optional, Bool) Whether the definition is an eval expression. Default: `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The macro ID.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_macro.example "macro_name"
```

## Usage in SPL

Reference macros in your SPL searches using backtick syntax:

```
`sysmon_events` EventCode=1
`threshold_alert(10, 50)`
```

## API Reference

This resource uses the **Splunk Enterprise REST API** (configuration endpoint for `macros.conf`):

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Create | `POST` | `/servicesNS/{owner}/{app}/configs/conf-macros` |
| Read | `GET` | `/servicesNS/{owner}/{app}/configs/conf-macros/{name}` |
| Update | `POST` | `/servicesNS/{owner}/{app}/configs/conf-macros/{name}` |
| Delete | `DELETE` | `/servicesNS/{owner}/{app}/configs/conf-macros/{name}` |

**Official Splunk Documentation:**

* [Knowledge endpoint descriptions](https://help.splunk.com/en/splunk-enterprise/rest-api-reference/9.4/knowledge-endpoints/knowledge-endpoint-descriptions)
* [Configuration endpoint descriptions](https://help.splunk.com/en/splunk-enterprise/leverage-rest-apis/rest-api-reference/10.0/configuration-endpoints/configuration-endpoint-descriptions) - `configs/conf-{file}` pattern
* [Use search macros in searches](https://docs.splunk.com/Documentation/Splunk/9.2.0/Knowledge/Usesearchmacros)
