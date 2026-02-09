---
page_title: "splunkes_investigation_note Resource - Splunk ES Provider"
description: |-
  Manages a note attached to an investigation in Splunk Enterprise Security.
---

# splunkes_investigation_note (Resource)

Manages notes attached to investigations in Splunk Enterprise Security. Notes can be used for runbooks, analysis documentation, and tracking investigation progress within an investigation.

## Example Usage

```hcl
resource "splunkes_investigation_note" "runbook" {
  investigation_id = splunkes_investigation.phishing_campaign.id

  content = <<-NOTE
    ## Phishing Investigation Runbook

    ### Triage Steps
    1. Verify sender domain against known-good domains
    2. Check for typosquatting
    3. Contact sender via phone

    ### Escalation
    - If credentials compromised: Reset all affected accounts
    - If ongoing campaign: Engage threat intelligence team
  NOTE
}
```

## Argument Reference

The following arguments are supported:

* `investigation_id` - (Required, String, ForceNew) The ID of the parent investigation. Changing this forces a new resource.
* `content` - (Required, String) The note content. Supports markdown formatting.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The note ID.
* `created_time` - The timestamp when the note was created.
* `modified_time` - The timestamp when the note was last modified.

## Import

Import is supported using the following syntax:

```shell
terraform import splunkes_investigation_note.example "investigation_id/note_id"
```

The import ID is a composite of the investigation ID and note ID, separated by `/`.
