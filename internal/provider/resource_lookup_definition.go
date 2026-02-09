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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &LookupDefinitionResource{}
	_ resource.ResourceWithConfigure   = &LookupDefinitionResource{}
	_ resource.ResourceWithImportState = &LookupDefinitionResource{}
)

// NewLookupDefinitionResource is a constructor that returns a new LookupDefinitionResource.
func NewLookupDefinitionResource() resource.Resource {
	return &LookupDefinitionResource{}
}

// LookupDefinitionResource implements the splunkes_lookup_definition resource.
type LookupDefinitionResource struct {
	client *sdk.SplunkClient
}

// LookupDefinitionResourceModel describes the resource data model.
type LookupDefinitionResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Filename           types.String `tfsdk:"filename"`
	App                types.String `tfsdk:"app"`
	Owner              types.String `tfsdk:"owner"`
	ExternalType       types.String `tfsdk:"external_type"`
	Collection         types.String `tfsdk:"collection"`
	FieldsList         types.String `tfsdk:"fields_list"`
	MaxMatches         types.Int64  `tfsdk:"max_matches"`
	MinMatches         types.Int64  `tfsdk:"min_matches"`
	DefaultMatch       types.String `tfsdk:"default_match"`
	CaseSensitiveMatch types.Bool   `tfsdk:"case_sensitive_match"`
	MatchType          types.String `tfsdk:"match_type"`
}

func (r *LookupDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_definition"
}

func (r *LookupDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Splunk lookup definition (transforms lookup).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this lookup definition resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the lookup definition.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"filename": schema.StringAttribute{
				Description: "The CSV lookup file name (e.g., my_lookup.csv).",
				Optional:    true,
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
			"external_type": schema.StringAttribute{
				Description: "The external type of the lookup. Use \"kvstore\" for KV store backed lookups.",
				Optional:    true,
			},
			"collection": schema.StringAttribute{
				Description: "The KV store collection name for KV store backed lookups.",
				Optional:    true,
			},
			"fields_list": schema.StringAttribute{
				Description: "A comma-separated list of fields in the lookup.",
				Optional:    true,
			},
			"max_matches": schema.Int64Attribute{
				Description: "The maximum number of matches for each input lookup value.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"min_matches": schema.Int64Attribute{
				Description: "The minimum number of matches for each input lookup value.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"default_match": schema.StringAttribute{
				Description: "The default value to use if fewer than min_matches matches are found.",
				Optional:    true,
			},
			"case_sensitive_match": schema.BoolAttribute{
				Description: "Whether the lookup match is case-sensitive. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"match_type": schema.StringAttribute{
				Description: "The match type for the lookup (e.g., WILDCARD(field), CIDR(field)).",
				Optional:    true,
			},
		},
	}
}

func (r *LookupDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LookupDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LookupDefinitionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()

	params := url.Values{}
	params.Set("name", plan.Name.ValueString())

	if !plan.Filename.IsNull() && !plan.Filename.IsUnknown() {
		params.Set("filename", plan.Filename.ValueString())
	}
	if !plan.ExternalType.IsNull() && !plan.ExternalType.IsUnknown() {
		params.Set("external_type", plan.ExternalType.ValueString())
	}
	if !plan.Collection.IsNull() && !plan.Collection.IsUnknown() {
		params.Set("collection", plan.Collection.ValueString())
	}
	if !plan.FieldsList.IsNull() && !plan.FieldsList.IsUnknown() {
		params.Set("fields_list", plan.FieldsList.ValueString())
	}
	if !plan.MaxMatches.IsNull() && !plan.MaxMatches.IsUnknown() {
		params.Set("max_matches", fmt.Sprintf("%d", plan.MaxMatches.ValueInt64()))
	}
	if !plan.MinMatches.IsNull() && !plan.MinMatches.IsUnknown() {
		params.Set("min_matches", fmt.Sprintf("%d", plan.MinMatches.ValueInt64()))
	}
	if !plan.DefaultMatch.IsNull() && !plan.DefaultMatch.IsUnknown() {
		params.Set("default_match", plan.DefaultMatch.ValueString())
	}
	if !plan.CaseSensitiveMatch.IsNull() && !plan.CaseSensitiveMatch.IsUnknown() {
		if plan.CaseSensitiveMatch.ValueBool() {
			params.Set("case_sensitive_match", "1")
		} else {
			params.Set("case_sensitive_match", "0")
		}
	}
	if !plan.MatchType.IsNull() && !plan.MatchType.IsUnknown() {
		params.Set("match_type", plan.MatchType.ValueString())
	}

	tflog.Debug(ctx, "Creating lookup definition", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.CreateLookupDefinition(owner, app, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lookup definition", err.Error())
		return
	}

	// Read back the created resource to populate all fields.
	readResp, err := r.client.ReadLookupDefinition(owner, app, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading lookup definition after creation", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LookupDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LookupDefinitionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading lookup definition", map[string]interface{}{"name": name})

	readResp, err := r.client.ReadLookupDefinition(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading lookup definition", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LookupDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LookupDefinitionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()
	name := plan.Name.ValueString()

	params := url.Values{}

	if !plan.Filename.IsNull() && !plan.Filename.IsUnknown() {
		params.Set("filename", plan.Filename.ValueString())
	}
	if !plan.ExternalType.IsNull() && !plan.ExternalType.IsUnknown() {
		params.Set("external_type", plan.ExternalType.ValueString())
	}
	if !plan.Collection.IsNull() && !plan.Collection.IsUnknown() {
		params.Set("collection", plan.Collection.ValueString())
	}
	if !plan.FieldsList.IsNull() && !plan.FieldsList.IsUnknown() {
		params.Set("fields_list", plan.FieldsList.ValueString())
	}
	if !plan.MaxMatches.IsNull() && !plan.MaxMatches.IsUnknown() {
		params.Set("max_matches", fmt.Sprintf("%d", plan.MaxMatches.ValueInt64()))
	}
	if !plan.MinMatches.IsNull() && !plan.MinMatches.IsUnknown() {
		params.Set("min_matches", fmt.Sprintf("%d", plan.MinMatches.ValueInt64()))
	}
	if !plan.DefaultMatch.IsNull() && !plan.DefaultMatch.IsUnknown() {
		params.Set("default_match", plan.DefaultMatch.ValueString())
	}
	if !plan.CaseSensitiveMatch.IsNull() && !plan.CaseSensitiveMatch.IsUnknown() {
		if plan.CaseSensitiveMatch.ValueBool() {
			params.Set("case_sensitive_match", "1")
		} else {
			params.Set("case_sensitive_match", "0")
		}
	}
	if !plan.MatchType.IsNull() && !plan.MatchType.IsUnknown() {
		params.Set("match_type", plan.MatchType.ValueString())
	}

	tflog.Debug(ctx, "Updating lookup definition", map[string]interface{}{"name": name})

	_, err := r.client.UpdateLookupDefinition(owner, app, name, params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lookup definition", err.Error())
		return
	}

	// Read back the updated resource.
	readResp, err := r.client.ReadLookupDefinition(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading lookup definition after update", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LookupDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LookupDefinitionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting lookup definition", map[string]interface{}{"name": name})

	err := r.client.DeleteLookupDefinition(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting lookup definition", err.Error())
		return
	}
}

func (r *LookupDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), "nobody")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), "search")...)
}

// mapResponseToModel maps a Splunk API response to the LookupDefinitionResourceModel.
func (r *LookupDefinitionResource) mapResponseToModel(response map[string]interface{}, model *LookupDefinitionResourceModel) {
	content, err := sdk.GetEntryContent(response)
	if err != nil {
		return
	}

	// Extract name from entry-level fields (not content).
	entries, _ := response["entry"].([]interface{})
	if len(entries) > 0 {
		entry, _ := entries[0].(map[string]interface{})
		if name, ok := entry["name"].(string); ok {
			model.Name = types.StringValue(name)
		}
	}

	model.ID = types.StringValue(model.Name.ValueString())

	if filename := sdk.ParseString(content, "filename"); filename != "" {
		model.Filename = types.StringValue(filename)
	}
	if externalType := sdk.ParseString(content, "external_type"); externalType != "" {
		model.ExternalType = types.StringValue(externalType)
	}
	if collection := sdk.ParseString(content, "collection"); collection != "" {
		model.Collection = types.StringValue(collection)
	}
	if fieldsList := sdk.ParseString(content, "fields_list"); fieldsList != "" {
		model.FieldsList = types.StringValue(fieldsList)
	}
	if defaultMatch := sdk.ParseString(content, "default_match"); defaultMatch != "" {
		model.DefaultMatch = types.StringValue(defaultMatch)
	}
	if matchType := sdk.ParseString(content, "match_type"); matchType != "" {
		model.MatchType = types.StringValue(matchType)
	}

	model.MaxMatches = types.Int64Value(sdk.ParseInt(content, "max_matches"))
	model.MinMatches = types.Int64Value(sdk.ParseInt(content, "min_matches"))
	model.CaseSensitiveMatch = types.BoolValue(sdk.ParseBool(content, "case_sensitive_match"))

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
