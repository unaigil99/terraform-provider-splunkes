package provider

import (
	"context"
	"fmt"

	"github.com/Agentric/terraform-provider-splunkes/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &MacroDataSource{}
	_ datasource.DataSourceWithConfigure = &MacroDataSource{}
)

// NewMacroDataSource is a constructor that returns a new MacroDataSource.
func NewMacroDataSource() datasource.DataSource {
	return &MacroDataSource{}
}

// MacroDataSource implements the splunkes_macro data source.
type MacroDataSource struct {
	client *sdk.SplunkClient
}

// MacroDataSourceModel describes the data source data model.
type MacroDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	App         types.String `tfsdk:"app"`
	Owner       types.String `tfsdk:"owner"`
	Definition  types.String `tfsdk:"definition"`
	Description types.String `tfsdk:"description"`
	Args        types.String `tfsdk:"args"`
	Validation  types.String `tfsdk:"validation"`
	IsEval      types.Bool   `tfsdk:"iseval"`
}

func (d *MacroDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_macro"
}

func (d *MacroDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Splunk search macro.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the macro.",
				Required:    true,
			},
			"app": schema.StringAttribute{
				Description: "The Splunk app context. Defaults to SplunkEnterpriseSecuritySuite.",
				Optional:    true,
				Computed:    true,
			},
			"owner": schema.StringAttribute{
				Description: "The Splunk user owner. Defaults to nobody.",
				Optional:    true,
				Computed:    true,
			},
			"definition": schema.StringAttribute{
				Description: "The SPL definition of the macro.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the macro.",
				Computed:    true,
			},
			"args": schema.StringAttribute{
				Description: "Comma-separated argument names for the macro.",
				Computed:    true,
			},
			"validation": schema.StringAttribute{
				Description: "A validation expression for the macro arguments.",
				Computed:    true,
			},
			"iseval": schema.BoolAttribute{
				Description: "Whether the macro definition is an eval expression.",
				Computed:    true,
			},
		},
	}
}

func (d *MacroDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.SplunkClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sdk.SplunkClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *MacroDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state MacroDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Apply defaults for optional fields.
	if state.App.IsNull() || state.App.IsUnknown() {
		state.App = types.StringValue("SplunkEnterpriseSecuritySuite")
	}
	if state.Owner.IsNull() || state.Owner.IsUnknown() {
		state.Owner = types.StringValue("nobody")
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading macro data source", map[string]interface{}{"name": name})

	readResp, err := d.client.ReadMacro(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading macro", err.Error())
		return
	}

	content, err := sdk.GetEntryContent(readResp)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing macro response", err.Error())
		return
	}

	state.Definition = types.StringValue(sdk.ParseString(content, "definition"))
	state.Description = types.StringValue(sdk.ParseString(content, "description"))
	state.Args = types.StringValue(sdk.ParseString(content, "args"))
	state.Validation = types.StringValue(sdk.ParseString(content, "validation"))
	state.IsEval = types.BoolValue(sdk.ParseBool(content, "iseval"))

	// Extract app and owner from ACL.
	acl, aclErr := sdk.GetEntryACL(readResp)
	if aclErr == nil {
		if appVal := sdk.ParseString(acl, "app"); appVal != "" {
			state.App = types.StringValue(appVal)
		}
		if ownerVal := sdk.ParseString(acl, "owner"); ownerVal != "" {
			state.Owner = types.StringValue(ownerVal)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
