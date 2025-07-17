import { SessionMetadata, SessionLog, SessionFilters } from '@/types';

const API_BASE = 'http://localhost:3001/api';

export class ApiService {
  async getSessions(filters: SessionFilters = {}): Promise<SessionMetadata[]> {
    const params = new URLSearchParams();
    
    if (filters.start_time) params.append('start_time', filters.start_time);
    if (filters.end_time) params.append('end_time', filters.end_time);
    if (filters.has_error !== undefined) params.append('has_error', String(filters.has_error));
    if (filters.search) params.append('search', filters.search);
    if (filters.limit) params.append('limit', String(filters.limit));
    if (filters.offset) params.append('offset', String(filters.offset));

    const response = await fetch(`${API_BASE}/sessions?${params}`);
    if (!response.ok) throw new Error('Failed to fetch sessions');
    return response.json();
  }

  async getSession(sessionId: string): Promise<SessionLog> {
    const response = await fetch(`${API_BASE}/sessions/${sessionId}`);
    if (!response.ok) throw new Error('Failed to fetch session');
    return response.json();
  }

  async getSessionFile(sessionId: string): Promise<SessionLog> {
    const response = await fetch(`${API_BASE}/sessions/${sessionId}/file`);
    if (!response.ok) throw new Error('Failed to fetch session file');
    return response.json();
  }
}

export const apiService = new ApiService();