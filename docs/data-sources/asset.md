---
page_title: "splunkes_asset Data Source - Splunk ES Provider"
description: |-
  Reads asset information from Splunk Enterprise Security.
---

# splunkes_asset (Data Source)

Use this data source to read asset (system/device) information from the Splunk Enterprise Security asset framework. This is a read-only API.

## Example Usage

```hcl
data "splunkes_asset" "server" {
  id = "web-server-01"
}

output "server_ip" {
  value = data.splunkes_asset.server.ip
}

output "server_priority" {
  value = data.splunkes_asset.server.priority
}

output "is_expected" {
  value = data.splunkes_asset.server.is_expected
}
```

## Argument Reference

* `id` - (Required) The asset ID to look up.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `ip` - IP address of the asset.
* `mac` - MAC address.
* `dns` - DNS hostname.
* `nt_host` - Windows hostname.
* `bunit` - Business unit.
* `category` - Category classification.
* `priority` - Priority level.
* `is_expected` - Whether the asset is expected/known (boolean).

## API Reference

* [Retrieve assets](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/assets_1/public_v2_assets)
* [ES API Overview](https://help.splunk.com/en/splunk-enterprise-security-8/rest-api-reference/8.0/overview/the-splunk-enterprise-security-api)
