package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Agentric/terraform-provider-splunkes/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &AnalyticStoryResource{}
	_ resource.ResourceWithConfigure   = &AnalyticStoryResource{}
	_ resource.ResourceWithImportState = &AnalyticStoryResource{}
)

// NewAnalyticStoryResource is a constructor that returns a new AnalyticStoryResource.
func NewAnalyticStoryResource() resource.Resource {
	return &AnalyticStoryResource{}
}

// AnalyticStoryResource implements the splunkes_analytic_story resource.
type AnalyticStoryResource struct {
	client *sdk.SplunkClient
}

// AnalyticStoryResourceModel describes the resource data model.
type AnalyticStoryResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	App                   types.String `tfsdk:"app"`
	Owner                 types.String `tfsdk:"owner"`
	Description           types.String `tfsdk:"description"`
	Category              types.List   `tfsdk:"category"`
	DataModels            types.List   `tfsdk:"data_models"`
	ProvidingTechnologies types.List   `tfsdk:"providing_technologies"`
	DetectionSearches     types.List   `tfsdk:"detection_searches"`
	InvestigativeSearches types.List   `tfsdk:"investigative_searches"`
	ContextualSearches    types.List   `tfsdk:"contextual_searches"`
}

func (r *AnalyticStoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_analytic_story"
}

func (r *AnalyticStoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Splunk ES analytic story.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this analytic story resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the analytic story.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app": schema.StringAttribute{
				Description: "The Splunk app context. Defaults to SplunkEnterpriseSecuritySuite.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("SplunkEnterpriseSecuritySuite"),
			},
			"owner": schema.StringAttribute{
				Description: "The Splunk user owner. Defaults to nobody.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("nobody"),
			},
			"description": schema.StringAttribute{
				Description: "A description of the analytic story.",
				Optional:    true,
			},
			"category": schema.ListAttribute{
				Description: "Categories for the analytic story (e.g., Adversary Tactics).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"data_models": schema.ListAttribute{
				Description: "Data models used by the analytic story (e.g., Endpoint, Network_Traffic).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"providing_technologies": schema.ListAttribute{
				Description: "Technologies providing data for the story (e.g., Sysmon, Carbon Black).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"detection_searches": schema.ListAttribute{
				Description: "Names of associated detection saved searches.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"investigative_searches": schema.ListAttribute{
				Description: "Names of associated investigative saved searches.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"contextual_searches": schema.ListAttribute{
				Description: "Names of associated contextual saved searches.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *AnalyticStoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AnalyticStoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AnalyticStoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()

	params := url.Values{}
	params.Set("name", plan.Name.ValueString())

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		params.Set("description", plan.Description.ValueString())
	}

	r.addListParam(ctx, params, "category", plan.Category, &resp.Diagnostics)
	r.addListParam(ctx, params, "data_models", plan.DataModels, &resp.Diagnostics)
	r.addListParam(ctx, params, "providing_technologies", plan.ProvidingTechnologies, &resp.Diagnostics)
	r.addListParam(ctx, params, "detection_searches", plan.DetectionSearches, &resp.Diagnostics)
	r.addListParam(ctx, params, "investigative_searches", plan.InvestigativeSearches, &resp.Diagnostics)
	r.addListParam(ctx, params, "contextual_searches", plan.ContextualSearches, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating analytic story", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.CreateAnalyticStory(owner, app, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating analytic story", err.Error())
		return
	}

	// Read back the created resource to populate all fields.
	readResp, err := r.client.ReadAnalyticStory(owner, app, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading analytic story after creation", err.Error())
		return
	}

	r.mapResponseToModel(ctx, readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AnalyticStoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AnalyticStoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading analytic story", map[string]interface{}{"name": name})

	readResp, err := r.client.ReadAnalyticStory(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading analytic story", err.Error())
		return
	}

	r.mapResponseToModel(ctx, readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AnalyticStoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AnalyticStoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()
	name := plan.Name.ValueString()

	params := url.Values{}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		params.Set("description", plan.Description.ValueString())
	}

	r.addListParam(ctx, params, "category", plan.Category, &resp.Diagnostics)
	r.addListParam(ctx, params, "data_models", plan.DataModels, &resp.Diagnostics)
	r.addListParam(ctx, params, "providing_technologies", plan.ProvidingTechnologies, &resp.Diagnostics)
	r.addListParam(ctx, params, "detection_searches", plan.DetectionSearches, &resp.Diagnostics)
	r.addListParam(ctx, params, "investigative_searches", plan.InvestigativeSearches, &resp.Diagnostics)
	r.addListParam(ctx, params, "contextual_searches", plan.ContextualSearches, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating analytic story", map[string]interface{}{"name": name})

	_, err := r.client.UpdateAnalyticStory(owner, app, name, params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating analytic story", err.Error())
		return
	}

	// Read back the updated resource.
	readResp, err := r.client.ReadAnalyticStory(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading analytic story after update", err.Error())
		return
	}

	r.mapResponseToModel(ctx, readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AnalyticStoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AnalyticStoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting analytic story", map[string]interface{}{"name": name})

	err := r.client.DeleteAnalyticStory(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting analytic story", err.Error())
		return
	}
}

func (r *AnalyticStoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), "nobody")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), "SplunkEnterpriseSecuritySuite")...)
}

// addListParam serializes a types.List to a JSON string and adds it to url.Values.
func (r *AnalyticStoryResource) addListParam(ctx context.Context, params url.Values, key string, list types.List, diags *diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return
	}

	var items []string
	diags.Append(list.ElementsAs(ctx, &items, false)...)
	if diags.HasError() {
		return
	}

	jsonBytes, err := json.Marshal(items)
	if err != nil {
		diags.AddError("Error serializing list field", fmt.Sprintf("Failed to serialize %s: %s", key, err.Error()))
		return
	}
	params.Set(key, string(jsonBytes))
}

// parseJSONListField parses a JSON-encoded list string from Splunk content into a types.List.
func parseJSONListField(ctx context.Context, content map[string]interface{}, key string) types.List {
	raw := sdk.ParseString(content, key)
	if raw == "" {
		return types.ListNull(types.StringType)
	}

	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		// If it is not valid JSON, treat it as a single-element list.
		items = []string{raw}
	}

	listVal, _ := types.ListValueFrom(ctx, types.StringType, items)
	return listVal
}

// mapResponseToModel maps a Splunk API response to the AnalyticStoryResourceModel.
func (r *AnalyticStoryResource) mapResponseToModel(ctx context.Context, response map[string]interface{}, model *AnalyticStoryResourceModel) {
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

	if desc := sdk.ParseString(content, "description"); desc != "" {
		model.Description = types.StringValue(desc)
	}

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

	// Parse JSON list fields from content.
	model.Category = parseJSONListField(ctx, content, "category")
	model.DataModels = parseJSONListField(ctx, content, "data_models")
	model.ProvidingTechnologies = parseJSONListField(ctx, content, "providing_technologies")
	model.DetectionSearches = parseJSONListField(ctx, content, "detection_searches")
	model.InvestigativeSearches = parseJSONListField(ctx, content, "investigative_searches")
	model.ContextualSearches = parseJSONListField(ctx, content, "contextual_searches")
}
