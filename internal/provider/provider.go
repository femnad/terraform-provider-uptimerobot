package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	uptimerobot "terraform-provider-uptimerobot/api"
)

const (
	apiKeyAttributeName = "api_key"
	apiKeyEnv           = "UPTIMEROBOT_API_KEY"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &uptimerobotProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &uptimerobotProvider{
			version: version,
		}
	}
}

type uptimerobotProvider struct {
	version string
}

type uptimerobotProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func (p uptimerobotProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "uptimerobot"
	resp.Version = p.version
}

func (p uptimerobotProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			apiKeyAttributeName: schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p uptimerobotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring UptimeRobot client")

	var config uptimerobotProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(apiKeyAttributeName),
			"Unknown UptimeRobot API key",
			fmt.Sprintf("The provider cannot create the UptimeRobot API client as there is an unknown "+
				"configuration value for the UptimeRobot API key. Either target apply the source of the value first, "+
				"set the value statically in the configuration, or use the %s environment variable.", apiKeyEnv),
		)
	}

	apiKey := os.Getenv(apiKeyEnv)
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root(apiKeyAttributeName),
			"Missing UptimeRobot API key",
			fmt.Sprintf("The provider cannot create the UptimeRobot API client as there is a missing or "+
				"empty value for the UptimeRobot API key. Set the %s value in the configuration or use the %s "+
				"environment variable. If either is already set, ensure the value is not empty.",
				apiKeyAttributeName, apiKeyEnv),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, apiKeyAttributeName, apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, apiKeyAttributeName)

	tflog.Debug(ctx, "Creating UptimeRobot client")

	c, err := uptimerobot.New(apiKey)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create UptimeRobot API Client",
			"An unexpected error occurred when creating the UptimeRobot API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"UptimeRobot Client Error: "+err.Error())
	}

	resp.DataSourceData = c
	resp.ResourceData = c

	tflog.Info(ctx, "Configured UptimeRobot client", map[string]any{"success": true})
}

func (p uptimerobotProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAlertContactsDataSource,
	}
}

func (p uptimerobotProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
