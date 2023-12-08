package mailjet

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/mailjet/mailjet-apiv3-go/v4"
)

var (
	_ provider.Provider = &mailjetProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mailjetProvider{
			version: version,
		}
	}
}

type mailjetProvider struct {
	version string
}

type mailjetProviderModel struct {
	BaseURL       types.String `tfsdk:"base_url"`
	PublicAPIKey  types.String `tfsdk:"api_key_public"`
	PrivateAPIKey types.String `tfsdk:"api_key_private"`
}

func (p *mailjetProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mailjet"
	resp.Version = p.version
}

func (p *mailjetProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL of the Mailjet API. Default to https://api.mailjet.com/v3.",
			},
			"api_key_public": schema.StringAttribute{
				Optional:    true,
				Description: "Public API key for Mailjet. Default to the value of the `MJ_APIKEY_PUBLIC` environment variable.",
			},
			"api_key_private": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Private API key for Mailjet. Default to the value of the `MJ_APIKEY_PRIVATE` environment variable.",
			},
		},
	}
}

func (p *mailjetProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config mailjetProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := "https://api.mailjet.com/v3"

	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	apiKeyPublic := os.Getenv("MJ_APIKEY_PUBLIC")
	if !config.PublicAPIKey.IsNull() {
		apiKeyPublic = config.PublicAPIKey.ValueString()
	}

	apiKeyPrivate := os.Getenv("MJ_APIKEY_PRIVATE")
	if !config.PrivateAPIKey.IsNull() {
		apiKeyPrivate = config.PrivateAPIKey.ValueString()
	}

	client := mailjet.NewMailjetClient(apiKeyPublic, apiKeyPrivate, baseURL)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *mailjetProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDNSDataSource,
	}
}

func (p *mailjetProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSenderResource,
	}
}
