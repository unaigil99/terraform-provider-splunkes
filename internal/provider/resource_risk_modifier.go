package provider

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &riskModifierResource{}
	_ resource.ResourceWithConfigure   = &riskModifierResource{}
	_ resource.ResourceWithImportState = &riskModifierResource{}
)

// NewRiskModifierResource returns a new resource.Resource for ES risk modifiers.
func NewRiskModifierResource() resource.Resource {
	return &riskModifierResource{}
}

type riskModifierResource struct {
	client *sdk.SplunkClient
}

type riskModifierResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Entity       types.String `tfsdk:"entity"`
	EntityType   types.String `tfsdk:"entity_type"`
	RiskModifier types.Int64  `tfsdk:"risk_modifier"`
	Description  types.String `tfsdk:"description"`
}

func (r *riskModifierResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_risk_modifier"
}

func (r *riskModifierResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a risk modifier for an entity in Splunk Enterprise Security (ES v2 API). " +
			"Risk modifiers adjust the risk score of an entity (user or system). " +
			"Note: risk modifiers cannot be deleted through the ES v2 API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The resource identifier (entity/entity_type).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entity": schema.StringAttribute{
				Description: "The risk entity name (user, host, etc.).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entity_type": schema.StringAttribute{
				Description: "The type of the entity: user or system.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("user", "system"),
				},
			},
			"risk_modifier": schema.Int64Attribute{
				Description: "The risk score modifier value.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The reason for the risk modifier.",
				Optional:    true,
			},
		},
	}
}

func (r *riskModifierResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *riskModifierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan riskModifierResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entity := plan.Entity.ValueString()
	entityType := plan.EntityType.ValueString()

	body := map[string]interface{}{
		"entity_type":   entityType,
		"risk_modifier": plan.RiskModifier.ValueInt64(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}

	tflog.Debug(ctx, "Creating risk modifier", map[string]interface{}{
		"entity":      entity,
		"entity_type": entityType,
	})

	_, err := r.client.AddRiskModifier(entity, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating risk modifier", err.Error())
		return
	}

	// Use entity/entity_type as the composite ID.
	plan.ID = types.StringValue(entity + "/" + entityType)

	// Read back the current state.
	r.refreshState(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *riskModifierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state riskModifierResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading risk modifier", map[string]interface{}{
		"entity":      state.Entity.ValueString(),
		"entity_type": state.EntityType.ValueString(),
	})

	r.refreshState(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *riskModifierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan riskModifierResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state riskModifierResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entity := state.Entity.ValueString()
	entityType := plan.EntityType.ValueString()

	body := map[string]interface{}{
		"entity_type":   entityType,
		"risk_modifier": plan.RiskModifier.ValueInt64(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}

	tflog.Debug(ctx, "Updating risk modifier", map[string]interface{}{
		"entity":      entity,
		"entity_type": entityType,
	})

	// AddRiskModifier is used for both create and update (additive operation).
	_, err := r.client.AddRiskModifier(entity, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating risk modifier", err.Error())
		return
	}

	plan.ID = types.StringValue(entity + "/" + entityType)
	plan.Entity = types.StringValue(entity)

	r.refreshState(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *riskModifierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state riskModifierResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Removing risk modifier from state (no DELETE endpoint)", map[string]interface{}{
		"entity":      state.Entity.ValueString(),
		"entity_type": state.EntityType.ValueString(),
	})

	// Risk modifiers in ES v2 are NOT deletable. Remove from Terraform state only.
	resp.Diagnostics.AddWarning(
		"Risk modifier not deleted from Splunk ES",
		fmt.Sprintf(
			"The Splunk ES v2 API does not support deleting risk modifiers. "+
				"Risk modifier for entity %q (%s) has been removed from Terraform state "+
				"but still exists in Splunk ES.",
			state.Entity.ValueString(),
			state.EntityType.ValueString(),
		),
	)
}

// ImportState handles "entity/entity_type" composite import IDs.
func (r *riskModifierResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the format 'entity/entity_type', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entity"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entity_type"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// refreshState reads the risk score for the entity from the ES v2 API and updates the model.
// The ES v2 ReadRiskScore endpoint returns the risk score object directly.
func (r *riskModifierResource) refreshState(_ context.Context, model *riskModifierResourceModel, diagnostics *diag.Diagnostics) {
	entity := model.Entity.ValueString()
	entityType := model.EntityType.ValueString()

	result, err := r.client.ReadRiskScore(entity, entityType)
	if err != nil {
		diagnostics.AddError("Error reading risk score", err.Error())
		return
	}

	// ES v2 ReadRiskScore returns the risk info directly (not wrapped).
	// If there is a wrapper, try to unwrap it.
	data := result
	if _, ok := data["risk_score"]; !ok {
		if extracted, extractErr := extractESv2ListEntry(data); extractErr == nil {
			data = extracted
		}
	}

	// The entity and entity_type are set from the plan/state, not from the response.
	model.ID = types.StringValue(entity + "/" + entityType)

	// Extract the risk modifier value from the response.
	// The API may return the modifier under "risk_modifier" or "risk_score".
	if _, hasModifier := data["risk_modifier"]; hasModifier {
		model.RiskModifier = types.Int64Value(sdk.ParseInt(data, "risk_modifier"))
	} else if _, hasScore := data["risk_score"]; hasScore {
		// Fall back to risk_score if risk_modifier is not present.
		model.RiskModifier = types.Int64Value(sdk.ParseInt(data, "risk_score"))
	}

	if v := sdk.ParseString(data, "description"); v != "" {
		model.Description = types.StringValue(v)
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}
}
