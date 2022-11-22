package provider

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PathValidator struct{}

func NewPathValidator() *PathValidator {
	return &PathValidator{}
}

func (v *PathValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v *PathValidator) MarkdownDescription(ctx context.Context) string {
	return "Value must be a valid CUE path"
}

func (v *PathValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	if req.AttributeConfig.IsNull() || req.AttributeConfig.IsUnknown() {
		return
	}

	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if str.IsNull() || str.IsUnknown() {
		return
	}

	path := cue.ParsePath(str.ValueString())
	if err := path.Err(); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Parsing Path Error",
			fmt.Sprintf("Parsing CUE path %q failed: %v", str.ValueString(), err),
		)
		return
	}

	return
}
