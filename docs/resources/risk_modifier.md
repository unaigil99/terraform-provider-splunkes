---
page_title: "splunkes_risk_modifier Resource - Splunk ES Provider"
description: |-
  Manages a risk score modifier for an entity in Splunk Enterprise Security.
---

# splunkes_risk_modifier (Resource)

Manages risk score modifiers for entities in Splunk Enterprise Security. Risk modifiers adjust the baseline risk score for specific users or systems, allowing security teams to reflect the relative importance or exposure of assets and identities.

This resource interacts with the ES v2 API at `/servicesNS/nobody/missioncontrol/public/v2/risks/risk_scores`.

## Example Usage

```hcl
resource "splunkes_risk_modifier" "cfo" {
  entity        = "cfo@company.com"
  entity_type   = "user"
  risk_modifier = 25
  description   = "C-suite executive with wire transfer authority"
}

resource "splunkes_risk_modifier" "txn_server" {
  entity        = "txn-server-01"
  entity_type   = "system"
  risk_modifier = 15
  description   = "Critical asset: transaction management server"
}
```

## Argument Reference

The following arguments are supported:

* `entity` - (Required, String, ForceNew) The entity name (username, hostname, IP address, etc.). Changing this forces a new resource.
* `entity_type` - (Required, String) Type of the entity. Valid values are `user`, `system`.
* `risk_modifier` - (Required, Int64) Risk modifier value. This value is added to the entity's base risk score.
* `description` - (Optional, String) Description of why this modifier exists. Default: `""`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The risk modifier ID.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_risk_modifier.example "entity/entity_type"
```

The import ID is a composite of the entity name and entity type, separated by `/`.

## Special Behavior

* **Destroy:** Removes the risk modifier from Terraform state only. The risk modifier remains in Splunk ES.

## API Reference

This resource uses the **Splunk Enterprise Security v2 API**:

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Create/Update | `POST` | `/servicesNS/nobody/missioncontrol/public/v2/risks/risk_scores/{entity}` |
| Read | `GET` | `/servicesNS/nobody/missioncontrol/public/v2/risks/risk_scores/{entity}?entity_type=...` |

**Official Splunk Documentation:**

* [Add risk modifiers](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/risks_1/public_v2_risk_entity_risk_scores_update)
* [Get risk scores](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/risks_1/public_v2_risk_entity_risk_scores_retrieve)
* [ES API Overview](https://help.splunk.com/en/splunk-enterprise-security-8/rest-api-reference/8.0/overview/the-splunk-enterprise-security-api)
