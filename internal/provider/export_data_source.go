package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ExportDataSource{}

func NewExportDataSource() datasource.DataSource {
	return &ExportDataSource{}
}

// ExportDataSource defines the data source implementation.
type ExportDataSource struct {
	client *Client
}

// ExportDataSourceModel describes the data source data model.
type ExportDataSourceModel struct {
	Dir      types.String `tfsdk:"dir"`
	Expr     types.String `tfsdk:"expr"`
	ID       types.String `tfsdk:"id"`
	Paths    types.List   `tfsdk:"paths"`
	Package  types.String `tfsdk:"pkg"`
	Rendered types.String `tfsdk:"rendered"`
	Tags     types.List   `tfsdk:"tags"`
}

func (d *ExportDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_export"
}

func (d *ExportDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "The export data source evaluates a CUE definition and renders the emit value as JSON encoded string",
		Attributes: map[string]tfsdk.Attribute{
			"dir": {
				MarkdownDescription: "Directory to use for CUE's evaluation. If omitted the current directory is used instead.",
				Type:                types.StringType,
				Optional:            true,
			},
			"expr": {
				MarkdownDescription: "Exrpession to lookup inside CUE value.",
				Type:                types.StringType,
				Optional:            true,
				Validators: []tfsdk.AttributeValidator{
					NewPathValidator(),
				},
			},
			"id": {
				MarkdownDescription: "SHA256 sum of the rendered emit value.",
				Type:                types.StringType,
				Computed:            true,
			},
			"paths": {
				MarkdownDescription: "List of paths to CUE instances to evaluate.",
				Type:                types.ListType{ElemType: types.StringType},
				Optional:            true,
			},
			"pkg": {
				MarkdownDescription: "Name of the package to be loaded. If not set it needs to be uniquely defined in it's context.",
				Type:                types.StringType,
				Optional:            true,
			},
			"rendered": {
				MarkdownDescription: "Emit value rendered as JSON encoded string.",
				Type:                types.StringType,
				Computed:            true,
			},
			"tags": {
				MarkdownDescription: "List of boolean tags or key-value pairs injected as values into fields during loading.",
				Type:                types.ListType{ElemType: types.StringType},
				Optional:            true,
			},
		},
	}, nil
}

func (d *ExportDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *provider.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ExportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExportDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var paths []string
	resp.Diagnostics.Append(data.Paths.ElementsAs(ctx, &paths, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values, err := d.client.Load(cuecontext.New(), paths, &load.Config{
		Dir:     data.Dir.ValueString(),
		Package: data.Package.ValueString(),
		Tags:    tags,
	})
	if err != nil {
		resp.Diagnostics.AddError("Loading Instances Error",
			fmt.Sprintf("Unable to load CUE values, got error: %s", err))
		return
	}

	val := values[0]
	if err := val.Validate(cue.Concrete(true), cue.Final()); err != nil {
		resp.Diagnostics.AddError("Value Validation Error",
			fmt.Sprintf("Unable to validate CUE value, got error: %s", err))
		return
	}

	if !data.Expr.IsNull() {
		path := cue.ParsePath(data.Expr.ValueString())
		if err := path.Err(); err != nil {
			resp.Diagnostics.AddError("Expression Parsing Error",
				fmt.Sprintf("Unable to parse CUE expression %q, got error: %s. "+
					"Please report this issue to the provider developers.",
					data.Expr.ValueString(), err,
				))
			return
		}

		val = val.LookupPath(path)
		if err := val.Err(); err != nil {
			resp.Diagnostics.AddError("Lookup Path Error",
				fmt.Sprintf("Unable to lookup CUE path in value, got error: %s", err))
			return
		}
	}

	rendered, err := json.Marshal(val)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected JSON Marshaling Error",
			fmt.Sprintf("Unable to marshal CUE value as JSON, got error: %s", err))
		return
	}

	data.Rendered = types.StringValue(string(rendered))
	data.ID = types.StringValue(hash(rendered))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func hash(in []byte) string {
	sum := sha256.Sum256(in)
	return hex.EncodeToString(sum[:])
}
