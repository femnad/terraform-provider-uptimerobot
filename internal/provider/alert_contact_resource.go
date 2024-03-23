package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	uptimerobot "terraform-provider-uptimerobot/api"
)

var (
	_ resource.Resource                = &alertContactResource{}
	_ resource.ResourceWithConfigure   = &alertContactResource{}
	_ resource.ResourceWithImportState = &alertContactResource{}
)

type alertContactResourceModel struct {
	FriendlyName types.String `tfsdk:"friendly_name"`
	ID           types.String `tfsdk:"id"`
	Status       types.String `tfsdk:"status"`
	Type         types.String `tfsdk:"type"`
	LastUpdated  types.String `tfsdk:"last_updated"`
	Value        types.String `tfsdk:"value"`
}

type alertContactResource struct {
	client *uptimerobot.Client
}

func NewAlertContactResource() resource.Resource {
	return &alertContactResource{}
}

func (a *alertContactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (a *alertContactResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*uptimerobot.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *uptimerobot.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	a.client = client
}

func (a *alertContactResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_contact"
}

func (a *alertContactResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	var validTypes []string
	for _, val := range uptimerobot.AlertContactTypes {
		validTypes = append(validTypes, val)
	}

	resp.Schema = schema.Schema{
		Description: "Manages an alert contact.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the alert contact",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the alert contact.",
				Computed:    true,
			},
			"friendly_name": schema.StringAttribute{
				Description: "Friendly name for the alert contact",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of alert contact",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of alert contact",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf(validTypes...)},
			},
			"value": schema.StringAttribute{
				Description: "Alert contact's contact details",
				Required:    true,
			},
		},
	}
}

func alertContactFromPlan(plan alertContactResourceModel) (uptimerobot.AlertContact, error) {
	var contact uptimerobot.AlertContact
	contactType, err := uptimerobot.AlertContactTypeToDesignator(plan.Type.ValueString())
	if err != nil {
		return contact, err
	}

	return uptimerobot.AlertContact{
		FriendlyName: plan.FriendlyName.ValueString(),
		Type:         contactType,
		Value:        plan.Value.ValueString(),
	}, nil
}

func updateFromAlertContact(model alertContactResourceModel, contact *uptimerobot.AlertContact) error {
	model.FriendlyName = types.StringValue(contact.FriendlyName)
	model.Value = types.StringValue(contact.Value)

	contactType, err := uptimerobot.AlertContactTypeToString(contact.Type)
	if err != nil {
		return err
	}

	status, err := uptimerobot.AlertContactStatusToString(contact.Status)
	if err != nil {
		return err
	}

	model.Status = types.StringValue(status)
	model.Type = types.StringValue(contactType)
	return nil
}

func (a *alertContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertContactResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	alertContact, err := alertContactFromPlan(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert contact from plan", err.Error())
		return
	}

	contact, err := a.client.CreateAlertContact(alertContact)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert contact", err.Error())
		return
	}
	plan.ID = types.StringValue(contact.ID)

	status, err := uptimerobot.AlertContactStatusToString(contact.Status)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up alert contact status", err.Error())
		return
	}
	plan.Status = types.StringValue(status)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (a *alertContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertContactResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	contacts, err := a.client.GetAlertContacts()
	if err != nil {
		resp.Diagnostics.AddError("Error getting alert contacts", err.Error())
		return
	}

	stateType := state.Type.ValueString()
	stateValue := state.Value.ValueString()
	typeInt, err := uptimerobot.AlertContactTypeToDesignator(stateType)
	if err != nil {
		resp.Diagnostics.AddError("Error determining alert contacts type", err.Error())
		return
	}

	var c *uptimerobot.AlertContact
	for _, contact := range contacts {
		if contact.Type == typeInt && contact.Value == stateValue {
			c = &contact
			break
		}
	}

	if c == nil {
		resp.Diagnostics.AddError("Unable to find alert contact", fmt.Sprintf(
			"No alert contact with type %s and value %s exists", stateType, stateValue))
		return
	}

	err = updateFromAlertContact(state, c)
	if err != nil {
		resp.Diagnostics.AddError("Unable updating state from alert contact", err.Error())
		return
	}
}

func (a *alertContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertContactResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	contactId := plan.ID.ValueString()
	alertContact, err := alertContactFromPlan(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error getting alert contact from plan", err.Error())
		return
	}
	alertContact.ID = contactId

	updated, err := a.client.UpdateAlertContact(alertContact)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert contact from plan", fmt.Sprintf(
			"Error updating alert contact for ID %s: %v", contactId, err))
		return
	}

	status, err := uptimerobot.AlertContactStatusToString(updated.Status)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert contact from plan", fmt.Sprintf(
			"Error determining alert contact status for ID %s: %v", contactId, err))
		return
	}
	plan.Status = types.StringValue(status)

	plan.FriendlyName = types.StringValue(updated.FriendlyName)
	plan.Value = types.StringValue(updated.Value)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (a *alertContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertContactResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	contactID := state.ID.ValueString()
	err := a.client.DeleteAlertContact(uptimerobot.AlertContact{ID: contactID})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting alert contact", fmt.Sprintf(
			"Could not delete alert contact with ID %s: %v", contactID, err))
		return
	}
}
