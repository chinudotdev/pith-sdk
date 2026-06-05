package pithsdk

// ModelSettings configures generation parameters for an agent run.
type ModelSettings struct {
	// Temperature controls randomness. Nil leaves the provider default.
	Temperature *float64
	// MaxTokens caps output tokens. Nil leaves the provider default.
	MaxTokens *int
}
