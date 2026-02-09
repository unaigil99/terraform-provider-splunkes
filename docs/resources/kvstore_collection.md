---
page_title: "splunkes_kvstore_collection Resource - Splunk ES Provider"
description: |-
  Manages a KV store collection in Splunk.
---

# splunkes_kvstore_collection (Resource)

Manages KV store collections in Splunk. KV store provides a way to store and retrieve key-value data, commonly used for lookup tables and application state.

## Example Usage

```hcl
resource "splunkes_kvstore_collection" "investigation_tracking" {
  name  = "investigation_tracking"
  app   = "SplunkEnterpriseSecuritySuite"
  owner = "nobody"
  fields = {
    "case_id"     = "string"
    "analyst"     = "string"
    "status"      = "string"
    "risk_score"  = "number"
    "created_at"  = "time"
    "is_resolved" = "bool"
  }
  accelerated_fields = {
    "idx_case" = "{\"case_id\": 1}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) Name of the KV store collection. Changing this forces a new resource.
* `app` - (Optional, String) Splunk app context. Default: `"search"`.
* `owner` - (Optional, String) Splunk user owner. Default: `"nobody"`.
* `fields` - (Optional, Map of String) Field definitions. Keys are field names, values are types. Valid types are `string`, `number`, `bool`, `time`.
* `accelerated_fields` - (Optional, Map of String) Accelerated field definitions. Keys are index names, values are JSON index specifications.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The KV store collection ID.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_kvstore_collection.example "collection_name"
```
