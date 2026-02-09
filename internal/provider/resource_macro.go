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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &MacroResource{}
	_ resource.ResourceWithConfigure   = &MacroResource{}
	_ resource.ResourceWithImportState = &MacroResource{}
)

// NewMacroResource is a constructor that returns a new MacroResource.
func NewMacroResource() resource.Resource {
	return &MacroResource{}
}

// MacroResource implements the splunkes_macro resource.
type MacroResource struct {
	client *sdk.SplunkClient
}

// MacroResourceModel describes the resource data model.
type MacroResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Definition  types.String `tfsdk:"definition"`
	Description types.String `tfsdk:"description"`
	App         types.String `tfsdk:"app"`
	Owner       types.String `tfsdk:"owner"`
	Args        types.String `tfsdk:"args"`
	Validation  types.String `tfsdk:"validation"`
	ErrorMsg    types.String `tfsdk:"errormsg"`
	IsEval      types.Bool   `tfsdk:"iseval"`
}

func (r *MacroResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_macro"
}

func (r *MacroResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Splunk search macro.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this macro resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the macro.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"definition": schema.StringAttribute{
				Description: "The SPL definition of the macro.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the macro.",
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
			"args": schema.StringAttribute{
				Description: "Comma-separated argument names for the macro.",
				Optional:    true,
			},
			"validation": schema.StringAttribute{
				Description: "A validation expression for the macro arguments.",
				Optional:    true,
			},
			"errormsg": schema.StringAttribute{
				Description: "The error message displayed when validation fails.",
				Optional:    true,
			},
			"iseval": schema.BoolAttribute{
				Description: "Whether the macro definition is an eval expression. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *MacroResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MacroResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MacroResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()

	params := url.Values{}
	params.Set("name", plan.Name.ValueString())
	params.Set("definition", plan.Definition.ValueString())

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		params.Set("description", plan.Description.ValueString())
	}
	if !plan.Args.IsNull() && !plan.Args.IsUnknown() {
		params.Set("args", plan.Args.ValueString())
	}
	if !plan.Validation.IsNull() && !plan.Validation.IsUnknown() {
		params.Set("validation", plan.Validation.ValueString())
	}
	if !plan.ErrorMsg.IsNull() && !plan.ErrorMsg.IsUnknown() {
		params.Set("errormsg", plan.ErrorMsg.ValueString())
	}
	if !plan.IsEval.IsNull() && !plan.IsEval.IsUnknown() {
		if plan.IsEval.ValueBool() {
			params.Set("iseval", "1")
		} else {
			params.Set("iseval", "0")
		}
	}

	tflog.Debug(ctx, "Creating macro", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.CreateMacro(owner, app, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating macro", err.Error())
		return
	}

	// Read back the created resource to populate all fields.
	readResp, err := r.client.ReadMacro(owner, app, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading macro after creation", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MacroResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MacroResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading macro", map[string]interface{}{"name": name})

	readResp, err := r.client.ReadMacro(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading macro", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MacroResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MacroResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()
	name := plan.Name.ValueString()

	params := url.Values{}
	params.Set("definition", plan.Definition.ValueString())

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		params.Set("description", plan.Description.ValueString())
	}
	if !plan.Args.IsNull() && !plan.Args.IsUnknown() {
		params.Set("args", plan.Args.ValueString())
	}
	if !plan.Validation.IsNull() && !plan.Validation.IsUnknown() {
		params.Set("validation", plan.Validation.ValueString())
	}
	if !plan.ErrorMsg.IsNull() && !plan.ErrorMsg.IsUnknown() {
		params.Set("errormsg", plan.ErrorMsg.ValueString())
	}
	if !plan.IsEval.IsNull() && !plan.IsEval.IsUnknown() {
		if plan.IsEval.ValueBool() {
			params.Set("iseval", "1")
		} else {
			params.Set("iseval", "0")
		}
	}

	tflog.Debug(ctx, "Updating macro", map[string]interface{}{"name": name})

	_, err := r.client.UpdateMacro(owner, app, name, params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating macro", err.Error())
		return
	}

	// Read back the updated resource.
	readResp, err := r.client.ReadMacro(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading macro after update", err.Error())
		return
	}

	r.mapResponseToModel(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MacroResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MacroResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting macro", map[string]interface{}{"name": name})

	err := r.client.DeleteMacro(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting macro", err.Error())
		return
	}
}

func (r *MacroResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), "nobody")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), "search")...)
}

// mapResponseToModel maps a Splunk API response to the MacroResourceModel.
func (r *MacroResource) mapResponseToModel(response map[string]interface{}, model *MacroResourceModel) {
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
	model.Definition = types.StringValue(sdk.ParseString(content, "definition"))

	if desc := sdk.ParseString(content, "description"); desc != "" {
		model.Description = types.StringValue(desc)
	}
	if args := sdk.ParseString(content, "args"); args != "" {
		model.Args = types.StringValue(args)
	}
	if validation := sdk.ParseString(content, "validation"); validation != "" {
		model.Validation = types.StringValue(validation)
	}
	if errormsg := sdk.ParseString(content, "errormsg"); errormsg != "" {
		model.ErrorMsg = types.StringValue(errormsg)
	}

	model.IsEval = types.BoolValue(sdk.ParseBool(content, "iseval"))

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
