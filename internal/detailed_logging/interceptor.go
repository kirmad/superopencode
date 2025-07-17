package detailed_logging

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPInterceptor wraps http.RoundTripper to log requests/responses
type HTTPInterceptor struct {
	transport http.RoundTripper
	logger    *DetailedLogger
}

// NewHTTPInterceptor creates a new HTTP interceptor
func NewHTTPInterceptor(transport http.RoundTripper, logger *DetailedLogger) *HTTPInterceptor {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &HTTPInterceptor{
		transport: transport,
		logger:    logger,
	}
}

// RoundTrip implements http.RoundTripper
func (h *HTTPInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	if h.logger == nil || !h.logger.IsEnabled() {
		return h.transport.RoundTrip(req)
	}

	// Create HTTP log entry
	httpLog := &HTTPLog{
		ID:        NewID(),
		SessionID: h.logger.sessionID,
		Method:    req.Method,
		URL:       req.URL.String(),
		Headers:   req.Header,
		StartTime: time.Now(),
	}

	// Capture request body if present
	if req.Body != nil && req.Method != "GET" && req.Method != "HEAD" {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			httpLog.Body = h.parseBody(bodyBytes, req.Header.Get("Content-Type"))
		}
	}

	// Set parent tool call if in context
	if toolID := h.logger.GetCurrentToolCall(); toolID != "" {
		httpLog.ParentToolCall = toolID
	}

	// Make the actual request
	resp, err := h.transport.RoundTrip(req)

	// Complete the log entry
	endTime := time.Now()
	httpLog.EndTime = &endTime
	httpLog.DurationMs = CalculateDuration(httpLog.StartTime, httpLog.EndTime)

	if err != nil {
		httpLog.Error = err.Error()
		h.logger.LogHTTP(httpLog)
		return nil, err
	}

	// Capture response
	httpLog.StatusCode = resp.StatusCode
	httpLog.ResponseHeaders = resp.Header

	// Capture response body if present
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			httpLog.ResponseBody = h.parseBody(bodyBytes, resp.Header.Get("Content-Type"))
		}
	}

	// Log the complete HTTP transaction
	h.logger.LogHTTP(httpLog)

	return resp, nil
}

// parseBody attempts to parse the body based on content type
func (h *HTTPInterceptor) parseBody(body []byte, contentType string) interface{} {
	if len(body) == 0 {
		return nil
	}

	// Check if it's JSON
	if strings.Contains(contentType, "application/json") {
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			return jsonData
		}
	}

	// Return as string for other content types
	return string(body)
}

// InstallGlobalInterceptor replaces http.DefaultTransport
func InstallGlobalInterceptor(logger *DetailedLogger) {
	http.DefaultTransport = NewHTTPInterceptor(http.DefaultTransport, logger)
}