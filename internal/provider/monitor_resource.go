package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	uptimerobot "terraform-provider-uptimerobot/api"
)

var (
	_ resource.Resource              = &monitorResource{}
	_ resource.ResourceWithConfigure = &monitorResource{}
)

type monitorResourceModel struct {
	FriendlyName types.String `tfsdk:"friendly_name"`
	ID           types.Int64  `tfsdk:"id"`
	Interval     types.Int64  `tfsdk:"interval"`
	Timeout      types.Int64  `tfsdk:"timeout"`
	Type         types.String `tfsdk:"type"`
	URL          types.String `tfsdk:"url"`
}

type monitorResource struct {
	client *uptimerobot.Client
}

func NewMonitorResource() resource.Resource {
	return &monitorResource{}
}

func (r *monitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *monitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *monitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	var validMonitorTypes []string
	for t, _ := range uptimerobot.MonitorTypes {
		validMonitorTypes = append(validMonitorTypes, t)
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"friendly_name": schema.StringAttribute{
				Required: true,
			},
			"url": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.OneOf(validMonitorTypes...)},
			},
			"interval": schema.Int64Attribute{
				Optional: true,
			},
			"timeout": schema.Int64Attribute{
				Optional: true,
			},
		},
	}
}

func (r *monitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan monitorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	monitor := uptimerobot.Monitor{
		FriendlyName: plan.FriendlyName.ValueString(),
		URL:          plan.URL.ValueString(),
		Interval:     plan.Interval.ValueInt64(),
		Timeout:      plan.Interval.ValueInt64(),
	}
	intType, err := uptimerobot.MonitorTypeToInt(plan.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating monitor",
			"Could not create order, error determining monitor type: "+err.Error())
		return
	}
	monitor.Type = intType

	monitor, err = r.client.CreateMonitor(monitor)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating monitor",
			"Could not create order, unexpected error: "+err.Error())
		return
	}

	plan.ID = types.Int64Value(monitor.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *monitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state monitorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueInt64()
	monitor, err := r.client.GetMonitor(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading UptimeRobot monitor",
			fmt.Sprintf("Could not read UptimeRobot monitor with ID %d: %v", id, err))
		return
	}

	monitorType, err := uptimerobot.MonitorTypeToStr(monitor.Type)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading UptimeRobot monitor",
			fmt.Sprintf("Could not read UptimeRobot monitor with ID %d, error determining monitoro type: %v",
				id, err))
		return
	}

	state.Type = types.StringValue(monitorType)
	state.FriendlyName = types.StringValue(monitor.FriendlyName)
	state.Interval = types.Int64Value(monitor.Interval)
	state.Timeout = types.Int64Value(monitor.Timeout)
	state.URL = types.StringValue(monitor.URL)
}

func (r *monitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *monitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
