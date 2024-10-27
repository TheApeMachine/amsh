package provider

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/theapemachine/amsh/berrt"
)

type RandomProvider struct {
	providers []Provider
	rng       *rand.Rand
}

func NewRandomProvider(apiKeys map[string]string) *RandomProvider {
	var providers []Provider

	// Initialize OpenAI
	if key := apiKeys["openai"]; key != "" {
		providers = append(providers, NewOpenAI(key, "gpt-4o-mini"))
	}

	// Initialize Anthropic
	if key := apiKeys["anthropic"]; key != "" {
		providers = append(providers, NewAnthropic(key, "claude-3-5-sonnet"))
	}

	// Initialize Google
	if key := apiKeys["google"]; key != "" {
		if provider, err := NewGoogle(key, "gemini-1.5-flash"); err == nil {
			providers = append(providers, provider)
		}
	}

	// Initialize Cohere
	if key := apiKeys["cohere"]; key != "" {
		if provider, err := NewCohere(key, "command-r"); err == nil {
			providers = append(providers, provider)
		}
	}

	if len(providers) == 0 {
		berrt.Error("RandomProvider", errors.New("no providers available"))
		return nil
	}

	return &RandomProvider{
		providers: providers,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *RandomProvider) Generate(ctx context.Context, messages []Message) <-chan Event {
	provider := r.providers[r.rng.Intn(len(r.providers))]
	return provider.Generate(ctx, messages)
}

func (r *RandomProvider) GenerateSync(ctx context.Context, messages []Message) (string, error) {
	provider := r.providers[r.rng.Intn(len(r.providers))]
	return provider.GenerateSync(ctx, messages)
}
