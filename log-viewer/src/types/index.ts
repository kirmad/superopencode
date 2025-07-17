export interface SessionMetadata {
  id: string;
  session_id: string;
  start_time: string;
  end_time?: string;
  llm_call_count: number;
  tool_call_count: number;
  http_call_count: number;
  total_tokens: number;
  total_cost: number;
  has_error: boolean;
  metadata?: string;
  created_at: string;
}

export interface TokenUsage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

export interface StreamEvent {
  event_type: string;
  data: any;
  timestamp: string;
}

export interface LLMCallLog {
  id: string;
  session_id: string;
  provider: string;
  model: string;
  start_time: string;
  end_time?: string;
  request: Record<string, any>;
  response?: Record<string, any>;
  stream_events?: StreamEvent[];
  error?: string;
  tokens_used?: TokenUsage;
  cost?: number;
  duration_ms: number;
  parent_tool_call?: string;
}

export interface ToolCallLog {
  id: string;
  session_id: string;
  name: string;
  start_time: string;
  end_time?: string;
  input: Record<string, any>;
  output?: any;
  error?: string;
  duration_ms: number;
  parent_id?: string;
  child_ids?: string[];
  parent_llm_call?: string;
}

export interface HTTPLog {
  id: string;
  session_id: string;
  method: string;
  url: string;
  headers: Record<string, string[]>;
  body?: any;
  status_code?: number;
  response_body?: any;
  response_headers?: Record<string, string[]>;
  start_time: string;
  end_time?: string;
  duration_ms: number;
  error?: string;
  parent_tool_call?: string;
}

export interface SessionLog {
  id: string;
  start_time: string;
  end_time?: string;
  metadata: Record<string, string>;
  llm_calls: LLMCallLog[];
  tool_calls: ToolCallLog[];
  http_calls: HTTPLog[];
  command_args: string[];
  user_id?: string;
}

export interface SessionFilters {
  start_time?: string;
  end_time?: string;
  has_error?: boolean;
  search?: string;
  limit?: number;
  offset?: number;
}