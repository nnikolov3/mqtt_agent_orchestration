package ai

import (
	"errors"
	"fmt"
)

// Common AI system errors
var (
	// Provider errors
	ErrProviderNotFound    = errors.New("provider not found")
	ErrProviderUnavailable = errors.New("provider unavailable")
	ErrProviderFailed      = errors.New("provider request failed")
	ErrProviderTimeout     = errors.New("provider request timeout")

	// Model errors
	ErrModelNotFound     = errors.New("model not found")
	ErrModelNotSupported = errors.New("model not supported by provider")
	ErrInvalidModel      = errors.New("invalid model configuration")
	ErrModelLoadFailed   = errors.New("failed to load model")

	// Request errors
	ErrInvalidRequest  = errors.New("invalid request")
	ErrInvalidResponse = errors.New("invalid response")
	ErrRequestTooLarge = errors.New("request exceeds size limits")
	ErrRateLimited     = errors.New("rate limit exceeded")

	// Cost errors
	ErrCostLimitExceeded = errors.New("cost limit exceeded")
	ErrInvalidCostConfig = errors.New("invalid cost configuration")

	// Resource errors
	ErrInsufficientMemory = errors.New("insufficient memory")
	ErrResourceExhausted  = errors.New("resource exhausted")

	// Configuration errors
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrMissingCredentials = errors.New("missing credentials")
)

// AIError wraps errors with additional context
type AIError struct {
	Code      string
	Message   string
	Provider  string
	Model     string
	RequestID string
	Cause     error
}

// Error implements the error interface
func (e *AIError) Error() string {
	if e.Provider != "" && e.Model != "" {
		return fmt.Sprintf("[%s] %s (provider: %s, model: %s)", e.Code, e.Message, e.Provider, e.Model)
	}
	if e.Provider != "" {
		return fmt.Sprintf("[%s] %s (provider: %s)", e.Code, e.Message, e.Provider)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AIError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches a target error
func (e *AIError) Is(target error) bool {
	if e.Cause != nil && errors.Is(e.Cause, target) {
		return true
	}
	return false
}

// NewAIError creates a new AIError
func NewAIError(code, message, provider, model, requestID string, cause error) *AIError {
	return &AIError{
		Code:      code,
		Message:   message,
		Provider:  provider,
		Model:     model,
		RequestID: requestID,
		Cause:     cause,
	}
}

// WrapProviderError wraps a provider error with context
func WrapProviderError(err error, provider, model, requestID string) error {
	if err == nil {
		return nil
	}

	var code, message string

	switch {
	case errors.Is(err, ErrProviderTimeout):
		code = "PROVIDER_TIMEOUT"
		message = "Provider request timed out"
	case errors.Is(err, ErrProviderUnavailable):
		code = "PROVIDER_UNAVAILABLE"
		message = "Provider is currently unavailable"
	case errors.Is(err, ErrRateLimited):
		code = "RATE_LIMITED"
		message = "Rate limit exceeded"
	case errors.Is(err, ErrModelNotSupported):
		code = "MODEL_NOT_SUPPORTED"
		message = "Model not supported by provider"
	case errors.Is(err, ErrInvalidRequest):
		code = "INVALID_REQUEST"
		message = "Invalid request format"
	case errors.Is(err, ErrInvalidResponse):
		code = "INVALID_RESPONSE"
		message = "Invalid response from provider"
	default:
		code = "PROVIDER_ERROR"
		message = "Provider request failed"
	}

	return NewAIError(code, message, provider, model, requestID, err)
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific retryable errors
	switch {
	case errors.Is(err, ErrProviderTimeout):
		return true
	case errors.Is(err, ErrProviderUnavailable):
		return true
	case errors.Is(err, ErrRateLimited):
		return false // Don't retry rate limits immediately
	case errors.Is(err, ErrResourceExhausted):
		return true
	default:
		return false
	}
}

// IsTemporaryError checks if an error is likely temporary
func IsTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrProviderTimeout) ||
		errors.Is(err, ErrProviderUnavailable) ||
		errors.Is(err, ErrResourceExhausted)
}

// IsCostError checks if an error is related to cost limits
func IsCostError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrCostLimitExceeded) ||
		errors.Is(err, ErrInvalidCostConfig)
}

// IsConfigurationError checks if an error is related to configuration
func IsConfigurationError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrInvalidConfig) ||
		errors.Is(err, ErrMissingCredentials) ||
		errors.Is(err, ErrInvalidModel)
}
