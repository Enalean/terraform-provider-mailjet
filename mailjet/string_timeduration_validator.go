package mailjet

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = timeDurationValidator{}

type timeDurationValidator struct {
}

func (validator timeDurationValidator) Description(_ context.Context) string {
	return `must be a string representing a duration of at least 1 second`
}

func (validator timeDurationValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator timeDurationValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	s := req.ConfigValue

	if s.IsUnknown() || s.IsNull() {
		return
	}

	duration, err := time.ParseDuration(s.ValueString())

	if err != nil {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"failed to parse the time duration",
			fmt.Sprintf("%q %s", s.ValueString(), validator.Description(ctx))),
		)
		return
	}

	if duration < time.Second {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"the time duration must be at least 1 second",
			fmt.Sprintf("%q %s", s.ValueString(), validator.Description(ctx))),
		)
		return
	}
}

func TimeDurationAtLeast1Sec() validator.String {
	return timeDurationValidator{}
}
