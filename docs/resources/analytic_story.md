---
page_title: "splunkes_analytic_story Resource - Splunk ES Provider"
description: |-
  Manages an analytic story in Splunk Enterprise Security.
---

# splunkes_analytic_story (Resource)

Manages analytic stories in Splunk Enterprise Security. Analytic stories group related detection searches and provide context for security investigations, helping analysts understand the full scope of a threat scenario.

## Example Usage

```hcl
resource "splunkes_analytic_story" "wire_fraud" {
  name        = "Real Estate Wire Fraud Detection"
  app         = "SplunkEnterpriseSecuritySuite"
  description = "Detects wire fraud attempts targeting real estate transactions"

  category              = ["Adversary Tactics", "Financial Fraud"]
  data_models           = ["Email", "Network_Traffic", "Web"]
  providing_technologies = ["Microsoft 365", "Proofpoint"]

  detection_searches = [
    splunkes_correlation_search.bec_detection.name,
    splunkes_correlation_search.credential_dump.name,
  ]

  investigative_searches = ["Get Email Details"]
  contextual_searches    = ["Get Risk Score For User"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String, ForceNew) The name of the analytic story. Changing this forces a new resource.
* `app` - (Optional, String) Splunk app context. Default: `"SplunkEnterpriseSecuritySuite"`.
* `owner` - (Optional, String) Splunk user owner. Default: `"nobody"`.
* `description` - (Optional, String) Description of the analytic story. Default: `""`.
* `category` - (Optional, List of String) Categories for the analytic story.
* `data_models` - (Optional, List of String) Related Splunk data models.
* `providing_technologies` - (Optional, List of String) Technologies that provide the data used by this story.
* `detection_searches` - (Optional, List of String) Names of detection searches included in this story.
* `investigative_searches` - (Optional, List of String) Names of investigative searches included in this story.
* `contextual_searches` - (Optional, List of String) Names of contextual searches included in this story.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The analytic story ID.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_analytic_story.example "story_name"
```

The import ID is the name of the analytic story.

~> **Note:** After import, `owner` defaults to `"nobody"` and `app` defaults to `"SplunkEnterpriseSecuritySuite"`.
