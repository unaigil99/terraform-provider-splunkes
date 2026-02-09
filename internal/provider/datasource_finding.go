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
	_ datasource.DataSource              = &FindingDataSource{}
	_ datasource.DataSourceWithConfigure = &FindingDataSource{}
)

// NewFindingDataSource is a constructor that returns a new FindingDataSource.
func NewFindingDataSource() datasource.DataSource {
	return &FindingDataSource{}
}

// FindingDataSource implements the splunkes_finding data source.
type FindingDataSource struct {
	client *sdk.SplunkClient
}

// FindingDataSourceModel describes the data source data model.
type FindingDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	RuleTitle      types.String `tfsdk:"rule_title"`
	SecurityDomain types.String `tfsdk:"security_domain"`
	RiskScore      types.String `tfsdk:"risk_score"`
	RiskObject     types.String `tfsdk:"risk_object"`
	RiskObjectType types.String `tfsdk:"risk_object_type"`
	Severity       types.String `tfsdk:"severity"`
}

func (d *FindingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_finding"
}

func (d *FindingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Splunk ES finding.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The finding ID.",
				Required:    true,
			},
			"rule_title": schema.StringAttribute{
				Description: "The title of the rule that generated this finding.",
				Computed:    true,
			},
			"security_domain": schema.StringAttribute{
				Description: "The security domain of the finding.",
				Computed:    true,
			},
			"risk_score": schema.StringAttribute{
				Description: "The risk score associated with the finding.",
				Computed:    true,
			},
			"risk_object": schema.StringAttribute{
				Description: "The risk object associated with the finding.",
				Computed:    true,
			},
			"risk_object_type": schema.StringAttribute{
				Description: "The type of the risk object (user, system, etc.).",
				Computed:    true,
			},
			"severity": schema.StringAttribute{
				Description: "The severity of the finding.",
				Computed:    true,
			},
		},
	}
}

func (d *FindingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state FindingDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	tflog.Debug(ctx, "Reading finding data source", map[string]interface{}{"id": id})

	readResp, err := d.client.ReadFinding(id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading finding", err.Error())
		return
	}

	// ES v2 API returns finding data directly.
	data := readResp

	state.RuleTitle = types.StringValue(sdk.ParseString(data, "rule_title"))
	state.SecurityDomain = types.StringValue(sdk.ParseString(data, "security_domain"))
	state.RiskScore = types.StringValue(sdk.ParseString(data, "risk_score"))
	state.RiskObject = types.StringValue(sdk.ParseString(data, "risk_object"))
	state.RiskObjectType = types.StringValue(sdk.ParseString(data, "risk_object_type"))
	state.Severity = types.StringValue(sdk.ParseString(data, "severity"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
