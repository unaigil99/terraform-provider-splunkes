package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Agentric/terraform-provider-splunkes/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &CorrelationSearchResource{}
	_ resource.ResourceWithImportState = &CorrelationSearchResource{}
)

// NewCorrelationSearchResource returns a new resource.Resource for correlation searches.
func NewCorrelationSearchResource() resource.Resource {
	return &CorrelationSearchResource{}
}

// CorrelationSearchResource implements the splunkes_correlation_search resource.
type CorrelationSearchResource struct {
	client *sdk.SplunkClient
}

// CorrelationSearchResourceModel describes the resource data model.
type CorrelationSearchResourceModel struct {
	// Identity
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	// Core search fields
	Search      types.String `tfsdk:"search"`
	Description types.String `tfsdk:"description"`
	App         types.String `tfsdk:"app"`
	Owner       types.String `tfsdk:"owner"`
	Disabled    types.Bool   `tfsdk:"disabled"`

	// Scheduling
	CronSchedule         types.String `tfsdk:"cron_schedule"`
	DispatchEarliestTime types.String `tfsdk:"dispatch_earliest_time"`
	DispatchLatestTime   types.String `tfsdk:"dispatch_latest_time"`
	Scheduling           types.String `tfsdk:"scheduling"`
	IsScheduled          types.Bool   `tfsdk:"is_scheduled"`

	// ES-specific
	CorrelationSearchEnabled types.Bool   `tfsdk:"correlation_search_enabled"`
	CorrelationSearchLabel   types.String `tfsdk:"correlation_search_label"`

	// Notable event action
	NotableEnabled            types.Bool   `tfsdk:"notable_enabled"`
	NotableRuleTitle          types.String `tfsdk:"notable_rule_title"`
	NotableRuleDescription    types.String `tfsdk:"notable_rule_description"`
	NotableSecurityDomain     types.String `tfsdk:"notable_security_domain"`
	NotableSeverity           types.String `tfsdk:"notable_severity"`
	NotableDrilldownName      types.String `tfsdk:"notable_drilldown_name"`
	NotableDrilldownSearch    types.String `tfsdk:"notable_drilldown_search"`
	NotableDefaultOwner       types.String `tfsdk:"notable_default_owner"`
	NotableRecommendedActions types.String `tfsdk:"notable_recommended_actions"`

	// Risk action
	RiskEnabled     types.Bool   `tfsdk:"risk_enabled"`
	RiskScore       types.Int64  `tfsdk:"risk_score"`
	RiskMessage     types.String `tfsdk:"risk_message"`
	RiskObjectField types.String `tfsdk:"risk_object_field"`
	RiskObjectType  types.String `tfsdk:"risk_object_type"`

	// MITRE ATT&CK and frameworks
	MitreAttackIDs  types.List `tfsdk:"mitre_attack_ids"`
	KillChainPhases types.List `tfsdk:"kill_chain_phases"`
	CIS20           types.List `tfsdk:"cis20"`
	NIST            types.List `tfsdk:"nist"`
	AnalyticStory   types.List `tfsdk:"analytic_story"`

	// Alert
	AlertSeverity       types.Int64  `tfsdk:"alert_severity"`
	AlertType           types.String `tfsdk:"alert_type"`
	AlertComparator     types.String `tfsdk:"alert_comparator"`
	AlertThreshold      types.String `tfsdk:"alert_threshold"`
	AlertSuppress       types.Bool   `tfsdk:"alert_suppress"`
	AlertSuppressPeriod types.String `tfsdk:"alert_suppress_period"`
	AlertSuppressFields types.String `tfsdk:"alert_suppress_fields"`
}

func (r *CorrelationSearchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_correlation_search"
}

func (r *CorrelationSearchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Splunk Enterprise Security correlation search. " +
			"Wraps the Splunk saved searches REST API with ES-specific parameters for " +
			"notable event creation, risk scoring, and MITRE ATT&CK annotations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The resource identifier (set to the correlation search name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the correlation search. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Core search fields
			"search": schema.StringAttribute{
				Description: "The SPL query for the correlation search.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the correlation search.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
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
			"disabled": schema.BoolAttribute{
				Description: "Whether the correlation search is disabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},

			// Scheduling
			"cron_schedule": schema.StringAttribute{
				Description: "The cron schedule for the correlation search (e.g., '*/5 * * * *').",
				Required:    true,
			},
			"dispatch_earliest_time": schema.StringAttribute{
				Description: "The earliest time for the search dispatch window. Defaults to '-70m@m'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("-70m@m"),
			},
			"dispatch_latest_time": schema.StringAttribute{
				Description: "The latest time for the search dispatch window. Defaults to '-10m@m'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("-10m@m"),
			},
			"scheduling": schema.StringAttribute{
				Description: "The schedule priority (schedule_priority). For example: default, higher, highest.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
			},
			"is_scheduled": schema.BoolAttribute{
				Description: "Whether the correlation search is scheduled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},

			// ES-specific
			"correlation_search_enabled": schema.BoolAttribute{
				Description: "Whether the correlation search action is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"correlation_search_label": schema.StringAttribute{
				Description: "The label for the correlation search.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},

			// Notable event action
			"notable_enabled": schema.BoolAttribute{
				Description: "Whether the notable event action is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"notable_rule_title": schema.StringAttribute{
				Description: "The title for notable events created by this search.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"notable_rule_description": schema.StringAttribute{
				Description: "The description for notable events created by this search.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"notable_security_domain": schema.StringAttribute{
				Description: "The security domain for notable events. Valid values: access, endpoint, network, identity, threat, audit.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("threat"),
			},
			"notable_severity": schema.StringAttribute{
				Description: "The severity for notable events. Valid values: informational, low, medium, high, critical.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("medium"),
			},
			"notable_drilldown_name": schema.StringAttribute{
				Description: "The name of the drilldown search for notable events.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"notable_drilldown_search": schema.StringAttribute{
				Description: "The SPL drilldown search for notable events.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"notable_default_owner": schema.StringAttribute{
				Description: "The default owner for notable events.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"notable_recommended_actions": schema.StringAttribute{
				Description: "Comma-separated list of recommended actions for notable events.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},

			// Risk action
			"risk_enabled": schema.BoolAttribute{
				Description: "Whether the risk action is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"risk_score": schema.Int64Attribute{
				Description: "The risk score to assign.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"risk_message": schema.StringAttribute{
				Description: "The risk message template.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"risk_object_field": schema.StringAttribute{
				Description: "The field name used as the risk object.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"risk_object_type": schema.StringAttribute{
				Description: "The type of risk object (e.g., system, user, other).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},

			// MITRE ATT&CK and security frameworks
			"mitre_attack_ids": schema.ListAttribute{
				Description: "List of MITRE ATT&CK technique IDs (e.g., T1059, T1078).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"kill_chain_phases": schema.ListAttribute{
				Description: "List of kill chain phases.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"cis20": schema.ListAttribute{
				Description: "List of CIS 20 control IDs.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"nist": schema.ListAttribute{
				Description: "List of NIST framework control IDs.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"analytic_story": schema.ListAttribute{
				Description: "List of analytic stories associated with this correlation search.",
				Optional:    true,
				ElementType: types.StringType,
			},

			// Alert
			"alert_severity": schema.Int64Attribute{
				Description: "The alert severity level (1-6).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3),
			},
			"alert_type": schema.StringAttribute{
				Description: "The alert trigger type (e.g., 'number of events').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("number of events"),
			},
			"alert_comparator": schema.StringAttribute{
				Description: "The alert comparator (e.g., 'greater than').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("greater than"),
			},
			"alert_threshold": schema.StringAttribute{
				Description: "The alert threshold value.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0"),
			},
			"alert_suppress": schema.BoolAttribute{
				Description: "Whether alert suppression is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"alert_suppress_period": schema.StringAttribute{
				Description: "The alert suppression period (e.g., '24h').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"alert_suppress_fields": schema.StringAttribute{
				Description: "Comma-separated list of fields for alert suppression.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
		},
	}
}

func (r *CorrelationSearchResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CorrelationSearchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CorrelationSearchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()
	name := plan.Name.ValueString()

	params := r.buildParams(ctx, &plan)

	tflog.Debug(ctx, "Creating correlation search", map[string]interface{}{
		"name":  name,
		"owner": owner,
		"app":   app,
	})

	_, err := r.client.CreateSavedSearch(owner, app, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Correlation Search",
			fmt.Sprintf("Could not create correlation search %q: %s", name, err.Error()),
		)
		return
	}

	// Read back the created resource to get server-side defaults
	readResp, err := r.client.ReadSavedSearch(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Correlation Search After Create",
			fmt.Sprintf("Could not read correlation search %q after creation: %s", name, err.Error()),
		)
		return
	}

	r.refreshState(ctx, readResp, &plan)
	plan.ID = types.StringValue(name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CorrelationSearchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CorrelationSearchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading correlation search", map[string]interface{}{
		"name":  name,
		"owner": owner,
		"app":   app,
	})

	readResp, err := r.client.ReadSavedSearch(owner, app, name)
	if err != nil {
		// If the resource no longer exists, remove it from state
		if strings.Contains(err.Error(), "404") {
			tflog.Warn(ctx, "Correlation search not found, removing from state", map[string]interface{}{
				"name": name,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Correlation Search",
			fmt.Sprintf("Could not read correlation search %q: %s", name, err.Error()),
		)
		return
	}

	r.refreshState(ctx, readResp, &state)
	state.ID = types.StringValue(name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CorrelationSearchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CorrelationSearchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := plan.Owner.ValueString()
	app := plan.App.ValueString()
	name := plan.Name.ValueString()

	params := r.buildParams(ctx, &plan)

	tflog.Debug(ctx, "Updating correlation search", map[string]interface{}{
		"name":  name,
		"owner": owner,
		"app":   app,
	})

	_, err := r.client.UpdateSavedSearch(owner, app, name, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Correlation Search",
			fmt.Sprintf("Could not update correlation search %q: %s", name, err.Error()),
		)
		return
	}

	// Read back the updated resource
	readResp, err := r.client.ReadSavedSearch(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Correlation Search After Update",
			fmt.Sprintf("Could not read correlation search %q after update: %s", name, err.Error()),
		)
		return
	}

	r.refreshState(ctx, readResp, &plan)
	plan.ID = types.StringValue(name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CorrelationSearchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CorrelationSearchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	app := state.App.ValueString()
	name := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting correlation search", map[string]interface{}{
		"name":  name,
		"owner": owner,
		"app":   app,
	})

	err := r.client.DeleteSavedSearch(owner, app, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Correlation Search",
			fmt.Sprintf("Could not delete correlation search %q: %s", name, err.Error()),
		)
		return
	}
}

func (r *CorrelationSearchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Set defaults for owner and app since they are not part of the import ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), "nobody")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), "SplunkEnterpriseSecuritySuite")...)
}

// buildParams converts the model into url.Values for the Splunk REST API.
func (r *CorrelationSearchResource) buildParams(ctx context.Context, model *CorrelationSearchResourceModel) url.Values {
	params := url.Values{}

	// Core parameters
	params.Set("name", model.Name.ValueString())
	params.Set("search", model.Search.ValueString())
	params.Set("description", model.Description.ValueString())

	// Scheduling
	params.Set("cron_schedule", model.CronSchedule.ValueString())
	params.Set("dispatch.earliest_time", model.DispatchEarliestTime.ValueString())
	params.Set("dispatch.latest_time", model.DispatchLatestTime.ValueString())
	params.Set("schedule_priority", model.Scheduling.ValueString())

	if model.IsScheduled.ValueBool() {
		params.Set("is_scheduled", "1")
	} else {
		params.Set("is_scheduled", "0")
	}

	if model.Disabled.ValueBool() {
		params.Set("disabled", "1")
	} else {
		params.Set("disabled", "0")
	}

	// Build the actions list
	var actions []string
	actions = append(actions, "correlationsearch")

	if model.NotableEnabled.ValueBool() {
		actions = append(actions, "notable")
	}
	if model.RiskEnabled.ValueBool() {
		actions = append(actions, "risk")
	}
	params.Set("actions", strings.Join(actions, ","))

	// ES correlation search action
	if model.CorrelationSearchEnabled.ValueBool() {
		params.Set("action.correlationsearch.enabled", "1")
	} else {
		params.Set("action.correlationsearch.enabled", "0")
	}
	if !model.CorrelationSearchLabel.IsNull() && !model.CorrelationSearchLabel.IsUnknown() {
		params.Set("action.correlationsearch.label", model.CorrelationSearchLabel.ValueString())
	}

	// Build annotations JSON for MITRE ATT&CK and frameworks
	annotations := r.buildAnnotations(ctx, model)
	if annotations != "" {
		params.Set("action.correlationsearch.annotations", annotations)
	}

	// Notable event action
	if model.NotableEnabled.ValueBool() {
		params.Set("action.notable", "1")
	} else {
		params.Set("action.notable", "0")
	}
	if !model.NotableRuleTitle.IsNull() && !model.NotableRuleTitle.IsUnknown() {
		params.Set("action.notable.param.rule_title", model.NotableRuleTitle.ValueString())
	}
	if !model.NotableRuleDescription.IsNull() && !model.NotableRuleDescription.IsUnknown() {
		params.Set("action.notable.param.rule_description", model.NotableRuleDescription.ValueString())
	}
	if !model.NotableSecurityDomain.IsNull() && !model.NotableSecurityDomain.IsUnknown() {
		params.Set("action.notable.param.security_domain", model.NotableSecurityDomain.ValueString())
	}
	if !model.NotableSeverity.IsNull() && !model.NotableSeverity.IsUnknown() {
		params.Set("action.notable.param.severity", model.NotableSeverity.ValueString())
	}
	if !model.NotableDrilldownName.IsNull() && !model.NotableDrilldownName.IsUnknown() {
		params.Set("action.notable.param.drilldown_name", model.NotableDrilldownName.ValueString())
	}
	if !model.NotableDrilldownSearch.IsNull() && !model.NotableDrilldownSearch.IsUnknown() {
		params.Set("action.notable.param.drilldown_search", model.NotableDrilldownSearch.ValueString())
	}
	if !model.NotableDefaultOwner.IsNull() && !model.NotableDefaultOwner.IsUnknown() {
		params.Set("action.notable.param.default_owner", model.NotableDefaultOwner.ValueString())
	}
	if !model.NotableRecommendedActions.IsNull() && !model.NotableRecommendedActions.IsUnknown() {
		params.Set("action.notable.param.recommended_actions", model.NotableRecommendedActions.ValueString())
	}

	// Risk action
	if model.RiskEnabled.ValueBool() {
		params.Set("action.risk", "1")
	} else {
		params.Set("action.risk", "0")
	}
	if !model.RiskMessage.IsNull() && !model.RiskMessage.IsUnknown() {
		params.Set("action.risk.param._risk_message", model.RiskMessage.ValueString())
	}
	if !model.RiskObjectField.IsNull() && !model.RiskObjectField.IsUnknown() {
		params.Set("action.risk.param._risk_object", model.RiskObjectField.ValueString())
	}
	if !model.RiskObjectType.IsNull() && !model.RiskObjectType.IsUnknown() {
		params.Set("action.risk.param._risk_object_type", model.RiskObjectType.ValueString())
	}
	params.Set("action.risk.param._risk_score", fmt.Sprintf("%d", model.RiskScore.ValueInt64()))

	// Alert settings
	params.Set("alert.severity", fmt.Sprintf("%d", model.AlertSeverity.ValueInt64()))
	if !model.AlertType.IsNull() && !model.AlertType.IsUnknown() {
		params.Set("alert_type", model.AlertType.ValueString())
	}
	if !model.AlertComparator.IsNull() && !model.AlertComparator.IsUnknown() {
		params.Set("alert_comparator", model.AlertComparator.ValueString())
	}
	if !model.AlertThreshold.IsNull() && !model.AlertThreshold.IsUnknown() {
		params.Set("alert_threshold", model.AlertThreshold.ValueString())
	}
	if model.AlertSuppress.ValueBool() {
		params.Set("alert.suppress", "1")
	} else {
		params.Set("alert.suppress", "0")
	}
	if !model.AlertSuppressPeriod.IsNull() && !model.AlertSuppressPeriod.IsUnknown() {
		params.Set("alert.suppress.period", model.AlertSuppressPeriod.ValueString())
	}
	if !model.AlertSuppressFields.IsNull() && !model.AlertSuppressFields.IsUnknown() {
		params.Set("alert.suppress.fields", model.AlertSuppressFields.ValueString())
	}

	return params
}

// buildAnnotations constructs the JSON annotations string for MITRE ATT&CK and framework mappings.
func (r *CorrelationSearchResource) buildAnnotations(ctx context.Context, model *CorrelationSearchResourceModel) string {
	annotationsMap := make(map[string][]string)

	if !model.MitreAttackIDs.IsNull() && !model.MitreAttackIDs.IsUnknown() {
		var ids []string
		model.MitreAttackIDs.ElementsAs(ctx, &ids, false)
		if len(ids) > 0 {
			annotationsMap["mitre_attack"] = ids
		}
	}

	if !model.KillChainPhases.IsNull() && !model.KillChainPhases.IsUnknown() {
		var phases []string
		model.KillChainPhases.ElementsAs(ctx, &phases, false)
		if len(phases) > 0 {
			annotationsMap["kill_chain_phases"] = phases
		}
	}

	if !model.CIS20.IsNull() && !model.CIS20.IsUnknown() {
		var controls []string
		model.CIS20.ElementsAs(ctx, &controls, false)
		if len(controls) > 0 {
			annotationsMap["cis20"] = controls
		}
	}

	if !model.NIST.IsNull() && !model.NIST.IsUnknown() {
		var controls []string
		model.NIST.ElementsAs(ctx, &controls, false)
		if len(controls) > 0 {
			annotationsMap["nist"] = controls
		}
	}

	if !model.AnalyticStory.IsNull() && !model.AnalyticStory.IsUnknown() {
		var stories []string
		model.AnalyticStory.ElementsAs(ctx, &stories, false)
		if len(stories) > 0 {
			annotationsMap["analytic_story"] = stories
		}
	}

	if len(annotationsMap) == 0 {
		return ""
	}

	jsonBytes, err := json.Marshal(annotationsMap)
	if err != nil {
		tflog.Warn(ctx, "Failed to marshal annotations JSON", map[string]interface{}{
			"error": err.Error(),
		})
		return ""
	}
	return string(jsonBytes)
}

// refreshState reads the API response and updates the model with server-side values.
func (r *CorrelationSearchResource) refreshState(ctx context.Context, response map[string]interface{}, model *CorrelationSearchResourceModel) {
	content, err := sdk.GetEntryContent(response)
	if err != nil {
		tflog.Warn(ctx, "Failed to extract entry content from response", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Core fields
	if v := sdk.ParseString(content, "search"); v != "" {
		model.Search = types.StringValue(v)
	}
	model.Description = types.StringValue(sdk.ParseString(content, "description"))
	model.Disabled = types.BoolValue(sdk.ParseBool(content, "disabled"))

	// Scheduling
	if v := sdk.ParseString(content, "cron_schedule"); v != "" {
		model.CronSchedule = types.StringValue(v)
	}
	model.DispatchEarliestTime = types.StringValue(sdk.ParseString(content, "dispatch.earliest_time"))
	model.DispatchLatestTime = types.StringValue(sdk.ParseString(content, "dispatch.latest_time"))
	model.Scheduling = types.StringValue(sdk.ParseString(content, "schedule_priority"))
	model.IsScheduled = types.BoolValue(sdk.ParseBool(content, "is_scheduled"))

	// ES correlation search
	model.CorrelationSearchEnabled = types.BoolValue(sdk.ParseBool(content, "action.correlationsearch.enabled"))
	model.CorrelationSearchLabel = types.StringValue(sdk.ParseString(content, "action.correlationsearch.label"))

	// Notable event action
	model.NotableEnabled = types.BoolValue(sdk.ParseBool(content, "action.notable"))
	model.NotableRuleTitle = types.StringValue(sdk.ParseString(content, "action.notable.param.rule_title"))
	model.NotableRuleDescription = types.StringValue(sdk.ParseString(content, "action.notable.param.rule_description"))
	model.NotableSecurityDomain = types.StringValue(sdk.ParseString(content, "action.notable.param.security_domain"))
	model.NotableSeverity = types.StringValue(sdk.ParseString(content, "action.notable.param.severity"))
	model.NotableDrilldownName = types.StringValue(sdk.ParseString(content, "action.notable.param.drilldown_name"))
	model.NotableDrilldownSearch = types.StringValue(sdk.ParseString(content, "action.notable.param.drilldown_search"))
	model.NotableDefaultOwner = types.StringValue(sdk.ParseString(content, "action.notable.param.default_owner"))
	model.NotableRecommendedActions = types.StringValue(sdk.ParseString(content, "action.notable.param.recommended_actions"))

	// Risk action
	model.RiskEnabled = types.BoolValue(sdk.ParseBool(content, "action.risk"))
	model.RiskScore = types.Int64Value(sdk.ParseInt(content, "action.risk.param._risk_score"))
	model.RiskMessage = types.StringValue(sdk.ParseString(content, "action.risk.param._risk_message"))
	model.RiskObjectField = types.StringValue(sdk.ParseString(content, "action.risk.param._risk_object"))
	model.RiskObjectType = types.StringValue(sdk.ParseString(content, "action.risk.param._risk_object_type"))

	// Parse annotations JSON for MITRE ATT&CK and frameworks
	r.refreshAnnotations(ctx, content, model)

	// Alert settings
	model.AlertSeverity = types.Int64Value(sdk.ParseInt(content, "alert.severity"))
	model.AlertType = types.StringValue(sdk.ParseString(content, "alert_type"))
	model.AlertComparator = types.StringValue(sdk.ParseString(content, "alert_comparator"))
	model.AlertThreshold = types.StringValue(sdk.ParseString(content, "alert_threshold"))
	model.AlertSuppress = types.BoolValue(sdk.ParseBool(content, "alert.suppress"))
	model.AlertSuppressPeriod = types.StringValue(sdk.ParseString(content, "alert.suppress.period"))
	model.AlertSuppressFields = types.StringValue(sdk.ParseString(content, "alert.suppress.fields"))

	// Read owner/app from ACL if available
	acl, aclErr := sdk.GetEntryACL(response)
	if aclErr == nil {
		if v := sdk.ParseString(acl, "owner"); v != "" {
			model.Owner = types.StringValue(v)
		}
		if v := sdk.ParseString(acl, "app"); v != "" {
			model.App = types.StringValue(v)
		}
	}
}

// refreshAnnotations parses the annotations JSON string and populates the list attributes.
func (r *CorrelationSearchResource) refreshAnnotations(ctx context.Context, content map[string]interface{}, model *CorrelationSearchResourceModel) {
	annotationsStr := sdk.ParseString(content, "action.correlationsearch.annotations")
	if annotationsStr == "" {
		return
	}

	var annotationsMap map[string]interface{}
	if err := json.Unmarshal([]byte(annotationsStr), &annotationsMap); err != nil {
		tflog.Warn(ctx, "Failed to parse annotations JSON", map[string]interface{}{
			"error":       err.Error(),
			"annotations": annotationsStr,
		})
		return
	}

	model.MitreAttackIDs = parseStringList(ctx, annotationsMap, "mitre_attack")
	model.KillChainPhases = parseStringList(ctx, annotationsMap, "kill_chain_phases")
	model.CIS20 = parseStringList(ctx, annotationsMap, "cis20")
	model.NIST = parseStringList(ctx, annotationsMap, "nist")
	model.AnalyticStory = parseStringList(ctx, annotationsMap, "analytic_story")
}

// parseStringList extracts a list of strings from an annotations map and returns a types.List.
func parseStringList(_ context.Context, data map[string]interface{}, key string) types.List {
	raw, ok := data[key]
	if !ok || raw == nil {
		return types.ListNull(types.StringType)
	}

	rawSlice, ok := raw.([]interface{})
	if !ok {
		return types.ListNull(types.StringType)
	}

	elements := make([]types.String, 0, len(rawSlice))
	for _, item := range rawSlice {
		if s, ok := item.(string); ok {
			elements = append(elements, types.StringValue(s))
		}
	}

	if len(elements) == 0 {
		return types.ListNull(types.StringType)
	}

	listVal, _ := types.ListValueFrom(context.Background(), types.StringType, elements)
	return listVal
}
