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
	Args     types.List   `tfsdk:"args"`
	Dir      types.String `tfsdk:"dir"`
	ID       types.String `tfsdk:"id"`
	Path     types.String `tfsdk:"path"`
	Rendered types.String `tfsdk:"rendered"`
}

func (d *ExportDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_export"
}

func (d *ExportDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "The export data source evaluates a CUE definition and renders the emit value as JSON encoded string",
		Attributes: map[string]tfsdk.Attribute{
			"args": {
				MarkdownDescription: "Command-line arguments passed to instances loading.",
				Type:                types.ListType{ElemType: types.StringType},
				Optional:            true,
			},
			"dir": {
				MarkdownDescription: "Directory to use for CUE's evaluation. If omitted the current directory is used instead.",
				Type:                types.StringType,
				Optional:            true,
			},
			"id": {
				MarkdownDescription: "SHA256 sum of the rendered emit value.",
				Type:                types.StringType,
				Computed:            true,
			},
			"path": {
				MarkdownDescription: "Path to lookup inside CUE value.",
				Type:                types.StringType,
				Optional:            true,
				Validators:          []tfsdk.AttributeValidator{NewPathValidator()},
			},
			"rendered": {
				MarkdownDescription: "Emit value rendered as JSON encoded string.",
				Type:                types.StringType,
				Computed:            true,
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

	var args []string
	resp.Diagnostics.Append(data.Args.ElementsAs(ctx, &args, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values, err := d.client.Load(cuecontext.New(), args, &load.Config{
		Dir: data.Dir.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unexpected CUE Loading Error",
			fmt.Sprintf("Unable to load CUE values, got error: %s", err))
		return
	}

	val := values[0]
	if err := val.Validate(cue.Concrete(true), cue.Final()); err != nil {
		resp.Diagnostics.AddError("Unexpected CUE Validation Error",
			fmt.Sprintf("Unable to validate CUE value, got error: %s", err))
		return
	}

	if !data.Path.IsNull() {
		path := cue.ParsePath(data.Path.ValueString())
		if err := path.Err(); err != nil {
			resp.Diagnostics.AddError("Unexpected CUE Parse Path Error",
				fmt.Sprintf("Unable to parse CUE path, got error: %s. Please report this issue to the provider developers.", err))
			return
		}

		val = val.LookupPath(path)
		if err := val.Err(); err != nil {
			resp.Diagnostics.AddError("Unexpected CUE Path Lookup Error",
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
