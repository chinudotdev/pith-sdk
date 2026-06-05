package wire

import (
	"context"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/loop"
	"github.com/chinudotdev/pith/protocol"
)

// Settings holds optional generation parameters applied at the gateway boundary.
type Settings struct {
	Temperature *float64
	MaxTokens   *int
}

// NewStreamFn returns a StreamFn that delegates to the gateway and merges settings.
func NewStreamFn(gw *gateway.LLMGateway, settings Settings) loop.StreamFn {
	return func(ctx context.Context, model protocol.ModelDescriptor, pctx protocol.Context, opts protocol.StreamOptions) (<-chan protocol.StreamEvent, error) {
		if settings.Temperature != nil {
			opts.Temperature = settings.Temperature
		}
		if settings.MaxTokens != nil {
			opts.MaxTokens = settings.MaxTokens
		}
		return gw.Stream(ctx, model, pctx, opts)
	}
}
