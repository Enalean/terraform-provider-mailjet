package mailjet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailjet/mailjet-apiv3-go/v3/resources"
	"github.com/mailjet/mailjet-apiv3-go/v4"
)

var (
	_ datasource.DataSource              = &dnsDataSource{}
	_ datasource.DataSourceWithConfigure = &dnsDataSource{}
)

func NewDNSDataSource() datasource.DataSource {
	return &dnsDataSource{}
}

type dnsDataSource struct {
	client *mailjet.Client
}

func (d *dnsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mailjet.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mailjet.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *dnsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns"
}

func (d *dnsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dns_id": schema.Int64Attribute{
				Required:    true,
				Description: "Unique numeric ID of the domain settings",
			},
			"entries": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"domain": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the domain linked to this DNS record.",
						},
						"ownership_token_record_name": schema.StringAttribute{
							Computed:    true,
							Description: "Value to use when configuring the TXT record (DNS) verification for the domain.",
						},
						"ownership_token": schema.StringAttribute{
							Computed:    true,
							Description: "Value of the token to verify the ownership of the domain.",
						},
						"spf_record_value": schema.StringAttribute{
							Computed:    true,
							Description: "Value to insert in the DNS SPF record for this domain.",
						},
						"dkim_record_name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the DNS DKIM record to insert for this domain.",
						},
						"dkim_record_value": schema.StringAttribute{
							Computed:    true,
							Description: "Value to insert in the DNS DKIM record for this domain.",
						},
					},
				},
			},
		},
	}
}

type dnsDataSourceModel struct {
	DNSID   types.Int64 `tfsdk:"dns_id"`
	Entries []dnsModel  `tfsdk:"entries"`
}

type dnsModel struct {
	ID                       types.Int64  `tfsdk:"id"`
	Domain                   types.String `tfsdk:"domain"`
	OwnerShipTokenRecordName types.String `tfsdk:"ownership_token_record_name"`
	OwnerShipToken           types.String `tfsdk:"ownership_token"`
	SPFRecordValue           types.String `tfsdk:"spf_record_value"`
	DKIMRecordName           types.String `tfsdk:"dkim_record_name"`
	DKIMRecordValue          types.String `tfsdk:"dkim_record_value"`
}

func (d *dnsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dnsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var responseData []resources.Dns
	mailjetRequest := &mailjet.Request{
		Resource: "dns",
		ID:       state.DNSID.ValueInt64(),
	}
	err := d.client.Get(mailjetRequest, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Mailjet DNS information",
			err.Error(),
		)
		return
	}

	for _, dnsEntry := range responseData {
		dnsEntryState := dnsModel{
			ID:                       types.Int64Value(dnsEntry.ID),
			Domain:                   types.StringValue(dnsEntry.Domain),
			OwnerShipTokenRecordName: types.StringValue(dnsEntry.OwnerShipTokenRecordName),
			OwnerShipToken:           types.StringValue(dnsEntry.OwnerShipToken),
			SPFRecordValue:           types.StringValue(dnsEntry.SPFRecordValue),
			DKIMRecordName:           types.StringValue(dnsEntry.DKIMRecordName),
			DKIMRecordValue:          types.StringValue(dnsEntry.DKIMRecordValue),
		}

		state.Entries = append(state.Entries, dnsEntryState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
