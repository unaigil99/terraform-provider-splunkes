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
	_ datasource.DataSource              = &InvestigationDataSource{}
	_ datasource.DataSourceWithConfigure = &InvestigationDataSource{}
)

// NewInvestigationDataSource is a constructor that returns a new InvestigationDataSource.
func NewInvestigationDataSource() datasource.DataSource {
	return &InvestigationDataSource{}
}

// InvestigationDataSource implements the splunkes_investigation data source.
type InvestigationDataSource struct {
	client *sdk.SplunkClient
}

// InvestigationDataSourceModel describes the data source data model.
type InvestigationDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Status   types.String `tfsdk:"status"`
	Assignee types.String `tfsdk:"assignee"`
	Priority types.String `tfsdk:"priority"`
}

func (d *InvestigationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_investigation"
}

func (d *InvestigationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Splunk ES investigation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The investigation ID (_key).",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name/title of the investigation.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the investigation.",
				Computed:    true,
			},
			"assignee": schema.StringAttribute{
				Description: "The user assigned to the investigation.",
				Computed:    true,
			},
			"priority": schema.StringAttribute{
				Description: "The priority of the investigation.",
				Computed:    true,
			},
		},
	}
}

func (d *InvestigationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InvestigationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state InvestigationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, "Reading investigation data source", map[string]interface{}{"id": id})

	readResp, err := d.client.ReadInvestigation(id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading investigation", err.Error())
		return
	}

	// ES v2 API returns investigation data directly or in a list.
	data := readResp
	if items, ok := readResp["items"].([]interface{}); ok && len(items) > 0 {
		if item, ok := items[0].(map[string]interface{}); ok {
			data = item
		}
	}

	state.Name = types.StringValue(sdk.ParseString(data, "title"))
	state.Status = types.StringValue(sdk.ParseString(data, "status"))
	state.Assignee = types.StringValue(sdk.ParseString(data, "assignee"))
	state.Priority = types.StringValue(sdk.ParseString(data, "priority"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
