package provider

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mergestat/timediff"
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
)

type ProviderStatus struct {
	name     string
	provider Provider
	occupied bool
	lastUsed time.Time
	failures int // Track consecutive failures
	mu       sync.Mutex
}

type BalancedProvider struct {
	providers   []*ProviderStatus
	selectIndex int
	// Add mutex for initialization state
	initMu      sync.Mutex
	initialized bool
}

var (
	balancedProviderInstance *BalancedProvider
	onceBalancedProvider     sync.Once
)

/*
NewBalancedProvider creates a new BalancedProvider as an ambient context,
so multiple calls to NewBalancedProvider will return the same instance.
*/
func NewBalancedProvider() *BalancedProvider {
	onceBalancedProvider.Do(func() {
		errnie.Info("new balanced provider")

		balancedProviderInstance = &BalancedProvider{
			providers: []*ProviderStatus{
				// {
				// 	name:     "llama3.2:3b",
				// 	provider: NewOllama("llama3.2:3b"),
				// 	occupied: false,
				// },
				{
					name:     "gpt-4o",
					provider: NewOpenAI(os.Getenv("OPENAI_API_KEY"), openai.ChatModelGPT4o2024_08_06),
					occupied: false,
				},
				{
					name:     "claude-3-5-sonnet",
					provider: NewAnthropic(os.Getenv("ANTHROPIC_API_KEY"), anthropic.ModelClaude3_5Sonnet20241022),
					occupied: false,
				},
				{
					name:     "gemini-1.5-flash",
					provider: NewGoogle(os.Getenv("GEMINI_API_KEY"), "gemini-1.5-flash"),
					occupied: false,
				},
				{
					name:     "command-r",
					provider: NewCohere(os.Getenv("COHERE_API_KEY"), "command-r"),
					occupied: false,
				},
				// {
				// 	name:     "LM Studio",
				// 	provider: NewLMStudio(os.Getenv("LM_STUDIO_API_KEY"), "bartowski/Llama-3.1-8B-Lexi-Uncensored-V2-GGUF"),
				// 	occupied: false,
				// },
				// {
				// 	name:     "NVIDIA",
				// 	provider: NewNVIDIA(os.Getenv("NVIDIA_API_KEY"), "nvidia/llama-3.1-nemotron-70b-instruct"),
				// 	occupied: false,
				// },
			},

			selectIndex: 0,
			initialized: false,
		}
	})

	return balancedProviderInstance
}

func (lb *BalancedProvider) Generate(ctx context.Context, params GenerationParams, messages []Message) <-chan Event {
	errnie.Info("generating with balanced provider")
	out := make(chan Event)

	go func() {
		defer close(out)

		ps := lb.getAvailableProvider()
		if ps == nil {
			return
		}

		defer func() {
			ps.mu.Lock()
			ps.occupied = false
			ps.mu.Unlock()
			errnie.Info("provider released")
		}()

		for event := range ps.provider.Generate(ctx, params, messages) {
			out <- event
		}
	}()

	return out
}

func (lb *BalancedProvider) getAvailableProvider() *ProviderStatus {
	errnie.Info("getting available provider")
	maxAttempts := 10
	cooldownPeriod := 60 * time.Second
	maxFailures := 3

	// Handle first request with random selection
	lb.initMu.Lock()
	if !lb.initialized {
		availableProviders := make([]*ProviderStatus, 0)
		for _, ps := range lb.providers {
			ps.mu.Lock()
			if !ps.occupied {
				availableProviders = append(availableProviders, ps)
			}
			ps.mu.Unlock()
		}

		if len(availableProviders) > 0 {
			// Select random provider for first request
			selectedIdx := rand.Intn(len(availableProviders))
			selected := availableProviders[selectedIdx]

			selected.mu.Lock()
			selected.occupied = true
			selected.lastUsed = time.Now()
			selected.mu.Unlock()

			lb.initialized = true
			lb.initMu.Unlock()

			errnie.Info("initial random provider selected: %s", selected.name)
			return selected
		}
		lb.initMu.Unlock()
	} else {
		lb.initMu.Unlock()
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		var bestProvider *ProviderStatus
		oldestUse := time.Now()

		for _, ps := range lb.providers {
			ps.mu.Lock()

			// Skip if provider is occupied
			if ps.occupied {
				ps.mu.Unlock()
				continue
			}

			// Skip if provider has exceeded failure threshold
			if ps.failures >= maxFailures && time.Since(ps.lastUsed) < cooldownPeriod {
				ps.mu.Unlock()
				continue
			}

			// Reset failures if cooldown period has passed
			if ps.failures >= maxFailures && time.Since(ps.lastUsed) >= cooldownPeriod {
				ps.failures = 0
			}

			// Select provider with lowest failure count and oldest last use
			if bestProvider == nil ||
				ps.failures < bestProvider.failures ||
				(ps.failures == bestProvider.failures && ps.lastUsed.Before(oldestUse)) {
				bestProvider = ps
				oldestUse = ps.lastUsed
			}

			ps.mu.Unlock()
		}

		if bestProvider != nil {
			bestProvider.mu.Lock()
			bestProvider.occupied = true
			bestProvider.lastUsed = time.Now()
			bestProvider.mu.Unlock()

			errnie.Info(
				"found available provider %s %d %s",
				bestProvider.name, bestProvider.failures, timediff.TimeDiff(bestProvider.lastUsed),
			)

			return bestProvider
		}

		// If all providers are busy, wait before trying again
		errnie.Warn("all providers occupied or in cooldown, attempt %d, waiting...", attempt+1)
		time.Sleep(1 * time.Second)
	}

	errnie.Error(errors.New("no providers available after maximum attempts"))
	return nil
}

func (lb *BalancedProvider) GenerateSync(ctx context.Context, params GenerationParams, messages []Message) (string, error) {
	errnie.Info("generating with balanced provider")

	events := lb.Generate(ctx, params, messages)
	var result string

	for event := range events {
		if event.Type == EventError {
			return "", errors.New(event.Content)
		}
		result += event.Content
	}

	return result, nil
}

func (lb *BalancedProvider) Configure(config map[string]interface{}) {
	// Configuration can be added here
}
