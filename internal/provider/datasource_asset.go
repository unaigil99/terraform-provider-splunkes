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
	_ datasource.DataSource              = &AssetDataSource{}
	_ datasource.DataSourceWithConfigure = &AssetDataSource{}
)

// NewAssetDataSource is a constructor that returns a new AssetDataSource.
func NewAssetDataSource() datasource.DataSource {
	return &AssetDataSource{}
}

// AssetDataSource implements the splunkes_asset data source.
type AssetDataSource struct {
	client *sdk.SplunkClient
}

// AssetDataSourceModel describes the data source data model.
type AssetDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	IP         types.String `tfsdk:"ip"`
	MAC        types.String `tfsdk:"mac"`
	DNS        types.String `tfsdk:"dns"`
	NTHost     types.String `tfsdk:"nt_host"`
	BUnit      types.String `tfsdk:"bunit"`
	Category   types.String `tfsdk:"category"`
	Priority   types.String `tfsdk:"priority"`
	IsExpected types.Bool   `tfsdk:"is_expected"`
}

func (d *AssetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset"
}

func (d *AssetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Splunk ES asset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The asset ID.",
				Required:    true,
			},
			"ip": schema.StringAttribute{
				Description: "The IP address of the asset.",
				Computed:    true,
			},
			"mac": schema.StringAttribute{
				Description: "The MAC address of the asset.",
				Computed:    true,
			},
			"dns": schema.StringAttribute{
				Description: "The DNS name of the asset.",
				Computed:    true,
			},
			"nt_host": schema.StringAttribute{
				Description: "The Windows hostname of the asset.",
				Computed:    true,
			},
			"bunit": schema.StringAttribute{
				Description: "The business unit of the asset.",
				Computed:    true,
			},
			"category": schema.StringAttribute{
				Description: "The category of the asset.",
				Computed:    true,
			},
			"priority": schema.StringAttribute{
				Description: "The priority of the asset.",
				Computed:    true,
			},
			"is_expected": schema.BoolAttribute{
				Description: "Whether the asset is expected on the network.",
				Computed:    true,
			},
		},
	}
}

func (d *AssetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AssetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AssetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, "Reading asset data source", map[string]interface{}{"id": id})

	readResp, err := d.client.ReadAsset(id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading asset", err.Error())
		return
	}

	// ES v2 API returns asset data directly.
	data := readResp

	state.IP = types.StringValue(sdk.ParseString(data, "ip"))
	state.MAC = types.StringValue(sdk.ParseString(data, "mac"))
	state.DNS = types.StringValue(sdk.ParseString(data, "dns"))
	state.NTHost = types.StringValue(sdk.ParseString(data, "nt_host"))
	state.BUnit = types.StringValue(sdk.ParseString(data, "bunit"))
	state.Category = types.StringValue(sdk.ParseString(data, "category"))
	state.Priority = types.StringValue(sdk.ParseString(data, "priority"))
	state.IsExpected = types.BoolValue(sdk.ParseBool(data, "is_expected"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
