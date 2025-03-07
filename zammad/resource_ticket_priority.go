// Copyright 2022 The Terraform Provider for Zammad Authors
// spdx-license-identifier: apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zammad

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/o11ydev/terraform-provider-zammad/internal/client"
)

func NewZammadTicketPriority() resource.Resource {
	return &resourceTicketPriority{}
}

type resourceTicketPriority struct {
	client *client.Client
}

// Ticket Priority Resource schema
func (r resourceTicketPriority) GetSchema(_ context.Context) (schema.Schema, diag.Diagnostics) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"note": schema.StringAttribute{
				Optional: true,
			},
			"ui_icon": schema.StringAttribute{
				Optional: true,
			},
			"ui_color": schema.StringAttribute{
				Optional: true,
			},
			"default_create": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"active": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{&defaultTrue{}, boolplanmodifier.UseStateForUnknown()},
			},
			"created_by_id": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"updated_by_id": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}, nil
}

func (r *resourceTicketPriority) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ticket_priority"
}

func (r *resourceTicketPriority) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client, _ = req.ProviderData.(*client.Client)
}

// Create a new resource
func (r resourceTicketPriority) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan TicketPriority
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tpreq := &client.TicketPriority{
		Name:          plan.Name.ValueString(),
		Note:          plan.Note.ValueString(),
		UIColor:       plan.UIColor.ValueString(),
		UIIcon:        plan.UIIcon.ValueString(),
		Active:        plan.Active.ValueBool(),
		DefaultCreate: plan.DefaultCreate.ValueBool(),
	}

	tp, err := r.client.CreateTicketPriority(tpreq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ticket_priotiry",
			"Could not create ticket_priority, unexpected error: "+err.Error(),
		)
		return
	}

	result := TicketPriority{
		ID:            types.StringValue(strconv.Itoa(tp.ID)),
		Name:          types.StringValue(tp.Name),
		Note:          types.StringValue(tp.Note),
		UIColor:       types.StringValue(tp.UIColor),
		UIIcon:        types.StringValue(tp.UIIcon),
		Active:        types.BoolValue(tp.Active),
		DefaultCreate: types.BoolValue(tp.DefaultCreate),
		CreatedByID:   types.Int64Value(int64(tp.CreatedByID)),
		UpdatedByID:   types.Int64Value(int64(tp.UpdatedByID)),
		CreatedAt:     types.StringValue(tp.CreatedAt),
		UpdatedAt:     types.StringValue(tp.UpdatedAt),
	}
	if plan.UIColor.IsNull() && tp.UIColor == "" {
		result.UIColor = types.StringNull()
	}
	if plan.UIIcon.IsNull() && tp.UIIcon == "" {
		result.UIIcon = types.StringNull()
	}
	if plan.Note.IsNull() && tp.Note == "" {
		result.Note = types.StringNull()
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r resourceTicketPriority) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TicketPriority
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tpID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ID",
			"Could convert id "+state.ID.ValueString()+": "+err.Error(),
		)
		return

	}

	newtp, err := r.client.GetTicketPriority(tpID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ticket_priority",
			"Could not read ticket_priority "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(newtp.Name)
	state.CreatedAt = types.StringValue(newtp.CreatedAt)
	state.Active = types.BoolValue(newtp.Active)
	state.DefaultCreate = types.BoolValue(newtp.DefaultCreate)
	state.UpdatedAt = types.StringValue(newtp.UpdatedAt)
	state.UpdatedByID = types.Int64Value(int64(newtp.UpdatedByID))
	state.CreatedByID = types.Int64Value(int64(newtp.CreatedByID))
	if state.UIColor.IsNull() && newtp.UIColor == "" {
		state.UIColor = types.StringNull()
	} else {
		state.UIColor = types.StringValue(newtp.UIColor)
	}
	if state.UIIcon.IsNull() && newtp.UIIcon == "" {
		state.UIIcon = types.StringNull()
	} else {
		state.UIIcon = types.StringValue(newtp.UIIcon)
	}
	if state.Note.IsNull() && newtp.Note == "" {
		state.Note = types.StringNull()
	} else {
		state.Note = types.StringValue(newtp.Note)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r resourceTicketPriority) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TicketPriority
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state TicketPriority
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tpID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ID",
			"Could convert id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	updatedTP := &client.TicketPriority{
		ID:            tpID,
		Name:          plan.Name.ValueString(),
		Note:          plan.Note.ValueString(),
		UIColor:       plan.UIColor.ValueString(),
		UIIcon:        plan.UIIcon.ValueString(),
		Active:        plan.Active.ValueBool(),
		DefaultCreate: plan.DefaultCreate.ValueBool(),
	}

	tp, err := r.client.UpdateTicketPriority(updatedTP)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error update order",
			"Could not update orderID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	result := TicketPriority{
		ID:            types.StringValue(strconv.Itoa(tp.ID)),
		Name:          types.StringValue(tp.Name),
		Note:          types.StringValue(tp.Note),
		UIColor:       types.StringValue(tp.UIColor),
		UIIcon:        types.StringValue(tp.UIIcon),
		Active:        types.BoolValue(tp.Active),
		DefaultCreate: types.BoolValue(tp.DefaultCreate),
		CreatedByID:   types.Int64Value(int64(tp.CreatedByID)),
		UpdatedByID:   types.Int64Value(int64(tp.UpdatedByID)),
		CreatedAt:     types.StringValue(tp.CreatedAt),
		UpdatedAt:     types.StringValue(tp.UpdatedAt),
	}
	if plan.UIColor.IsNull() && tp.UIColor == "" {
		result.UIColor = types.StringNull()
	}
	if plan.UIIcon.IsNull() && tp.UIIcon == "" {
		result.UIIcon = types.StringNull()
	}
	if plan.Note.IsNull() && tp.Note == "" {
		result.Note = types.StringNull()
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceTicketPriority) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TicketPriority
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tpID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ID",
			"Could convert id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	err = r.client.DeleteTicketPriority(&client.TicketPriority{ID: tpID})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting ticket_priority",
			"Could not delete ticket_priority "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

// Import resource
func (r resourceTicketPriority) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type defaultTrue struct{}

func (m defaultTrue) Description(ctx context.Context) string {
	return "If value is not configured, defaults to true"
}

func (m defaultTrue) MarkdownDescription(ctx context.Context) string {
	return "If value is not configured, defaults to `true`"
}

func (m defaultTrue) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	var str types.Bool
	diags := tfsdk.ValueAs(ctx, req.PlanValue, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if !str.IsUnknown() {
		return
	}

	resp.PlanValue = types.BoolValue(true)
}
