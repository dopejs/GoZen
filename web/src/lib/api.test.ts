import { describe, it, expect, vi, beforeEach } from 'vitest'
import { providersApi, profilesApi, settingsApi, logsApi, usageApi, syncApi, authApi, healthApi, ApiError, sessionsApi, webhooksApi, middlewareApi, budgetApi, providerHealthApi, botApi } from './api'

describe('API utilities', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('providersApi', () => {
    it('lists providers', async () => {
      const providers = await providersApi.list()
      expect(providers).toHaveLength(2)
      expect(providers[0].name).toBe('anthropic')
    })

    it('gets a provider by name', async () => {
      const provider = await providersApi.get('anthropic')
      expect(provider.name).toBe('anthropic')
      expect(provider.base_url).toBe('https://api.anthropic.com')
    })

    it('creates a provider', async () => {
      const provider = await providersApi.create({
        name: 'new-provider',
        base_url: 'https://api.new.com',
        auth_token: 'token',
      })
      expect(provider.name).toBe('new-provider')
    })

    it('updates a provider', async () => {
      const provider = await providersApi.update('anthropic', {
        base_url: 'https://updated.com',
      })
      expect(provider.base_url).toBe('https://updated.com')
    })

    it('deletes a provider', async () => {
      const result = await providersApi.delete('anthropic')
      expect(result.success).toBeDefined()
    })
  })

  describe('profilesApi', () => {
    it('lists profiles', async () => {
      const profiles = await profilesApi.list()
      expect(profiles).toHaveLength(2)
      expect(profiles[0].name).toBe('default')
    })

    it('gets a profile by name', async () => {
      const profile = await profilesApi.get('default')
      expect(profile.name).toBe('default')
    })

    it('creates a profile', async () => {
      const profile = await profilesApi.create({
        name: 'new-profile',
        providers: ['anthropic'],
      })
      expect(profile.name).toBe('new-profile')
    })

    it('updates a profile', async () => {
      const profile = await profilesApi.update('default', {
        providers: ['anthropic'],
      })
      expect(profile.providers).toContain('anthropic')
    })

    it('deletes a profile', async () => {
      const result = await profilesApi.delete('work')
      expect(result.success).toBeDefined()
    })
  })

  describe('logsApi', () => {
    it('lists logs', async () => {
      const result = await logsApi.list()
      expect(result.entries).toBeDefined()
      expect(result.total).toBeDefined()
    })

    it('lists logs with filters', async () => {
      const result = await logsApi.list({
        provider: 'anthropic',
        errors_only: true,
        limit: 50,
      })
      expect(result).toBeDefined()
    })
  })

  describe('usageApi', () => {
    it('gets usage summary', async () => {
      const summary = await usageApi.summary()
      expect(summary.total_input_tokens).toBe(10000)
      expect(summary.total_cost).toBe(1.5)
    })

    it('gets usage summary with params', async () => {
      const summary = await usageApi.summary({ period: 'week' })
      expect(summary).toBeDefined()
    })

    it('gets hourly usage', async () => {
      const hourly = await usageApi.hourly()
      expect(hourly).toBeDefined()
    })

    it('gets hourly usage with groupBy', async () => {
      const hourly = await usageApi.hourly({ groupBy: 'provider' })
      expect(hourly).toBeDefined()
    })
  })

  describe('usageApi', () => {
    it('creates error with status and message', () => {
      const error = new ApiError(404, 'Not found')
      expect(error.status).toBe(404)
      expect(error.message).toBe('Not found')
      expect(error.name).toBe('ApiError')
    })
  })

  describe('authApi', () => {
    it('checks auth status', async () => {
      const result = await authApi.check()
      expect(result.authenticated).toBeDefined()
    })

    it('logs in', async () => {
      const result = await authApi.login('password123')
      expect(result).toBeDefined()
    })

    it('logs in with encrypted password', async () => {
      const result = await authApi.login('password123', 'encrypted-data')
      expect(result).toBeDefined()
    })

    it('logs out', async () => {
      const result = await authApi.logout()
      expect(result.success).toBeDefined()
    })

    it('gets public key', async () => {
      const result = await authApi.getPubKey()
      expect(result.public_key).toBe('test-pubkey')
    })
  })

  describe('healthApi', () => {
    it('checks health', async () => {
      const result = await healthApi.check()
      expect(result.status).toBe('ok')
      expect(result.version).toBe('3.0.0')
    })

    it('reloads config', async () => {
      const result = await healthApi.reload()
      expect(result.status).toBe('ok')
    })
  })

  describe('sessionsApi', () => {
    it('lists sessions', async () => {
      const sessions = await sessionsApi.list()
      expect(sessions).toHaveLength(1)
      expect(sessions[0].id).toBe('sess-1')
    })

    it('gets a session by id', async () => {
      const session = await sessionsApi.get('sess-1')
      expect(session.id).toBe('sess-1')
    })

    it('deletes a session', async () => {
      const result = await sessionsApi.delete('sess-1')
      expect(result.success).toBeDefined()
    })
  })

  describe('webhooksApi', () => {
    it('lists webhooks', async () => {
      const webhooks = await webhooksApi.list()
      expect(webhooks).toHaveLength(1)
    })

    it('creates a webhook', async () => {
      const webhook = await webhooksApi.create({
        name: 'new', url: 'https://example.com', events: [], enabled: true,
      })
      expect(webhook.id).toBe('wh-2')
    })

    it('updates a webhook', async () => {
      const webhook = await webhooksApi.update('wh-1', { name: 'updated' })
      expect(webhook.name).toBe('updated')
    })

    it('deletes a webhook', async () => {
      const result = await webhooksApi.delete('wh-1')
      expect(result.success).toBeDefined()
    })

    it('tests a webhook', async () => {
      const result = await webhooksApi.test('wh-1')
      expect(result.success).toBe(true)
    })
  })

  describe('middlewareApi', () => {
    it('gets middleware config', async () => {
      const config = await middlewareApi.get()
      expect(config.enabled).toBe(false)
    })

    it('updates middleware config', async () => {
      const result = await middlewareApi.update({ enabled: true, middlewares: [] })
      expect(result.status).toBe('updated')
    })

    it('enables a middleware', async () => {
      const result = await middlewareApi.enable('test-mw')
      expect(result.status).toBe('enabled')
    })

    it('disables a middleware', async () => {
      const result = await middlewareApi.disable('test-mw')
      expect(result.status).toBe('disabled')
    })

    it('reloads middleware', async () => {
      const result = await middlewareApi.reload()
      expect(result.status).toBe('reloaded')
    })
  })

  describe('budgetApi', () => {
    it('gets budget', async () => {
      const budget = await budgetApi.get()
      expect(budget.enabled).toBe(true)
    })

    it('updates budget', async () => {
      const budget = await budgetApi.update({ daily_limit: 20 })
      expect(budget).toBeDefined()
    })

    it('gets budget status', async () => {
      const status = await budgetApi.status()
      expect(status.daily_used).toBe(1.5)
    })
  })

  describe('providerHealthApi', () => {
    it('lists provider health', async () => {
      const health = await providerHealthApi.list()
      expect(health).toHaveLength(1)
      expect(health[0].name).toBe('anthropic')
    })

    it('gets provider health by name', async () => {
      const health = await providerHealthApi.get('anthropic')
      expect(health.name).toBe('anthropic')
      expect(health.status).toBe('healthy')
    })
  })

  describe('settingsApi', () => {
    it('gets settings', async () => {
      const settings = await settingsApi.get()
      expect(settings.default_profile).toBe('default')
      expect(settings.clients).toContain('claude')
    })

    it('updates settings', async () => {
      const settings = await settingsApi.update({
        default_profile: 'work',
      })
      expect(settings.default_profile).toBe('default')
    })

    it('changes password', async () => {
      const result = await settingsApi.changePassword('old', 'new123')
      expect(result.success).toBeDefined()
    })
  })

  describe('syncApi', () => {
    it('gets sync config', async () => {
      const config = await syncApi.getConfig()
      expect(config).toBeDefined()
    })

    it('updates sync config', async () => {
      const config = await syncApi.updateConfig({ backend: 'gist' })
      expect(config).toBeDefined()
    })

    it('gets sync status', async () => {
      const status = await syncApi.status()
      expect(status).toBeDefined()
    })

    it('pulls sync', async () => {
      const result = await syncApi.pull()
      expect(result.success).toBe(true)
    })

    it('pushes sync', async () => {
      const result = await syncApi.push()
      expect(result.success).toBe(true)
    })

    it('tests sync', async () => {
      const result = await syncApi.test()
      expect(result.success).toBe(true)
    })

    it('creates gist', async () => {
      const result = await syncApi.createGist()
      expect(result.gist_id).toBe('abc123')
    })
  })

  describe('usageApi extended', () => {
    it('gets usage summary with date range', async () => {
      const summary = await usageApi.summary({ since: '2026-01-01', until: '2026-02-01' })
      expect(summary).toBeDefined()
    })

    it('gets usage summary with project', async () => {
      const summary = await usageApi.summary({ project: 'test-project' })
      expect(summary).toBeDefined()
    })

    it('gets hourly usage with date range', async () => {
      const hourly = await usageApi.hourly({ since: '2026-01-01', until: '2026-02-01' })
      expect(hourly).toBeDefined()
    })

    it('gets hourly usage with hours', async () => {
      const hourly = await usageApi.hourly({ hours: 24 })
      expect(hourly).toBeDefined()
    })
  })

  describe('botApi', () => {
    it('streams chat response with SSE', async () => {
      const deltas: string[] = []
      let receivedSessionId = ''

      const result = await botApi.chat(
        'hello',
        undefined,
        (content) => deltas.push(content),
        (sid) => { receivedSessionId = sid }
      )

      expect(result.content).toBe('Hello world')
      expect(result.sessionId).toBe('sess-123')
      expect(receivedSessionId).toBe('sess-123')
      expect(deltas).toContain('Hello')
    })

    it('streams chat with existing session', async () => {
      const result = await botApi.chat('hello', 'existing-sess')
      expect(result.content).toBe('Hello world')
    })

    it('clears chat session', async () => {
      const result = await botApi.clearChat('sess-123')
      expect(result.session_id).toBe('sess-123')
      expect(result.status).toBe('cleared')
    })
  })

  describe('middlewareApi.upload', () => {
    it('uploads a middleware plugin', async () => {
      const file = new File(['test'], 'plugin.so', { type: 'application/octet-stream' })
      const result = await middlewareApi.upload(file, 'test-plugin')
      expect(result.status).toBe('uploaded')
      expect(result.name).toBe('test')
    })

    it('uploads without name', async () => {
      const file = new File(['test'], 'plugin.so', { type: 'application/octet-stream' })
      const result = await middlewareApi.upload(file)
      expect(result.status).toBe('uploaded')
    })
  })
})
