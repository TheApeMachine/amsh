package provider

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
)

type ProviderStatus struct {
	provider Provider
	occupied bool
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

		// Find available provider or wait
		ps := lb.getAvailableProvider()
		if ps == nil {
			resultChan <- Event{
				Type:    EventError,
				Content: "no providers available",
			}
			return
		}

		// Ensure we release the provider when done
		defer func() {
			ps.mu.Lock()
			ps.occupied = false
			ps.mu.Unlock()
			log.Info("provider released")
		}()

		// Create a separate context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Create done channel to handle cleanup
		done := make(chan struct{})

		// Start provider generation in separate goroutine
		go func() {
			defer close(done)
			for event := range ps.provider.Generate(timeoutCtx, params, messages) {
				select {
				case resultChan <- event:
				case <-timeoutCtx.Done():
					return
				}
			}
		}()

		// Wait for either completion or timeout
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

	for attempt := 0; attempt < maxAttempts; attempt++ {
		ps := lb.providers[lb.selectIndex%len(lb.providers)]
		lb.selectIndex++

		if !ps.occupied {
			ps.occupied = true
			log.Info("found available provider", "provider", ps.provider)
			return ps
		}

		// If all providers are busy, wait a bit before trying again
		log.Warn("all providers occupied, waiting...", "attempt", attempt+1)
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
