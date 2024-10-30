package provider

import (
	"context"
	"math/rand"
	"os"
	"time"
)

type RandomProvider struct {
	providers []Provider
	rng       *rand.Rand
}

func NewRandomProvider() *RandomProvider {
	return &RandomProvider{
		providers: []Provider{
			NewOpenAI(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini"),
			NewAnthropic(os.Getenv("ANTHROPIC_API_KEY"), "claude-3-5-sonnet"),
			NewGoogle(os.Getenv("GOOGLE_API_KEY"), "gemini-1.5-flash"),
			NewCohere(os.Getenv("COHERE_API_KEY"), "command-r"),
		},
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *RandomProvider) Generate(ctx context.Context, params GenerationParams, messages []Message) <-chan Event {
	provider := r.providers[r.rng.Intn(len(r.providers))]
	return provider.Generate(ctx, params, messages)
}

func (r *RandomProvider) GenerateSync(ctx context.Context, params GenerationParams, messages []Message) (string, error) {
	provider := r.providers[r.rng.Intn(len(r.providers))]
	return provider.GenerateSync(ctx, params, messages)
}

func (provider *RandomProvider) Configure(config map[string]interface{}) {}
