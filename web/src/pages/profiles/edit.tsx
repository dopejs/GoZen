import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { ArrowLeft, Plus, Trash2, GripVertical, ChevronDown, ChevronUp } from 'lucide-react'
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
import { useProfile, useCreateProfile, useUpdateProfile } from '@/hooks/use-profiles'
import { useProviders } from '@/hooks/use-providers'
import {
  SCENARIOS,
  SCENARIO_LABELS,
  LOAD_BALANCE_STRATEGIES,
  type Profile,
  type Scenario,
  type ScenarioRoute,
  type ProviderRoute,
  type LoadBalanceStrategy,
} from '@/types/api'

export function ProfileEditPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { name } = useParams<{ name: string }>()
  const isNew = !name || name === 'new'

  const { data: existingProfile, isLoading } = useProfile(name || '')
  const { data: providers } = useProviders()
  const createProfile = useCreateProfile()
  const updateProfile = useUpdateProfile()

  // Form state
  const [formData, setFormData] = useState<Partial<Profile>>({
    name: '',
    providers: [],
    strategy: 'failover',
    long_context_threshold: 100000,
    routing: {},
  })

  // Scenario routing expanded state
  const [expandedScenarios, setExpandedScenarios] = useState<Record<Scenario, boolean>>({} as Record<Scenario, boolean>)

  // Initialize form with existing data
  useEffect(() => {
    if (existingProfile && !isNew) {
      setFormData(existingProfile)
    }
  }, [existingProfile, isNew])
  const handleSave = async () => {
    if (!formData.name) {
      toast.error(t('profiles.nameRequired'))
      return
    }
    try {
      if (isNew) {
        await createProfile.mutateAsync(formData as Profile)
      } else {
        await updateProfile.mutateAsync({ name: name!, profile: formData })
      }
      toast.success(t('common.success'))
      navigate('/profiles')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const availableProviders = providers?.filter((p) => !formData.providers?.includes(p.name)) || []

  const addProvider = (providerName: string) => {
    setFormData((prev) => ({
      ...prev,
      providers: [...(prev.providers || []), providerName],
    }))
  }

  const removeProvider = (providerName: string) => {
    setFormData((prev) => ({
      ...prev,
      providers: (prev.providers || []).filter((p) => p !== providerName),
    }))
  }

  const moveProvider = (index: number, direction: 'up' | 'down') => {
    const newProviders = [...(formData.providers || [])]
    const newIndex = direction === 'up' ? index - 1 : index + 1
    if (newIndex < 0 || newIndex >= newProviders.length) return
    ;[newProviders[index], newProviders[newIndex]] = [newProviders[newIndex], newProviders[index]]
    setFormData((prev) => ({ ...prev, providers: newProviders }))
  }

  const toggleScenario = (scenario: Scenario) => {
    setExpandedScenarios((prev) => ({ ...prev, [scenario]: !prev[scenario] }))
  }

  const updateScenarioRoute = (scenario: Scenario, route: ScenarioRoute | undefined) => {
    setFormData((prev) => {
      const newRouting = { ...(prev.routing || {}) }
      if (route && route.providers.length > 0) {
        newRouting[scenario] = route
      } else {
        delete newRouting[scenario]
      }
      return { ...prev, routing: newRouting }
    })
  }

  if (isLoading && !isNew) {
    return <div className="flex justify-center p-8">{t('common.loading')}</div>
  }

  const strategyDescriptions: Record<LoadBalanceStrategy, string> = {
    failover: t('profiles.strategyFailoverDesc'),
    'round-robin': t('profiles.strategyRoundRobinDesc'),
    'least-latency': t('profiles.strategyLeastLatencyDesc'),
    'least-cost': t('profiles.strategyLeastCostDesc'),
  }
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => navigate('/profiles')}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold">
            {isNew ? t('profiles.addProfile') : t('profiles.editProfile')}
          </h1>
          <p className="text-muted-foreground">
            {isNew ? t('profiles.addProfileDesc') : name}
          </p>
        </div>
      </div>

      <Tabs defaultValue="basic">
        <TabsList>
          <TabsTrigger value="basic">{t('profiles.basicSettings')}</TabsTrigger>
          <TabsTrigger value="providers">{t('profiles.providers')}</TabsTrigger>
          <TabsTrigger value="routing">{t('profiles.scenarioRouting')}</TabsTrigger>
        </TabsList>

        <TabsContent value="basic" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">{t('profiles.basicSettings')}</CardTitle>
              <CardDescription>{t('profiles.basicSettingsDesc')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="name">{t('profiles.name')}</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  disabled={!isNew}
                  placeholder="my-profile"
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="strategy">{t('profiles.strategy')}</Label>
                <Select
                  value={formData.strategy || 'failover'}
                  onValueChange={(value) => setFormData({ ...formData, strategy: value as LoadBalanceStrategy })}
                >
                  <SelectTrigger id="strategy">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {LOAD_BALANCE_STRATEGIES.map((s) => (
                      <SelectItem key={s} value={s}>
                        {t(`profiles.strategy${s.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join('')}`)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  {strategyDescriptions[formData.strategy || 'failover']}
                </p>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="long_context_threshold">{t('profiles.longContextThreshold')}</Label>
                <Input
                  id="long_context_threshold"
                  type="number"
                  value={formData.long_context_threshold || ''}
                  onChange={(e) => setFormData({ ...formData, long_context_threshold: parseInt(e.target.value) || undefined })}
                  placeholder="100000"
                />
                <p className="text-xs text-muted-foreground">{t('profiles.longContextThresholdHint')}</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="providers" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">{t('profiles.providers')}</CardTitle>
              <CardDescription>{t('profiles.providersDesc')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <Label>{t('profiles.selectedProviders')}</Label>
                {formData.providers && formData.providers.length > 0 ? (
                  <div className="space-y-2">
                    {formData.providers.map((providerName, index) => (
                      <div key={providerName} className="flex items-center gap-2 p-2 border rounded-md bg-muted/50">
                        <GripVertical className="h-4 w-4 text-muted-foreground" />
                        <span className="flex-1 font-mono text-sm">{providerName}</span>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => moveProvider(index, 'up')}
                          disabled={index === 0}
                        >
                          <ChevronUp className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => moveProvider(index, 'down')}
                          disabled={index === formData.providers!.length - 1}
                        >
                          <ChevronDown className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" onClick={() => removeProvider(providerName)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">{t('profiles.noProvidersSelected')}</p>
                )}
              </div>

              {availableProviders.length > 0 && (
                <div className="grid gap-2">
                  <Label>{t('profiles.availableProviders')}</Label>
                  <div className="flex flex-wrap gap-2">
                    {availableProviders.map((provider) => (
                      <Button
                        key={provider.name}
                        variant="outline"
                        size="sm"
                        onClick={() => addProvider(provider.name)}
                      >
                        <Plus className="h-4 w-4 mr-1" />
                        {provider.name}
                      </Button>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="routing" className="mt-4">
          <div className="space-y-4">
            {SCENARIOS.filter((s) => s !== 'default').map((scenario) => (
              <ScenarioCard
                key={scenario}
                scenario={scenario}
                route={formData.routing?.[scenario]}
                providers={providers || []}
                expanded={expandedScenarios[scenario] || false}
                onToggle={() => toggleScenario(scenario)}
                onUpdate={(route) => updateScenarioRoute(scenario, route)}
              />
            ))}
          </div>
        </TabsContent>
      </Tabs>

      <div className="flex gap-2">
        <Button onClick={handleSave} disabled={createProfile.isPending || updateProfile.isPending}>
          {t('common.save')}
        </Button>
        <Button variant="outline" onClick={() => navigate('/profiles')}>
          {t('common.cancel')}
        </Button>
      </div>
    </div>
  )
}

interface ScenarioCardProps {
  scenario: Scenario
  route?: ScenarioRoute
  providers: { name: string }[]
  expanded: boolean
  onToggle: () => void
  onUpdate: (route: ScenarioRoute | undefined) => void
}

function ScenarioCard({ scenario, route, providers, expanded, onToggle, onUpdate }: ScenarioCardProps) {
  const { t } = useTranslation()
  const hasRoute = route && route.providers.length > 0

  const addScenarioProvider = () => {
    const newProviders: ProviderRoute[] = [...(route?.providers || []), { name: '' }]
    onUpdate({ providers: newProviders })
  }

  const updateScenarioProvider = (index: number, providerRoute: ProviderRoute) => {
    const newProviders = [...(route?.providers || [])]
    newProviders[index] = providerRoute
    onUpdate({ providers: newProviders })
  }

  const removeScenarioProvider = (index: number) => {
    const newProviders = (route?.providers || []).filter((_, i) => i !== index)
    onUpdate(newProviders.length > 0 ? { providers: newProviders } : undefined)
  }

  return (
    <Card>
      <CardHeader
        className="cursor-pointer"
        onClick={onToggle}
      >
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-base">{SCENARIO_LABELS[scenario]}</CardTitle>
            <CardDescription>
              {hasRoute
                ? `${route.providers.length} ${t('profiles.scenarioProviders').toLowerCase()}`
                : t('profiles.inheritFromProfile')}
            </CardDescription>
          </div>
          {expanded ? <ChevronUp className="h-5 w-5" /> : <ChevronDown className="h-5 w-5" />}
        </div>
      </CardHeader>
      {expanded && (
        <CardContent className="space-y-3">
          {route?.providers.map((providerRoute, index) => (
            <div key={index} className="flex gap-2 items-center">
              <Select
                value={providerRoute.name}
                onValueChange={(value) => updateScenarioProvider(index, { ...providerRoute, name: value })}
              >
                <SelectTrigger className="flex-1">
                  <SelectValue placeholder={t('profiles.providers')} />
                </SelectTrigger>
                <SelectContent>
                  {providers.map((p) => (
                    <SelectItem key={p.name} value={p.name}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Input
                value={providerRoute.model || ''}
                onChange={(e) => updateScenarioProvider(index, { ...providerRoute, model: e.target.value || undefined })}
                placeholder={t('profiles.modelOverride')}
                className="flex-1"
              />
              <Button variant="ghost" size="icon" onClick={() => removeScenarioProvider(index)}>
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
          <Button variant="outline" size="sm" onClick={addScenarioProvider}>
            <Plus className="h-4 w-4 mr-1" />
            {t('profiles.addScenarioProvider')}
          </Button>
        </CardContent>
      )}
    </Card>
  )
}
