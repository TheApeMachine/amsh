package provider

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type ProviderStatus struct {
	provider Provider
	occupied bool
	lastUsed time.Time
	requests int64
	failures int64
	mu       sync.Mutex
}

type BalancedProvider struct {
	providers []*ProviderStatus
	queue     chan queuedRequest
	mu        sync.RWMutex
}

type queuedRequest struct {
	ctx       context.Context
	messages  []Message
	response  chan<- Event
	timestamp time.Time
}

var (
	balancedProviderInstance *BalancedProvider
	onceBalancedProvider     sync.Once
)

func NewBalancedProvider() *BalancedProvider {
	log.Info("NewBalancedProvider")

	apiKeys := map[string]string{
		"openai":    os.Getenv("OPENAI_API_KEY"),
		"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
		"google":    os.Getenv("GEMINI_API_KEY"),
		"cohere":    os.Getenv("COHERE_API_KEY"),
	}

	lb := &BalancedProvider{
		providers: []*ProviderStatus{
			{
				provider: NewOpenAI(apiKeys["openai"], "gpt-4-mini"),
				occupied: false,
				lastUsed: time.Now(),
				requests: 0,
				failures: 0,
			},
			{
				provider: NewAnthropic(apiKeys["anthropic"], "claude-3-5-sonnet"),
				occupied: false,
				lastUsed: time.Now(),
				requests: 0,
				failures: 0,
			},
			{
				provider: NewGoogle(apiKeys["google"], "gemini-1.5-flash"),
				occupied: false,
				lastUsed: time.Now(),
				requests: 0,
				failures: 0,
			},
			{
				provider: NewCohere(apiKeys["cohere"], "command-r"),
				occupied: false,
				lastUsed: time.Now(),
				requests: 0,
				failures: 0,
			},
		},
		queue: make(chan queuedRequest, 100), // Buffer size configurable
	}

	// Start queue processor
	go lb.processQueue()

	return lb
}

func (lb *BalancedProvider) Generate(ctx context.Context, messages []Message) <-chan Event {
	log.Info("Generate")

	resultChan := make(chan Event)

	// Queue the request
	lb.queue <- queuedRequest{
		ctx:       ctx,
		messages:  messages,
		response:  resultChan,
		timestamp: time.Now(),
	}

	return resultChan
}

func (lb *BalancedProvider) processQueue() {
	log.Info("Starting queue processor")

	for request := range lb.queue {
		// Find available provider or wait
		provider := lb.getAvailableProvider()

		go func(req queuedRequest, ps *ProviderStatus) {
			// Mark provider as occupied
			ps.mu.Lock()
			ps.occupied = true
			ps.lastUsed = time.Now()
			ps.requests++
			ps.mu.Unlock()

			// Process request
			for event := range ps.provider.Generate(req.ctx, req.messages) {
				select {
				case req.response <- event:
				case <-req.ctx.Done():
					close(req.response)
					ps.mu.Lock()
					ps.failures++
					ps.occupied = false
					ps.mu.Unlock()
					return
				}
			}

			// Mark provider as free
			ps.mu.Lock()
			ps.occupied = false
			ps.mu.Unlock()

			close(req.response)
		}(request, provider)
	}
}

func (lb *BalancedProvider) getAvailableProvider() *ProviderStatus {
	log.Info("getAvailableProvider")

	for {
		lb.mu.RLock()
		// First try: find any non-occupied provider
		for _, ps := range lb.providers {
			ps.mu.Lock()
			if !ps.occupied {
				lb.mu.RUnlock()
				return ps
			}
			ps.mu.Unlock()
		}

		// All providers are busy, find least recently used
		var leastRecent *ProviderStatus
		oldestTime := time.Now()

		for _, ps := range lb.providers {
			ps.mu.Lock()
			if ps.lastUsed.Before(oldestTime) {
				oldestTime = ps.lastUsed
				leastRecent = ps
			}
			ps.mu.Unlock()
		}

		if leastRecent != nil {
			lb.mu.RUnlock()
			// Wait a bit before reusing a busy provider
			time.Sleep(100 * time.Millisecond)
			return leastRecent
		}

		lb.mu.RUnlock()
		// If all providers are busy, wait a bit before trying again
		time.Sleep(100 * time.Millisecond)
	}
}

func (lb *BalancedProvider) GenerateSync(ctx context.Context, messages []Message) (string, error) {
	log.Info("GenerateSync")

	events := lb.Generate(ctx, messages)
	var result string

	for event := range events {
		if event.Type == EventError {
			return "", errors.New(event.Content)
		}
		result += event.Content
	}

	return result, nil
}

// Add monitoring methods
func (lb *BalancedProvider) GetStats() []ProviderStats {
	log.Info("GetStats")

	var stats []ProviderStats
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	for _, ps := range lb.providers {
		ps.mu.Lock()
		stats = append(stats, ProviderStats{
			Occupied: ps.occupied,
			Requests: ps.requests,
			Failures: ps.failures,
			LastUsed: ps.lastUsed,
		})
		ps.mu.Unlock()
	}

	return stats
}

type ProviderStats struct {
	Occupied bool
	Requests int64
	Failures int64
	LastUsed time.Time
}

func (lb *BalancedProvider) Configure(config map[string]interface{}) {
	// Configuration can be added here
}
