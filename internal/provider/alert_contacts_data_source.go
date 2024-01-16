package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	uptimerobot "terraform-provider-uptimerobot/api"
)

var (
	_ datasource.DataSource              = &alertContactsDataSource{}
	_ datasource.DataSourceWithConfigure = &alertContactsDataSource{}
)

type alertContactsDataSource struct {
	client *uptimerobot.Client
}

type alertContactsDataSourceModel struct {
	AlertContacts []alertContactModel `tfsdk:"alert_contacts"`
}

type alertContactModel struct {
	ID           types.String `tfsdk:"id"`
	FriendlyName types.String `tfsdk:"friendly_name"`
	Type         types.Int64  `tfsdk:"type"`
	Status       types.Int64  `tfsdk:"status"`
	Value        types.String `tfsdk:"value"`
}

func (d *alertContactsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *alertContactsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_contacts"
}

func (d *alertContactsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches alert contacts defined for the account",
		Attributes: map[string]schema.Attribute{
			"alert_contacts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Identifier of the alert contact.",
							Computed:    true,
						},
						"friendly_name": schema.StringAttribute{
							Description: "Friendly name for the alert contact.",
							Computed:    true,
						},
						"type": schema.Int64Attribute{
							Description: "Type of the alert contact.",
							Computed:    true,
						},
						"status": schema.Int64Attribute{
							Description: "Status of the alert contact.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the alert contact.",
							Computed:    true,
						},
					},
				},
			},
		}}
}

func (d *alertContactsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state alertContactsDataSourceModel

	alertContacts, err := d.client.GetAlertContacts()
	if err != nil {
		resp.Diagnostics.AddError("Unable to read UptimeRobot alert contacts", err.Error())
		return
	}

	for _, contact := range alertContacts {
		alertContactState := alertContactModel{
			ID:           types.StringValue(contact.ID),
			FriendlyName: types.StringValue(contact.FriendlyName),
			Type:         types.Int64Value(contact.Type),
			Status:       types.Int64Value(contact.Status),
			Value:        types.StringValue(contact.Value),
		}

		state.AlertContacts = append(state.AlertContacts, alertContactState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func NewAlertContactsDataSource() datasource.DataSource {
	return &alertContactsDataSource{}
}
