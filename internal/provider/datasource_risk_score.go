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
	_ datasource.DataSource              = &RiskScoreDataSource{}
	_ datasource.DataSourceWithConfigure = &RiskScoreDataSource{}
)

// NewRiskScoreDataSource is a constructor that returns a new RiskScoreDataSource.
func NewRiskScoreDataSource() datasource.DataSource {
	return &RiskScoreDataSource{}
}

// RiskScoreDataSource implements the splunkes_risk_score data source.
type RiskScoreDataSource struct {
	client *sdk.SplunkClient
}

// RiskScoreDataSourceModel describes the data source data model.
type RiskScoreDataSourceModel struct {
	Entity     types.String `tfsdk:"entity"`
	EntityType types.String `tfsdk:"entity_type"`
	RiskScore  types.Int64  `tfsdk:"risk_score"`
	RiskLevel  types.String `tfsdk:"risk_level"`
}

func (d *RiskScoreDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_risk_score"
}

func (d *RiskScoreDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Splunk ES risk score for an entity.",
		Attributes: map[string]schema.Attribute{
			"entity": schema.StringAttribute{
				Description: "The entity name to look up the risk score for.",
				Required:    true,
			},
			"entity_type": schema.StringAttribute{
				Description: "The entity type (user, system). Defaults to user.",
				Optional:    true,
				Computed:    true,
			},
			"risk_score": schema.Int64Attribute{
				Description: "The computed risk score for the entity.",
				Computed:    true,
			},
			"risk_level": schema.StringAttribute{
				Description: "The risk level classification (low, medium, high, critical).",
				Computed:    true,
			},
		},
	}
}

func (d *RiskScoreDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RiskScoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state RiskScoreDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Apply default for entity_type.
	if state.EntityType.IsNull() || state.EntityType.IsUnknown() {
		state.EntityType = types.StringValue("user")
	}

	entity := state.Entity.ValueString()
	entityType := state.EntityType.ValueString()

	tflog.Debug(ctx, "Reading risk score data source", map[string]interface{}{"entity": entity, "entity_type": entityType})

	readResp, err := d.client.ReadRiskScore(entity, entityType)
	if err != nil {
		resp.Diagnostics.AddError("Error reading risk score", err.Error())
		return
	}

	// ES v2 API returns risk score data directly.
	state.RiskScore = types.Int64Value(sdk.ParseInt(readResp, "risk_score"))
	state.RiskLevel = types.StringValue(sdk.ParseString(readResp, "risk_level"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
