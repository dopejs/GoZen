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

  http.post('/api/v1/bot/chat', async ({ request }) => {
    const body = await request.json() as { clear?: boolean; session_id?: string }
    if (body.clear) {
      return HttpResponse.json({ session_id: body.session_id, status: 'cleared' })
    }
    // SSE streaming response
    const encoder = new TextEncoder()
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(encoder.encode('event: session\ndata: {"session_id":"sess-123"}\n\n'))
        controller.enqueue(encoder.encode('event: delta\ndata: {"content":"Hello"}\n\n'))
        controller.enqueue(encoder.encode('event: done\ndata: {"content":"Hello world"}\n\n'))
        controller.close()
      },
    })
    return new HttpResponse(stream, {
      headers: { 'Content-Type': 'text/event-stream' },
    })
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
    return HttpResponse.json({ success: true })
  }),

  // Auth operations
  http.post('/api/v1/auth/login', () => {
    return HttpResponse.json({ status: 'ok' })
  }),

  http.post('/api/v1/auth/logout', () => {
    return HttpResponse.json({ success: true })
  }),

  http.get('/api/v1/auth/pubkey', () => {
    return HttpResponse.json({ public_key: 'test-pubkey' })
  }),

  // Sessions
  http.get('/api/v1/sessions', () => {
    return HttpResponse.json([
      { id: 'sess-1', provider: 'anthropic', created_at: new Date().toISOString() },
    ])
  }),

  http.get('/api/v1/sessions/:id', ({ params }) => {
    return HttpResponse.json({ id: params.id, provider: 'anthropic' })
  }),

  http.delete('/api/v1/sessions/:id', () => {
    return HttpResponse.json({ success: true })
  }),

  // Webhooks
  http.get('/api/v1/webhooks', () => {
    return HttpResponse.json([
      { id: 'wh-1', name: 'test', url: 'https://example.com/hook', events: ['failover'], enabled: true },
    ])
  }),

  http.post('/api/v1/webhooks', () => {
    return HttpResponse.json({ id: 'wh-2', name: 'new', url: 'https://example.com', events: [], enabled: true }, { status: 201 })
  }),

  http.put('/api/v1/webhooks/:id', ({ params }) => {
    return HttpResponse.json({ id: params.id, name: 'updated', url: 'https://example.com', events: [], enabled: true })
  }),

  http.delete('/api/v1/webhooks/:id', () => {
    return HttpResponse.json({ success: true })
  }),

  http.post('/api/v1/webhooks/test', () => {
    return HttpResponse.json({ success: true })
  }),

  // Middleware
  http.get('/api/v1/middleware', () => {
    return HttpResponse.json({ enabled: false, middlewares: [] })
  }),

  http.put('/api/v1/middleware', () => {
    return HttpResponse.json({ status: 'updated' })
  }),

  http.post('/api/v1/middleware/:name/enable', () => {
    return HttpResponse.json({ status: 'enabled' })
  }),

  http.post('/api/v1/middleware/:name/disable', () => {
    return HttpResponse.json({ status: 'disabled' })
  }),

  http.post('/api/v1/middleware/reload', () => {
    return HttpResponse.json({ status: 'reloaded' })
  }),

  http.post('/api/v1/middleware/upload', () => {
    return HttpResponse.json({ status: 'uploaded', name: 'test', path: '/tmp/test.so', checksum: 'abc123' })
  }),

  // Reload
  http.post('/api/v1/reload', () => {
    return HttpResponse.json({ status: 'ok' })
  }),

  // Auto-Permission
  http.get('/api/v1/auto-permission', () => {
    return HttpResponse.json({
      claude: null,
      codex: null,
      opencode: null,
    })
  }),

  http.get('/api/v1/auto-permission/:client', () => {
    return HttpResponse.json({ enabled: false, mode: '' })
  }),

  http.put('/api/v1/auto-permission/:client', async ({ request }) => {
    const body = await request.json() as { enabled: boolean; mode: string }
    return HttpResponse.json({ enabled: body.enabled, mode: body.mode })
  }),

  // Sync test & create-gist
  http.post('/api/v1/sync/test', () => {
    return HttpResponse.json({ success: true })
  }),

  http.post('/api/v1/sync/create-gist', () => {
    return HttpResponse.json({ gist_id: 'abc123', gist_url: 'https://gist.github.com/abc123' })
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

  // Monitoring/Requests
  http.get('/api/v1/monitoring/requests', () => {
    return HttpResponse.json({
      requests: [
        {
          id: 'req-123',
          timestamp: new Date().toISOString(),
          provider: 'anthropic',
          model: 'claude-sonnet-4',
          status_code: 200,
          duration_ms: 1500,
          input_tokens: 100,
          output_tokens: 50,
          cost_usd: 0.005,
        },
      ],
      total: 1,
    })
  }),

  http.get('/api/v1/monitoring/requests/:id', ({ params }) => {
    return HttpResponse.json({
      id: params.id,
      timestamp: new Date().toISOString(),
      provider: 'anthropic',
      model: 'claude-sonnet-4',
      status_code: 200,
      duration_ms: 1500,
      input_tokens: 100,
      output_tokens: 50,
      cost_usd: 0.005,
      session_id: 'sess-123',
      client_type: 'claude',
      request_format: 'anthropic',
      request_size: 1024,
      failover_chain: [],
    })
  }),
]
