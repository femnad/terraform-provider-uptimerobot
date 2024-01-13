package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func (p uptimerobotProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "uptimerobot"
	resp.Version = p.version
}

func (p uptimerobotProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p uptimerobotProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
}

func (p uptimerobotProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p uptimerobotProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
