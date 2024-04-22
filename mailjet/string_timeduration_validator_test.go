package mailjet

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTimeDuration(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}

	tests := map[string]testCase{
		"unknown": {
			val: types.StringUnknown(),
		},
		"null": {
			val: types.StringNull(),
		},
		"valid": {
			val: types.StringValue("30s"),
		},
		"invalid": {
			val: types.StringValue("30wrong"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"failed to parse the time duration",
					`"30wrong" must be a string representing a duration of at least 1 second`,
				),
			},
		},
		"invalid_too_small": {
			val: types.StringValue("0s"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"the time duration must be at least 1 second",
					`"0s" must be a string representing a duration of at least 1 second`,
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}

			response := validator.StringResponse{}

			TimeDurationAtLeast1Sec().ValidateString(context.Background(), request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
