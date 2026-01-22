package ai

import (
	"context"
	"fmt"
)

// MockClient is a thread-safe mock for the AI Client interface.
type MockClient struct {
	// Mock responses
	GenerateResponse *GenerateResponse
	GenerateError    error

	// Spies
	GenerateCalls []GenerateRequest
}

// GenerateContent records the call and returns the pre-configured response.
func (m *MockClient) GenerateContent(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	m.GenerateCalls = append(m.GenerateCalls, req)
	if m.GenerateError != nil {
		return nil, m.GenerateError
	}
	if m.GenerateResponse == nil {
		return nil, fmt.Errorf("mock response not configured")
	}
	return m.GenerateResponse, nil
}

// Close is a no-op for the mock.
func (m *MockClient) Close() error {
	return nil
}
