package provider

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/data"
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
	pr          *io.PipeReader
	pw          *io.PipeWriter
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
	errnie.Trace("%s", "provider.NewBalancedProvider", "new")

	onceBalancedProvider.Do(func() {
		pr, pw := io.Pipe()
		balancedProviderInstance = &BalancedProvider{
			pr: pr,
			pw: pw,
			providers: []*ProviderStatus{
				{
					name: "gpt-4o",
					provider: NewOpenAI(
						"https://api.openai.com/v1",
						os.Getenv("OPENAI_API_KEY"),
						openai.ChatModelGPT4oMini2024_07_18,
					),
					occupied: false,
				},
				// {
				// 	name:     "claude-3-5-sonnet",
				// 	provider: NewAnthropic(os.Getenv("ANTHROPIC_API_KEY"), anthropic.ModelClaude3_5Sonnet20241022),
				// 	occupied: false,
				// },
				// {
				// 	name:     "gemini-1.5-flash",
				// 	provider: NewGoogle(os.Getenv("GEMINI_API_KEY"), "gemini-1.5-flash"),
				// 	occupied: false,
				// },
				// {
				// 	name:     "command-r",
				// 	provider: NewCohere(os.Getenv("COHERE_API_KEY"), "command-r"),
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

func (lb *BalancedProvider) Read(p []byte) (n int, err error) {
	errnie.Trace("BalancedProvider.Read START with buffer size: %d", len(p))

	if lb.pr == nil {
		errnie.Trace("BalancedProvider.Read pipe reader is nil")
		return 0, io.EOF
	}

	n, err = lb.pr.Read(p)
	errnie.Trace("BalancedProvider.Read got %d bytes, err: %v", n, err)
	return n, err
}

func (lb *BalancedProvider) Write(p []byte) (n int, err error) {
	errnie.Trace("BalancedProvider.Write START with %d bytes", len(p))
	ps := lb.getAvailableProvider()
	if ps == nil {
		return 0, errors.New("no provider available")
	}

	errnie.Trace("BalancedProvider.Write got provider: %s", ps.name)

	// Start a goroutine to handle the response
	go func() {
		defer func() {
			errnie.Trace("BalancedProvider.Write goroutine CLEANUP")
			ps.mu.Lock()
			ps.occupied = false
			ps.mu.Unlock()
		}()

		// Write to provider
		errnie.Trace("BalancedProvider.Write WRITING to provider")
		if _, err := ps.provider.Write(p); err != nil {
			errnie.Error(err)
			errnie.Trace("BalancedProvider.Write failed to write to provider: %v", err)
			return
		}

		errnie.Trace("BalancedProvider.Write STARTING to read response")

		// Read from provider and write to balanced provider's pipe
		buf := make([]byte, 1024)
		for {
			errnie.Trace("BalancedProvider.Write reading loop START")
			n, err := ps.provider.Read(buf)
			errnie.Trace("BalancedProvider.Write read %d bytes, err: %v", n, err)
			
			if err != nil {
				if err != io.EOF {
					errnie.Error(err)
				}
				errnie.Trace("BalancedProvider.Write closing pipe due to error/EOF")
				lb.pw.Close()
				return
			}

			if n > 0 {
				errnie.Trace("BalancedProvider.Write writing %d bytes to pipe", n)
				if _, err := lb.pw.Write(buf[:n]); err != nil {
					errnie.Error(err)
					errnie.Trace("BalancedProvider.Write failed to write to pipe: %v", err)
					lb.pw.Close()
					return
				}
				errnie.Trace("BalancedProvider.Write successfully wrote to pipe")
			}
		}
	}()

	errnie.Trace("BalancedProvider.Write returning with len: %d", len(p))
	return len(p), nil
}

func (lb *BalancedProvider) Close() error {
	errnie.Trace("%s", "BalancedProvider.Close", "close")

	if lb.pw != nil {
		lb.pw.Close()
	}
	if lb.pr != nil {
		lb.pr.Close()
	}
	return nil
}

func (lb *BalancedProvider) Generate(ctx context.Context, params GenerationParams) <-chan Event {
	errnie.Trace("%s", "BalancedProvider.Generate", "generate")

	out := make(chan Event)

	go func() {
		defer close(out)

		// Create artifact from params
		artifact := data.New("balanced", "user", "generate", []byte(params.String()))

		// Write artifact to provider
		if _, err := io.Copy(lb, artifact); err != nil {
			errnie.Error(err)
			return
		}

		// Read responses
		buf := make([]byte, 1024)
		for {
			n, err := lb.Read(buf)
			if err != nil {
				if err != io.EOF {
					errnie.Error(err)
				}
				break
			}

			// Convert response to event
			responseArtifact := data.Empty()
			responseArtifact.Unmarshal(buf[:n])
			if responseArtifact != nil {
				out <- Event{
					Type:    EventToken,
					Content: responseArtifact.Peek("payload"),
				}
			}
		}

		out <- Event{Type: EventDone}
	}()

	return out
}

func (lb *BalancedProvider) getAvailableProvider() *ProviderStatus {
	errnie.Trace("%s", "BalancedProvider.getAvailableProvider", "get")

	// Handle first request with random selection
	if provider := lb.handleFirstRequest(); provider != nil {
		return provider
	}

	return lb.findBestAvailableProvider()
}

func (lb *BalancedProvider) handleFirstRequest() *ProviderStatus {
	errnie.Trace("%s", "BalancedProvider.handleFirstRequest", "handle")

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
	errnie.Trace("%s", "BalancedProvider.findBestAvailableProvider", "find")

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
	errnie.Trace("%s", "BalancedProvider.selectBestProvider", "select")

	var bestProvider *ProviderStatus
	oldestUse := time.Now()
	cooldownPeriod := 60 * time.Second
	maxFailures := 3

	for _, ps := range lb.providers {
		ps.mu.Lock()
		defer ps.mu.Unlock()

		if !lb.isProviderAvailable(ps, cooldownPeriod, maxFailures) {
			continue
		}

		if lb.isBetterProvider(ps, bestProvider, oldestUse) {
			bestProvider = ps
			oldestUse = ps.lastUsed
		}
	}

	if bestProvider != nil {
		lb.markProviderAsOccupied(bestProvider)
	}

	return bestProvider
}

func (lb *BalancedProvider) getUnoccupiedProviders() []*ProviderStatus {
	errnie.Trace("%s", "BalancedProvider.getUnoccupiedProviders", "get")

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
	errnie.Trace("%s", "BalancedProvider.isProviderAvailable", "is")

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
	errnie.Trace("%s", "BalancedProvider.isBetterProvider", "is")

	return current == nil ||
		candidate.failures < current.failures ||
		(candidate.failures == current.failures && candidate.lastUsed.Before(oldestUse))
}

func (lb *BalancedProvider) markProviderAsOccupied(ps *ProviderStatus) {
	errnie.Trace("%s", "BalancedProvider.markProviderAsOccupied", "mark")

	ps.occupied = true
	ps.lastUsed = time.Now()
}

func (lb *BalancedProvider) Configure(config map[string]interface{}) {
	errnie.Trace("%s", "BalancedProvider.Configure", "configure")

	// Configuration can be added here
}
