package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceExport() *schema.Resource {
	return &schema.Resource{
		Description: "The `cue_export` data source evaluates a CUE configuration and renders the emit value as JSON " +
			"encoded string.",

		ReadContext: DataSourceExportRead,

		Schema: map[string]*schema.Schema{
			"expression": {
				Description: "If set only export this single expression.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"files": {
				Description: "Files to evaluate and emit. Multiple files are combined at the top-level. " +
					"Order doesn't matter. In absence of files, the working directory is loaded as a package instance.",
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"working_dir": {
				Description: "Working directory of the program. If not supplied, the program will run in the current " +
					"directory.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"rendered": {
				Description: "The final rendered JSON encoded string of the emit value.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func DataSourceExportRead(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	ctx := cuecontext.New()
	files := expandStringList(data.Get("files").([]interface{}))
	dir := data.Get("working_dir").(string)

	values, err := client.Load(ctx, files, &load.Config{
		Dir: dir,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	val := Unify(values)

	if expression, ok := data.GetOk("expression"); ok {
		val = val.LookupPath(cue.ParsePath(expression.(string)))
		if err := val.Err(); err != nil {
			diag.FromErr(fmt.Errorf("failed to lookup path: %w", err))
		}
	}

	if err := val.Validate(cue.Concrete(true), cue.Final()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to validate value: %w", err))
	}

	rendered, err := json.Marshal(val)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to marshal value: %w", err))
	}

	if err := data.Set("rendered", string(rendered)); err != nil {
		return diag.FromErr(fmt.Errorf("failed setting rendered: %w", err))
	}

	data.SetId(hashBytes(rendered))
	return nil
}
