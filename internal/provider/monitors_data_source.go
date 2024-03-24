package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	uptimerobot "terraform-provider-uptimerobot/api"
)

var (
	_ datasource.DataSource              = &monitorsDataSource{}
	_ datasource.DataSourceWithConfigure = &monitorsDataSource{}
)

type monitorsDataSource struct {
	client *uptimerobot.Client
}

type monitorsDataSourceModel struct {
	Monitors []monitorModel `tfsdk:"monitors"`
}

type monitorModel struct {
	ID           types.String `tfsdk:"id"`
	FriendlyName types.String `tfsdk:"friendly_name"`
	URL          types.String `tfsdk:"url"`
	Type         types.String `tfsdk:"type"`
	Interval     types.Int64  `tfsdk:"interval"`
	Timeout      types.Int64  `tfsdk:"timeout"`
}

func (d *monitorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *monitorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

func (d *monitorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches monitors defined for the account",
		Attributes: map[string]schema.Attribute{
			"monitors": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The identifier of the monitor.",
							Computed:    true,
						},
						"friendly_name": schema.StringAttribute{
							Description: "Friendly name for the monitor.",
							Computed:    true,
						},
						"url": schema.StringAttribute{
							Description: "The URL or IP the monitor.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the monitor.",
							Computed:    true,
						},
						"interval": schema.Int64Attribute{
							Description: "The interval for the monitoring check.",
							Computed:    true,
						},
						"timeout": schema.Int64Attribute{
							Description: "Timeout for the monitoring check.",
							Computed:    true,
						},
					},
				},
			},
		}}
}

func (d *monitorsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state monitorsDataSourceModel

	monitors, err := d.client.GetMonitors()
	if err != nil {
		resp.Diagnostics.AddError("Unable to read UptimeRobot monitors", err.Error())
		return
	}

	state.Monitors = make([]monitorModel, 0)
	for _, monitor := range monitors {
		monitorID := strconv.Itoa(int(monitor.ID))
		var monitorType string
		monitorType, err = uptimerobot.MonitorTypeToStr(monitor.ID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read UptimeRobot monitor, error determining monitor type",
				err.Error())
			return
		}

		stateMonitor := monitorModel{
			ID:           types.StringValue(monitorID),
			FriendlyName: types.StringValue(monitor.FriendlyName),
			URL:          types.StringValue(monitor.URL),
			Type:         types.StringValue(monitorType),
			Interval:     types.Int64Value(monitor.Interval),
			Timeout:      types.Int64Value(monitor.Timeout),
		}

		state.Monitors = append(state.Monitors, stateMonitor)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func NewMonitorsDataSource() datasource.DataSource {
	return &monitorsDataSource{}
}
