package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		value   attr.Value
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
		},
		"invalid type": {
			value:   types.Int64Value(101),
			wantErr: true,
		}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			resp := tfsdk.ValidateAttributeResponse{}
			NewPathValidator().Validate(context.TODO(),
				tfsdk.ValidateAttributeRequest{
					AttributePath:           path.Root("test"),
					AttributePathExpression: path.MatchRoot("test"),
					AttributeConfig:         test.value,
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
