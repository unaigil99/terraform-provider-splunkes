---
page_title: "splunkes_investigation_note Resource - Splunk ES Provider"
description: |-
  Manages a note attached to an investigation in Splunk Enterprise Security.
---

# splunkes_investigation_note (Resource)

Manages notes attached to investigations in Splunk Enterprise Security. Notes can be used for runbooks, analysis documentation, and tracking investigation progress within an investigation.

This resource interacts with the ES v2 API at `/servicesNS/nobody/missioncontrol/public/v2/investigations/{id}/notes`.

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

## API Reference

This resource uses the **Splunk Enterprise Security v2 API**:

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Create | `POST` | `/servicesNS/nobody/missioncontrol/public/v2/investigations/{id}/notes` |
| Read | `GET` | `/servicesNS/nobody/missioncontrol/public/v2/investigations/{id}/notes` |
| Update | `POST` | `/servicesNS/nobody/missioncontrol/public/v2/investigations/{id}/notes/{noteID}` |
| Delete | `DELETE` | `/servicesNS/nobody/missioncontrol/public/v2/investigations/{id}/notes/{noteID}` |

**Official Splunk Documentation:**

* [Create note in investigation](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/notes_1/public_v2_create_note_in_investigation)
* [Get notes from investigation](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/notes_1/public_v2_get_notes_from_investigation)
* [Update note in investigation](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/notes_1/public_v2_update_note_in_investigation)
* [Delete note from investigation](https://help.splunk.com/en/splunk-enterprise-security-8/api-reference/8.3/splunk-enterprise-security-api-reference/notes_1/public_v2_delete_note_in_investigation)
* [ES API Overview](https://help.splunk.com/en/splunk-enterprise-security-8/rest-api-reference/8.0/overview/the-splunk-enterprise-security-api)
