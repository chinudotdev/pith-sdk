// Package pithsdk is a minimal, OpenAI Agents-style SDK for Go app developers.
// It wraps github.com/chinudotdev/pith primitives so you can define an agent,
// run it, and get text back without wiring gateways, ModelDescriptors, or EventBus.
//
// Quick start:
//
//	client, _ := pithsdk.NewClient(pithsdk.ClientConfig{APIKey: os.Getenv("OPENAI_API_KEY")})
//	agent, _ := pithsdk.NewAgent(pithsdk.AgentConfig{
//	    Instructions: "You are helpful.",
//	    Model:        "gpt-4o-mini",
//	})
//	session, _ := client.NewSession(agent)
//	result, _ := session.Run(ctx, "Hello!")
//
// Import with the recommended alias:
//
//	import pithsdk "github.com/chinudotdev/pith-sdk"
//
// See the examples/ directory and https://github.com/chinudotdev/pith-sdk for more.
package pithsdk
