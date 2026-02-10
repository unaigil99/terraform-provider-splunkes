---
page_title: "splunkes_lookup_definition Resource - Splunk ES Provider"
description: |-
  Manages a lookup definition in Splunk.
---

# splunkes_lookup_definition (Resource)

Manages lookup definitions in Splunk. Lookup definitions map to either CSV files or KV store collections, enabling data enrichment in searches.

This resource interacts with the Splunk REST API at `/servicesNS/{owner}/{app}/data/transforms/lookups`.

## Example Usage

```hcl
# CSV-backed lookup
resource "splunkes_lookup_definition" "threat_ips" {
  name                 = "threat_ips_lookup"
  filename             = splunkes_lookup_table.threat_ips.name
  app                  = "SplunkEnterpriseSecuritySuite"
  case_sensitive_match = false
  fields_list          = "ip,threat_category,confidence,description"
}

# KV store-backed lookup
resource "splunkes_lookup_definition" "investigation_tracking" {
  name          = "investigation_tracking_lookup"
  external_type = "kvstore"
  collection    = splunkes_kvstore_collection.tracking.name
  app           = "SplunkEnterpriseSecuritySuite"
  fields_list   = "case_id,analyst,status,risk_score"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) Name of the lookup definition. Changing this forces a new resource.
* `filename` - (Optional, String) CSV filename for CSV-backed lookups. Mutually exclusive with `external_type` and `collection`.
* `app` - (Optional, String) Splunk app context. Default: `"search"`.
* `owner` - (Optional, String) Splunk user owner. Default: `"nobody"`.
* `external_type` - (Optional, String) External lookup type. Set to `"kvstore"` for KV store-backed lookups.
* `collection` - (Optional, String) KV store collection name. Required when `external_type` is `"kvstore"`.
* `fields_list` - (Optional, String) Comma-separated field names. Default: `""`.
* `max_matches` - (Optional, Int64) Maximum number of matches. Default: computed by Splunk.
* `min_matches` - (Optional, Int64) Minimum number of matches. Default: computed by Splunk.
* `default_match` - (Optional, String) Default value when no match is found. Default: `""`.
* `case_sensitive_match` - (Optional, Bool) Whether matching is case-sensitive. Default: `true`.
* `match_type` - (Optional, String) Match type (e.g., `"WILDCARD(field)"`, `"CIDR(field)"`). Default: `""`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The lookup definition ID.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_lookup_definition.example "lookup_name"
```

## API Reference

This resource uses the **Splunk Enterprise REST API**:

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Create | `POST` | `/servicesNS/{owner}/{app}/data/transforms/lookups` |
| Read | `GET` | `/servicesNS/{owner}/{app}/data/transforms/lookups/{name}` |
| Update | `POST` | `/servicesNS/{owner}/{app}/data/transforms/lookups/{name}` |
| Delete | `DELETE` | `/servicesNS/{owner}/{app}/data/transforms/lookups/{name}` |

**Official Splunk Documentation:**

* [Knowledge endpoint descriptions](https://help.splunk.com/en/splunk-enterprise/rest-api-reference/9.4/knowledge-endpoints/knowledge-endpoint-descriptions) - `data/transforms/lookups` endpoint
* [Managing knowledge objects](https://help.splunk.com/en/splunk-cloud-platform/leverage-rest-apis/rest-api-tutorials/9.2.2406/rest-api-tutorials/managing-knowledge-objects)
