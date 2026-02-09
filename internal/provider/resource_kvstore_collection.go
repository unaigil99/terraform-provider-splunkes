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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &KVStoreCollectionResource{}
	_ resource.ResourceWithConfigure   = &KVStoreCollectionResource{}
	_ resource.ResourceWithImportState = &KVStoreCollectionResource{}
)

// NewKVStoreCollectionResource is a constructor that returns a new KVStoreCollectionResource.
func NewKVStoreCollectionResource() resource.Resource {
	return &KVStoreCollectionResource{}
}

// KVStoreCollectionResource implements the splunkes_kvstore_collection resource.
type KVStoreCollectionResource struct {
	client *sdk.SplunkClient
}

// KVStoreCollectionResourceModel describes the resource data model.
type KVStoreCollectionResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	App               types.String `tfsdk:"app"`
	Owner             types.String `tfsdk:"owner"`
	Fields            types.Map    `tfsdk:"fields"`
	AcceleratedFields types.Map    `tfsdk:"accelerated_fields"`
}

func (r *KVStoreCollectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kvstore_collection"
}

func (r *KVStoreCollectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Splunk KV Store collection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this KV store collection resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the KV store collection.",
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
			"fields": schema.MapAttribute{
				Description: "A map of field names to their types (number, bool, string, time).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"accelerated_fields": schema.MapAttribute{
				Description: "A map of accelerated field names to their JSON index definitions.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *KVStoreCollectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KVStoreCollectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KVStoreCollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()

	params := url.Values{}
	params.Set("name", plan.Name.ValueString())

	r.addFieldParams(ctx, &plan, params, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating KV store collection", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.CreateKVStoreCollection(owner, app, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating KV store collection", err.Error())
		return
	}

	// Read back the created resource to populate all fields.
	readResp, err := r.client.ReadKVStoreCollection(owner, app, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading KV store collection after creation", err.Error())
		return
	}

	r.mapResponseToModel(ctx, readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KVStoreCollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KVStoreCollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading KV store collection", map[string]interface{}{"name": name})

	readResp, err := r.client.ReadKVStoreCollection(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading KV store collection", err.Error())
		return
	}

	r.mapResponseToModel(ctx, readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KVStoreCollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan KVStoreCollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()
	name := plan.Name.ValueString()

	params := url.Values{}

	r.addFieldParams(ctx, &plan, params, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating KV store collection", map[string]interface{}{"name": name})

	_, err := r.client.UpdateKVStoreCollection(owner, app, name, params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating KV store collection", err.Error())
		return
	}

	// Read back the updated resource.
	readResp, err := r.client.ReadKVStoreCollection(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading KV store collection after update", err.Error())
		return
	}

	r.mapResponseToModel(ctx, readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KVStoreCollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KVStoreCollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting KV store collection", map[string]interface{}{"name": name})

	err := r.client.DeleteKVStoreCollection(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting KV store collection", err.Error())
		return
	}
}

func (r *KVStoreCollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), "nobody")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), "search")...)
}

// addFieldParams adds field.* and accelerated_fields.* parameters to url.Values.
func (r *KVStoreCollectionResource) addFieldParams(ctx context.Context, model *KVStoreCollectionResourceModel, params url.Values, diags *diag.Diagnostics) {
	if !model.Fields.IsNull() && !model.Fields.IsUnknown() {
		fields := make(map[string]string)
		diags.Append(model.Fields.ElementsAs(ctx, &fields, false)...)
		for k, v := range fields {
			params.Set(fmt.Sprintf("field.%s", k), v)
		}
	}

	if !model.AcceleratedFields.IsNull() && !model.AcceleratedFields.IsUnknown() {
		accelFields := make(map[string]string)
		diags.Append(model.AcceleratedFields.ElementsAs(ctx, &accelFields, false)...)
		for k, v := range accelFields {
			params.Set(fmt.Sprintf("accelerated_fields.%s", k), v)
		}
	}
}

// mapResponseToModel maps a Splunk API response to the KVStoreCollectionResourceModel.
func (r *KVStoreCollectionResource) mapResponseToModel(ctx context.Context, response map[string]interface{}, model *KVStoreCollectionResourceModel) {
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

	// Parse fields (field.*) and accelerated_fields (accelerated_fields.*) from content.
	fields := make(map[string]string)
	accelFields := make(map[string]string)

	fieldPrefix := "field."
	accelPrefix := "accelerated_fields."

	for key, val := range content {
		if strings.HasPrefix(key, fieldPrefix) {
			fieldName := key[len(fieldPrefix):]
			if strVal, ok := val.(string); ok {
				fields[fieldName] = strVal
			}
		}
		if strings.HasPrefix(key, accelPrefix) {
			accelName := key[len(accelPrefix):]
			if strVal, ok := val.(string); ok {
				accelFields[accelName] = strVal
			}
		}
	}

	if len(fields) > 0 {
		mapVal, _ := types.MapValueFrom(ctx, types.StringType, fields)
		model.Fields = mapVal
	}
	if len(accelFields) > 0 {
		mapVal, _ := types.MapValueFrom(ctx, types.StringType, accelFields)
		model.AcceleratedFields = mapVal
	}
}
