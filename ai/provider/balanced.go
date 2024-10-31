package provider

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
)

type ProviderStatus struct {
	provider Provider
	occupied bool
	lastUsed time.Time
	failures int // Track consecutive failures
	mu       sync.Mutex
}

type BalancedProvider struct {
	providers   []*ProviderStatus
	selectIndex int
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
		log.Info("NewBalancedProvider")

		balancedProviderInstance = &BalancedProvider{
			providers: []*ProviderStatus{
				{
					provider: NewOpenAI(os.Getenv("OPENAI_API_KEY"), openai.ChatModelGPT4oMini2024_07_18),
					occupied: false,
				},
				{
					provider: NewAnthropic(os.Getenv("ANTHROPIC_API_KEY"), anthropic.ModelClaude3_5Sonnet20241022),
					occupied: false,
				},
				{
					provider: NewGoogle(os.Getenv("GEMINI_API_KEY"), "gemini-1.5-flash"),
					occupied: false,
				},
				{
					provider: NewCohere(os.Getenv("COHERE_API_KEY"), "command-r"),
					occupied: false,
				},
				{
					provider: NewOllama("llama3.2:3b"),
					occupied: false,
				},
			},

			selectIndex: 0,
		}
	})

	return balancedProviderInstance
}

func (lb *BalancedProvider) Generate(ctx context.Context, params GenerationParams, messages []Message) <-chan Event {
	log.Info("Generate")
	resultChan := make(chan Event)

	go func() {
		defer close(resultChan)

		ps := lb.getAvailableProvider()
		if ps == nil {
			resultChan <- Event{
				Type:    EventError,
				Content: "no providers available",
			}
			return
		}

		defer func() {
			ps.mu.Lock()
			ps.occupied = false
			ps.mu.Unlock()
			log.Info("provider released")
		}()

		timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		done := make(chan struct{})

		go func() {
			defer close(done)
			for event := range ps.provider.Generate(timeoutCtx, params, messages) {
				if event.Type == EventError &&
					(strings.Contains(event.Error.Error(), "429") ||
						strings.Contains(event.Error.Error(), "rate_limit")) {
					// Handle rate limit error
					ps.mu.Lock()
					ps.failures++
					ps.mu.Unlock()
					log.Warn("rate limit hit, increasing failure count",
						"provider", ps.provider,
						"failures", ps.failures)
				}

				select {
				case resultChan <- event:
				case <-timeoutCtx.Done():
					return
				}
			}
		}()

		select {
		case <-done:
			return
		case <-timeoutCtx.Done():
			resultChan <- Event{
				Type:    EventError,
				Content: "provider timeout",
			}
			return
		}
	}()

	return resultChan
}

func (lb *BalancedProvider) getAvailableProvider() *ProviderStatus {
	log.Info("getAvailableProvider")
	maxAttempts := 10
	cooldownPeriod := 60 * time.Second // Cooldown after rate limit

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Find the provider with the lowest failure count and oldest last use
		var bestProvider *ProviderStatus
		oldestUse := time.Now()

		for _, ps := range lb.providers {
			ps.mu.Lock()

			// Skip if provider is occupied
			if ps.occupied {
				ps.mu.Unlock()
				continue
			}

			// Skip if provider is in cooldown from rate limit
			if ps.failures > 0 && time.Since(ps.lastUsed) < cooldownPeriod {
				ps.mu.Unlock()
				continue
			}

			// Reset failures if cooldown period has passed
			if ps.failures > 0 && time.Since(ps.lastUsed) >= cooldownPeriod {
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

			log.Info("found available provider",
				"provider", bestProvider.provider,
				"failures", bestProvider.failures,
				"lastUsed", bestProvider.lastUsed)
			return bestProvider
		}

		// If all providers are busy, wait before trying again
		log.Warn("all providers occupied or in cooldown, waiting...", "attempt", attempt+1)
		time.Sleep(1 * time.Second)
	}

	log.Error("no providers available after maximum attempts")
	return nil
}

func (lb *BalancedProvider) GenerateSync(ctx context.Context, params GenerationParams, messages []Message) (string, error) {
	log.Info("GenerateSync")

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
