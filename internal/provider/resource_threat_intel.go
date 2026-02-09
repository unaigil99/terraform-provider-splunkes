package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Agentric/terraform-provider-splunkes/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ThreatIntelResource{}
	_ resource.ResourceWithConfigure   = &ThreatIntelResource{}
	_ resource.ResourceWithImportState = &ThreatIntelResource{}
)

// NewThreatIntelResource is a constructor that returns a new ThreatIntelResource.
func NewThreatIntelResource() resource.Resource {
	return &ThreatIntelResource{}
}

// ThreatIntelResource implements the splunkes_threat_intel resource.
type ThreatIntelResource struct {
	client *sdk.SplunkClient
}

// ThreatIntelResourceModel describes the resource data model.
type ThreatIntelResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Collection  types.String `tfsdk:"collection"`
	Description types.String `tfsdk:"description"`
	Fields      types.Map    `tfsdk:"fields"`
}

func (r *ThreatIntelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_threat_intel"
}

func (r *ThreatIntelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Splunk ES threat intelligence item.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The item key (_key) for this threat intel item.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"collection": schema.StringAttribute{
				Description: "The threat intel collection name (ip_intel, domain_intel, file_intel, email_intel, etc.).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description for this threat intel item.",
				Optional:    true,
			},
			"fields": schema.MapAttribute{
				Description: "Key-value pairs for the intel item (ip, domain, weight, etc.).",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *ThreatIntelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ThreatIntelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ThreatIntelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	collection := plan.Collection.ValueString()

	params := r.buildParams(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating threat intel item", map[string]interface{}{"collection": collection})

	createResp, err := r.client.CreateThreatIntelItem(collection, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating threat intel item", err.Error())
		return
	}

	// Extract the _key from the response.
	key := r.extractKey(createResp)
	if key == "" {
		resp.Diagnostics.AddError("Error creating threat intel item", "No _key returned in response")
		return
	}

	plan.ID = types.StringValue(key)

	// Read back the created resource.
	readResp, err := r.client.ReadThreatIntelItem(collection, key)
	if err != nil {
		resp.Diagnostics.AddError("Error reading threat intel item after creation", err.Error())
		return
	}

	r.mapResponseToModel(ctx, readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ThreatIntelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ThreatIntelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	collection := state.Collection.ValueString()
	key := state.ID.ValueString()

	tflog.Debug(ctx, "Reading threat intel item", map[string]interface{}{"collection": collection, "key": key})

	readResp, err := r.client.ReadThreatIntelItem(collection, key)
	if err != nil {
		resp.Diagnostics.AddError("Error reading threat intel item", err.Error())
		return
	}

	r.mapResponseToModel(ctx, readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ThreatIntelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ThreatIntelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	collection := plan.Collection.ValueString()
	key := plan.ID.ValueString()

	params := r.buildParams(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating threat intel item", map[string]interface{}{"collection": collection, "key": key})

	_, err := r.client.UpdateThreatIntelItem(collection, key, params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating threat intel item", err.Error())
		return
	}

	// Read back the updated resource.
	readResp, err := r.client.ReadThreatIntelItem(collection, key)
	if err != nil {
		resp.Diagnostics.AddError("Error reading threat intel item after update", err.Error())
		return
	}

	r.mapResponseToModel(ctx, readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ThreatIntelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ThreatIntelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	collection := state.Collection.ValueString()
	key := state.ID.ValueString()

	tflog.Debug(ctx, "Deleting threat intel item", map[string]interface{}{"collection": collection, "key": key})

	err := r.client.DeleteThreatIntelItem(collection, key)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting threat intel item", err.Error())
		return
	}
}

func (r *ThreatIntelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID format: collection/key, got: %s", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("collection"), parts[0])...)
}

// buildParams constructs url.Values from the model fields.
func (r *ThreatIntelResource) buildParams(ctx context.Context, model *ThreatIntelResourceModel, diags *diag.Diagnostics) url.Values {
	params := url.Values{}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		params.Set("description", model.Description.ValueString())
	}

	if !model.Fields.IsNull() && !model.Fields.IsUnknown() {
		fields := make(map[string]string)
		diags.Append(model.Fields.ElementsAs(ctx, &fields, false)...)
		for k, v := range fields {
			params.Set(k, v)
		}
	}

	return params
}

// extractKey extracts the _key from a Splunk threat intel API response.
func (r *ThreatIntelResource) extractKey(response map[string]interface{}) string {
	// The threat intel API may return _key at the top level or within entry content.
	if key, ok := response["_key"].(string); ok {
		return key
	}

	// Try standard entry format.
	content, err := sdk.GetEntryContent(response)
	if err == nil {
		if key := sdk.ParseString(content, "_key"); key != "" {
			return key
		}
	}

	return ""
}

// mapResponseToModel maps a Splunk API response to the ThreatIntelResourceModel.
func (r *ThreatIntelResource) mapResponseToModel(ctx context.Context, response map[string]interface{}, model *ThreatIntelResourceModel) {
	// Threat intel API may return data in entry content or directly.
	content, err := sdk.GetEntryContent(response)
	if err != nil {
		// Fallback: response may be the content directly.
		content = response
	}

	if key := sdk.ParseString(content, "_key"); key != "" {
		model.ID = types.StringValue(key)
	}

	if desc := sdk.ParseString(content, "description"); desc != "" {
		model.Description = types.StringValue(desc)
	}

	// Rebuild fields map from content, excluding internal keys.
	fields := make(map[string]string)
	internalKeys := map[string]bool{
		"_key": true, "_user": true, "_time": true,
		"description": true, "threat_key": true,
	}
	for k, v := range content {
		if internalKeys[k] {
			continue
		}
		if strVal, ok := v.(string); ok {
			fields[k] = strVal
		} else if v != nil {
			fields[k] = fmt.Sprintf("%v", v)
		}
	}

	if len(fields) > 0 {
		mapVal, _ := types.MapValueFrom(ctx, types.StringType, fields)
		model.Fields = mapVal
	}
}
