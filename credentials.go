package pithsdk

import (
	"fmt"
	"os"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
)

type credentialResolver func() (protocol.Credential, error)

type credentialRegistry struct {
	resolvers map[protocol.ProviderId]credentialResolver
}

func newCredentialRegistry() credentialRegistry {
	return credentialRegistry{resolvers: make(map[protocol.ProviderId]credentialResolver)}
}

func (r *credentialRegistry) set(pid protocol.ProviderId, resolver credentialResolver) {
	r.resolvers[pid] = resolver
}

func (r credentialRegistry) credentialProvider() gateway.CredentialProvider {
	return gateway.CredentialProviderFunc(func(pid protocol.ProviderId) (protocol.Credential, error) {
		resolver, ok := r.resolvers[pid]
		if !ok {
			return nil, &protocol.Error{Code: protocol.ErrAuth, Message: fmt.Sprintf("no credentials for provider %q", pid)}
		}
		return resolver()
	})
}

func openAIResolver(explicitKey string) credentialResolver {
	return func() (protocol.Credential, error) {
		key := explicitKey
		if key == "" {
			key = os.Getenv("OPENAI_API_KEY")
		}
		if key == "" {
			return nil, &protocol.Error{Code: protocol.ErrAuth, Message: "API key required: set ClientConfig.APIKey or OPENAI_API_KEY"}
		}
		return protocol.ApiKey{Key: key}, nil
	}
}

func apiKeyResolver(key string) credentialResolver {
	return func() (protocol.Credential, error) {
		if key == "" {
			return nil, &protocol.Error{Code: protocol.ErrAuth, Message: "API key is empty"}
		}
		return protocol.ApiKey{Key: key}, nil
	}
}

func apiKeyEnvResolver(env string) credentialResolver {
	return func() (protocol.Credential, error) {
		key := os.Getenv(env)
		if key == "" {
			return nil, &protocol.Error{Code: protocol.ErrAuth, Message: fmt.Sprintf("%s not set", env)}
		}
		return protocol.ApiKey{Key: key}, nil
	}
}

func credentialFuncResolver(fn func(providerID string) (string, error), providerID string) credentialResolver {
	return func() (protocol.Credential, error) {
		key, err := fn(providerID)
		if err != nil {
			return nil, &protocol.Error{Code: protocol.ErrAuth, Message: err.Error(), Cause: err}
		}
		if key == "" {
			return nil, &protocol.Error{Code: protocol.ErrAuth, Message: fmt.Sprintf("credential function returned empty key for provider %q", providerID)}
		}
		return protocol.ApiKey{Key: key}, nil
	}
}
