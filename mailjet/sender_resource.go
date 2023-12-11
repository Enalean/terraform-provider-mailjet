package mailjet

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailjet/mailjet-apiv3-go/v3/resources"
	"github.com/mailjet/mailjet-apiv3-go/v4"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ resource.Resource                = &senderResource{}
	_ resource.ResourceWithConfigure   = &senderResource{}
	_ resource.ResourceWithImportState = &senderResource{}
)

func NewSenderResource() resource.Resource {
	return &senderResource{}
}

type senderResource struct {
	client *mailjet.Client
}

func (r *senderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *senderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sender"
}

func (r *senderResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The email address for this sender. To register a domain use *@example.com.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "User-provided name for this sender.",
			},
			"is_default_sender": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates whether this is the default sender or not.",
			},
			"email_type": schema.StringAttribute{
				Required:    true,
				Description: "Type of emails this sender will send. This is for purely informative purposes - the values do not place any sending restrictions on the sender email or domain. Can be transactional, bulk or unknown",
			},
			"dns_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Unique numeric ID of the DNS domain to which sender belongs.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Unique numeric ID of this sender.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp indicating when this sender object was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

type senderResourceModel struct {
	Email           types.String `tfsdk:"email"`
	Name            types.String `tfsdk:"name"`
	IsDefaultSender types.Bool   `tfsdk:"is_default_sender"`
	EmailType       types.String `tfsdk:"email_type"`
	ID              types.Int64  `tfsdk:"id"`
	DNSID           types.Int64  `tfsdk:"dns_id"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (r *senderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan senderResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Name.IsNull() {
		plan.Name = types.StringValue("")
	}
	if plan.EmailType.IsNull() {
		plan.EmailType = types.StringValue("unknown")
	}
	if plan.IsDefaultSender.IsNull() {
		plan.IsDefaultSender = types.BoolValue(false)
	}

	senderToCreate := resources.Sender{
		Email:           plan.Email.ValueString(),
		Name:            plan.Name.ValueString(),
		EmailType:       plan.EmailType.ValueString(),
		IsDefaultSender: plan.IsDefaultSender.ValueBool(),
	}

	senderSearchRequest := &mailjet.Request{
		Resource: "sender",
		AltID:    senderToCreate.Email,
	}
	var responseDataSearch []resources.Sender
	err := r.client.Get(senderSearchRequest, &responseDataSearch)

	if err == nil && len(responseDataSearch) == 1 && responseDataSearch[0].Status == "Deleted" {
		plan.ID = types.Int64Value(responseDataSearch[0].ID)
		err := r.updateSender(&senderToCreate)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update existing Mailjet sender information",
				"Could not update sender "+senderToCreate.Email+": "+err.Error(),
			)
			return
		}

		r.updateStateWithFetchedSenderInformation(&plan, &resp.Diagnostics)
	} else {
		mailjetFullUpdateRequest := &mailjet.FullRequest{
			Info: &mailjet.Request{
				Resource: "sender",
			},
			Payload: senderToCreate,
		}
		var responseData []resources.Sender

		err = r.client.Post(mailjetFullUpdateRequest, &responseData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to to create a Mailjet sender",
				err.Error(),
			)
			return
		}

		if len(responseData) != 1 {
			resp.Diagnostics.AddError(
				"Sender creation response is not coherent",
				fmt.Sprintf("Expected 1 response entry, got %d", len(responseData)),
			)
			return
		}

		r.refreshState(&responseData[0], &plan)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *senderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state senderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateStateWithFetchedSenderInformation(&state, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *senderResource) updateStateWithFetchedSenderInformation(state *senderResourceModel, diags *diag.Diagnostics) {
	var responseData []resources.Sender
	mailjetRequest := &mailjet.Request{
		Resource: "sender",
		ID:       state.ID.ValueInt64(),
	}
	err := r.client.Get(mailjetRequest, &responseData)
	if err != nil {
		diags.AddError(
			"Unable to read Mailjet sender information",
			"Could not read sender #"+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	if len(responseData) != 1 {
		diags.AddError(
			"Retrieved Mailjet sender information are not coherent",
			"Could not read sender #"+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+fmt.Sprintf("Could not read sender #Expected 1 response entry, got %d", len(responseData)),
		)
		return
	}

	r.refreshState(&responseData[0], state)
}

func (r *senderResource) refreshState(responseData *resources.Sender, state *senderResourceModel) {
	state.Email = types.StringValue(responseData.Email)
	state.EmailType = types.StringValue(responseData.EmailType)
	state.Name = types.StringValue(responseData.Name)
	state.ID = types.Int64Value(responseData.ID)
	state.DNSID = types.Int64Value(responseData.DNSID)
	state.CreatedAt = types.StringValue(responseData.CreatedAt.Format(time.RFC850))
}

func (r *senderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan senderResourceModel
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	senderUpdate := resources.Sender{
		Email:           plan.Email.ValueString(),
		Name:            plan.Name.ValueString(),
		EmailType:       plan.EmailType.ValueString(),
		IsDefaultSender: plan.IsDefaultSender.ValueBool(),
	}
	err := r.updateSender(&senderUpdate)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update Mailjet sender information",
			"Could not update sender #"+strconv.FormatInt(plan.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	r.updateStateWithFetchedSenderInformation(&plan, &resp.Diagnostics)
}

func (r *senderResource) updateSender(sender *resources.Sender) error {
	mailjetRequest := &mailjet.Request{
		Resource: "sender",
		AltID:    sender.Email,
	}
	mailjetFullRequest := &mailjet.FullRequest{
		Info: mailjetRequest,
		Payload: resources.Sender{
			Name:            sender.Name,
			EmailType:       sender.EmailType,
			IsDefaultSender: sender.IsDefaultSender,
		},
	}

	return r.client.Put(mailjetFullRequest, []string{"Name", "EmailType", "IsDefaultSender"})
}

func (r *senderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state senderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mailjetRequest := &mailjet.Request{
		Resource: "sender",
		ID:       state.ID.ValueInt64(),
	}
	err := r.client.Delete(mailjetRequest)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting a Mailjet sender",
			"Could not the sender, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *senderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
