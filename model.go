package pithsdk

// ModelSettings configures generation parameters for an agent run.
type ModelSettings struct {
	Temperature *float64
	MaxTokens   *int
}
