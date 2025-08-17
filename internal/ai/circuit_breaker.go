package ai

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int           // Number of failures before opening
	ResetTimeout     time.Duration // Time to wait before transitioning to half-open
	SuccessThreshold int           // Number of successes in half-open to close
	RequestThreshold int           // Minimum requests before evaluating failure rate
	WindowSize       time.Duration // Rolling window for failure tracking
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		SuccessThreshold: 3,
		RequestThreshold: 10,
		WindowSize:       60 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern for provider resilience
type CircuitBreaker struct {
	mu              sync.RWMutex
	config          CircuitBreakerConfig
	state           CircuitState
	failures        int
	successes       int
	requests        int
	lastFailureTime time.Time
	lastTransition  time.Time
	requestWindow   []requestRecord
	providerName    string
}

// requestRecord tracks individual request outcomes
type requestRecord struct {
	timestamp time.Time
	success   bool
}

// NewCircuitBreaker creates a new circuit breaker for a provider
func NewCircuitBreaker(providerName string, config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config:         config,
		state:          CircuitClosed,
		lastTransition: time.Now(),
		requestWindow:  make([]requestRecord, 0),
		providerName:   providerName,
	}
}

// Execute wraps a function call with circuit breaker logic
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if !cb.AllowRequest() {
		return NewCircuitBreakerError(cb.providerName, cb.state, "circuit breaker is open")
	}

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	cb.RecordResult(err == nil)

	if err != nil {
		return fmt.Errorf("provider %s failed after %v: %w", cb.providerName, duration, err)
	}

	return nil
}

// AllowRequest checks if a request should be allowed through
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		return time.Since(cb.lastTransition) >= cb.config.ResetTimeout
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordResult records the outcome of a request and updates circuit state
func (cb *CircuitBreaker) RecordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	cb.requests++

	// Add to rolling window
	cb.requestWindow = append(cb.requestWindow, requestRecord{
		timestamp: now,
		success:   success,
	})

	// Clean old entries from window
	cb.cleanWindow(now)

	if success {
		cb.onSuccess()
	} else {
		cb.onFailure(now)
	}
}

// onSuccess handles successful request outcomes
func (cb *CircuitBreaker) onSuccess() {
	cb.successes++

	switch cb.state {
	case CircuitHalfOpen:
		if cb.successes >= cb.config.SuccessThreshold {
			cb.setState(CircuitClosed)
			cb.reset()
		}
	case CircuitOpen:
		// Transition to half-open on first success after timeout
		if time.Since(cb.lastTransition) >= cb.config.ResetTimeout {
			cb.setState(CircuitHalfOpen)
			cb.successes = 1
		}
	}
}

// onFailure handles failed request outcomes
func (cb *CircuitBreaker) onFailure(timestamp time.Time) {
	cb.failures++
	cb.lastFailureTime = timestamp

	switch cb.state {
	case CircuitClosed:
		if cb.shouldOpen() {
			cb.setState(CircuitOpen)
		}
	case CircuitHalfOpen:
		// Any failure in half-open state immediately opens circuit
		cb.setState(CircuitOpen)
	}
}

// shouldOpen determines if circuit should transition to open state
func (cb *CircuitBreaker) shouldOpen() bool {
	// Need minimum requests before considering opening
	if cb.requests < cb.config.RequestThreshold {
		return false
	}

	// Check failure rate in current window
	windowFailures := 0
	for _, record := range cb.requestWindow {
		if !record.success {
			windowFailures++
		}
	}

	failureRate := float64(windowFailures) / float64(len(cb.requestWindow))
	thresholdRate := float64(cb.config.FailureThreshold) / float64(cb.config.RequestThreshold)

	return failureRate >= thresholdRate
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitState) {
	oldState := cb.state
	cb.state = newState
	cb.lastTransition = time.Now()

	// Log state transitions for monitoring
	if oldState != newState {
		fmt.Printf("Circuit breaker for %s: %s -> %s\n", cb.providerName, oldState, newState)
	}
}

// reset clears counters when circuit is closed
func (cb *CircuitBreaker) reset() {
	cb.failures = 0
	cb.successes = 0
	cb.requests = 0
	cb.requestWindow = cb.requestWindow[:0]
}

// cleanWindow removes old entries from the rolling window
func (cb *CircuitBreaker) cleanWindow(now time.Time) {
	cutoff := now.Add(-cb.config.WindowSize)

	// Find first entry within window
	start := 0
	for i, record := range cb.requestWindow {
		if record.timestamp.After(cutoff) {
			start = i
			break
		}
	}

	// Keep only recent entries
	if start > 0 {
		copy(cb.requestWindow, cb.requestWindow[start:])
		cb.requestWindow = cb.requestWindow[:len(cb.requestWindow)-start]
	}
}

// GetState returns current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	windowFailures := 0
	for _, record := range cb.requestWindow {
		if !record.success {
			windowFailures++
		}
	}

	failureRate := 0.0
	if len(cb.requestWindow) > 0 {
		failureRate = float64(windowFailures) / float64(len(cb.requestWindow))
	}

	return CircuitBreakerStats{
		Provider:        cb.providerName,
		State:           cb.state,
		Failures:        cb.failures,
		Successes:       cb.successes,
		Requests:        cb.requests,
		WindowFailures:  windowFailures,
		WindowRequests:  len(cb.requestWindow),
		FailureRate:     failureRate,
		LastFailureTime: cb.lastFailureTime,
		LastTransition:  cb.lastTransition,
		TimeInState:     time.Since(cb.lastTransition),
	}
}

// CircuitBreakerStats contains circuit breaker metrics
type CircuitBreakerStats struct {
	Provider        string        `json:"provider"`
	State           CircuitState  `json:"state"`
	Failures        int           `json:"failures"`
	Successes       int           `json:"successes"`
	Requests        int           `json:"requests"`
	WindowFailures  int           `json:"window_failures"`
	WindowRequests  int           `json:"window_requests"`
	FailureRate     float64       `json:"failure_rate"`
	LastFailureTime time.Time     `json:"last_failure_time"`
	LastTransition  time.Time     `json:"last_transition"`
	TimeInState     time.Duration `json:"time_in_state"`
}

// CircuitBreakerError represents circuit breaker related errors
type CircuitBreakerError struct {
	Provider string
	State    CircuitState
	Message  string
}

func (e *CircuitBreakerError) Error() string {
	return fmt.Sprintf("circuit breaker error for %s (state: %s): %s", e.Provider, e.State, e.Message)
}

func (e *CircuitBreakerError) IsCircuitOpen() bool {
	return e.State == CircuitOpen
}

// NewCircuitBreakerError creates a new circuit breaker error
func NewCircuitBreakerError(provider string, state CircuitState, message string) *CircuitBreakerError {
	return &CircuitBreakerError{
		Provider: provider,
		State:    state,
		Message:  message,
	}
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	config   CircuitBreakerConfig
}

// NewCircuitBreakerManager creates a new manager for circuit breakers
func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		config:   config,
	}
}

// GetBreaker returns or creates a circuit breaker for a provider
func (m *CircuitBreakerManager) GetBreaker(providerName string) *CircuitBreaker {
	m.mu.RLock()
	breaker, exists := m.breakers[providerName]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check pattern
	if breaker, exists := m.breakers[providerName]; exists {
		return breaker
	}

	breaker = NewCircuitBreaker(providerName, m.config)
	m.breakers[providerName] = breaker
	return breaker
}

// ExecuteWithBreaker executes a function with circuit breaker protection
func (m *CircuitBreakerManager) ExecuteWithBreaker(ctx context.Context, providerName string, fn func(ctx context.Context) error) error {
	breaker := m.GetBreaker(providerName)
	return breaker.Execute(ctx, fn)
}

// GetAllStats returns stats for all circuit breakers
func (m *CircuitBreakerManager) GetAllStats() map[string]CircuitBreakerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]CircuitBreakerStats)
	for name, breaker := range m.breakers {
		stats[name] = breaker.GetStats()
	}
	return stats
}

// ResetBreaker manually resets a circuit breaker
func (m *CircuitBreakerManager) ResetBreaker(providerName string) {
	m.mu.RLock()
	breaker, exists := m.breakers[providerName]
	m.mu.RUnlock()

	if exists {
		breaker.mu.Lock()
		breaker.setState(CircuitClosed)
		breaker.reset()
		breaker.mu.Unlock()
	}
}

// GetHealthyProviders returns list of providers with closed circuits
func (m *CircuitBreakerManager) GetHealthyProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthy := make([]string, 0)
	for name, breaker := range m.breakers {
		if breaker.GetState() == CircuitClosed {
			healthy = append(healthy, name)
		}
	}
	return healthy
}
