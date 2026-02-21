import { useState, useEffect } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { ArrowLeft, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useProvider, useCreateProvider, useUpdateProvider } from '@/hooks/use-providers'
import { AVAILABLE_CLIENTS, CLIENT_ENV_HINTS, type ClientType, type Provider } from '@/types/api'

export function ProviderEditPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { name } = useParams<{ name: string }>()
  const [searchParams, setSearchParams] = useSearchParams()
  const isNew = !name || name === 'new'

  const currentTab = searchParams.get('s') || 'basic'
  const setCurrentTab = (tab: string) => {
    setSearchParams({ s: tab })
  }

  const { data: existingProvider, isLoading } = useProvider(name || '')
  const createProvider = useCreateProvider()
  const updateProvider = useUpdateProvider()

  // Form state
  const [formData, setFormData] = useState<Partial<Provider>>({
    name: '',
    type: 'anthropic',
    base_url: 'https://api.anthropic.com',
    auth_token: '',
    model: '',
    reasoning_model: '',
    haiku_model: '',
    opus_model: '',
    sonnet_model: '',
    env_vars: {},
    claude_env_vars: {},
    codex_env_vars: {},
    opencode_env_vars: {},
  })

  // Initialize form with existing data
  useEffect(() => {
    if (existingProvider && !isNew) {
      setFormData({
        ...existingProvider,
        auth_token: '', // Don't show masked token
      })
    }
  }, [existingProvider, isNew])

  const handleSave = async () => {
    if (!formData.name) {
      toast.error(t('providers.nameRequired'))
      return
    }
    try {
      if (isNew) {
        await createProvider.mutateAsync(formData as Provider)
      } else {
        await updateProvider.mutateAsync({ name: name!, provider: formData })
      }
      toast.success(t('common.success'))
      navigate('/providers')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const updateEnvVar = (client: ClientType | 'legacy', key: string, value: string) => {
    const fieldMap: Record<string, keyof Provider> = {
      legacy: 'env_vars',
      claude: 'claude_env_vars',
      codex: 'codex_env_vars',
      opencode: 'opencode_env_vars',
    }
    const field = fieldMap[client]
    setFormData((prev) => ({
      ...prev,
      [field]: { ...(prev[field] as Record<string, string>), [key]: value },
    }))
  }

  const removeEnvVar = (client: ClientType | 'legacy', key: string) => {
    const fieldMap: Record<string, keyof Provider> = {
      legacy: 'env_vars',
      claude: 'claude_env_vars',
      codex: 'codex_env_vars',
      opencode: 'opencode_env_vars',
    }
    const field = fieldMap[client]
    const current = { ...(formData[field] as Record<string, string>) }
    delete current[key]
    setFormData((prev) => ({ ...prev, [field]: current }))
  }

  const addEnvVar = (client: ClientType | 'legacy') => {
    updateEnvVar(client, '', '')
  }

  if (isLoading && !isNew) {
    return <div className="flex justify-center p-8">{t('common.loading')}</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => navigate('/providers')}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold">
            {isNew ? t('providers.addProvider') : t('providers.editProvider')}
          </h1>
          <p className="text-muted-foreground">
            {isNew ? t('providers.addProviderDesc') : name}
          </p>
        </div>
      </div>

      <Tabs value={currentTab} onValueChange={setCurrentTab}>
        <TabsList>
          <TabsTrigger value="basic">{t('providers.basicSettings')}</TabsTrigger>
          <TabsTrigger value="models">{t('providers.modelSettings')}</TabsTrigger>
          <TabsTrigger value="envvars">{t('providers.envVars')}</TabsTrigger>
        </TabsList>

        <TabsContent value="basic" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">{t('providers.basicSettings')}</CardTitle>
              <CardDescription>{t('providers.basicSettingsDesc')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="name">{t('providers.name')}</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  disabled={!isNew}
                  placeholder="my-provider"
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="type">{t('providers.type')}</Label>
                <Select
                  value={formData.type || 'anthropic'}
                  onValueChange={(value) => setFormData({ ...formData, type: value as 'anthropic' | 'openai' })}
                >
                  <SelectTrigger id="type">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="anthropic">Anthropic</SelectItem>
                    <SelectItem value="openai">OpenAI Compatible</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="base_url">{t('providers.baseUrl')}</Label>
                <Input
                  id="base_url"
                  value={formData.base_url}
                  onChange={(e) => setFormData({ ...formData, base_url: e.target.value })}
                  placeholder="https://api.anthropic.com"
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="auth_token">{t('providers.authToken')}</Label>
                <Input
                  id="auth_token"
                  type="password"
                  value={formData.auth_token}
                  onChange={(e) => setFormData({ ...formData, auth_token: e.target.value })}
                  placeholder={isNew ? '' : t('providers.leaveEmptyToKeep')}
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="models" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">{t('providers.modelSettings')}</CardTitle>
              <CardDescription>{t('providers.modelSettingsDesc')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="model">{t('providers.defaultModel')}</Label>
                <Input
                  id="model"
                  value={formData.model}
                  onChange={(e) => setFormData({ ...formData, model: e.target.value })}
                  placeholder="claude-sonnet-4-20250514"
                />
                <p className="text-xs text-muted-foreground">{t('providers.defaultModelHint')}</p>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="sonnet_model">{t('providers.sonnetModel')}</Label>
                <Input
                  id="sonnet_model"
                  value={formData.sonnet_model}
                  onChange={(e) => setFormData({ ...formData, sonnet_model: e.target.value })}
                  placeholder="claude-sonnet-4-20250514"
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="opus_model">{t('providers.opusModel')}</Label>
                <Input
                  id="opus_model"
                  value={formData.opus_model}
                  onChange={(e) => setFormData({ ...formData, opus_model: e.target.value })}
                  placeholder="claude-opus-4-20250514"
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="haiku_model">{t('providers.haikuModel')}</Label>
                <Input
                  id="haiku_model"
                  value={formData.haiku_model}
                  onChange={(e) => setFormData({ ...formData, haiku_model: e.target.value })}
                  placeholder="claude-haiku-3-5-20241022"
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="reasoning_model">{t('providers.reasoningModel')}</Label>
                <Input
                  id="reasoning_model"
                  value={formData.reasoning_model}
                  onChange={(e) => setFormData({ ...formData, reasoning_model: e.target.value })}
                  placeholder="claude-sonnet-4-20250514"
                />
                <p className="text-xs text-muted-foreground">{t('providers.reasoningModelHint')}</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="envvars" className="mt-4">
          <div className="space-y-4">
            {/* Legacy env vars */}
            <EnvVarsCard
              title={t('providers.legacyEnvVars')}
              description={t('providers.legacyEnvVarsDesc')}
              envVars={formData.env_vars || {}}
              hints={[]}
              onUpdate={(key, value) => updateEnvVar('legacy', key, value)}
              onRemove={(key) => removeEnvVar('legacy', key)}
              onAdd={() => addEnvVar('legacy')}
            />

            {/* Client-specific env vars */}
            {AVAILABLE_CLIENTS.map((client) => (
              <EnvVarsCard
                key={client}
                title={t(`providers.${client}EnvVars`)}
                description={t(`providers.${client}EnvVarsDesc`)}
                envVars={(formData[`${client}_env_vars` as keyof Provider] as Record<string, string>) || {}}
                hints={CLIENT_ENV_HINTS[client]}
                onUpdate={(key, value) => updateEnvVar(client, key, value)}
                onRemove={(key) => removeEnvVar(client, key)}
                onAdd={() => addEnvVar(client)}
              />
            ))}
          </div>
        </TabsContent>
      </Tabs>

      <div className="flex gap-2">
        <Button onClick={handleSave} disabled={createProvider.isPending || updateProvider.isPending}>
          {t('common.save')}
        </Button>
        <Button variant="outline" onClick={() => navigate('/providers')}>
          {t('common.cancel')}
        </Button>
      </div>
    </div>
  )
}

interface EnvVarsCardProps {
  title: string
  description: string
  envVars: Record<string, string>
  hints: string[]
  onUpdate: (key: string, value: string) => void
  onRemove: (key: string) => void
  onAdd: () => void
}

function EnvVarsCard({ title, description, envVars, hints, onUpdate, onRemove, onAdd }: EnvVarsCardProps) {
  const { t } = useTranslation()
  const entries = Object.entries(envVars)
  const existingKeys = new Set(Object.keys(envVars))

  const handleHintClick = (hint: string) => {
    if (!existingKeys.has(hint)) {
      // Check if there's an empty key row, remove it first so the new one goes to the end
      if (existingKeys.has('')) {
        onRemove('')
      }
      onUpdate(hint, '')
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {hints.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            <span className="text-xs text-muted-foreground">{t('providers.commonVars')}:</span>
            {hints.map((hint) => (
              <code
                key={hint}
                className={`text-xs px-1 rounded ${
                  existingKeys.has(hint)
                    ? 'bg-muted text-muted-foreground line-through'
                    : 'bg-muted hover:bg-primary/20 cursor-pointer'
                }`}
                onClick={() => handleHintClick(hint)}
              >
                {hint}
              </code>
            ))}
          </div>
        )}

        {entries.map(([key, value], index) => (
          <div key={index} className="flex gap-2 items-center">
            <Input
              value={key}
              onChange={(e) => {
                const newKey = e.target.value
                onRemove(key)
                onUpdate(newKey, value)
              }}
              placeholder="KEY"
              className="flex-1"
            />
            <span className="text-muted-foreground">=</span>
            <Input
              value={value}
              onChange={(e) => onUpdate(key, e.target.value)}
              placeholder="value"
              className="flex-[2]"
            />
            <Button variant="ghost" size="icon" onClick={() => onRemove(key)}>
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ))}

        <Button variant="outline" size="sm" onClick={onAdd}>
          <Plus className="h-4 w-4 mr-1" />
          {t('providers.addEnvVar')}
        </Button>
      </CardContent>
    </Card>
  )
}
