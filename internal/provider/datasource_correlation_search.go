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
	_ datasource.DataSource              = &CorrelationSearchDataSource{}
	_ datasource.DataSourceWithConfigure = &CorrelationSearchDataSource{}
)

// NewCorrelationSearchDataSource is a constructor that returns a new CorrelationSearchDataSource.
func NewCorrelationSearchDataSource() datasource.DataSource {
	return &CorrelationSearchDataSource{}
}

// CorrelationSearchDataSource implements the splunkes_correlation_search data source.
type CorrelationSearchDataSource struct {
	client *sdk.SplunkClient
}

// CorrelationSearchDataSourceModel describes the data source data model.
type CorrelationSearchDataSourceModel struct {
	Name            types.String `tfsdk:"name"`
	App             types.String `tfsdk:"app"`
	Owner           types.String `tfsdk:"owner"`
	Search          types.String `tfsdk:"search"`
	Description     types.String `tfsdk:"description"`
	Disabled        types.Bool   `tfsdk:"disabled"`
	CronSchedule    types.String `tfsdk:"cron_schedule"`
	NotableEnabled  types.Bool   `tfsdk:"notable_enabled"`
	SecurityDomain  types.String `tfsdk:"security_domain"`
	Severity        types.String `tfsdk:"severity"`
	RiskScore       types.String `tfsdk:"risk_score"`
}

func (d *CorrelationSearchDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_correlation_search"
}

func (d *CorrelationSearchDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Splunk ES correlation search.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the correlation search.",
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
			"search": schema.StringAttribute{
				Description: "The SPL search query.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the correlation search.",
				Computed:    true,
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the correlation search is disabled.",
				Computed:    true,
			},
			"cron_schedule": schema.StringAttribute{
				Description: "The cron schedule for the correlation search.",
				Computed:    true,
			},
			"notable_enabled": schema.BoolAttribute{
				Description: "Whether notable event creation is enabled.",
				Computed:    true,
			},
			"security_domain": schema.StringAttribute{
				Description: "The security domain for the correlation search.",
				Computed:    true,
			},
			"severity": schema.StringAttribute{
				Description: "The severity of the correlation search.",
				Computed:    true,
			},
			"risk_score": schema.StringAttribute{
				Description: "The default risk score for the correlation search.",
				Computed:    true,
			},
		},
	}
}

func (d *CorrelationSearchDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CorrelationSearchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CorrelationSearchDataSourceModel
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

	tflog.Debug(ctx, "Reading correlation search data source", map[string]interface{}{"name": name})

	readResp, err := d.client.ReadSavedSearch(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading correlation search", err.Error())
		return
	}

	content, err := sdk.GetEntryContent(readResp)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing correlation search response", err.Error())
		return
	}

	state.Search = types.StringValue(sdk.ParseString(content, "search"))
	state.Description = types.StringValue(sdk.ParseString(content, "description"))
	state.Disabled = types.BoolValue(sdk.ParseBool(content, "disabled"))
	state.CronSchedule = types.StringValue(sdk.ParseString(content, "cron_schedule"))

	// Notable and security domain fields are stored under action.correlationsearch.*
	state.NotableEnabled = types.BoolValue(sdk.ParseBool(content, "action.notable"))
	state.SecurityDomain = types.StringValue(sdk.ParseString(content, "action.correlationsearch.label"))
	state.Severity = types.StringValue(sdk.ParseString(content, "alert.severity"))
	state.RiskScore = types.StringValue(sdk.ParseString(content, "action.risk.param._risk_score"))

	// If security_domain not found in correlationsearch.label, try alternative field.
	if state.SecurityDomain.ValueString() == "" {
		state.SecurityDomain = types.StringValue(sdk.ParseString(content, "security_domain"))
	}

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
