package provider

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Unified  types.Bool   `tfsdk:"unified"`
}

func (d *ExportDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_export"
}

func (d *ExportDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The export data source evaluates a CUE definition and renders the emit value as JSON encoded string",
		Attributes: map[string]schema.Attribute{
			"dir": schema.StringAttribute{
				MarkdownDescription: "Directory to use for CUE's evaluation. If omitted the current directory is used instead.",
				Optional:            true,
			},
			"expr": schema.StringAttribute{
				MarkdownDescription: "Exrpession to lookup inside CUE value.",
				Optional:            true,
				Validators: []validator.String{
					NewPathValidator(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "FNV-128a sum of the rendered emit value.",
				Computed:            true,
			},
			"paths": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of paths to CUE instances to evaluate.",
				Optional:            true,
			},
			"pkg": schema.StringAttribute{
				MarkdownDescription: "Name of the package to be loaded. If not set it needs to be uniquely defined in it's context.",
				Optional:            true,
			},
			"rendered": schema.StringAttribute{
				MarkdownDescription: "Emit value rendered as JSON encoded string.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of boolean tags or key-value pairs injected as values into fields during loading.",
				Optional:            true,
			},
			"unified": schema.BoolAttribute{
				MarkdownDescription: "Unify multiple values into a single one. If false only the first value is emitted. (Default `true`)",
				Optional:            true,
			},
		},
	}
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
	if data.Unified.IsNull() || data.Unified.IsUnknown() || data.Unified.ValueBool() {
		for _, w := range values[1:] {
			val = val.Unify(w)
		}
	}

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
	hash := fnv.New128a()
	_, _ = hash.Write(in)
	var buf []byte
	return hex.EncodeToString(hash.Sum(buf))
}
