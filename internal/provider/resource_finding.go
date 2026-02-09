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
	_ resource.Resource                = &findingResource{}
	_ resource.ResourceWithConfigure   = &findingResource{}
	_ resource.ResourceWithImportState = &findingResource{}
)

// NewFindingResource returns a new resource.Resource for ES findings.
func NewFindingResource() resource.Resource {
	return &findingResource{}
}

type findingResource struct {
	client *sdk.SplunkClient
}

type findingResourceModel struct {
	ID              types.String `tfsdk:"id"`
	RuleTitle       types.String `tfsdk:"rule_title"`
	RuleDescription types.String `tfsdk:"rule_description"`
	SecurityDomain  types.String `tfsdk:"security_domain"`
	RiskScore       types.Int64  `tfsdk:"risk_score"`
	RiskObject      types.String `tfsdk:"risk_object"`
	RiskObjectType  types.String `tfsdk:"risk_object_type"`
	Severity        types.String `tfsdk:"severity"`
	Owner           types.String `tfsdk:"owner"`
	Status          types.String `tfsdk:"status"`
	Urgency         types.String `tfsdk:"urgency"`
	Disposition     types.String `tfsdk:"disposition"`
	Time            types.String `tfsdk:"time"`
}

func (r *findingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_finding"
}

func (r *findingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a finding (notable event) in Splunk Enterprise Security (ES v2 API). " +
			"Findings are created via the API but cannot be updated or deleted through the ES v2 API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The finding ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rule_title": schema.StringAttribute{
				Description: "The title of the detection rule or finding.",
				Required:    true,
			},
			"rule_description": schema.StringAttribute{
				Description: "A description of the detection rule.",
				Optional:    true,
			},
			"security_domain": schema.StringAttribute{
				Description: "The security domain: access, endpoint, network, identity, threat, or audit.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("access", "endpoint", "network", "identity", "threat", "audit"),
				},
			},
			"risk_score": schema.Int64Attribute{
				Description: "The risk score for this finding.",
				Required:    true,
			},
			"risk_object": schema.StringAttribute{
				Description: "The entity (user, host, etc.) that was found risky.",
				Required:    true,
			},
			"risk_object_type": schema.StringAttribute{
				Description: "The type of the risk object: user, system, or other.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("user", "system", "other"),
				},
			},
			"severity": schema.StringAttribute{
				Description: "Finding severity: informational, low, medium, high, or critical.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("informational", "low", "medium", "high", "critical"),
				},
			},
			"owner": schema.StringAttribute{
				Description: "The assigned owner of the finding.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Finding status: new, in_progress, pending, resolved, or closed.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("new", "in_progress", "pending", "resolved", "closed"),
				},
			},
			"urgency": schema.StringAttribute{
				Description: "The urgency level of the finding.",
				Optional:    true,
			},
			"disposition": schema.StringAttribute{
				Description: "The disposition of the finding.",
				Optional:    true,
			},
			"time": schema.StringAttribute{
				Description: "The time the finding was created.",
				Computed:    true,
			},
		},
	}
}

func (r *findingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *findingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan findingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"rule_title":       plan.RuleTitle.ValueString(),
		"security_domain":  plan.SecurityDomain.ValueString(),
		"risk_score":       plan.RiskScore.ValueInt64(),
		"risk_object":      plan.RiskObject.ValueString(),
		"risk_object_type": plan.RiskObjectType.ValueString(),
	}

	if !plan.RuleDescription.IsNull() && !plan.RuleDescription.IsUnknown() {
		body["rule_description"] = plan.RuleDescription.ValueString()
	}
	if !plan.Severity.IsNull() && !plan.Severity.IsUnknown() {
		body["severity"] = plan.Severity.ValueString()
	}
	if !plan.Owner.IsNull() && !plan.Owner.IsUnknown() {
		body["owner"] = plan.Owner.ValueString()
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		body["status"] = plan.Status.ValueString()
	}
	if !plan.Urgency.IsNull() && !plan.Urgency.IsUnknown() {
		body["urgency"] = plan.Urgency.ValueString()
	}
	if !plan.Disposition.IsNull() && !plan.Disposition.IsUnknown() {
		body["disposition"] = plan.Disposition.ValueString()
	}

	tflog.Debug(ctx, "Creating ES finding", map[string]interface{}{
		"rule_title": plan.RuleTitle.ValueString(),
	})

	result, err := r.client.CreateFinding(body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating finding", err.Error())
		return
	}

	// ES v2 returns _key or id directly in the response JSON.
	id := sdk.ParseString(result, "_key")
	if id == "" {
		id = sdk.ParseString(result, "id")
	}
	if id == "" {
		id = sdk.ParseString(result, "event_id")
	}
	if id == "" {
		resp.Diagnostics.AddError(
			"Error creating finding",
			"No _key, id, or event_id returned in the API response.",
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

func (r *findingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state findingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading ES finding", map[string]interface{}{"id": state.ID.ValueString()})

	r.refreshState(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *findingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Findings in ES v2 are NOT updatable via the API.
	resp.Diagnostics.AddWarning(
		"Findings cannot be updated",
		"The Splunk ES v2 API does not support updating findings. "+
			"The finding will retain its original values. "+
			"To apply changes, destroy and recreate the resource.",
	)

	// Persist the planned state so Terraform does not lose track of the resource.
	var state findingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *findingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state findingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Removing finding from state (no DELETE endpoint)", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Findings in ES v2 are NOT deletable. Remove from Terraform state only.
	resp.Diagnostics.AddWarning(
		"Finding not deleted from Splunk ES",
		fmt.Sprintf(
			"The Splunk ES v2 API does not support deleting findings. "+
				"Finding %s has been removed from Terraform state but still exists in Splunk ES.",
			state.ID.ValueString(),
		),
	)
}

func (r *findingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// refreshState reads the finding from the ES v2 API and updates the model.
// The ES v2 ReadFinding endpoint returns the finding object directly.
func (r *findingResource) refreshState(_ context.Context, model *findingResourceModel, diagnostics *diag.Diagnostics) {
	id := model.ID.ValueString()

	result, err := r.client.ReadFinding(id)
	if err != nil {
		diagnostics.AddError("Error reading finding", err.Error())
		return
	}

	// ES v2 ReadFinding returns the finding JSON directly (not wrapped).
	// If there is a wrapper, try to unwrap it.
	data := result
	if _, ok := data["_key"]; !ok {
		if extracted, extractErr := extractESv2ListEntry(data); extractErr == nil {
			data = extracted
		}
	}

	if k := sdk.ParseString(data, "_key"); k != "" {
		model.ID = types.StringValue(k)
	} else if k := sdk.ParseString(data, "id"); k != "" {
		model.ID = types.StringValue(k)
	} else if k := sdk.ParseString(data, "event_id"); k != "" {
		model.ID = types.StringValue(k)
	}

	if v := sdk.ParseString(data, "rule_title"); v != "" {
		model.RuleTitle = types.StringValue(v)
	}

	if v := sdk.ParseString(data, "rule_description"); v != "" {
		model.RuleDescription = types.StringValue(v)
	} else if model.RuleDescription.IsNull() {
		model.RuleDescription = types.StringNull()
	}

	if v := sdk.ParseString(data, "security_domain"); v != "" {
		model.SecurityDomain = types.StringValue(v)
	}

	model.RiskScore = types.Int64Value(sdk.ParseInt(data, "risk_score"))

	if v := sdk.ParseString(data, "risk_object"); v != "" {
		model.RiskObject = types.StringValue(v)
	}

	if v := sdk.ParseString(data, "risk_object_type"); v != "" {
		model.RiskObjectType = types.StringValue(v)
	}

	if v := sdk.ParseString(data, "severity"); v != "" {
		model.Severity = types.StringValue(v)
	} else if model.Severity.IsNull() {
		model.Severity = types.StringNull()
	}

	if v := sdk.ParseString(data, "owner"); v != "" {
		model.Owner = types.StringValue(v)
	} else if model.Owner.IsNull() {
		model.Owner = types.StringNull()
	}

	if v := sdk.ParseString(data, "status"); v != "" {
		model.Status = types.StringValue(v)
	} else if model.Status.IsNull() {
		model.Status = types.StringNull()
	}

	if v := sdk.ParseString(data, "urgency"); v != "" {
		model.Urgency = types.StringValue(v)
	} else if model.Urgency.IsNull() {
		model.Urgency = types.StringNull()
	}

	if v := sdk.ParseString(data, "disposition"); v != "" {
		model.Disposition = types.StringValue(v)
	} else if model.Disposition.IsNull() {
		model.Disposition = types.StringNull()
	}

	if v := sdk.ParseString(data, "time"); v != "" {
		model.Time = types.StringValue(v)
	} else if v := sdk.ParseString(data, "_time"); v != "" {
		model.Time = types.StringValue(v)
	} else {
		model.Time = types.StringNull()
	}
}
