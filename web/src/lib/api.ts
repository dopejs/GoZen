import type {
  Provider,
  Profile,
  LogsResponse,
  UsageSummary,
  HourlyUsage,
  Budget,
  BudgetStatus,
  ProviderHealth,
  Settings,
  Binding,
  SyncConfig,
  SyncStatus,
  Webhook,
  Session,
  AuthCheckResponse,
  LoginResponse,
  HealthResponse,
  BotConfig,
} from '@/types/api'

const API_BASE = '/api/v1'

class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE}${endpoint}`
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include',
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }))
    throw new ApiError(response.status, error.error || `HTTP ${response.status}`)
  }

  return response.json()
}

// Auth API
export const authApi = {
  check: () => request<AuthCheckResponse>('/auth/check'),
  login: (password: string, encrypted?: string) =>
    request<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ password, encrypted }),
    }),
  logout: () => request<{ success: boolean }>('/auth/logout', { method: 'POST' }),
  getPubKey: () => request<{ public_key: string }>('/auth/pubkey'),
}

// Health API
export const healthApi = {
  check: () => request<HealthResponse>('/health'),
  reload: () => request<{ status: string }>('/reload', { method: 'POST' }),
}

// Providers API
export const providersApi = {
  list: () => request<Provider[]>('/providers'),
  get: (name: string) => request<Provider>(`/providers/${encodeURIComponent(name)}`),
  create: (provider: Provider) =>
    request<Provider>('/providers', {
      method: 'POST',
      body: JSON.stringify(provider),
    }),
  update: (name: string, provider: Partial<Provider>) =>
    request<Provider>(`/providers/${encodeURIComponent(name)}`, {
      method: 'PUT',
      body: JSON.stringify(provider),
    }),
  delete: (name: string) =>
    request<{ success: boolean }>(`/providers/${encodeURIComponent(name)}`, {
      method: 'DELETE',
    }),
}

// Profiles API
export const profilesApi = {
  list: () => request<Profile[]>('/profiles'),
  get: (name: string) => request<Profile>(`/profiles/${encodeURIComponent(name)}`),
  create: (profile: Profile) =>
    request<Profile>('/profiles', {
      method: 'POST',
      body: JSON.stringify(profile),
    }),
  update: (name: string, profile: Partial<Profile>) =>
    request<Profile>(`/profiles/${encodeURIComponent(name)}`, {
      method: 'PUT',
      body: JSON.stringify(profile),
    }),
  delete: (name: string) =>
    request<{ success: boolean }>(`/profiles/${encodeURIComponent(name)}`, {
      method: 'DELETE',
    }),
}

// Logs API
export const logsApi = {
  list: (params?: {
    provider?: string
    session_id?: string
    client_type?: string
    errors_only?: boolean
    limit?: number
  }) => {
    const searchParams = new URLSearchParams()
    if (params?.provider) searchParams.set('provider', params.provider)
    if (params?.session_id) searchParams.set('session_id', params.session_id)
    if (params?.client_type) searchParams.set('client_type', params.client_type)
    if (params?.errors_only) searchParams.set('errors_only', 'true')
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    const query = searchParams.toString()
    return request<LogsResponse>(`/logs${query ? `?${query}` : ''}`)
  },
}

// Usage API
export const usageApi = {
  summary: (period?: 'today' | 'week' | 'month') => {
    const query = period ? `?period=${period}` : ''
    return request<UsageSummary>(`/usage/summary${query}`)
  },
  hourly: (date?: string) => {
    const query = date ? `?date=${date}` : ''
    return request<HourlyUsage[]>(`/usage/hourly${query}`)
  },
}

// Budget API
export const budgetApi = {
  get: () => request<Budget>('/budget'),
  update: (budget: Partial<Budget>) =>
    request<Budget>('/budget', {
      method: 'PUT',
      body: JSON.stringify(budget),
    }),
  status: () => request<BudgetStatus>('/budget/status'),
}

// Provider Health API
export const providerHealthApi = {
  list: () => request<ProviderHealth[]>('/health/providers'),
  get: (name: string) =>
    request<ProviderHealth>(`/health/providers/${encodeURIComponent(name)}`),
}

// Settings API
export const settingsApi = {
  get: () => request<Settings>('/settings'),
  update: (settings: Partial<Settings>) =>
    request<Settings>('/settings', {
      method: 'PUT',
      body: JSON.stringify(settings),
    }),
  changePassword: (currentPassword: string, newPassword: string) =>
    request<{ success: boolean }>('/settings/password', {
      method: 'POST',
      body: JSON.stringify({
        current_password: currentPassword,
        new_password: newPassword,
      }),
    }),
}

// Bindings API
export const bindingsApi = {
  list: () => request<Binding[]>('/bindings'),
  create: (binding: { path: string; profile: string }) =>
    request<Binding>('/bindings', {
      method: 'POST',
      body: JSON.stringify(binding),
    }),
  delete: (path: string) =>
    request<{ success: boolean }>(`/bindings/${encodeURIComponent(path)}`, {
      method: 'DELETE',
    }),
}

// Sync API
export const syncApi = {
  getConfig: () => request<SyncConfig>('/sync/config'),
  updateConfig: (config: Partial<SyncConfig>) =>
    request<SyncConfig>('/sync/config', {
      method: 'PUT',
      body: JSON.stringify(config),
    }),
  status: () => request<SyncStatus>('/sync/status'),
  pull: () => request<{ success: boolean }>('/sync/pull', { method: 'POST' }),
  push: () => request<{ success: boolean }>('/sync/push', { method: 'POST' }),
  test: () => request<{ success: boolean; error?: string }>('/sync/test', { method: 'POST' }),
  createGist: () => request<{ gist_id: string; gist_url: string }>('/sync/create-gist', { method: 'POST' }),
}

// Webhooks API
export const webhooksApi = {
  list: () => request<Webhook[]>('/webhooks'),
  create: (webhook: Omit<Webhook, 'id'>) =>
    request<Webhook>('/webhooks', {
      method: 'POST',
      body: JSON.stringify(webhook),
    }),
  update: (id: string, webhook: Partial<Webhook>) =>
    request<Webhook>(`/webhooks/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(webhook),
    }),
  delete: (id: string) =>
    request<{ success: boolean }>(`/webhooks/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),
  test: (id: string) =>
    request<{ success: boolean; error?: string }>('/webhooks/test', {
      method: 'POST',
      body: JSON.stringify({ id }),
    }),
}

// Sessions API
export const sessionsApi = {
  list: () => request<Session[]>('/sessions'),
  get: (id: string) => request<Session>(`/sessions/${encodeURIComponent(id)}`),
  delete: (id: string) =>
    request<{ success: boolean }>(`/sessions/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),
}

// Bot API
export const botApi = {
  get: () => request<BotConfig>('/bot'),
  update: (bot: Partial<BotConfig>) =>
    request<BotConfig>('/bot', {
      method: 'PUT',
      body: JSON.stringify(bot),
    }),
}

export { ApiError }
