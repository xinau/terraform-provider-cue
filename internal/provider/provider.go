package provider

import (
	"context"
	"fmt"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"cue_export": DataSourceExport(),
			},
		}

		p.ConfigureContextFunc = Configure(version, p)

		return p
	}
}

func Configure(_ string, _ *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return &Client{}, nil
	}
}

// Client is a workaround for some concurrency problems inside cue
// These are triggered i.e. when loading instances concurrently
// See https://github.com/cue-lang/cue/issues/460
type Client struct {
	mtx sync.Mutex
}

func (c *Client) Load(ctx *cue.Context, args []string, cfg *load.Config) ([]cue.Value, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	var values []cue.Value
	for _, i := range load.Instances(args, cfg) {
		if err := i.Err; err != nil {
			return nil, fmt.Errorf("failed to load instance: %w", err)
		}

		val := ctx.BuildInstance(i)
		if err := val.Err(); err != nil {
			return nil, fmt.Errorf("failed to build instance: %w", err)
		}

		values = append(values, val)
	}

	return values, nil
}

func Unify(values []cue.Value) cue.Value {
	if len(values) == 0 {
		return cue.Value{}
	}

	val := values[0]
	for _, v := range values[1:] {
		val = val.Unify(v)
	}

	return val
}
