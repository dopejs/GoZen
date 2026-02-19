// Provider types
export interface Provider {
  name: string
  base_url: string
  api_key: string
  api_key_env?: string
  model?: string
  priority?: number
  weight?: number
  enabled?: boolean
}

// Profile types
export interface Profile {
  name: string
  providers: string[]
  fallback?: string[]
  routing?: RoutingConfig
  is_default?: boolean
}

export interface RoutingConfig {
  strategy?: 'priority' | 'round-robin' | 'weighted' | 'least-latency'
  health_check?: boolean
  retry_count?: number
}

// Log types
export interface LogEntry {
  id: string
  timestamp: string
  provider: string
  model: string
  status: number
  latency_ms: number
  input_tokens: number
  output_tokens: number
  error?: string
  session_id?: string
  client_type?: string
}

export interface LogsResponse {
  entries: LogEntry[]
  total: number
  providers: string[]
}

// Usage types
export interface UsageSummary {
  total_requests: number
  total_input_tokens: number
  total_output_tokens: number
  total_cost: number
  by_provider: Record<string, ProviderUsage>
  by_model: Record<string, ModelUsage>
}

export interface ProviderUsage {
  requests: number
  input_tokens: number
  output_tokens: number
  cost: number
}

export interface ModelUsage {
  requests: number
  input_tokens: number
  output_tokens: number
  cost: number
}

export interface HourlyUsage {
  hour: string
  requests: number
  input_tokens: number
  output_tokens: number
  cost: number
}

// Budget types
export interface Budget {
  enabled: boolean
  monthly_limit: number
  daily_limit: number
  alert_threshold: number
}

export interface BudgetStatus {
  monthly_used: number
  monthly_limit: number
  monthly_remaining: number
  daily_used: number
  daily_limit: number
  daily_remaining: number
  alert_triggered: boolean
}

// Health types
export interface ProviderHealth {
  name: string
  status: 'healthy' | 'degraded' | 'unhealthy'
  latency_ms: number
  success_rate: number
  last_check: string
  error?: string
}

// Settings types
export interface Settings {
  web_port: number
  proxy_port: number
  log_level: string
  max_retries: number
  timeout_ms: number
}

// Binding types
export interface Binding {
  path: string
  profile: string
  created_at: string
}

// Sync types
export interface SyncConfig {
  enabled: boolean
  gist_id?: string
  github_token?: string
  auto_sync: boolean
  sync_interval_minutes: number
}

export interface SyncStatus {
  enabled: boolean
  last_sync?: string
  last_error?: string
  gist_url?: string
}

// Webhook types
export interface Webhook {
  id: string
  name: string
  url: string
  events: string[]
  enabled: boolean
  secret?: string
}

// Session types
export interface Session {
  id: string
  provider: string
  model: string
  created_at: string
  last_activity: string
  request_count: number
  total_tokens: number
}

// Auth types
export interface AuthCheckResponse {
  authenticated: boolean
  password_set: boolean
}

export interface LoginRequest {
  password: string
  encrypted?: string
}

export interface LoginResponse {
  success: boolean
  error?: string
}

// API response wrapper
export interface ApiResponse<T> {
  data?: T
  error?: string
}

// Health check response
export interface HealthResponse {
  status: string
  version: string
}
