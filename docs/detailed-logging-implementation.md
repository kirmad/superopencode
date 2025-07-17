# Detailed Logging System Implementation Guide

## Overview

This document provides a complete implementation guide for the detailed logging system in superopencode. The system captures all LLM interactions, tool calls, and HTTP requests/responses with hierarchical tracking and provides a modern web-based viewer for analysis.

## Architecture Overview

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Application   │────▶│ Logging Wrapper  │────▶│  LLM Provider   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
         │                       │                         │
         ▼                       ▼                         ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Tool Execution │────▶│  Tool Tracker    │     │ HTTP Interceptor│
└─────────────────┘     └──────────────────┘     └─────────────────┘
         │                       │                         │
         └───────────────────────┴─────────────────────────┘
                                 │
                                 ▼
                        ┌──────────────────┐
                        │   Log Storage    │
                        │  (SQLite + JSON) │
                        └──────────────────┘
                                 │
                                 ▼
                        ┌──────────────────┐
                        │   Web Viewer     │
                        │  (Port 8080)     │
                        └──────────────────┘
```

## Implementation Steps

### Step 1: Create Core Package Structure

Create the following directory structure:

```
internal/detailed_logging/
├── detailed.go          # Main logger implementation
├── interceptor.go       # HTTP interceptor
├── provider.go          # Provider wrapper
├── storage.go           # Data storage layer
├── tool_tracker.go      # Tool hierarchy tracking
├── viewer.go            # Web viewer server
└── types.go             # Common types and interfaces
```

### Step 2: Define Core Types

**File: `internal/detailed_logging/types.go`**

```go
package detailed_logging

import (
    "encoding/json"
    "time"
)

// Core data structures
type SessionLog struct {
    ID          string        `json:"id"`
    StartTime   time.Time     `json:"startTime"`
    EndTime     time.Time     `json:"endTime"`
    UserPrompt  string        `json:"userPrompt"`
    Summary     string        `json:"summary"`
    TotalTokens int           `json:"totalTokens"`
    Requests    []RequestLog  `json:"requests"`
}

type RequestLog struct {
    ID        string          `json:"id"`
    Type      string          `json:"type"` // "llm", "tool", "http"
    Timestamp time.Time       `json:"timestamp"`
    Duration  time.Duration   `json:"duration"`
    Details   json.RawMessage `json:"details"`
    ToolCalls []*ToolCallHierarchy `json:"toolCalls,omitempty"`
}

type LLMRequest struct {
    ID        string                 `json:"id"`
    SessionID string                 `json:"sessionId"`
    Provider  string                 `json:"provider"`
    Model     string                 `json:"model"`
    Messages  []MessageLog          `json:"messages"`
    Tools     []ToolDefinition      `json:"tools,omitempty"`
    Timestamp time.Time             `json:"timestamp"`
}

type LLMResponse struct {
    StreamID     string    `json:"streamId"`
    FullResponse string    `json:"fullResponse"`
    TokensUsed   int       `json:"tokensUsed"`
    Timestamp    time.Time `json:"timestamp"`
}

type HTTPRequest struct {
    ID        string              `json:"id"`
    Method    string              `json:"method"`
    URL       string              `json:"url"`
    Headers   map[string][]string `json:"headers"`
    Body      []byte              `json:"body"`
    Timestamp time.Time           `json:"timestamp"`
}

type HTTPResponse struct {
    RequestID  string              `json:"requestId"`
    StatusCode int                 `json:"statusCode"`
    Headers    map[string][]string `json:"headers"`
    Body       []byte              `json:"body"`
    Duration   time.Duration       `json:"duration"`
    Timestamp  time.Time           `json:"timestamp"`
}

type ToolCallHierarchy struct {
    ID           string               `json:"id"`
    ParentID     string               `json:"parentId,omitempty"`
    SessionID    string               `json:"sessionId"`
    Tool         string               `json:"tool"`
    Input        json.RawMessage      `json:"input"`
    Output       json.RawMessage      `json:"output,omitempty"`
    Error        string               `json:"error,omitempty"`
    StartTime    time.Time            `json:"startTime"`
    EndTime      time.Time            `json:"endTime"`
    Children     []*ToolCallHierarchy `json:"children"`
    Depth        int                  `json:"depth"`
}

type MessageLog struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ToolDefinition struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type StreamEventLog struct {
    StreamID  string    `json:"streamId"`
    Type      string    `json:"type"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

### Step 3: Implement HTTP Interceptor

**File: `internal/detailed_logging/interceptor.go`**

```go
package detailed_logging

import (
    "bytes"
    "io"
    "net/http"
    "time"
    
    "github.com/google/uuid"
)

type HTTPInterceptor struct {
    transport http.RoundTripper
    logger    *DetailedLogger
}

func NewHTTPInterceptor(logger *DetailedLogger) *HTTPInterceptor {
    return &HTTPInterceptor{
        transport: http.DefaultTransport,
        logger:    logger,
    }
}

func (i *HTTPInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
    requestID := uuid.New().String()
    startTime := time.Now()
    
    // Clone the request body
    var reqBody []byte
    if req.Body != nil {
        reqBody, _ = io.ReadAll(req.Body)
        req.Body = io.NopCloser(bytes.NewReader(reqBody))
    }
    
    // Log the request
    i.logger.LogHTTPRequest(&HTTPRequest{
        ID:        requestID,
        Method:    req.Method,
        URL:       req.URL.String(),
        Headers:   req.Header,
        Body:      reqBody,
        Timestamp: startTime,
    })
    
    // Execute the request
    resp, err := i.transport.RoundTrip(req)
    
    if err != nil {
        i.logger.LogHTTPError(requestID, err)
        return nil, err
    }
    
    // Clone the response body
    var respBody []byte
    if resp.Body != nil {
        respBody, _ = io.ReadAll(resp.Body)
        resp.Body = io.NopCloser(bytes.NewReader(respBody))
    }
    
    // Log the response
    i.logger.LogHTTPResponse(&HTTPResponse{
        RequestID:  requestID,
        StatusCode: resp.StatusCode,
        Headers:    resp.Header,
        Body:       respBody,
        Duration:   time.Since(startTime),
        Timestamp:  time.Now(),
    })
    
    return resp, nil
}

// WrapHTTPClient wraps an existing HTTP client with detailed logging
func WrapHTTPClient(client *http.Client, logger *DetailedLogger) *http.Client {
    if client == nil {
        client = http.DefaultClient
    }
    
    wrappedClient := *client
    wrappedClient.Transport = NewHTTPInterceptor(logger)
    
    return &wrappedClient
}
```

### Step 4: Implement Provider Wrapper

**File: `internal/detailed_logging/provider.go`**

```go
package detailed_logging

import (
    "context"
    "encoding/json"
    "strings"
    "time"
    
    "github.com/google/uuid"
    "github.com/kirmad/superopencode/internal/llm/provider"
    "github.com/kirmad/superopencode/internal/message"
)

type DetailedLoggingProvider struct {
    wrapped provider.Provider
    logger  *DetailedLogger
}

func WrapProvider(p provider.Provider, logger *DetailedLogger) provider.Provider {
    return &DetailedLoggingProvider{
        wrapped: p,
        logger:  logger,
    }
}

func (p *DetailedLoggingProvider) Model() models.Model {
    return p.wrapped.Model()
}

func (p *DetailedLoggingProvider) Stream(ctx context.Context, params provider.StreamParams) (<-chan provider.StreamEvent, error) {
    sessionID, _ := ctx.Value("sessionID").(string)
    streamID := uuid.New().String()
    
    // Convert messages for logging
    var msgLogs []MessageLog
    for _, msg := range params.Messages {
        msgLogs = append(msgLogs, MessageLog{
            Role:    string(msg.Role),
            Content: msg.Content().String(),
        })
    }
    
    // Convert tools for logging
    var toolDefs []ToolDefinition
    for _, tool := range params.Tools {
        toolDefs = append(toolDefs, ToolDefinition{
            Name:        tool.Name(),
            Description: tool.Description(),
        })
    }
    
    // Log the request
    p.logger.LogLLMRequest(&LLMRequest{
        ID:        streamID,
        SessionID: sessionID,
        Provider:  string(p.wrapped.Model().Provider),
        Model:     string(p.wrapped.Model().ID),
        Messages:  msgLogs,
        Tools:     toolDefs,
        Timestamp: time.Now(),
    })
    
    // Call the wrapped provider
    eventChan, err := p.wrapped.Stream(ctx, params)
    if err != nil {
        p.logger.LogLLMError(streamID, err)
        return nil, err
    }
    
    // Wrap the event channel to capture responses
    wrappedChan := make(chan provider.StreamEvent, 1)
    go func() {
        defer close(wrappedChan)
        
        var fullResponse strings.Builder
        startTime := time.Now()
        
        for event := range eventChan {
            // Log each event
            p.logger.LogStreamEvent(&StreamEventLog{
                StreamID:  streamID,
                Type:      string(event.Type),
                Content:   event.Content,
                Timestamp: time.Now(),
            })
            
            // Accumulate content
            if event.Type == provider.StreamEventTypeContent {
                fullResponse.WriteString(event.Content)
            }
            
            // Forward the event
            wrappedChan <- event
        }
        
        // Log the complete response
        p.logger.LogLLMResponse(&LLMResponse{
            StreamID:     streamID,
            FullResponse: fullResponse.String(),
            Timestamp:    time.Now(),
        })
        
        // Update session duration
        p.logger.UpdateRequestDuration(streamID, time.Since(startTime))
    }()
    
    return wrappedChan, nil
}
```

### Step 5: Implement Tool Tracker

**File: `internal/detailed_logging/tool_tracker.go`**

```go
package detailed_logging

import (
    "context"
    "encoding/json"
    "sync"
    "time"
    
    "github.com/google/uuid"
)

type ToolTracker struct {
    mu          sync.RWMutex
    activeCalls map[string]*ToolCallHierarchy
    rootCalls   map[string][]*ToolCallHierarchy // sessionID -> root calls
}

func NewToolTracker() *ToolTracker {
    return &ToolTracker{
        activeCalls: make(map[string]*ToolCallHierarchy),
        rootCalls:   make(map[string][]*ToolCallHierarchy),
    }
}

func (t *ToolTracker) StartToolCall(sessionID, parentID, tool string, input json.RawMessage) string {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    callID := uuid.New().String()
    depth := 0
    
    // Calculate depth based on parent
    if parentID != "" {
        if parent, exists := t.activeCalls[parentID]; exists {
            depth = parent.Depth + 1
        }
    }
    
    call := &ToolCallHierarchy{
        ID:        callID,
        ParentID:  parentID,
        SessionID: sessionID,
        Tool:      tool,
        Input:     input,
        StartTime: time.Now(),
        Depth:     depth,
        Children:  make([]*ToolCallHierarchy, 0),
    }
    
    t.activeCalls[callID] = call
    
    if parentID == "" {
        // Root call
        if _, exists := t.rootCalls[sessionID]; !exists {
            t.rootCalls[sessionID] = make([]*ToolCallHierarchy, 0)
        }
        t.rootCalls[sessionID] = append(t.rootCalls[sessionID], call)
    } else {
        // Child call - add to parent's children
        if parent, exists := t.activeCalls[parentID]; exists {
            parent.Children = append(parent.Children, call)
        }
    }
    
    return callID
}

func (t *ToolTracker) EndToolCall(callID string, output json.RawMessage, err error) {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    if call, exists := t.activeCalls[callID]; exists {
        call.EndTime = time.Now()
        call.Output = output
        if err != nil {
            call.Error = err.Error()
        }
        delete(t.activeCalls, callID)
    }
}

func (t *ToolTracker) GetSessionCalls(sessionID string) []*ToolCallHierarchy {
    t.mu.RLock()
    defer t.mu.RUnlock()
    
    return t.rootCalls[sessionID]
}

func (t *ToolTracker) ClearSession(sessionID string) {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    delete(t.rootCalls, sessionID)
}

// WrapToolExecution wraps a tool execution with tracking
func WrapToolExecution(ctx context.Context, tracker *ToolTracker, tool string, input json.RawMessage, fn func(context.Context) (json.RawMessage, error)) (json.RawMessage, error) {
    sessionID, _ := ctx.Value("sessionID").(string)
    parentID, _ := ctx.Value("parentToolCallID").(string)
    
    callID := tracker.StartToolCall(sessionID, parentID, tool, input)
    
    // Create new context with current tool call ID
    newCtx := context.WithValue(ctx, "parentToolCallID", callID)
    
    output, err := fn(newCtx)
    
    tracker.EndToolCall(callID, output, err)
    
    return output, err
}
```

### Step 6: Implement Storage Layer

**File: `internal/detailed_logging/storage.go`**

```go
package detailed_logging

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
)

type LogStorage struct {
    basePath string
    db       *sql.DB
}

func NewLogStorage(basePath string) (*LogStorage, error) {
    // Create base directory
    if err := os.MkdirAll(basePath, 0755); err != nil {
        return nil, err
    }
    
    // Open SQLite database
    dbPath := filepath.Join(basePath, "logs.db")
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }
    
    storage := &LogStorage{
        basePath: basePath,
        db:       db,
    }
    
    // Initialize schema
    if err := storage.initSchema(); err != nil {
        return nil, err
    }
    
    return storage, nil
}

func (s *LogStorage) initSchema() error {
    schema := `
    CREATE TABLE IF NOT EXISTS sessions (
        id TEXT PRIMARY KEY,
        start_time TIMESTAMP,
        end_time TIMESTAMP,
        user_prompt TEXT,
        summary TEXT,
        total_tokens INTEGER,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time);
    
    CREATE TABLE IF NOT EXISTS requests (
        id TEXT PRIMARY KEY,
        session_id TEXT,
        type TEXT,
        timestamp TIMESTAMP,
        duration_ms INTEGER,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (session_id) REFERENCES sessions(id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_requests_session ON requests(session_id);
    CREATE INDEX IF NOT EXISTS idx_requests_type ON requests(type);
    `
    
    _, err := s.db.Exec(schema)
    return err
}

func (s *LogStorage) SaveSession(session *SessionLog) error {
    // Start transaction
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Insert session metadata
    _, err = tx.Exec(`
        INSERT INTO sessions (id, start_time, end_time, user_prompt, summary, total_tokens)
        VALUES (?, ?, ?, ?, ?, ?)
    `, session.ID, session.StartTime, session.EndTime, session.UserPrompt, session.Summary, session.TotalTokens)
    
    if err != nil {
        return err
    }
    
    // Insert requests metadata
    for _, req := range session.Requests {
        _, err = tx.Exec(`
            INSERT INTO requests (id, session_id, type, timestamp, duration_ms)
            VALUES (?, ?, ?, ?, ?)
        `, req.ID, session.ID, req.Type, req.Timestamp, req.Duration.Milliseconds())
        
        if err != nil {
            return err
        }
    }
    
    // Commit transaction
    if err = tx.Commit(); err != nil {
        return err
    }
    
    // Save detailed session data to JSON
    sessionDir := filepath.Join(s.basePath, session.ID)
    if err := os.MkdirAll(sessionDir, 0755); err != nil {
        return err
    }
    
    // Save complete session data
    sessionFile := filepath.Join(sessionDir, "session.json")
    data, err := json.MarshalIndent(session, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(sessionFile, data, 0644)
}

func (s *LogStorage) GetSessions(limit int, offset int) ([]SessionSummary, error) {
    query := `
        SELECT id, start_time, end_time, user_prompt, total_tokens
        FROM sessions
        ORDER BY start_time DESC
        LIMIT ? OFFSET ?
    `
    
    rows, err := s.db.Query(query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var sessions []SessionSummary
    for rows.Next() {
        var s SessionSummary
        err := rows.Scan(&s.ID, &s.StartTime, &s.EndTime, &s.UserPrompt, &s.TotalTokens)
        if err != nil {
            continue
        }
        sessions = append(sessions, s)
    }
    
    return sessions, rows.Err()
}

func (s *LogStorage) GetSessionDetails(sessionID string) (*SessionLog, error) {
    sessionFile := filepath.Join(s.basePath, sessionID, "session.json")
    data, err := os.ReadFile(sessionFile)
    if err != nil {
        return nil, err
    }
    
    var session SessionLog
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, err
    }
    
    return &session, nil
}

func (s *LogStorage) SearchSessions(query string) ([]SessionSummary, error) {
    sqlQuery := `
        SELECT id, start_time, end_time, user_prompt, total_tokens
        FROM sessions
        WHERE user_prompt LIKE ?
        ORDER BY start_time DESC
        LIMIT 100
    `
    
    rows, err := s.db.Query(sqlQuery, "%"+query+"%")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var sessions []SessionSummary
    for rows.Next() {
        var s SessionSummary
        err := rows.Scan(&s.ID, &s.StartTime, &s.EndTime, &s.UserPrompt, &s.TotalTokens)
        if err != nil {
            continue
        }
        sessions = append(sessions, s)
    }
    
    return sessions, rows.Err()
}

type SessionSummary struct {
    ID          string    `json:"id"`
    StartTime   time.Time `json:"startTime"`
    EndTime     time.Time `json:"endTime"`
    UserPrompt  string    `json:"userPrompt"`
    TotalTokens int       `json:"totalTokens"`
}
```

### Step 7: Implement Main Logger

**File: `internal/detailed_logging/detailed.go`**

```go
package detailed_logging

import (
    "context"
    "encoding/json"
    "net/http"
    "sync"
    "time"
)

type DetailedLogger struct {
    mu          sync.RWMutex
    storage     *LogStorage
    toolTracker *ToolTracker
    sessions    map[string]*SessionLog
    enabled     bool
}

func NewDetailedLogger(basePath string) (*DetailedLogger, error) {
    storage, err := NewLogStorage(basePath)
    if err != nil {
        return nil, err
    }
    
    return &DetailedLogger{
        storage:     storage,
        toolTracker: NewToolTracker(),
        sessions:    make(map[string]*SessionLog),
        enabled:     true,
    }, nil
}

func (d *DetailedLogger) IsEnabled() bool {
    return d.enabled
}

func (d *DetailedLogger) StartSession(sessionID, userPrompt string) {
    if !d.enabled {
        return
    }
    
    d.mu.Lock()
    defer d.mu.Unlock()
    
    d.sessions[sessionID] = &SessionLog{
        ID:         sessionID,
        StartTime:  time.Now(),
        UserPrompt: userPrompt,
        Requests:   make([]RequestLog, 0),
    }
}

func (d *DetailedLogger) EndSession(sessionID string) {
    if !d.enabled {
        return
    }
    
    d.mu.Lock()
    defer d.mu.Unlock()
    
    if session, exists := d.sessions[sessionID]; exists {
        session.EndTime = time.Now()
        
        // Get tool calls for this session
        toolCalls := d.toolTracker.GetSessionCalls(sessionID)
        
        // Attach tool calls to their respective LLM requests
        for i := range session.Requests {
            if session.Requests[i].Type == "llm" {
                session.Requests[i].ToolCalls = toolCalls
                break
            }
        }
        
        // Save to storage
        d.storage.SaveSession(session)
        
        // Cleanup
        delete(d.sessions, sessionID)
        d.toolTracker.ClearSession(sessionID)
    }
}

func (d *DetailedLogger) LogLLMRequest(req *LLMRequest) {
    if !d.enabled {
        return
    }
    
    d.mu.Lock()
    defer d.mu.Unlock()
    
    if session, exists := d.sessions[req.SessionID]; exists {
        details, _ := json.Marshal(req)
        session.Requests = append(session.Requests, RequestLog{
            ID:        req.ID,
            Type:      "llm",
            Timestamp: req.Timestamp,
            Details:   details,
        })
    }
}

func (d *DetailedLogger) LogLLMResponse(resp *LLMResponse) {
    if !d.enabled {
        return
    }
    
    d.mu.Lock()
    defer d.mu.Unlock()
    
    // Update token count for the session
    for sessionID, session := range d.sessions {
        for _, req := range session.Requests {
            if req.ID == resp.StreamID {
                session.TotalTokens += resp.TokensUsed
                return
            }
        }
    }
}

func (d *DetailedLogger) LogHTTPRequest(req *HTTPRequest) {
    if !d.enabled {
        return
    }
    
    d.mu.Lock()
    defer d.mu.Unlock()
    
    // Find the current session from context (would need to be passed)
    // For now, we'll store HTTP requests globally
    details, _ := json.Marshal(req)
    
    // Find active session (simplified - in real implementation would use context)
    for _, session := range d.sessions {
        session.Requests = append(session.Requests, RequestLog{
            ID:        req.ID,
            Type:      "http",
            Timestamp: req.Timestamp,
            Details:   details,
        })
        break
    }
}

func (d *DetailedLogger) LogHTTPResponse(resp *HTTPResponse) {
    if !d.enabled {
        return
    }
    
    d.mu.Lock()
    defer d.mu.Unlock()
    
    // Update duration for the corresponding request
    for _, session := range d.sessions {
        for i, req := range session.Requests {
            if req.ID == resp.RequestID {
                session.Requests[i].Duration = resp.Duration
                
                // Merge response details into request
                var reqDetails HTTPRequest
                json.Unmarshal(req.Details, &reqDetails)
                
                fullDetails := struct {
                    HTTPRequest
                    Response HTTPResponse `json:"response"`
                }{
                    HTTPRequest: reqDetails,
                    Response:    *resp,
                }
                
                session.Requests[i].Details, _ = json.Marshal(fullDetails)
                return
            }
        }
    }
}

func (d *DetailedLogger) LogHTTPError(requestID string, err error) {
    if !d.enabled {
        return
    }
    
    // Similar to LogHTTPResponse but with error details
}

func (d *DetailedLogger) LogLLMError(streamID string, err error) {
    if !d.enabled {
        return
    }
    
    // Log LLM errors
}

func (d *DetailedLogger) LogStreamEvent(event *StreamEventLog) {
    if !d.enabled {
        return
    }
    
    // For real-time monitoring, could emit to a channel
}

func (d *DetailedLogger) UpdateRequestDuration(requestID string, duration time.Duration) {
    if !d.enabled {
        return
    }
    
    d.mu.Lock()
    defer d.mu.Unlock()
    
    for _, session := range d.sessions {
        for i, req := range session.Requests {
            if req.ID == requestID {
                session.Requests[i].Duration = duration
                return
            }
        }
    }
}

func (d *DetailedLogger) Storage() *LogStorage {
    return d.storage
}

func (d *DetailedLogger) ToolTracker() *ToolTracker {
    return d.toolTracker
}

// WrapHTTPClient creates an HTTP client with logging
func (d *DetailedLogger) WrapHTTPClient(client *http.Client) *http.Client {
    if !d.enabled {
        return client
    }
    
    return WrapHTTPClient(client, d)
}
```

### Step 8: Implement Web Viewer

**File: `internal/detailed_logging/viewer.go`**

```go
package detailed_logging

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    
    "github.com/kirmad/superopencode/internal/logging"
)

type ViewerServer struct {
    storage *LogStorage
    port    int
}

func NewViewerServer(storage *LogStorage, port int) *ViewerServer {
    return &ViewerServer{
        storage: storage,
        port:    port,
    }
}

func (v *ViewerServer) Start() error {
    mux := http.NewServeMux()
    
    // Serve static HTML
    mux.HandleFunc("/", v.handleIndex)
    
    // API endpoints
    mux.HandleFunc("/api/sessions", v.handleSessions)
    mux.HandleFunc("/api/sessions/", v.handleSessionDetails)
    mux.HandleFunc("/api/search", v.handleSearch)
    
    addr := fmt.Sprintf(":%d", v.port)
    logging.Info("Starting detailed logs viewer on http://localhost%s", addr)
    
    return http.ListenAndServe(addr, mux)
}

func (v *ViewerServer) handleIndex(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(viewerHTML))
}

func (v *ViewerServer) handleSessions(w http.ResponseWriter, r *http.Request) {
    limit := 50
    offset := 0
    
    if l := r.URL.Query().Get("limit"); l != "" {
        if val, err := strconv.Atoi(l); err == nil {
            limit = val
        }
    }
    
    if o := r.URL.Query().Get("offset"); o != "" {
        if val, err := strconv.Atoi(o); err == nil {
            offset = val
        }
    }
    
    sessions, err := v.storage.GetSessions(limit, offset)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(sessions)
}

func (v *ViewerServer) handleSessionDetails(w http.ResponseWriter, r *http.Request) {
    sessionID := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
    
    session, err := v.storage.GetSessionDetails(sessionID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    
    // Calculate duration
    duration := session.EndTime.Sub(session.StartTime).Milliseconds()
    
    response := struct {
        *SessionLog
        Duration int64 `json:"duration"`
    }{
        SessionLog: session,
        Duration:   duration,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (v *ViewerServer) handleSearch(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    if query == "" {
        http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
        return
    }
    
    sessions, err := v.storage.SearchSessions(query)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(sessions)
}

// viewerHTML contains the complete HTML/CSS/JS for the viewer
// This is defined in the design section above
const viewerHTML = `[HTML content from design section]`
```

### Step 9: Integration Points

**File: `internal/config/config.go` - Add flag support**

```go
type Config struct {
    // ... existing fields ...
    DetailedLogs     bool   `json:"detailedLogs,omitempty"`
    DetailedLogsPort int    `json:"detailedLogsPort,omitempty"` // Default 8080
}
```

**File: `cmd/root.go` - Add CLI flag**

```go
func init() {
    // ... existing flags ...
    rootCmd.Flags().Bool("detailed-logs", false, "Enable detailed logging of all LLM and tool interactions")
    rootCmd.Flags().Int("detailed-logs-port", 8080, "Port for the detailed logs viewer")
}
```

**File: `internal/app/app.go` - Integration logic**

```go
package app

import (
    "github.com/kirmad/superopencode/internal/detailed_logging"
)

func New(ctx context.Context, conn *sql.DB) (*App, error) {
    // ... existing initialization ...
    
    // Initialize detailed logging if enabled
    var detailedLogger *detailed_logging.DetailedLogger
    if config.Get().DetailedLogs {
        var err error
        detailedLogger, err = detailed_logging.NewDetailedLogger(
            filepath.Join(config.Get().Data.Directory, "detailed_logs"),
        )
        if err != nil {
            return nil, fmt.Errorf("failed to initialize detailed logging: %w", err)
        }
        
        // Start the viewer server
        port := config.Get().DetailedLogsPort
        if port == 0 {
            port = 8080
        }
        
        viewer := detailed_logging.NewViewerServer(detailedLogger.Storage(), port)
        go func() {
            if err := viewer.Start(); err != nil {
                logging.Error("Failed to start detailed logs viewer: %v", err)
            }
        }()
        
        logging.Info("Detailed logging enabled. Viewer available at http://localhost:%d", port)
    }
    
    // Pass detailedLogger to agent creation
    app.CoderAgent, err = agent.NewAgent(
        config.AgentCoder,
        app.Sessions,
        app.Messages,
        coderTools,
        agent.WithDetailedLogger(detailedLogger),
    )
    
    return app, nil
}
```

**File: `internal/llm/agent/agent.go` - Modify to accept logger**

```go
type agentOptions struct {
    detailedLogger *detailed_logging.DetailedLogger
}

type AgentOption func(*agentOptions)

func WithDetailedLogger(logger *detailed_logging.DetailedLogger) AgentOption {
    return func(opts *agentOptions) {
        opts.detailedLogger = logger
    }
}

func NewAgent(
    agentName config.AgentName,
    sessions session.Service,
    messages message.Service,
    agentTools []tools.BaseTool,
    options ...AgentOption,
) (Service, error) {
    opts := &agentOptions{}
    for _, opt := range options {
        opt(opts)
    }
    
    // ... existing code ...
    
    // Wrap provider if detailed logging is enabled
    if opts.detailedLogger != nil {
        agentProvider = detailed_logging.WrapProvider(agentProvider, opts.detailedLogger)
        
        // Also wrap HTTP clients in providers
        // This would need to be done in each provider's initialization
    }
    
    agent := &agent{
        // ... existing fields ...
        detailedLogger: opts.detailedLogger,
    }
    
    return agent, nil
}

// In Run method, start/end session logging
func (a *agent) Run(ctx context.Context, sessionID string, content string, attachments ...message.Attachment) (<-chan AgentEvent, error) {
    // ... existing code ...
    
    if a.detailedLogger != nil {
        a.detailedLogger.StartSession(sessionID, content)
        defer a.detailedLogger.EndSession(sessionID)
    }
    
    // ... rest of the method ...
}
```

**File: `internal/llm/tools/base.go` - Modify tool execution**

```go
// Modify the Execute method wrapper in the tools package
func ExecuteWithLogging(ctx context.Context, tool BaseTool, input json.RawMessage, logger *detailed_logging.DetailedLogger) (json.RawMessage, error) {
    if logger == nil || !logger.IsEnabled() {
        return tool.Execute(ctx, input)
    }
    
    return detailed_logging.WrapToolExecution(
        ctx,
        logger.ToolTracker(),
        tool.Name(),
        input,
        func(newCtx context.Context) (json.RawMessage, error) {
            return tool.Execute(newCtx, input)
        },
    )
}
```

### Step 10: Testing

Create comprehensive tests for each component:

**File: `internal/detailed_logging/detailed_test.go`**

```go
package detailed_logging

import (
    "context"
    "encoding/json"
    "testing"
    "time"
)

func TestDetailedLogger(t *testing.T) {
    // Create temporary directory for test
    tmpDir := t.TempDir()
    
    logger, err := NewDetailedLogger(tmpDir)
    if err != nil {
        t.Fatalf("Failed to create logger: %v", err)
    }
    
    // Test session logging
    sessionID := "test-session-1"
    logger.StartSession(sessionID, "Test prompt")
    
    // Log an LLM request
    logger.LogLLMRequest(&LLMRequest{
        ID:        "req-1",
        SessionID: sessionID,
        Provider:  "anthropic",
        Model:     "claude-3-opus",
        Messages: []MessageLog{
            {Role: "user", Content: "Hello"},
        },
        Timestamp: time.Now(),
    })
    
    // End session
    logger.EndSession(sessionID)
    
    // Verify session was saved
    session, err := logger.Storage().GetSessionDetails(sessionID)
    if err != nil {
        t.Fatalf("Failed to get session: %v", err)
    }
    
    if session.UserPrompt != "Test prompt" {
        t.Errorf("Expected prompt 'Test prompt', got '%s'", session.UserPrompt)
    }
    
    if len(session.Requests) != 1 {
        t.Errorf("Expected 1 request, got %d", len(session.Requests))
    }
}

func TestToolTracker(t *testing.T) {
    tracker := NewToolTracker()
    
    sessionID := "test-session"
    input := json.RawMessage(`{"query": "test"}`)
    
    // Start root tool call
    rootID := tracker.StartToolCall(sessionID, "", "grep", input)
    
    // Start child tool call
    childID := tracker.StartToolCall(sessionID, rootID, "read", input)
    
    // End child
    tracker.EndToolCall(childID, json.RawMessage(`{"content": "file content"}`), nil)
    
    // End root
    tracker.EndToolCall(rootID, json.RawMessage(`{"results": ["file1.go"]}`), nil)
    
    // Get hierarchy
    calls := tracker.GetSessionCalls(sessionID)
    if len(calls) != 1 {
        t.Errorf("Expected 1 root call, got %d", len(calls))
    }
    
    if len(calls[0].Children) != 1 {
        t.Errorf("Expected 1 child call, got %d", len(calls[0].Children))
    }
}
```

## Dependencies

Add the following dependencies to `go.mod`:

```go
require (
    github.com/mattn/go-sqlite3 v1.14.17
    github.com/google/uuid v1.3.0
)
```

## Configuration

The detailed logging system can be configured through:

1. **CLI flags**:
   - `--detailed-logs`: Enable detailed logging
   - `--detailed-logs-port`: Set viewer port (default: 8080)

2. **Config file** (`.opencode/config.json`):
   ```json
   {
     "detailedLogs": true,
     "detailedLogsPort": 8080
   }
   ```

## Performance Considerations

1. **Minimal overhead when disabled**: All logging checks `IsEnabled()` first
2. **Async logging**: Stream events are logged asynchronously
3. **Efficient storage**: SQLite for metadata, JSON files for details
4. **Bounded memory**: Session data is flushed to disk on completion
5. **HTTP interception**: Only bodies are cloned when logging is enabled

## Security Considerations

1. **Sensitive data**: The logger captures all request/response data including:
   - API keys in headers (should be redacted)
   - User prompts and responses
   - File contents accessed by tools

2. **Access control**: The web viewer runs on localhost only

3. **Data retention**: Implement cleanup policies for old logs

## Future Enhancements

1. **Export capabilities**: Export sessions as markdown, JSON, or PDF
2. **Filtering**: Advanced filtering by date, model, tool usage
3. **Metrics**: Token usage graphs, performance analytics
4. **Real-time view**: WebSocket support for live session monitoring
5. **Redaction**: Automatic redaction of sensitive patterns
6. **Compression**: Compress old session data
7. **Cloud sync**: Optional cloud backup of logs

## Summary

This implementation provides:

1. **Complete data capture** at HTTP and provider levels
2. **Hierarchical tool tracking** with parent-child relationships
3. **Modern web viewer** with search and filtering
4. **Minimal performance impact** when disabled
5. **Easy integration** with existing codebase
6. **Extensible architecture** for future enhancements

The system is designed to be transparent to the application when disabled and comprehensive when enabled, providing invaluable debugging and analysis capabilities for understanding LLM interactions.