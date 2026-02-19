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
  default_profile?: string
  default_client?: string
  web_port: number
  proxy_port?: number
  profiles?: string[]
  clients?: string[]
}

// Binding types
export interface Binding {
  path: string
  profile?: string
  cli?: string
}

// Sync types
export interface SyncConfig {
  configured: boolean
  enabled?: boolean
  backend?: string
  gist_id?: string
  gist_token?: string
  auto_pull?: boolean
  pull_interval?: number
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

// Bot types
export interface BotConfig {
  enabled: boolean
  profile?: string
  socket_path?: string
  platforms?: BotPlatformsConfig
  interaction?: BotInteractionConfig
  aliases?: Record<string, string>
  notify?: BotNotifyConfig
}

export interface BotPlatformsConfig {
  telegram?: BotTelegramConfig
  discord?: BotDiscordConfig
  slack?: BotSlackConfig
  lark?: BotLarkConfig
  fbmessenger?: BotFBMessengerConfig
}

export interface BotTelegramConfig {
  enabled: boolean
  token: string
  allowed_users?: string[]
  allowed_chats?: string[]
}

export interface BotDiscordConfig {
  enabled: boolean
  token: string
  allowed_users?: string[]
  allowed_channels?: string[]
  allowed_guilds?: string[]
}

export interface BotSlackConfig {
  enabled: boolean
  bot_token: string
  app_token: string
  allowed_users?: string[]
  allowed_channels?: string[]
}

export interface BotLarkConfig {
  enabled: boolean
  app_id: string
  app_secret: string
  allowed_users?: string[]
  allowed_chats?: string[]
}

export interface BotFBMessengerConfig {
  enabled: boolean
  page_token: string
  verify_token: string
  app_secret?: string
  allowed_users?: string[]
}

export interface BotInteractionConfig {
  require_mention?: boolean
  mention_keywords?: string[]
  direct_message_mode?: string
  channel_mode?: string
}

export interface BotNotifyConfig {
  default_platform?: string
  default_chat_id?: string
  quiet_hours_start?: string
  quiet_hours_end?: string
  quiet_hours_zone?: string
}
