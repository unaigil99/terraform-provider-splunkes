---
page_title: "splunkes_threat_intel Resource - Splunk ES Provider"
description: |-
  Manages a threat intelligence indicator in Splunk Enterprise Security.
---

# splunkes_threat_intel (Resource)

Manages threat intelligence indicators in Splunk Enterprise Security. Indicators can be added to threat intel collections such as `ip_intel`, `domain_intel`, `email_intel`, `file_intel`, and others, enabling automated correlation against incoming security data.

## Example Usage

```hcl
resource "splunkes_threat_intel" "phishing_domain" {
  collection  = "domain_intel"
  description = "Phishing domain targeting real estate closings"
  fields = {
    "domain"          = "secure-closing-docs.com"
    "threat_category" = "phishing"
    "weight"          = "4"
    "description"     = "Domain used in BEC attacks against title companies"
  }
}

resource "splunkes_threat_intel" "c2_ip" {
  collection  = "ip_intel"
  description = "C2 server for BEC campaign"
  fields = {
    "ip"              = "198.51.100.42"
    "threat_category" = "c2"
    "weight"          = "5"
  }
}
```

## Argument Reference

The following arguments are supported:

* `collection` - (Required, String, ForceNew) The threat intel collection name. Common values are `ip_intel`, `domain_intel`, `email_intel`, `file_intel`. Changing this forces a new resource.
* `fields` - (Required, Map of String) Field values for the indicator. The keys depend on the collection type (see below).
* `description` - (Optional, String) Description of the indicator. Default: `""`.

### Common Collections and Key Fields

| Collection     | Key Fields                                       |
|----------------|--------------------------------------------------|
| `ip_intel`     | `ip`, `threat_category`, `weight`                |
| `domain_intel` | `domain`, `threat_category`, `weight`            |
| `email_intel`  | `src_user`, `threat_category`, `weight`          |
| `file_intel`   | `file_hash`, `file_name`, `threat_category`, `weight` |

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `_key` of the threat intel item in the collection.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_threat_intel.example "collection/key"
```

The import ID is a composite of the collection name and the item `_key`, separated by `/`.
