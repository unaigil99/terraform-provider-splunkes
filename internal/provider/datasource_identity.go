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
	_ datasource.DataSource              = &IdentityDataSource{}
	_ datasource.DataSourceWithConfigure = &IdentityDataSource{}
)

// NewIdentityDataSource is a constructor that returns a new IdentityDataSource.
func NewIdentityDataSource() datasource.DataSource {
	return &IdentityDataSource{}
}

// IdentityDataSource implements the splunkes_identity data source.
type IdentityDataSource struct {
	client *sdk.SplunkClient
}

// IdentityDataSourceModel describes the data source data model.
type IdentityDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Email     types.String `tfsdk:"email"`
	BUnit     types.String `tfsdk:"bunit"`
	Category  types.String `tfsdk:"category"`
	Priority  types.String `tfsdk:"priority"`
	Watchlist types.Bool   `tfsdk:"watchlist"`
}

func (d *IdentityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity"
}

func (d *IdentityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Splunk ES identity.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The identity ID.",
				Required:    true,
			},
			"first_name": schema.StringAttribute{
				Description: "The first name of the identity.",
				Computed:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "The last name of the identity.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the identity.",
				Computed:    true,
			},
			"bunit": schema.StringAttribute{
				Description: "The business unit of the identity.",
				Computed:    true,
			},
			"category": schema.StringAttribute{
				Description: "The category of the identity.",
				Computed:    true,
			},
			"priority": schema.StringAttribute{
				Description: "The priority of the identity.",
				Computed:    true,
			},
			"watchlist": schema.BoolAttribute{
				Description: "Whether the identity is on the watchlist.",
				Computed:    true,
			},
		},
	}
}

func (d *IdentityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IdentityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state IdentityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, "Reading identity data source", map[string]interface{}{"id": id})

	readResp, err := d.client.ReadIdentity(id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading identity", err.Error())
		return
	}

	// ES v2 API returns identity data directly.
	data := readResp

	state.FirstName = types.StringValue(sdk.ParseString(data, "first_name"))
	state.LastName = types.StringValue(sdk.ParseString(data, "last_name"))
	state.Email = types.StringValue(sdk.ParseString(data, "email"))
	state.BUnit = types.StringValue(sdk.ParseString(data, "bunit"))
	state.Category = types.StringValue(sdk.ParseString(data, "category"))
	state.Priority = types.StringValue(sdk.ParseString(data, "priority"))
	state.Watchlist = types.BoolValue(sdk.ParseBool(data, "watchlist"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
