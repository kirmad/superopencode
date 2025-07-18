CREATE TABLE sessions (
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

INSERT INTO sessions VALUES 
('sess_1', 'session-2024-01-15-001', '2024-01-15 10:30:00', '2024-01-15 10:35:00', 3, 5, 8, 1200, 0.05, 0, '{}', '2024-01-15 10:30:00'),
('sess_2', 'session-2024-01-15-002', '2024-01-15 11:00:00', '2024-01-15 11:08:00', 2, 3, 5, 800, 0.03, 0, '{}', '2024-01-15 11:00:00'),
('sess_3', 'session-2024-01-15-003', '2024-01-15 12:15:00', NULL, 1, 2, 3, 400, 0.02, 1, '{}', '2024-01-15 12:15:00');