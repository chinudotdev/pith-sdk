package resolve

import (
	"fmt"
	"strings"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
)

// Model resolves a bare model ID against the default provider catalog.
func Model(gw *gateway.LLMGateway, providerID, modelID string) (protocol.ModelDescriptor, error) {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return protocol.ModelDescriptor{}, fmt.Errorf("model ID is required")
	}
	if strings.Contains(modelID, "/") {
		return protocol.ModelDescriptor{}, fmt.Errorf("provider/model syntax is not supported yet; use a bare model ID with the default provider")
	}

	model, ok := gw.Catalog.Get(protocol.ProviderId(providerID), modelID)
	if !ok {
		return protocol.ModelDescriptor{}, fmt.Errorf("unknown model %q: not found in provider %q", modelID, providerID)
	}
	return model, nil
}
