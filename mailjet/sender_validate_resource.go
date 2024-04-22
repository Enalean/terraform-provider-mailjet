package mailjet

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailjet/mailjet-apiv3-go/v3/resources"
	"github.com/mailjet/mailjet-apiv3-go/v4"
)

var (
	_ resource.Resource              = &senderValidateResource{}
	_ resource.ResourceWithConfigure = &senderValidateResource{}
)

func NewSenderValidateResource() resource.Resource {
	return &senderValidateResource{}
}

type senderValidateResource struct {
	client *mailjet.Client
}

func (r *senderValidateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *senderValidateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sender_validate"
}

func (r *senderValidateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:    true,
				Description: "Unique numeric ID for the sender you want to validate.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"wait_for": schema.StringAttribute{
				Optional:    true,
				Description: "When specified, the provider will make multiple attempts to validate the resource until the specified duration is reached. One attempt is made per second.",
				Validators: []validator.String{
					TimeDurationAtLeast1Sec(),
				},
			},
		},
	}
}

type senderValidateResourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	WaitFor types.String `tfsdk:"wait_for"`
}

func (r *senderValidateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state senderValidateResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	senderSearchRequest := &mailjet.Request{
		Resource: "sender",
		ID:       state.ID.ValueInt64(),
	}

	durationString := state.WaitFor.ValueString()
	if durationString == "" {
		durationString = "0s"
	}

	waitForDuration, err := time.ParseDuration(durationString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse wait_for",
			err.Error(),
		)
		return
	}

	var responseDataSearch []resources.Sender
	err = r.client.Get(senderSearchRequest, &responseDataSearch)

	if err == nil && len(responseDataSearch) == 1 && responseDataSearch[0].Status == "Active" {
		diags := resp.State.Set(ctx, state)
		resp.Diagnostics.Append(diags...)
		return
	}

	mailjetValidateRequest := &mailjet.Request{
		Resource: "sender",
		ID:       state.ID.ValueInt64(),
		Action:   "validate",
	}

	mailjetValidateFullRequest := &mailjet.FullRequest{
		Info: mailjetValidateRequest,
	}

	var responseDataValidation []resources.SenderValidate

	startAttempt := time.Now()
	for {
		err = r.client.Post(mailjetValidateFullRequest, responseDataValidation)

		if err == nil && len(responseDataValidation) == 1 && responseDataValidation[0].GlobalError == "" {
			return
		}
		if time.Since(startAttempt) > waitForDuration {
			break
		}
		time.Sleep(time.Second)
	}

	err = r.client.Post(mailjetValidateFullRequest, responseDataValidation)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error validating the Mailjet sender",
			"Unexpected error while validating the sender: "+err.Error(),
		)
		return
	}

	if len(responseDataValidation) == 0 {
		resp.Diagnostics.AddError(
			"Error validating the Mailjet sender",
			"No validation methods were found for this sender.",
		)
		return
	}

	if len(responseDataValidation) > 1 {
		resp.Diagnostics.AddError(
			"Error validating the Mailjet sender",
			"Multiple data validation information where provided unexpectedly",
		)
		return
	}

	if responseDataValidation[0].GlobalError != "" {
		resp.Diagnostics.AddError(
			"Error validating the Mailjet sender",
			"Could not validate the sender: "+responseDataValidation[0].GlobalError,
		)
		return
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *senderValidateResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// No need to do anything, the resource does not really exist on the Mailjet side
}

func (r *senderValidateResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// No need to do anything
}

func (r *senderValidateResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No need to do anything
}
