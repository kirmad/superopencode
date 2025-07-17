import initSqlJs, { Database } from 'sql.js';
import { SessionMetadata, SessionLog, SessionFilters } from '@/types';

class DatabaseService {
  private db: Database | null = null;
  private initialized = false;

  async initialize(dbPath?: string) {
    if (this.initialized) return;

    const SQL = await initSqlJs({
      locateFile: (file) => `https://sql.js.org/dist/${file}`
    });

    if (dbPath) {
      const response = await fetch(dbPath);
      const arrayBuffer = await response.arrayBuffer();
      this.db = new SQL.Database(new Uint8Array(arrayBuffer));
    } else {
      this.db = new SQL.Database();
    }

    this.initialized = true;
  }

  async listSessions(filters: SessionFilters = {}): Promise<SessionMetadata[]> {
    if (!this.db) throw new Error('Database not initialized');

    let query = `
      SELECT id, session_id, start_time, end_time, llm_call_count, 
             tool_call_count, http_call_count, total_tokens, total_cost,
             has_error, metadata, created_at
      FROM sessions
      WHERE 1=1
    `;

    const params: any[] = [];

    if (filters.start_time) {
      query += ' AND start_time >= ?';
      params.push(filters.start_time);
    }

    if (filters.end_time) {
      query += ' AND start_time <= ?';
      params.push(filters.end_time);
    }

    if (filters.has_error !== undefined) {
      query += ' AND has_error = ?';
      params.push(filters.has_error ? 1 : 0);
    }

    if (filters.search) {
      query += ' AND (session_id LIKE ? OR metadata LIKE ?)';
      params.push(`%${filters.search}%`, `%${filters.search}%`);
    }

    query += ' ORDER BY start_time DESC';

    if (filters.limit) {
      query += ' LIMIT ?';
      params.push(filters.limit);
    }

    if (filters.offset) {
      query += ' OFFSET ?';
      params.push(filters.offset);
    }

    const stmt = this.db.prepare(query);
    const results = stmt.getAsObject(params);
    stmt.free();

    return Array.isArray(results) ? results as SessionMetadata[] : [results as SessionMetadata];
  }

  async loadSessionDetails(sessionId: string): Promise<SessionLog | null> {
    try {
      const response = await fetch(`/api/sessions/${sessionId}.json`);
      if (!response.ok) return null;
      return await response.json();
    } catch (error) {
      console.error('Failed to load session details:', error);
      return null;
    }
  }
}

export const databaseService = new DatabaseService();