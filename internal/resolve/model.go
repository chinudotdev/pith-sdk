package resolve

import (
	"fmt"
	"strings"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
)

// Model resolves a model string against the gateway catalog.
// Bare IDs use defaultProvider; "provider/model-id" selects an explicit provider.
func Model(gw *gateway.LLMGateway, defaultProvider, modelString string) (protocol.ModelDescriptor, error) {
	modelString = strings.TrimSpace(modelString)
	if modelString == "" {
		return protocol.ModelDescriptor{}, fmt.Errorf("model ID is required")
	}

	providerID, modelID := splitModelString(defaultProvider, modelString)
	fullName := modelString
	if providerID != defaultProvider || strings.Contains(modelString, "/") {
		fullName = providerID + "/" + modelID
	}

	if !providerRegistered(gw, providerID) {
		return protocol.ModelDescriptor{}, fmt.Errorf("unknown provider %q: not registered", providerID)
	}

	model, ok := gw.Catalog.Get(protocol.ProviderId(providerID), modelID)
	if !ok {
		if strings.Contains(modelString, "/") {
			return protocol.ModelDescriptor{}, fmt.Errorf(
				"unknown model %q: provider %q registered but model not found",
				fullName, providerID,
			)
		}
		return protocol.ModelDescriptor{}, fmt.Errorf("unknown model %q: not found in provider %q", modelID, providerID)
	}
	return model, nil
}

func splitModelString(defaultProvider, modelString string) (providerID, modelID string) {
	if idx := strings.Index(modelString, "/"); idx >= 0 {
		return modelString[:idx], modelString[idx+1:]
	}
	return defaultProvider, modelString
}

func providerRegistered(gw *gateway.LLMGateway, providerID string) bool {
	for _, p := range gw.Catalog.GetProviders() {
		if string(p) == providerID {
			return true
		}
	}
	return false
}
