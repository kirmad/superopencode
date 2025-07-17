package detailed_logging

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
)

// Storage handles persisting session logs
type Storage struct {
	db       *sql.DB
	dataDir  string
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) (*Storage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open SQLite database
	dbPath := filepath.Join(dataDir, "sessions.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &Storage{
		db:      db,
		dataDir: dataDir,
	}

	// Initialize schema
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return s, nil
}

// initSchema creates the necessary tables
func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		llm_call_count INTEGER DEFAULT 0,
		tool_call_count INTEGER DEFAULT 0,
		http_call_count INTEGER DEFAULT 0,
		total_tokens INTEGER DEFAULT 0,
		total_cost REAL DEFAULT 0,
		has_error BOOLEAN DEFAULT 0,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time);
	CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_has_error ON sessions(has_error);
	`

	_, err := s.db.Exec(schema)
	return err
}

// SaveSession persists a complete session log
func (s *Storage) SaveSession(session *SessionLog) error {
	// Calculate metadata
	metadata := s.calculateMetadata(session)

	// Save to SQLite - use REPLACE to handle updates
	query := `
	INSERT OR REPLACE INTO sessions (
		id, session_id, start_time, end_time,
		llm_call_count, tool_call_count, http_call_count,
		total_tokens, total_cost, has_error, metadata
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	metadataJSON, _ := json.Marshal(session.Metadata)
	
	_, err := s.db.Exec(query,
		session.ID,
		session.ID,
		session.StartTime,
		session.EndTime,
		metadata.LLMCallCount,
		metadata.ToolCallCount,
		metadata.HTTPCallCount,
		metadata.TotalTokens,
		metadata.TotalCost,
		metadata.HasError,
		string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to save session metadata: %w", err)
	}

	// Save detailed JSON to file
	jsonPath := filepath.Join(s.dataDir, fmt.Sprintf("%s.json", session.ID))
	jsonData, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession retrieves a session by ID
func (s *Storage) LoadSession(sessionID string) (*SessionLog, error) {
	// Load from JSON file
	jsonPath := filepath.Join(s.dataDir, fmt.Sprintf("%s.json", sessionID))
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		// Don't wrap os.IsNotExist errors so they can be properly detected
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session SessionLog
	if err := json.Unmarshal(jsonData, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &session, nil
}

// ListSessions returns session metadata based on filters
func (s *Storage) ListSessions(filters SessionFilters) ([]StorageMetadata, error) {
	query := `
	SELECT 
		id, session_id, start_time, end_time,
		llm_call_count, tool_call_count, http_call_count,
		total_tokens, total_cost, has_error
	FROM sessions
	WHERE 1=1
	`
	args := []interface{}{}

	if filters.StartTime != nil {
		query += " AND start_time >= ?"
		args = append(args, *filters.StartTime)
	}

	if filters.EndTime != nil {
		query += " AND start_time <= ?"
		args = append(args, *filters.EndTime)
	}

	if filters.HasError != nil {
		query += " AND has_error = ?"
		args = append(args, *filters.HasError)
	}

	query += " ORDER BY start_time DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var results []StorageMetadata
	for rows.Next() {
		var meta StorageMetadata
		err := rows.Scan(
			&meta.ID,
			&meta.SessionID,
			&meta.StartTime,
			&meta.EndTime,
			&meta.LLMCallCount,
			&meta.ToolCallCount,
			&meta.HTTPCallCount,
			&meta.TotalTokens,
			&meta.TotalCost,
			&meta.HasError,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, meta)
	}

	return results, nil
}

// calculateMetadata computes summary statistics for a session
func (s *Storage) calculateMetadata(session *SessionLog) StorageMetadata {
	meta := StorageMetadata{
		ID:            session.ID,
		SessionID:     session.ID,
		StartTime:     session.StartTime,
		EndTime:       session.EndTime,
		LLMCallCount:  len(session.LLMCalls),
		ToolCallCount: len(session.ToolCalls),
		HTTPCallCount: len(session.HTTPCalls),
	}

	// Calculate totals
	for _, llm := range session.LLMCalls {
		if llm.TokensUsed != nil {
			meta.TotalTokens += llm.TokensUsed.Total
		}
		if llm.Cost != nil {
			meta.TotalCost += *llm.Cost
		}
		if llm.Error != "" {
			meta.HasError = true
		}
	}

	for _, tool := range session.ToolCalls {
		if tool.Error != "" {
			meta.HasError = true
		}
	}

	for _, http := range session.HTTPCalls {
		if http.Error != "" {
			meta.HasError = true
		}
	}

	return meta
}

// DeleteOldSessions removes sessions older than the retention period
func (s *Storage) DeleteOldSessions(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	// Get sessions to delete
	rows, err := s.db.Query(
		"SELECT id FROM sessions WHERE start_time < ?",
		cutoffTime,
	)
	if err != nil {
		return fmt.Errorf("failed to query old sessions: %w", err)
	}
	defer rows.Close()

	var sessionsToDelete []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		sessionsToDelete = append(sessionsToDelete, id)
	}

	// Delete from database
	_, err = s.db.Exec(
		"DELETE FROM sessions WHERE start_time < ?",
		cutoffTime,
	)
	if err != nil {
		return fmt.Errorf("failed to delete old sessions: %w", err)
	}

	// Delete JSON files
	for _, id := range sessionsToDelete {
		jsonPath := filepath.Join(s.dataDir, fmt.Sprintf("%s.json", id))
		os.Remove(jsonPath) // Ignore errors
	}

	return nil
}

// Close closes the storage
func (s *Storage) Close() error {
	return s.db.Close()
}

// SessionFilters defines filters for listing sessions
type SessionFilters struct {
	StartTime *time.Time
	EndTime   *time.Time
	HasError  *bool
	Limit     int
}