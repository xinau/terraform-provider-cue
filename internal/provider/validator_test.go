package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		value   types.String
		wantErr bool
	}

	tests := map[string]testCase{
		"single path": {
			value:   types.StringValue("foo"),
			wantErr: false,
		},
		"mulitipart path": {
			value:   types.StringValue("foo.bar"),
			wantErr: false,
		},
		"invalid path": {
			value:   types.StringValue("foo,bar"),
			wantErr: true,
		},
		"null path": {
			value:   types.StringNull(),
			wantErr: false,
		}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			resp := validator.StringResponse{}
			NewPathValidator().ValidateString(context.TODO(),
				validator.StringRequest{
					Path:           path.Root("test"),
					PathExpression: path.MatchRoot("test"),
					ConfigValue:    test.value,
				},
				&resp,
			)

			if !resp.Diagnostics.HasError() && test.wantErr {
				t.Fatal("expected error, got no error")
			}

			if resp.Diagnostics.HasError() && !test.wantErr {
				t.Fatalf("got unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}
