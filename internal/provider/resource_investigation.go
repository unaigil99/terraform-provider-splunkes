package provider

import (
	"context"
	"fmt"

	"github.com/Agentric/terraform-provider-splunkes/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &investigationResource{}
	_ resource.ResourceWithConfigure   = &investigationResource{}
	_ resource.ResourceWithImportState = &investigationResource{}
)

// NewInvestigationResource returns a new resource.Resource for ES investigations.
func NewInvestigationResource() resource.Resource {
	return &investigationResource{}
}

type investigationResource struct {
	client *sdk.SplunkClient
}

type investigationResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Status       types.String `tfsdk:"status"`
	Assignee     types.String `tfsdk:"assignee"`
	Priority     types.String `tfsdk:"priority"`
	Tags         types.List   `tfsdk:"tags"`
	CreatedTime  types.String `tfsdk:"created_time"`
	ModifiedTime types.String `tfsdk:"modified_time"`
}

func (r *investigationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_investigation"
}

func (r *investigationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an investigation in Splunk Enterprise Security (ES v2 API).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ES investigation ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The investigation title.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the investigation.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Investigation status: new, in_progress, pending, on_hold, resolved, or closed.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("new", "in_progress", "pending", "on_hold", "resolved", "closed"),
				},
			},
			"assignee": schema.StringAttribute{
				Description: "The user assigned to the investigation.",
				Optional:    true,
			},
			"priority": schema.StringAttribute{
				Description: "Investigation priority: informational, low, medium, high, critical, or unknown.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("informational", "low", "medium", "high", "critical", "unknown"),
				},
			},
			"tags": schema.ListAttribute{
				Description: "A list of tags for the investigation.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"created_time": schema.StringAttribute{
				Description: "The time the investigation was created.",
				Computed:    true,
			},
			"modified_time": schema.StringAttribute{
				Description: "The time the investigation was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *investigationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.SplunkClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.SplunkClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *investigationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan investigationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"title": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		body["status"] = plan.Status.ValueString()
	}
	if !plan.Assignee.IsNull() && !plan.Assignee.IsUnknown() {
		body["assignee"] = plan.Assignee.ValueString()
	}
	if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
		body["urgency"] = plan.Priority.ValueString()
	}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		body["tags"] = tags
	}

	tflog.Debug(ctx, "Creating ES investigation", map[string]interface{}{"title": plan.Name.ValueString()})

	result, err := r.client.CreateInvestigation(body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating investigation", err.Error())
		return
	}

	// ES v2 returns _key or id directly in the response JSON.
	id := sdk.ParseString(result, "_key")
	if id == "" {
		id = sdk.ParseString(result, "id")
	}
	if id == "" {
		resp.Diagnostics.AddError(
			"Error creating investigation",
			"No _key or id returned in the API response.",
		)
		return
	}

	plan.ID = types.StringValue(id)

	// Read back the full state from the API.
	r.refreshState(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *investigationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state investigationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading ES investigation", map[string]interface{}{"id": state.ID.ValueString()})

	r.refreshState(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *investigationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan investigationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state investigationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body := map[string]interface{}{
		"title": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	} else {
		body["description"] = ""
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		body["status"] = plan.Status.ValueString()
	}
	if !plan.Assignee.IsNull() && !plan.Assignee.IsUnknown() {
		body["assignee"] = plan.Assignee.ValueString()
	} else {
		body["assignee"] = ""
	}
	if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
		body["urgency"] = plan.Priority.ValueString()
	}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		body["tags"] = tags
	} else {
		body["tags"] = []string{}
	}

	tflog.Debug(ctx, "Updating ES investigation", map[string]interface{}{"id": id})

	_, err := r.client.UpdateInvestigation(id, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating investigation", err.Error())
		return
	}

	plan.ID = types.StringValue(id)

	r.refreshState(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *investigationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state investigationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	tflog.Debug(ctx, "Closing ES investigation (no DELETE endpoint)", map[string]interface{}{"id": id})

	// ES investigations do not have a DELETE endpoint. Close the investigation instead.
	body := map[string]interface{}{
		"status": "closed",
	}
	_, err := r.client.UpdateInvestigation(id, body)
	if err != nil {
		resp.Diagnostics.AddError("Error closing investigation", err.Error())
		return
	}
}

func (r *investigationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// refreshState reads the investigation from the ES v2 API and updates the model.
// The ES v2 ReadInvestigation endpoint uses a filter query and returns a list-style
// JSON response. We extract the first matching entry.
func (r *investigationResource) refreshState(ctx context.Context, model *investigationResourceModel, diagnostics *diag.Diagnostics) {
	id := model.ID.ValueString()

	result, err := r.client.ReadInvestigation(id)
	if err != nil {
		diagnostics.AddError("Error reading investigation", err.Error())
		return
	}

	// The ES v2 filter endpoint returns a JSON object. The matching investigations
	// may appear directly in a top-level list or under a key. Try common patterns.
	data, extractErr := extractESv2ListEntry(result)
	if extractErr != nil {
		diagnostics.AddError(
			"Error parsing investigation response",
			fmt.Sprintf("Could not parse ES v2 response for investigation %s: %s", id, extractErr.Error()),
		)
		return
	}

	model.ID = types.StringValue(sdk.ParseString(data, "_key"))
	model.Name = types.StringValue(sdk.ParseString(data, "title"))

	if desc := sdk.ParseString(data, "description"); desc != "" {
		model.Description = types.StringValue(desc)
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	if status := sdk.ParseString(data, "status"); status != "" {
		model.Status = types.StringValue(status)
	} else if model.Status.IsNull() {
		model.Status = types.StringNull()
	}

	if assignee := sdk.ParseString(data, "assignee"); assignee != "" {
		model.Assignee = types.StringValue(assignee)
	} else if model.Assignee.IsNull() {
		model.Assignee = types.StringNull()
	}

	if priority := sdk.ParseString(data, "urgency"); priority != "" {
		model.Priority = types.StringValue(priority)
	} else if model.Priority.IsNull() {
		model.Priority = types.StringNull()
	}

	// Parse tags from the response.
	if rawTags, ok := data["tags"]; ok && rawTags != nil {
		if tagSlice, ok := rawTags.([]interface{}); ok {
			tagStrings := make([]string, len(tagSlice))
			for i, t := range tagSlice {
				tagStrings[i] = fmt.Sprintf("%v", t)
			}
			tagList, d := types.ListValueFrom(ctx, types.StringType, tagStrings)
			diagnostics.Append(d...)
			model.Tags = tagList
		} else {
			model.Tags = types.ListNull(types.StringType)
		}
	} else if model.Tags.IsNull() {
		model.Tags = types.ListNull(types.StringType)
	}

	if createdTime := sdk.ParseString(data, "create_time"); createdTime != "" {
		model.CreatedTime = types.StringValue(createdTime)
	} else {
		model.CreatedTime = types.StringNull()
	}

	if modifiedTime := sdk.ParseString(data, "modify_time"); modifiedTime != "" {
		model.ModifiedTime = types.StringValue(modifiedTime)
	} else {
		model.ModifiedTime = types.StringNull()
	}
}

// extractESv2ListEntry extracts the first object from an ES v2 list-style response.
// The ES v2 API returns either:
//   - A JSON array at the top level (the response map wraps it under a key)
//   - A JSON object with a "data" or "items" key containing the array
//   - The object itself if it was a direct single-object response
func extractESv2ListEntry(result map[string]interface{}) (map[string]interface{}, error) {
	// If the response itself has _key, it is the object directly.
	if _, ok := result["_key"]; ok {
		return result, nil
	}

	// Try common list wrapper keys used by ES v2.
	for _, key := range []string{"data", "items", "results", "entry"} {
		if raw, ok := result[key]; ok {
			if items, ok := raw.([]interface{}); ok && len(items) > 0 {
				if item, ok := items[0].(map[string]interface{}); ok {
					return item, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no investigation entry found in response")
}
