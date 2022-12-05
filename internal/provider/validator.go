package provider

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func (v *PathValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	str := req.ConfigValue
	if str.IsNull() || str.IsUnknown() {
		return
	}

	path := cue.ParsePath(str.ValueString())
	if err := path.Err(); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Expression Parsing Error",
			fmt.Sprintf("Parsing CUE expression %q failed: %v", str.ValueString(), err),
		)
		return
	}

	return
}
