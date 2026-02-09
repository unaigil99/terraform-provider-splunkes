package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Agentric/terraform-provider-splunkes/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &SavedSearchResource{}
	_ resource.ResourceWithConfigure   = &SavedSearchResource{}
	_ resource.ResourceWithImportState = &SavedSearchResource{}
)

// NewSavedSearchResource is a constructor that returns a new SavedSearchResource.
func NewSavedSearchResource() resource.Resource {
	return &SavedSearchResource{}
}

// SavedSearchResource implements the splunkes_saved_search resource.
type SavedSearchResource struct {
	client *sdk.SplunkClient
}

// SavedSearchResourceModel describes the resource data model.
type SavedSearchResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Search               types.String `tfsdk:"search"`
	Description          types.String `tfsdk:"description"`
	App                  types.String `tfsdk:"app"`
	Owner                types.String `tfsdk:"owner"`
	Disabled             types.Bool   `tfsdk:"disabled"`
	IsScheduled          types.Bool   `tfsdk:"is_scheduled"`
	CronSchedule         types.String `tfsdk:"cron_schedule"`
	DispatchEarliestTime types.String `tfsdk:"dispatch_earliest_time"`
	DispatchLatestTime   types.String `tfsdk:"dispatch_latest_time"`
	Actions              types.String `tfsdk:"actions"`
	ActionEmailTo        types.String `tfsdk:"action_email_to"`
	ActionEmailSubject   types.String `tfsdk:"action_email_subject"`
	AlertType            types.String `tfsdk:"alert_type"`
	AlertComparator      types.String `tfsdk:"alert_comparator"`
	AlertThreshold       types.String `tfsdk:"alert_threshold"`
	AlertSeverity        types.Int64  `tfsdk:"alert_severity"`
}

func (r *SavedSearchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saved_search"
}

func (r *SavedSearchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Splunk saved search.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this saved search resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the saved search.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"search": schema.StringAttribute{
				Description: "The SPL search query.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the saved search.",
				Optional:    true,
			},
			"app": schema.StringAttribute{
				Description: "The Splunk app context. Defaults to search.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("search"),
			},
			"owner": schema.StringAttribute{
				Description: "The Splunk user owner. Defaults to nobody.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("nobody"),
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the saved search is disabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_scheduled": schema.BoolAttribute{
				Description: "Whether the saved search is scheduled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"cron_schedule": schema.StringAttribute{
				Description: "The cron schedule for the saved search.",
				Optional:    true,
			},
			"dispatch_earliest_time": schema.StringAttribute{
				Description: "The earliest time for the search dispatch.",
				Optional:    true,
			},
			"dispatch_latest_time": schema.StringAttribute{
				Description: "The latest time for the search dispatch.",
				Optional:    true,
			},
			"actions": schema.StringAttribute{
				Description: "Comma-separated list of action names to enable.",
				Optional:    true,
			},
			"action_email_to": schema.StringAttribute{
				Description: "Email address(es) to send results to.",
				Optional:    true,
			},
			"action_email_subject": schema.StringAttribute{
				Description: "Subject line for email actions.",
				Optional:    true,
			},
			"alert_type": schema.StringAttribute{
				Description: "The alert type (always, number of events, etc.).",
				Optional:    true,
			},
			"alert_comparator": schema.StringAttribute{
				Description: "The alert comparator (greater than, less than, etc.).",
				Optional:    true,
			},
			"alert_threshold": schema.StringAttribute{
				Description: "The alert threshold value.",
				Optional:    true,
			},
			"alert_severity": schema.Int64Attribute{
				Description: "The alert severity level (1-5). Defaults to 3.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3),
			},
		},
	}
}

func (r *SavedSearchResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SavedSearchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SavedSearchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()

	params := r.buildParams(&plan, true)

	tflog.Debug(ctx, "Creating saved search", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.CreateSavedSearch(owner, app, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating saved search", err.Error())
		return
	}

	// Read back the created resource to populate all fields.
	readResp, err := r.client.ReadSavedSearch(owner, app, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading saved search after creation", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SavedSearchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SavedSearchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading saved search", map[string]interface{}{"name": name})

	readResp, err := r.client.ReadSavedSearch(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading saved search", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SavedSearchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SavedSearchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()
	name := plan.Name.ValueString()

	params := r.buildParams(&plan, false)

	tflog.Debug(ctx, "Updating saved search", map[string]interface{}{"name": name})

	_, err := r.client.UpdateSavedSearch(owner, app, name, params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating saved search", err.Error())
		return
	}

	// Read back the updated resource.
	readResp, err := r.client.ReadSavedSearch(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading saved search after update", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SavedSearchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SavedSearchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting saved search", map[string]interface{}{"name": name})

	err := r.client.DeleteSavedSearch(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting saved search", err.Error())
		return
	}
}

func (r *SavedSearchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), "nobody")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), "search")...)
}

// buildParams constructs url.Values from the SavedSearchResourceModel.
func (r *SavedSearchResource) buildParams(model *SavedSearchResourceModel, isCreate bool) url.Values {
	params := url.Values{}

	if isCreate {
		params.Set("name", model.Name.ValueString())
	}

	params.Set("search", model.Search.ValueString())

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		params.Set("description", model.Description.ValueString())
	}
	if !model.Disabled.IsNull() && !model.Disabled.IsUnknown() {
		if model.Disabled.ValueBool() {
			params.Set("disabled", "1")
		} else {
			params.Set("disabled", "0")
		}
	}
	if !model.IsScheduled.IsNull() && !model.IsScheduled.IsUnknown() {
		if model.IsScheduled.ValueBool() {
			params.Set("is_scheduled", "1")
		} else {
			params.Set("is_scheduled", "0")
		}
	}
	if !model.CronSchedule.IsNull() && !model.CronSchedule.IsUnknown() {
		params.Set("cron_schedule", model.CronSchedule.ValueString())
	}
	if !model.DispatchEarliestTime.IsNull() && !model.DispatchEarliestTime.IsUnknown() {
		params.Set("dispatch.earliest_time", model.DispatchEarliestTime.ValueString())
	}
	if !model.DispatchLatestTime.IsNull() && !model.DispatchLatestTime.IsUnknown() {
		params.Set("dispatch.latest_time", model.DispatchLatestTime.ValueString())
	}
	if !model.Actions.IsNull() && !model.Actions.IsUnknown() {
		params.Set("actions", model.Actions.ValueString())
	}
	if !model.ActionEmailTo.IsNull() && !model.ActionEmailTo.IsUnknown() {
		params.Set("action.email.to", model.ActionEmailTo.ValueString())
	}
	if !model.ActionEmailSubject.IsNull() && !model.ActionEmailSubject.IsUnknown() {
		params.Set("action.email.subject", model.ActionEmailSubject.ValueString())
	}
	if !model.AlertType.IsNull() && !model.AlertType.IsUnknown() {
		params.Set("alert_type", model.AlertType.ValueString())
	}
	if !model.AlertComparator.IsNull() && !model.AlertComparator.IsUnknown() {
		params.Set("alert_comparator", model.AlertComparator.ValueString())
	}
	if !model.AlertThreshold.IsNull() && !model.AlertThreshold.IsUnknown() {
		params.Set("alert_threshold", model.AlertThreshold.ValueString())
	}
	if !model.AlertSeverity.IsNull() && !model.AlertSeverity.IsUnknown() {
		params.Set("alert.severity", fmt.Sprintf("%d", model.AlertSeverity.ValueInt64()))
	}

	return params
}

// mapResponseToModel maps a Splunk API response to the SavedSearchResourceModel.
func (r *SavedSearchResource) mapResponseToModel(response map[string]interface{}, model *SavedSearchResourceModel) {
	content, err := sdk.GetEntryContent(response)
	if err != nil {
		return
	}

	// Extract name from entry-level fields.
	entries, _ := response["entry"].([]interface{})
	if len(entries) > 0 {
		entry, _ := entries[0].(map[string]interface{})
		if name, ok := entry["name"].(string); ok {
			model.Name = types.StringValue(name)
		}
	}

	model.ID = types.StringValue(model.Name.ValueString())
	model.Search = types.StringValue(sdk.ParseString(content, "search"))

	if desc := sdk.ParseString(content, "description"); desc != "" {
		model.Description = types.StringValue(desc)
	}

	model.Disabled = types.BoolValue(sdk.ParseBool(content, "disabled"))
	model.IsScheduled = types.BoolValue(sdk.ParseBool(content, "is_scheduled"))

	if cronSchedule := sdk.ParseString(content, "cron_schedule"); cronSchedule != "" {
		model.CronSchedule = types.StringValue(cronSchedule)
	}
	if det := sdk.ParseString(content, "dispatch.earliest_time"); det != "" {
		model.DispatchEarliestTime = types.StringValue(det)
	}
	if dlt := sdk.ParseString(content, "dispatch.latest_time"); dlt != "" {
		model.DispatchLatestTime = types.StringValue(dlt)
	}
	if actions := sdk.ParseString(content, "actions"); actions != "" {
		model.Actions = types.StringValue(actions)
	}
	if emailTo := sdk.ParseString(content, "action.email.to"); emailTo != "" {
		model.ActionEmailTo = types.StringValue(emailTo)
	}
	if emailSubject := sdk.ParseString(content, "action.email.subject"); emailSubject != "" {
		model.ActionEmailSubject = types.StringValue(emailSubject)
	}
	if alertType := sdk.ParseString(content, "alert_type"); alertType != "" {
		model.AlertType = types.StringValue(alertType)
	}
	if alertComparator := sdk.ParseString(content, "alert_comparator"); alertComparator != "" {
		model.AlertComparator = types.StringValue(alertComparator)
	}
	if alertThreshold := sdk.ParseString(content, "alert_threshold"); alertThreshold != "" {
		model.AlertThreshold = types.StringValue(alertThreshold)
	}

	model.AlertSeverity = types.Int64Value(sdk.ParseInt(content, "alert.severity"))

	// Extract app and owner from ACL.
	acl, aclErr := sdk.GetEntryACL(response)
	if aclErr == nil {
		if appVal := sdk.ParseString(acl, "app"); appVal != "" {
			model.App = types.StringValue(appVal)
		}
		if ownerVal := sdk.ParseString(acl, "owner"); ownerVal != "" {
			model.Owner = types.StringValue(ownerVal)
		}
	}
}
