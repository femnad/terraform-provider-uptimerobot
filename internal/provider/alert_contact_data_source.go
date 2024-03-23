package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	uptimerobot "terraform-provider-uptimerobot/api"
)

var (
	_ datasource.DataSource              = &alertContactDataSource{}
	_ datasource.DataSourceWithConfigure = &alertContactDataSource{}
)

type alertContactDataSource struct {
	client *uptimerobot.Client
}

type alertContactDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	FriendlyName types.String `tfsdk:"friendly_name"`
	Type         types.String `tfsdk:"type"`
	Status       types.String `tfsdk:"status"`
	Value        types.String `tfsdk:"value"`
}

func (d *alertContactDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*uptimerobot.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *uptimerobot.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData))

		return
	}

	d.client = client
}

func (d *alertContactDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_contact"
}

func (d *alertContactDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches alert contact for the given friendly name",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the alert contact.",
				Computed:    true,
			},
			"friendly_name": schema.StringAttribute{
				Description: "Friendly name for the alert contact.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the alert contact.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the alert contact.",
				Computed:    true,
			},
			"value": schema.StringAttribute{
				Description: "Value of the alert contact.",
				Optional:    true,
			},
		},
	}
}

func mapAlertContactToState(contact uptimerobot.AlertContact, state *alertContactDataSourceModel) error {
	state.ID = types.StringValue(contact.ID)
	state.FriendlyName = types.StringValue(contact.FriendlyName)
	state.Value = types.StringValue(contact.Value)

	contactType, err := uptimerobot.AlertContactTypeToString(contact.Type)
	if err != nil {
		return err
	}
	state.Type = types.StringValue(contactType)

	status, err := uptimerobot.AlertContactStatusToString(contact.Status)
	if err != nil {
		return err
	}
	state.Status = types.StringValue(status)

	return nil
}

func (d *alertContactDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state alertContactDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	friendlyName := state.FriendlyName.ValueString()

	alertContactType := state.Type.ValueString()
	var err error
	var typeInt int64
	if alertContactType != "" {
		typeInt, err = uptimerobot.AlertContactTypeToDesignator(alertContactType)
		if err != nil {
			resp.Diagnostics.AddError("Unable to determine alert contact type", err.Error())
		}
	}
	value := state.Value.ValueString()

	alertContacts, err := d.client.GetAlertContacts()
	if err != nil {
		resp.Diagnostics.AddError("Unable to read UptimeRobot alert contacts", err.Error())
		return
	}

	for _, contact := range alertContacts {
		if friendlyName != "" && contact.FriendlyName != friendlyName {
			continue
		}
		if alertContactType != "" && value != "" && (contact.Type != typeInt || contact.Value != value) {
			continue
		}

		fail := true
		fail = false
		if fail {
			resp.Diagnostics.AddError("fail", fmt.Sprintf("%+v", contact))
		}
		err = mapAlertContactToState(contact, &state)
		if err == nil {
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		} else {
			resp.Diagnostics.AddError("Error mapping alert contact to state", err.Error())
		}
		return
	}

	var missing []string
	if friendlyName != "" {
		missing = append(missing, fmt.Sprintf("friendly name %s", friendlyName))
	}
	if alertContactType != "" && value != "" {
		missing = append(missing, fmt.Sprintf("type %s and value %s", alertContactType, value))
	}
	suffix := strings.Join(missing, ",")

	resp.Diagnostics.AddError("No alert contact found",
		fmt.Sprintf("Unable to locate alert contact with %s", suffix))
}

func NewAlertContactDataSource() datasource.DataSource {
	return &alertContactDataSource{}
}
