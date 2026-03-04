// Provider types
export interface Provider {
  name: string
  type?: 'anthropic' | 'openai'
  base_url: string
  auth_token: string
  proxy_url?: string
  model?: string
  reasoning_model?: string
  haiku_model?: string
  opus_model?: string
  sonnet_model?: string
  env_vars?: Record<string, string>
  claude_env_vars?: Record<string, string>
  codex_env_vars?: Record<string, string>
  opencode_env_vars?: Record<string, string>
}

// Available clients
export const AVAILABLE_CLIENTS = ['claude', 'codex', 'opencode'] as const
export type ClientType = (typeof AVAILABLE_CLIENTS)[number]

// Common environment variables per client (excluding API keys and models which are set by provider)
export const CLIENT_ENV_HINTS: Record<ClientType, string[]> = {
  claude: [
    'CLAUDE_CODE_MAX_TOKENS',
    'CLAUDE_CODE_USE_BEDROCK',
    'CLAUDE_CODE_USE_VERTEX',
    'CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC',
    'BASH_MAX_TIMEOUT_MS',
    'BASH_DEFAULT_TIMEOUT_MS',
  ],
  codex: [
    'CODEX_SANDBOX_TYPE',
    'CODEX_UNSAFE_ALLOW_NO_SANDBOX',
  ],
  opencode: [
    'OPENCODE_PROVIDER',
    'OPENCODE_AUTO_COMPACT',
  ],
}

// Scenarios
export type Scenario = 'think' | 'image' | 'longContext' | 'webSearch' | 'code' | 'background' | 'default'

export const SCENARIOS: Scenario[] = ['default', 'think', 'image', 'longContext', 'code', 'webSearch', 'background']

export const SCENARIO_LABELS: Record<Scenario, string> = {
  default: 'Default',
  think: 'Extended Thinking',
  image: 'Image Processing',
  longContext: 'Long Context',
  code: 'Code',
  webSearch: 'Web Search',
  background: 'Background Tasks',
}

// Provider route for scenario
export interface ProviderRoute {
  name: string
  model?: string
}

// Scenario route
export interface ScenarioRoute {
  providers: ProviderRoute[]
}

// Load balance strategy
export type LoadBalanceStrategy = 'failover' | 'round-robin' | 'least-latency' | 'least-cost'

export const LOAD_BALANCE_STRATEGIES: LoadBalanceStrategy[] = [
  'failover',
  'round-robin',
  'least-latency',
  'least-cost',
]

// Profile types
export interface Profile {
  name: string
  providers: string[]
  routing?: Partial<Record<Scenario, ScenarioRoute>>
  long_context_threshold?: number
  strategy?: LoadBalanceStrategy
  is_default?: boolean
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

// Request monitoring types
export interface ProviderAttempt {
  provider: string
  status_code: number
  error_message?: string
  duration_ms: number
  skipped?: boolean
  skip_reason?: string
}

export interface RequestRecord {
  id: string
  timestamp: string
  session_id: string
  client_type: string
  provider: string
  model: string
  request_format: string
  status_code: number
  duration_ms: number
  input_tokens: number
  output_tokens: number
  cost_usd: number
  request_size: number
  failover_chain?: ProviderAttempt[]
  error_message?: string
}

export interface RequestsResponse {
  requests: RequestRecord[]
  total: number
  limit: number
}

// Usage types
export interface UsageSummary {
  total_requests: number
  total_input_tokens: number
  total_output_tokens: number
  total_cost: number
  request_count: number
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

export interface HourlyUsageByDimension {
  hour: string
  dimension: string
  input_tokens: number
  output_tokens: number
  cost: number
  request_count: number
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
  // Gist
  gist_id?: string
  // Repo
  repo_owner?: string
  repo_name?: string
  repo_path?: string
  repo_branch?: string
  // GitHub (Gist & Repo)
  token?: string
  // S3
  endpoint?: string
  bucket?: string
  region?: string
  access_key?: string
  secret_key?: string
  // WebDAV
  username?: string
  // Encryption
  passphrase?: string
  // Common
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

// Middleware types
export interface MiddlewareEntry {
  name: string
  enabled: boolean
  source: string
  path?: string
  url?: string
  version?: string
  description?: string
  priority?: number
  config?: Record<string, unknown>
}

export interface MiddlewareConfig {
  enabled: boolean
  middlewares: MiddlewareEntry[]
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
  recent_paths?: string[]
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
