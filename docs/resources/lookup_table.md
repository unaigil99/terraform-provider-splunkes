---
page_title: "splunkes_lookup_table Resource - Splunk ES Provider"
description: |-
  Manages a CSV lookup table file in Splunk.
---

# splunkes_lookup_table (Resource)

Manages CSV lookup table files in Splunk. Upload and manage CSV data that can be used by lookup definitions for search enrichment.

## Example Usage

```hcl
resource "splunkes_lookup_table" "threat_ips" {
  name  = "threat_ips.csv"
  app   = "SplunkEnterpriseSecuritySuite"
  owner = "nobody"
  content = <<-CSV
    ip,threat_category,confidence,description
    10.0.0.100,c2,high,Known C2 server
    192.168.1.200,scanner,medium,Port scanner
  CSV
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) CSV filename (e.g., `"my_lookup.csv"`). Changing this forces a new resource.
* `app` - (Optional, String) Splunk app context. Default: `"search"`.
* `owner` - (Optional, String) Splunk user owner. Default: `"nobody"`.
* `content` - (Optional, String) CSV content body.

~> **Note:** Splunk does not return CSV content on read, so the `content` field is write-only. Changes made to the CSV content outside of Terraform will not be detected.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The lookup table file ID.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_lookup_table.example "filename.csv"
```
