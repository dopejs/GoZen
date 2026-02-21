import { http, HttpResponse } from 'msw'

export const handlers = [
  // Providers
  http.get('/api/v1/providers', () => {
    return HttpResponse.json([
      { name: 'anthropic', base_url: 'https://api.anthropic.com', auth_token: '****', model: 'claude-sonnet-4-5' },
      { name: 'openai', base_url: 'https://api.openai.com', auth_token: '****', model: 'gpt-4' },
    ])
  }),

  http.get('/api/v1/providers/:name', ({ params }) => {
    return HttpResponse.json({
      name: params.name,
      base_url: 'https://api.anthropic.com',
      auth_token: '****',
      model: 'claude-sonnet-4-5',
    })
  }),

  http.post('/api/v1/providers', async ({ request }) => {
    const body = await request.json() as { name: string }
    return HttpResponse.json({ name: body.name, base_url: '', auth_token: '****' }, { status: 201 })
  }),

  http.put('/api/v1/providers/:name', ({ params }) => {
    return HttpResponse.json({ name: params.name, base_url: 'https://updated.com', auth_token: '****' })
  }),

  http.delete('/api/v1/providers/:name', () => {
    return HttpResponse.json({ success: true })
  }),

  // Profiles
  http.get('/api/v1/profiles', () => {
    return HttpResponse.json([
      { name: 'default', providers: ['anthropic', 'openai'] },
      { name: 'work', providers: ['anthropic'] },
    ])
  }),

  http.get('/api/v1/profiles/:name', ({ params }) => {
    return HttpResponse.json({ name: params.name, providers: ['anthropic'] })
  }),

  http.post('/api/v1/profiles', async ({ request }) => {
    const body = await request.json() as { name: string; providers: string[] }
    return HttpResponse.json({ name: body.name, providers: body.providers }, { status: 201 })
  }),

  http.put('/api/v1/profiles/:name', ({ params }) => {
    return HttpResponse.json({ name: params.name, providers: ['anthropic'] })
  }),

  http.delete('/api/v1/profiles/:name', () => {
    return HttpResponse.json({ success: true })
  }),

  // Settings
  http.get('/api/v1/settings', () => {
    return HttpResponse.json({
      default_profile: 'default',
      default_client: 'claude',
      web_port: 18080,
      profiles: ['default', 'work'],
      clients: ['claude', 'cline', 'cursor'],
    })
  }),

  http.put('/api/v1/settings', () => {
    return HttpResponse.json({
      default_profile: 'default',
      default_client: 'claude',
      web_port: 18080,
    })
  }),

  // Bindings
  http.get('/api/v1/bindings', () => {
    return HttpResponse.json({
      bindings: [{ path: '/test/project', profile: 'default', cli: 'claude' }],
      profiles: ['default', 'work'],
      clients: ['claude', 'cline'],
    })
  }),

  http.post('/api/v1/bindings', () => {
    return HttpResponse.json({ path: '/new/project', profile: 'default' }, { status: 201 })
  }),

  http.delete('/api/v1/bindings/:path', () => {
    return HttpResponse.json({ success: true })
  }),

  // Bot
  http.get('/api/v1/bot', () => {
    return HttpResponse.json({
      enabled: true,
      profile: 'default',
      socket_path: '/tmp/zen.sock',
      platforms: {
        telegram: { enabled: true, token: '****', allowed_users: ['user1'] },
      },
      interaction: { require_mention: true, mention_keywords: ['@zen'] },
      aliases: { api: '/path/to/api' },
      notify: { default_platform: 'telegram' },
    })
  }),

  http.put('/api/v1/bot', () => {
    return HttpResponse.json({ enabled: true, profile: 'default' })
  }),

  // Usage
  http.get('/api/v1/usage/summary', () => {
    return HttpResponse.json({
      total_input_tokens: 10000,
      total_output_tokens: 5000,
      total_cost: 1.5,
      request_count: 100,
      by_provider: { anthropic: { input_tokens: 10000, output_tokens: 5000, cost: 1.5, request_count: 100 } },
      by_model: {},
      by_project: {},
    })
  }),

  http.get('/api/v1/usage/hourly', () => {
    return HttpResponse.json([])
  }),

  // Logs
  http.get('/api/v1/logs', () => {
    return HttpResponse.json({ entries: [], total: 0, providers: [] })
  }),

  // Health
  http.get('/api/v1/health', () => {
    return HttpResponse.json({ status: 'ok', version: '3.0.0' })
  }),

  // Sync
  http.get('/api/v1/sync/config', () => {
    return HttpResponse.json({ configured: false, enabled: false })
  }),

  http.get('/api/v1/sync/status', () => {
    return HttpResponse.json({ enabled: false })
  }),

  // Auth
  http.get('/api/v1/auth/check', () => {
    return HttpResponse.json({ authenticated: true, password_set: false })
  }),

  http.post('/api/v1/settings/password', () => {
    return HttpResponse.json({ status: 'password updated' })
  }),

  // Budget
  http.get('/api/v1/budget', () => {
    return HttpResponse.json({
      daily_limit: 10,
      weekly_limit: 50,
      monthly_limit: 200,
      enabled: true,
    })
  }),

  http.put('/api/v1/budget', () => {
    return HttpResponse.json({ daily_limit: 10, enabled: true })
  }),

  http.get('/api/v1/budget/status', () => {
    return HttpResponse.json({
      daily_used: 1.5,
      weekly_used: 10,
      monthly_used: 50,
      daily_remaining: 8.5,
    })
  }),

  // Provider Health
  http.get('/api/v1/health/providers', () => {
    return HttpResponse.json([
      { name: 'anthropic', status: 'healthy', latency_ms: 100, last_check: new Date().toISOString() },
    ])
  }),

  http.get('/api/v1/health/providers/:name', ({ params }) => {
    return HttpResponse.json({
      name: params.name,
      status: 'healthy',
      latency_ms: 100,
      last_check: new Date().toISOString(),
    })
  }),

  // Sync operations
  http.put('/api/v1/sync/config', () => {
    return HttpResponse.json({ backend: 'gist', configured: true })
  }),

  http.post('/api/v1/sync/pull', () => {
    return HttpResponse.json({ success: true })
  }),

  http.post('/api/v1/sync/push', () => {
    return HttpResponse.json({ success: true })
  }),
]
