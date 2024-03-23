package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	uptimerobot "terraform-provider-uptimerobot/api"
)

var (
	_ resource.Resource                = &monitorResource{}
	_ resource.ResourceWithConfigure   = &monitorResource{}
	_ resource.ResourceWithImportState = &monitorResource{}
)

type monitorResourceModel struct {
	FriendlyName  types.String                      `tfsdk:"friendly_name"`
	ID            types.String                      `tfsdk:"id"`
	Interval      types.Int64                       `tfsdk:"interval"`
	LastUpdated   types.String                      `tfsdk:"last_updated"`
	Timeout       types.Int64                       `tfsdk:"timeout"`
	Type          types.String                      `tfsdk:"type"`
	URL           types.String                      `tfsdk:"url"`
	AlertContacts []uptimerobot.MonitorAlertContact `tfsdk:"alert_contact"`
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

func (r *monitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *monitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *monitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	var validMonitorTypes []string
	for t := range uptimerobot.MonitorTypes {
		validMonitorTypes = append(validMonitorTypes, t)
	}

	resp.Schema = schema.Schema{
		Description: "Manages a monitor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the monitor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the monitor.",
				Computed:    true,
			},
			"friendly_name": schema.StringAttribute{
				Description: "Friendly name of the monitor",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL to monitor",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the monitor",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf(validMonitorTypes...)},
			},
			"interval": schema.Int64Attribute{
				Description: "Monitor check interval",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Monitor check timeout",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"alert_contact": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Alert contact ID",
							Required:    true,
						},
						"threshold": schema.Int64Attribute{
							Computed:    true,
							Description: "Threshold for alerting (minutes)",
							Default:     int64default.StaticInt64(0),
						},
						"recurrence": schema.Int64Attribute{
							Computed:    true,
							Description: "Repetition interval for alerts (minutes)",
							Default:     int64default.StaticInt64(0),
						},
					},
				},
			},
		},
	}
}

func monitorFromPlan(plan monitorResourceModel) (uptimerobot.Monitor, error) {
	monitor := uptimerobot.Monitor{
		FriendlyName: plan.FriendlyName.ValueString(),
		URL:          plan.URL.ValueString(),
		Interval:     plan.Interval.ValueInt64(),
		Timeout:      plan.Interval.ValueInt64(),
	}
	monitor.AlertContacts = uptimerobot.SerializeMonitorAlertContacts(plan.AlertContacts)

	intType, err := uptimerobot.MonitorTypeToInt(plan.Type.ValueString())
	if err != nil {
		return monitor, err
	}

	monitor.Type = intType
	return monitor, nil
}

func (r *monitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan monitorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	monitor, err := monitorFromPlan(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating monitor",
			"Could not create monitor, error determining monitor type: "+err.Error())
		return
	}

	monitor, err = r.client.CreateMonitor(monitor)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating monitor",
			"Could not create monitor, unexpected error: "+err.Error())
		return
	}

	planID := strconv.Itoa(int(monitor.ID))
	plan.ID = types.StringValue(planID)

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

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading UptimeRobot monitor",
			fmt.Sprintf("Could not determine ID of the monitor %d: %v", id, err))
		return
	}
	monitorId := int64(id)

	monitor, err := r.client.GetMonitor(monitorId)
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
	var plan monitorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitor, err := monitorFromPlan(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating monitor",
			fmt.Sprintf("Could not update monitor %v", err))
	}

	monitorID, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating monitor",
			fmt.Sprintf("Could not determine monitor ID %v", err))
	}
	monitor.ID = int64(monitorID)

	monitor, err = r.client.UpdateMonitor(monitor)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating monitor",
			fmt.Sprintf("Could not update monitor %v", err))
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *monitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state monitorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitorID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting monitor",
			fmt.Sprintf("Could not determine monitor ID: %v", err))
		return
	}

	err = r.client.DeleteMonitor(int64(monitorID))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting monitor",
			fmt.Sprintf("Could not delete monitor with ID %d: %v", monitorID, err))
		return
	}
}
