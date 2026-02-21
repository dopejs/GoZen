import { describe, it, expect, vi, beforeEach } from 'vitest'
import { providersApi, profilesApi, settingsApi, bindingsApi, logsApi, usageApi, botApi, syncApi, authApi, healthApi, ApiError } from './api'

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
  })

  describe('bindingsApi', () => {
    it('lists bindings', async () => {
      const result = await bindingsApi.list()
      expect(result.bindings).toHaveLength(1)
      expect(result.profiles).toContain('default')
    })

    it('creates a binding', async () => {
      const binding = await bindingsApi.create('/new/project', 'default')
      expect(binding.path).toBe('/new/project')
    })

    it('deletes a binding', async () => {
      const result = await bindingsApi.delete('/test/project')
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

  describe('botApi', () => {
    it('gets bot config', async () => {
      const bot = await botApi.get()
      expect(bot.enabled).toBe(true)
      expect(bot.profile).toBe('default')
    })

    it('updates bot config', async () => {
      const bot = await botApi.update({ enabled: false })
      expect(bot.enabled).toBe(true)
    })
  })

  describe('syncApi', () => {
    it('gets sync config', async () => {
      const config = await syncApi.getConfig()
      expect(config).toBeDefined()
    })

    it('gets sync status', async () => {
      const status = await syncApi.status()
      expect(status).toBeDefined()
    })
  })

  describe('authApi', () => {
    it('checks auth status', async () => {
      const result = await authApi.check()
      expect(result.authenticated).toBeDefined()
    })
  })

  describe('healthApi', () => {
    it('checks health', async () => {
      const result = await healthApi.check()
      expect(result.status).toBe('ok')
      expect(result.version).toBe('3.0.0')
    })
  })

  describe('ApiError', () => {
    it('creates error with status and message', () => {
      const error = new ApiError(404, 'Not found')
      expect(error.status).toBe(404)
      expect(error.message).toBe('Not found')
      expect(error.name).toBe('ApiError')
    })
  })
})
