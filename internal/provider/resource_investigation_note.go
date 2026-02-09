package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Agentric/terraform-provider-splunkes/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &investigationNoteResource{}
	_ resource.ResourceWithConfigure   = &investigationNoteResource{}
	_ resource.ResourceWithImportState = &investigationNoteResource{}
)

// NewInvestigationNoteResource returns a new resource.Resource for ES investigation notes.
func NewInvestigationNoteResource() resource.Resource {
	return &investigationNoteResource{}
}

type investigationNoteResource struct {
	client *sdk.SplunkClient
}

type investigationNoteResourceModel struct {
	ID              types.String `tfsdk:"id"`
	InvestigationID types.String `tfsdk:"investigation_id"`
	Content         types.String `tfsdk:"content"`
	CreatedTime     types.String `tfsdk:"created_time"`
	ModifiedTime    types.String `tfsdk:"modified_time"`
}

func (r *investigationNoteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_investigation_note"
}

func (r *investigationNoteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a note on an investigation in Splunk Enterprise Security (ES v2 API).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The note ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"investigation_id": schema.StringAttribute{
				Description: "The parent investigation ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Description: "The note content text.",
				Required:    true,
			},
			"created_time": schema.StringAttribute{
				Description: "The time the note was created.",
				Computed:    true,
			},
			"modified_time": schema.StringAttribute{
				Description: "The time the note was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *investigationNoteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *investigationNoteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan investigationNoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	investigationID := plan.InvestigationID.ValueString()

	body := map[string]interface{}{
		"content": plan.Content.ValueString(),
	}

	tflog.Debug(ctx, "Creating investigation note", map[string]interface{}{
		"investigation_id": investigationID,
	})

	result, err := r.client.CreateInvestigationNote(investigationID, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating investigation note", err.Error())
		return
	}

	// ES v2 returns _key or id directly in the response JSON.
	noteID := sdk.ParseString(result, "_key")
	if noteID == "" {
		noteID = sdk.ParseString(result, "id")
	}
	if noteID == "" {
		resp.Diagnostics.AddError(
			"Error creating investigation note",
			"No _key or id returned in the API response.",
		)
		return
	}

	plan.ID = types.StringValue(noteID)

	// Read back the full state from the API.
	r.refreshState(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *investigationNoteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state investigationNoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading investigation note", map[string]interface{}{
		"investigation_id": state.InvestigationID.ValueString(),
		"note_id":          state.ID.ValueString(),
	})

	r.refreshState(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *investigationNoteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan investigationNoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state investigationNoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	investigationID := state.InvestigationID.ValueString()
	noteID := state.ID.ValueString()

	body := map[string]interface{}{
		"content": plan.Content.ValueString(),
	}

	tflog.Debug(ctx, "Updating investigation note", map[string]interface{}{
		"investigation_id": investigationID,
		"note_id":          noteID,
	})

	_, err := r.client.UpdateInvestigationNote(investigationID, noteID, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating investigation note", err.Error())
		return
	}

	plan.ID = types.StringValue(noteID)
	plan.InvestigationID = types.StringValue(investigationID)

	r.refreshState(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *investigationNoteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state investigationNoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	investigationID := state.InvestigationID.ValueString()
	noteID := state.ID.ValueString()

	tflog.Debug(ctx, "Deleting investigation note", map[string]interface{}{
		"investigation_id": investigationID,
		"note_id":          noteID,
	})

	err := r.client.DeleteInvestigationNote(investigationID, noteID)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting investigation note", err.Error())
		return
	}
}

// ImportState handles "investigation_id/note_id" composite import IDs.
func (r *investigationNoteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the format 'investigation_id/note_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("investigation_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// refreshState reads all notes for the investigation and finds the specific note by ID.
// The ES v2 API ReadInvestigationNotes returns all notes for an investigation.
func (r *investigationNoteResource) refreshState(_ context.Context, model *investigationNoteResourceModel, diagnostics *diag.Diagnostics) {
	investigationID := model.InvestigationID.ValueString()
	noteID := model.ID.ValueString()

	result, err := r.client.ReadInvestigationNotes(investigationID)
	if err != nil {
		diagnostics.AddError("Error reading investigation notes", err.Error())
		return
	}

	// Find the specific note by ID in the response.
	data, findErr := findNoteByID(result, noteID)
	if findErr != nil {
		diagnostics.AddError(
			"Error finding investigation note",
			fmt.Sprintf("Note %s not found in investigation %s: %s", noteID, investigationID, findErr.Error()),
		)
		return
	}

	model.ID = types.StringValue(sdk.ParseString(data, "_key"))
	if model.ID.ValueString() == "" {
		model.ID = types.StringValue(sdk.ParseString(data, "id"))
	}

	if content := sdk.ParseString(data, "content"); content != "" {
		model.Content = types.StringValue(content)
	}

	if createdTime := sdk.ParseString(data, "create_time"); createdTime != "" {
		model.CreatedTime = types.StringValue(createdTime)
	} else {
		model.CreatedTime = types.StringNull()
	}

	if modifiedTime := sdk.ParseString(data, "modify_time"); modifiedTime != "" {
		model.ModifiedTime = types.StringValue(modifiedTime)
	} else {
		model.ModifiedTime = types.StringNull()
	}
}

// findNoteByID searches an ES v2 notes list response for a note matching the given ID.
func findNoteByID(result map[string]interface{}, noteID string) (map[string]interface{}, error) {
	// Try common list wrapper keys used by ES v2.
	for _, key := range []string{"data", "items", "results", "entry", ""} {
		var items []interface{}
		if key == "" {
			// Check if the result itself is the note (single object with _key).
			if k := sdk.ParseString(result, "_key"); k == noteID {
				return result, nil
			}
			if k := sdk.ParseString(result, "id"); k == noteID {
				return result, nil
			}
			continue
		}
		raw, ok := result[key]
		if !ok {
			continue
		}
		items, ok = raw.([]interface{})
		if !ok {
			continue
		}
		for _, item := range items {
			note, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			k := sdk.ParseString(note, "_key")
			if k == "" {
				k = sdk.ParseString(note, "id")
			}
			if k == noteID {
				return note, nil
			}
		}
	}

	return nil, fmt.Errorf("note with ID %s not found", noteID)
}
