package provider

import (
	"context"
	"os"

	"github.com/Agentric/terraform-provider-splunkes/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &SplunkESProvider{}

// SplunkESProvider implements the Terraform provider for Splunk Enterprise Security.
type SplunkESProvider struct {
	version string
}

// SplunkESProviderModel maps the provider schema to Go types.
type SplunkESProviderModel struct {
	URL                types.String `tfsdk:"url"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	AuthToken          types.String `tfsdk:"auth_token"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	Timeout            types.Int64  `tfsdk:"timeout"`
}

// New returns a provider.Provider constructor function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SplunkESProvider{
			version: version,
		}
	}
}

func (p *SplunkESProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "splunkes"
	resp.Version = p.version
}

func (p *SplunkESProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Splunk Enterprise Security. Manages detection rules, " +
			"correlation searches, investigations, findings, risk modifiers, macros, lookups, " +
			"KV store collections, threat intelligence, and analytic stories.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Splunk management URL (e.g., https://localhost:8089). " +
					"Can also be set via SPLUNK_URL environment variable.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				Description: "Splunk admin username. Can also be set via SPLUNK_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Splunk admin password. Can also be set via SPLUNK_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"auth_token": schema.StringAttribute{
				Description: "Splunk authentication token (Bearer). Takes precedence over username/password. " +
					"Can also be set via SPLUNK_AUTH_TOKEN environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Defaults to false. " +
					"Can also be set via SPLUNK_INSECURE_SKIP_VERIFY environment variable.",
				Optional: true,
			},
			"timeout": schema.Int64Attribute{
				Description: "HTTP request timeout in seconds. Defaults to 60. " +
					"Can also be set via SPLUNK_TIMEOUT environment variable.",
				Optional: true,
			},
		},
	}
}

func (p *SplunkESProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SplunkESProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve configuration from env vars with schema overrides
	splunkURL := envOrDefault("SPLUNK_URL", "https://localhost:8089")
	if !config.URL.IsNull() && !config.URL.IsUnknown() {
		splunkURL = config.URL.ValueString()
	}

	username := os.Getenv("SPLUNK_USERNAME")
	if !config.Username.IsNull() && !config.Username.IsUnknown() {
		username = config.Username.ValueString()
	}

	password := os.Getenv("SPLUNK_PASSWORD")
	if !config.Password.IsNull() && !config.Password.IsUnknown() {
		password = config.Password.ValueString()
	}

	authToken := os.Getenv("SPLUNK_AUTH_TOKEN")
	if !config.AuthToken.IsNull() && !config.AuthToken.IsUnknown() {
		authToken = config.AuthToken.ValueString()
	}

	insecure := os.Getenv("SPLUNK_INSECURE_SKIP_VERIFY") == "true"
	if !config.InsecureSkipVerify.IsNull() && !config.InsecureSkipVerify.IsUnknown() {
		insecure = config.InsecureSkipVerify.ValueBool()
	}

	timeout := 60
	if !config.Timeout.IsNull() && !config.Timeout.IsUnknown() {
		timeout = int(config.Timeout.ValueInt64())
	}

	if authToken == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError(
			"Missing Splunk Credentials",
			"Either auth_token or both username and password must be provided. "+
				"Set them in the provider configuration or via environment variables "+
				"(SPLUNK_AUTH_TOKEN, SPLUNK_USERNAME, SPLUNK_PASSWORD).",
		)
		return
	}

	client, err := sdk.NewClient(splunkURL, username, password, authToken, insecure, timeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Splunk Client",
			"Could not authenticate with the Splunk instance: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SplunkESProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCorrelationSearchResource,
		NewMacroResource,
		NewLookupDefinitionResource,
		NewLookupTableResource,
		NewKVStoreCollectionResource,
		NewInvestigationResource,
		NewInvestigationNoteResource,
		NewFindingResource,
		NewRiskModifierResource,
		NewThreatIntelResource,
		NewAnalyticStoryResource,
		NewSavedSearchResource,
	}
}

func (p *SplunkESProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCorrelationSearchDataSource,
		NewInvestigationDataSource,
		NewFindingDataSource,
		NewRiskScoreDataSource,
		NewIdentityDataSource,
		NewAssetDataSource,
		NewMacroDataSource,
	}
}

func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
