package provider

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const PathValidatorDescription = "value must be a valid CUE path"

type PathValidator struct{}

func NewPathValidator() *PathValidator {
	return &PathValidator{}
}

func (pv *PathValidator) Description(_ context.Context) string {
	return PathValidatorDescription
}

func (pv *PathValidator) MarkdownDescription(_ context.Context) string {
	return PathValidatorDescription
}

func (pv *PathValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	typ := req.AttributeConfig.Type(ctx)
	if typ != types.StringType {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeTypeDiagnostic(
			req.AttributePath,
			"expected value of type string",
			typ.String(),
		))
		return
	}

	str := req.AttributeConfig.(types.String)
	if str.IsUnknown() || str.IsNull() {
		return
	}

	path := cue.ParsePath(str.ValueString())
	if err := path.Err(); err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.AttributePath,
			pv.Description(ctx),
			fmt.Sprintf("invalid path %q: %s", str.ValueString(), err),
		))
		return
	}

	return
}
