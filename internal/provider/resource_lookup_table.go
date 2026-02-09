package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Agentric/terraform-provider-splunkes/internal/sdk"
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
	_ resource.Resource                = &LookupTableResource{}
	_ resource.ResourceWithConfigure   = &LookupTableResource{}
	_ resource.ResourceWithImportState = &LookupTableResource{}
)

// NewLookupTableResource is a constructor that returns a new LookupTableResource.
func NewLookupTableResource() resource.Resource {
	return &LookupTableResource{}
}

// LookupTableResource implements the splunkes_lookup_table resource.
type LookupTableResource struct {
	client *sdk.SplunkClient
}

// LookupTableResourceModel describes the resource data model.
type LookupTableResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	App     types.String `tfsdk:"app"`
	Owner   types.String `tfsdk:"owner"`
	Content types.String `tfsdk:"content"`
}

func (r *LookupTableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_table"
}

func (r *LookupTableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Splunk lookup table file (CSV).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this lookup table resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The filename of the lookup table (e.g., my_lookup.csv).",
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
			"content": schema.StringAttribute{
				Description: "The CSV content of the lookup table.",
				Optional:    true,
			},
		},
	}
}

func (r *LookupTableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LookupTableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LookupTableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()
	name := plan.Name.ValueString()

	// Create the lookup table file metadata.
	params := url.Values{}
	params.Set("eai:data", name)

	tflog.Debug(ctx, "Creating lookup table", map[string]interface{}{"name": name})

	_, err := r.client.CreateLookupTable(owner, app, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lookup table", err.Error())
		return
	}

	// If content is provided, upload the CSV content.
	if !plan.Content.IsNull() && !plan.Content.IsUnknown() && plan.Content.ValueString() != "" {
		tflog.Debug(ctx, "Uploading lookup table content", map[string]interface{}{"name": name})

		_, err := r.client.UpdateLookupTableContent(owner, app, name, plan.Content.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error uploading lookup table content", err.Error())
			return
		}
	}

	// Read back the created resource to populate all fields.
	readResp, err := r.client.ReadLookupTable(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading lookup table after creation", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LookupTableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LookupTableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading lookup table", map[string]interface{}{"name": name})

	readResp, err := r.client.ReadLookupTable(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading lookup table", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LookupTableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LookupTableResourceModel
	var state LookupTableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()
	name := plan.Name.ValueString()

	// Check if content has changed.
	contentChanged := !plan.Content.Equal(state.Content)

	if contentChanged && !plan.Content.IsNull() && !plan.Content.IsUnknown() {
		tflog.Debug(ctx, "Updating lookup table content", map[string]interface{}{"name": name})

		_, err := r.client.UpdateLookupTableContent(owner, app, name, plan.Content.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error updating lookup table content", err.Error())
			return
		}
	} else if !contentChanged {
		// Metadata-only update (e.g., if app or owner could change without replace).
		params := url.Values{}
		params.Set("eai:data", name)

		tflog.Debug(ctx, "Updating lookup table metadata", map[string]interface{}{"name": name})

		_, err := r.client.UpdateLookupTable(owner, app, name, params)
		if err != nil {
			resp.Diagnostics.AddError("Error updating lookup table metadata", err.Error())
			return
		}
	}

	// Read back the updated resource.
	readResp, err := r.client.ReadLookupTable(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading lookup table after update", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LookupTableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LookupTableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting lookup table", map[string]interface{}{"name": name})

	err := r.client.DeleteLookupTable(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting lookup table", err.Error())
		return
	}
}

func (r *LookupTableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), "nobody")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), "search")...)
}

// mapResponseToModel maps a Splunk API response to the LookupTableResourceModel.
func (r *LookupTableResource) mapResponseToModel(response map[string]interface{}, model *LookupTableResourceModel) {
	// Extract name from entry-level fields.
	entries, _ := response["entry"].([]interface{})
	if len(entries) > 0 {
		entry, _ := entries[0].(map[string]interface{})
		if name, ok := entry["name"].(string); ok {
			model.Name = types.StringValue(name)
		}
	}

	model.ID = types.StringValue(model.Name.ValueString())

	// Note: The Splunk lookup-table-files API does not return CSV content in the
	// metadata response. We preserve the content value from the plan/state to
	// avoid unnecessary diffs. The content attribute is write-only in practice.

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
