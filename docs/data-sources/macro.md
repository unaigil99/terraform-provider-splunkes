---
page_title: "splunkes_macro Data Source - Splunk ES Provider"
description: |-
  Reads an existing search macro from Splunk.
---

# splunkes_macro (Data Source)

Use this data source to read an existing search macro from Splunk. This is useful for referencing macros that are managed outside of Terraform, such as built-in CIM macros.

## Example Usage

```hcl
data "splunkes_macro" "cim_endpoints" {
  name = "cim_Endpoint_Processes_indexes"
  app  = "SplunkEnterpriseSecuritySuite"
}

output "macro_definition" {
  value = data.splunkes_macro.cim_endpoints.definition
}
```

## Argument Reference

* `name` - (Required) The name of the macro.
* `app` - (Optional) The Splunk app context. Defaults to `SplunkEnterpriseSecuritySuite`.
* `owner` - (Optional) The Splunk user owner. Defaults to `nobody`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `definition` - The SPL definition of the macro.
* `description` - A description of the macro.
* `args` - Comma-separated argument names.
* `validation` - A validation expression for the macro arguments.
* `iseval` - Whether the macro definition is an eval expression.

## API Reference

* [Knowledge endpoint descriptions](https://help.splunk.com/en/splunk-enterprise/rest-api-reference/9.4/knowledge-endpoints/knowledge-endpoint-descriptions)
* [Configuration endpoint descriptions](https://help.splunk.com/en/splunk-enterprise/leverage-rest-apis/rest-api-reference/10.0/configuration-endpoints/configuration-endpoint-descriptions) - `configs/conf-macros` endpoint
