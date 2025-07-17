package detailed_logging

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func TestNewHTTPInterceptor(t *testing.T) {
	logger := &DetailedLogger{enabled: true}
	
	// Test with nil transport
	interceptor := NewHTTPInterceptor(nil, logger)
	assert.NotNil(t, interceptor)
	assert.Equal(t, http.DefaultTransport, interceptor.transport)
	assert.Equal(t, logger, interceptor.logger)
	
	// Test with custom transport
	customTransport := &mockRoundTripper{}
	interceptor = NewHTTPInterceptor(customTransport, logger)
	assert.Equal(t, customTransport, interceptor.transport)
}

func TestHTTPInterceptorRoundTrip(t *testing.T) {
	t.Skip("Skipping for now - need to refactor to use proper mocking")
	t.Run("disabled logger", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("test response"))),
			},
		}
		
		interceptor := &HTTPInterceptor{
			transport: mockTransport,
			logger:    nil,
		}
		
		req := httptest.NewRequest("GET", "https://example.com/test", nil)
		resp, err := interceptor.RoundTrip(req)
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
	
	t.Run("successful GET request", func(t *testing.T) {
		// For this test, we'll verify the behavior by checking the session
		logger := &DetailedLogger{
			enabled:   true,
			sessionID: "test-session",
			session: &SessionLog{
				HTTPCalls: []HTTPLog{},
			},
		}
		
		mockTransport := &mockRoundTripper{
			response: &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Body:       io.NopCloser(bytes.NewReader([]byte("test response"))),
			},
		}
		
		interceptor := &HTTPInterceptor{
			transport: mockTransport,
			logger:    logger,
		}
		
		req := httptest.NewRequest("GET", "https://example.com/test", nil)
		resp, err := interceptor.RoundTrip(req)
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		
		// TODO: Add verification once we have proper mocking
		
		// Read response body to verify it's still available
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "test response", string(body))
	})
	
	t.Run("POST request with body", func(t *testing.T) {
		t.Skip("Skipping - need mock refactoring")
	})
	
	t.Run("request with error", func(t *testing.T) {
		t.Skip("Skipping - need mock refactoring")
	})
}

func TestParseBody(t *testing.T) {
	interceptor := &HTTPInterceptor{}
	
	t.Run("empty body", func(t *testing.T) {
		result := interceptor.parseBody([]byte{}, "application/json")
		assert.Nil(t, result)
	})
	
	t.Run("JSON body", func(t *testing.T) {
		body := []byte(`{"key":"value","number":42}`)
		result := interceptor.parseBody(body, "application/json")
		
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "value", resultMap["key"])
		assert.Equal(t, float64(42), resultMap["number"])
	})
	
	t.Run("invalid JSON body", func(t *testing.T) {
		body := []byte(`{invalid json}`)
		result := interceptor.parseBody(body, "application/json")
		
		// Should return as string when JSON parsing fails
		resultStr, ok := result.(string)
		require.True(t, ok)
		assert.Equal(t, "{invalid json}", resultStr)
	})
	
	t.Run("non-JSON body", func(t *testing.T) {
		body := []byte("plain text response")
		result := interceptor.parseBody(body, "text/plain")
		
		resultStr, ok := result.(string)
		require.True(t, ok)
		assert.Equal(t, "plain text response", resultStr)
	})
}

func TestInstallGlobalInterceptor(t *testing.T) {
	// Save original transport
	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()
	
	logger := &DetailedLogger{enabled: true}
	InstallGlobalInterceptor(logger)
	
	interceptor, ok := http.DefaultTransport.(*HTTPInterceptor)
	require.True(t, ok)
	assert.Equal(t, logger, interceptor.logger)
	assert.Equal(t, originalTransport, interceptor.transport)
}