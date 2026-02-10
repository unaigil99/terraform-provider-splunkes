---
page_title: "splunkes_finding Resource - Splunk ES Provider"
description: |-
  Manages a security finding in Splunk Enterprise Security.
---

# splunkes_finding (Resource)

Manages security findings in Splunk Enterprise Security. Findings represent manual security observations that analysts create to document threats, suspicious activity, or policy violations.

This resource interacts with the ES v2 API at `/servicesNS/nobody/missioncontrol/public/v2/findings`.

~> **Note:** Findings cannot be updated or deleted via the ES v2 API. Attempting to update a finding will produce a warning and retain the original values. Destroying the resource removes it from Terraform state only; the finding remains in Splunk ES.

## Example Usage

```hcl
resource "splunkes_finding" "wire_fraud" {
  rule_title       = "Suspicious Wire Transfer Request"
  rule_description = "Manual finding: Wire transfer request from unverified email domain"
  security_domain  = "threat"
  risk_score       = 85
  risk_object      = "finance-director@company.com"
  risk_object_type = "user"
  severity         = "critical"
}
```

## Argument Reference

The following arguments are supported:

* `rule_title` - (Required, String) Title of the finding.
* `security_domain` - (Required, String) Security domain. Valid values are `access`, `endpoint`, `network`, `identity`, `threat`, `audit`.
* `risk_score` - (Required, Int64) Risk score from 0 to 100.
* `risk_object` - (Required, String) The entity affected by the finding (e.g., a username, hostname, or IP address).
* `risk_object_type` - (Required, String) Type of the risk object. Valid values are `user`, `system`, `other`.
* `rule_description` - (Optional, String) Description of the finding. Default: `""`.
* `severity` - (Optional, String) Severity level. Valid values are `informational`, `low`, `medium`, `high`, `critical`.
* `owner` - (Optional, String) Owner or analyst assigned to the finding. Default: `""`.
* `status` - (Optional, String) Status of the finding. Default: `""`.
* `urgency` - (Optional, String) Urgency level. Default: `""`.
* `disposition` - (Optional, String) Disposition of the finding. Default: `""`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The finding ID.
* `time` - The timestamp when the finding was created.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_finding.example "finding_id"
```

The import ID is the finding ID from Splunk ES.

## Special Behavior

* **Update:** Not supported by the ES v2 API. Terraform will display a warning and retain the original values. To change a finding, destroy and recreate the resource.
* **Destroy:** Removes the finding from Terraform state only. The finding remains in Splunk ES and cannot be deleted via the API.

## API Reference

This resource uses the **Splunk Enterprise Security v2 API**:

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Create | `POST` | `/servicesNS/nobody/missioncontrol/public/v2/findings` |
| Read | `GET` | `/servicesNS/nobody/missioncontrol/public/v2/findings/{id}` |

**Official Splunk Documentation:**

* [Create manual finding](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/findings_1/public_v2_create_manual_finding)
* [Retrieve finding by ID](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/findings_1/public_v2_get_finding_by_id)
* [Retrieve findings](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/findings_1/public_v2_get_findings)
* [ES API Overview](https://help.splunk.com/en/splunk-enterprise-security-8/rest-api-reference/8.0/overview/the-splunk-enterprise-security-api)
