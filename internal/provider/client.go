package provider

import (
	"fmt"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
)

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
			return nil, fmt.Errorf("loading instance %q: %w", i.ID(), err)
		}

		if i.Incomplete {
			return nil, fmt.Errorf("loading instance %q dependencies", i.ID())
		}

		val := ctx.BuildInstance(i)
		if err := val.Err(); err != nil {
			return nil, fmt.Errorf("building instance: %w", err)
		}

		values = append(values, val)
	}

	return values, nil
}
