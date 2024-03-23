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
	_ datasource.DataSource              = &accountDetailsDataSource{}
	_ datasource.DataSourceWithConfigure = &accountDetailsDataSource{}
)

type accountDetailsDataSource struct {
	client *uptimerobot.Client
}

type accountDetailsDataSourceModel struct {
	Email           types.String `tfsdk:"email"`
	MonitorLimit    types.Int64  `tfsdk:"monitor_limit"`
	MonitorInterval types.Int64  `tfsdk:"monitor_interval"`
	UpMonitors      types.Int64  `tfsdk:"up_monitors"`
	DownMonitors    types.Int64  `tfsdk:"down_monitors"`
	PausedMonitors  types.Int64  `tfsdk:"paused_monitors"`
}

func (d *accountDetailsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *accountDetailsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_details"
}

func (d *accountDetailsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details for the account",
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Description: "The account email.",
				Computed:    true,
			},
			"monitor_limit": schema.Int64Attribute{
				Description: "The maximum number of monitors that can be created for the account.",
				Computed:    true,
			},
			"monitor_interval": schema.Int64Attribute{
				Description: "The minimum monitoring interval supported by the account (in seconds).",
				Computed:    true,
			},
			"up_monitors": schema.Int64Attribute{
				Description: "Number of up monitors.",
				Computed:    true,
			},
			"down_monitors": schema.Int64Attribute{
				Description: "Number of down monitors.",
				Computed:    true,
			},
			"paused_monitors": schema.Int64Attribute{
				Description: "Number of paused monitors.",
				Computed:    true,
			},
		},
	}
}

func (d *accountDetailsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accountDetailsDataSourceModel

	account, err := d.client.GetAccount()
	if err != nil {
		resp.Diagnostics.AddError("Unable to read UptimeRobot account details", err.Error())
		return
	}

	state.Email = types.StringValue(account.Email)
	state.MonitorLimit = types.Int64Value(account.MonitorLimit)
	state.MonitorInterval = types.Int64Value(account.MonitorInterval)
	state.UpMonitors = types.Int64Value(account.UpMonitors)
	state.DownMonitors = types.Int64Value(account.DownMonitors)
	state.PausedMonitors = types.Int64Value(account.PausedMonitors)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func NewAccountDetailsDataSource() datasource.DataSource {
	return &accountDetailsDataSource{}
}
