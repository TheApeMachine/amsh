package provider

import (
	"errors"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/errnie"
)

type ProviderStatus struct {
	name     string
	provider Provider
	occupied bool
	lastUsed time.Time
	failures int
	mu       sync.Mutex
}

type BalancedProvider struct {
	providers   []*ProviderStatus
	selectIndex int
	initMu      sync.Mutex
	initialized bool
}

var (
	balancedProviderInstance *BalancedProvider
	onceBalancedProvider     sync.Once
)

func NewBalancedProvider() *BalancedProvider {
	onceBalancedProvider.Do(func() {
		balancedProviderInstance = &BalancedProvider{
			providers: []*ProviderStatus{
				{
					name: "gpt-4o-mini",
					provider: NewOpenAI(
						os.Getenv("OPENAI_API_KEY"),
						openai.ChatModelGPT4oMini,
					),
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
				// 	name:     "llama3.2:3b",
				// 	provider: NewOllama("llama3.2:3b"),
				// 	occupied: false,
				// },
				// {
				// 	name:     "LM Studio",
				// 	provider: NewOpenAI(
				// 		os.Getenv("LM_STUDIO_API_KEY"),
				// 		"https://api.openai.com/v1",
				// 		"bartowski/Llama-3.1-8B-Lexi-Uncensored-V2-GGUF",
				// 	),
				// 	occupied: false,
				// },
				// {
				// 	name:     "NVIDIA",
				// 	provider: NewOpenAI(
				// 		os.Getenv("NVIDIA_API_KEY"),
				// 		"https://api.openai.com/v1",
				// 		"nvidia/llama-3.1-nemotron-70b-instruct",
				// 	),
				// 	occupied: false,
				// },
			},
			selectIndex: 0,
			initialized: false,
		}
	})

	return balancedProviderInstance
}

func (lb *BalancedProvider) Generate(artifacts []*data.Artifact) <-chan *data.Artifact {
	return twoface.NewAccumulator(
		"balanced",
		"provider",
		"completion",
		artifacts...,
	).Yield(func(accumulator *twoface.Accumulator) {
		provider := lb.getAvailableProvider()
		if provider == nil {
			errnie.Error(errors.New("no available provider found"))
		}

		provider.mu.Lock()
		provider.occupied = true
		provider.lastUsed = time.Now()
		provider.mu.Unlock()

		defer close(accumulator.Out)

		for artifact := range provider.provider.Generate(artifacts) {
			accumulator.Out <- artifact
		}

		provider.mu.Lock()
		provider.occupied = false
		provider.mu.Unlock()
	}).Generate()
}

func (lb *BalancedProvider) getAvailableProvider() *ProviderStatus {
	if provider := lb.handleFirstRequest(); provider != nil {
		return provider
	}
	return lb.findBestAvailableProvider()
}

func (lb *BalancedProvider) handleFirstRequest() *ProviderStatus {
	lb.initMu.Lock()
	defer lb.initMu.Unlock()

	if lb.initialized {
		return nil
	}

	availableProviders := lb.getUnoccupiedProviders()
	if len(availableProviders) == 0 {
		return nil
	}

	selected := availableProviders[rand.Intn(len(availableProviders))]
	lb.markProviderAsOccupied(selected)
	lb.initialized = true

	return selected
}

func (lb *BalancedProvider) findBestAvailableProvider() *ProviderStatus {
	maxAttempts := 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if provider := lb.selectBestProvider(); provider != nil {
			return provider
		}
		errnie.Warn("all providers occupied or in cooldown, attempt %d, waiting...", attempt+1)
		time.Sleep(1 * time.Second)
	}

	errnie.Error(errors.New("no providers available after maximum attempts"))
	return nil
}

func (lb *BalancedProvider) selectBestProvider() *ProviderStatus {
	var bestProvider *ProviderStatus
	oldestUse := time.Now()
	cooldownPeriod := 60 * time.Second
	maxFailures := 3

	for _, ps := range lb.providers {
		ps.mu.Lock()

		if !lb.isProviderAvailable(ps, cooldownPeriod, maxFailures) {
			ps.mu.Unlock()
			continue
		}

		if lb.isBetterProvider(ps, bestProvider, oldestUse) {
			bestProvider = ps
			oldestUse = ps.lastUsed
		}
		ps.mu.Unlock()
	}

	if bestProvider != nil {
		lb.markProviderAsOccupied(bestProvider)
	}

	return bestProvider
}

func (lb *BalancedProvider) getUnoccupiedProviders() []*ProviderStatus {
	available := make([]*ProviderStatus, 0)
	for _, ps := range lb.providers {
		ps.mu.Lock()
		if !ps.occupied {
			available = append(available, ps)
		}
		ps.mu.Unlock()
	}
	return available
}

func (lb *BalancedProvider) isProviderAvailable(ps *ProviderStatus, cooldownPeriod time.Duration, maxFailures int) bool {
	if ps.occupied {
		return false
	}

	if ps.failures >= maxFailures && time.Since(ps.lastUsed) < cooldownPeriod {
		return false
	}

	if ps.failures >= maxFailures && time.Since(ps.lastUsed) >= cooldownPeriod {
		ps.failures = 0
	}

	return true
}

func (lb *BalancedProvider) isBetterProvider(candidate, current *ProviderStatus, oldestUse time.Time) bool {
	return current == nil ||
		candidate.failures < current.failures ||
		(candidate.failures == current.failures && candidate.lastUsed.Before(oldestUse))
}

func (lb *BalancedProvider) markProviderAsOccupied(ps *ProviderStatus) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.occupied = true
	ps.lastUsed = time.Now()
}
