import type {
  Provider,
  Profile,
  LogsResponse,
  RequestsResponse,
  RequestRecord,
  UsageSummary,
  HourlyUsage,
  HourlyUsageByDimension,
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
  MiddlewareConfig,
  AutoPermissionConfig,
  AutoPermissionAll,
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
  create: (provider: Provider, addToProfiles?: string[]) => {
    const { name, ...config } = provider
    return request<Provider>('/providers', {
      method: 'POST',
      body: JSON.stringify({ name, config, add_to_profiles: addToProfiles }),
    })
  },
  update: (name: string, provider: Partial<Provider>) =>
    request<Provider>(`/providers/${encodeURIComponent(name)}`, {
      method: 'PUT',
      body: JSON.stringify(provider),
    }),
  delete: (name: string) =>
    request<{ success: boolean }>(`/providers/${encodeURIComponent(name)}`, {
      method: 'DELETE',
    }),
  disable: (name: string, type: 'today' | 'month' | 'permanent') =>
    request<{ provider: string; disabled: boolean; type: string; created_at: string; expires_at: string }>(
      `/providers/${encodeURIComponent(name)}/disable`,
      { method: 'POST', body: JSON.stringify({ type }) }
    ),
  enable: (name: string) =>
    request<{ provider: string; disabled: boolean }>(
      `/providers/${encodeURIComponent(name)}/enable`,
      { method: 'POST' }
    ),
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

// Request monitoring API
export const requestsApi = {
  list: (params?: {
    provider?: string
    session?: string
    model?: string
    status_min?: number
    status_max?: number
    limit?: number
  }) => {
    const searchParams = new URLSearchParams()
    if (params?.provider) searchParams.set('provider', params.provider)
    if (params?.session) searchParams.set('session', params.session)
    if (params?.model) searchParams.set('model', params.model)
    if (params?.status_min) searchParams.set('status_min', params.status_min.toString())
    if (params?.status_max) searchParams.set('status_max', params.status_max.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    const query = searchParams.toString()
    return request<RequestsResponse>(`/monitoring/requests${query ? `?${query}` : ''}`)
  },
  get: (id: string) => request<RequestRecord>(`/monitoring/requests/${id}`),
}

// Usage API
export const usageApi = {
  summary: (params?: {
    period?: 'today' | 'week' | 'month'
    since?: string
    until?: string
    project?: string
  }) => {
    const searchParams = new URLSearchParams()
    if (params?.period) searchParams.set('period', params.period)
    if (params?.since) searchParams.set('since', params.since)
    if (params?.until) searchParams.set('until', params.until)
    if (params?.project) searchParams.set('project', params.project)
    const query = searchParams.toString()
    return request<UsageSummary>(`/usage/summary${query ? `?${query}` : ''}`)
  },
  hourly: (params?: {
    hours?: number
    since?: string
    until?: string
    groupBy?: 'provider' | 'model'
  }) => {
    const searchParams = new URLSearchParams()
    if (params?.hours) searchParams.set('hours', params.hours.toString())
    if (params?.since) searchParams.set('since', params.since)
    if (params?.until) searchParams.set('until', params.until)
    if (params?.groupBy) searchParams.set('group_by', params.groupBy)
    const query = searchParams.toString()
    if (params?.groupBy) {
      return request<HourlyUsageByDimension[]>(`/usage/hourly${query ? `?${query}` : ''}`)
    }
    return request<HourlyUsage[]>(`/usage/hourly${query ? `?${query}` : ''}`)
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

// Auto-Permission API
export const autoPermissionApi = {
  getAll: () => request<AutoPermissionAll>('/auto-permission'),
  get: (client: string) =>
    request<AutoPermissionConfig>(`/auto-permission/${encodeURIComponent(client)}`),
  update: (client: string, config: { enabled: boolean; mode: string }) =>
    request<AutoPermissionConfig>(`/auto-permission/${encodeURIComponent(client)}`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),
}

// Bindings API
export const bindingsApi = {
  list: () => request<{ bindings: Binding[]; profiles: string[]; clients: string[] }>('/bindings'),
  create: (path: string, profile?: string, cli?: string) =>
    request<Binding>('/bindings', {
      method: 'POST',
      body: JSON.stringify({ path, profile, cli }),
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
  chat: async (
    message: string,
    sessionId?: string,
    onDelta?: (content: string) => void,
    onSession?: (sessionId: string) => void
  ): Promise<{ content: string; sessionId: string }> => {
    const response = await fetch(`${API_BASE}/bot/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ message, session_id: sessionId }),
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new ApiError(response.status, error.error || `HTTP ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) throw new Error('No response body')

    const decoder = new TextDecoder()
    let fullContent = ''
    let newSessionId = sessionId || ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      const chunk = decoder.decode(value, { stream: true })
      const lines = chunk.split('\n')

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i]
        if (line.startsWith('event: ')) {
          const event = line.slice(7)
          const dataLine = lines[i + 1]
          if (dataLine?.startsWith('data: ')) {
            const data = JSON.parse(dataLine.slice(6))
            if (event === 'session') {
              newSessionId = data.session_id
              onSession?.(newSessionId)
            } else if (event === 'delta') {
              fullContent += data.content
              onDelta?.(data.content)
            } else if (event === 'done') {
              fullContent = data.content
            } else if (event === 'error') {
              throw new ApiError(500, data.error)
            }
          }
        }
      }
    }

    return { content: fullContent, sessionId: newSessionId }
  },
  clearChat: (sessionId: string) =>
    request<{ session_id: string; status: string }>('/bot/chat', {
      method: 'POST',
      body: JSON.stringify({ session_id: sessionId, clear: true }),
    }),
}

// Middleware API
export const middlewareApi = {
  get: () => request<MiddlewareConfig>('/middleware'),
  update: (config: MiddlewareConfig) =>
    request<{ status: string }>('/middleware', {
      method: 'PUT',
      body: JSON.stringify(config),
    }),
  enable: (name: string) =>
    request<{ status: string }>(`/middleware/${encodeURIComponent(name)}/enable`, {
      method: 'POST',
    }),
  disable: (name: string) =>
    request<{ status: string }>(`/middleware/${encodeURIComponent(name)}/disable`, {
      method: 'POST',
    }),
  reload: () =>
    request<{ status: string }>('/middleware/reload', {
      method: 'POST',
    }),
  upload: async (file: File, name?: string): Promise<{ status: string; name: string; path: string; checksum: string }> => {
    const formData = new FormData()
    formData.append('plugin', file)
    if (name) {
      formData.append('name', name)
    }
    const response = await fetch(`${API_BASE}/middleware/upload`, {
      method: 'POST',
      body: formData,
      credentials: 'include',
    })
    if (!response.ok) {
      const text = await response.text()
      throw new ApiError(response.status, text || response.statusText)
    }
    return response.json()
  },
}

export { ApiError }
