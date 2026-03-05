import { useState, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { RefreshCw, Power, PowerOff, Settings2, Puzzle, Plus, Trash2, Upload, Globe } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  useMiddleware,
  useUpdateMiddleware,
  useEnableMiddleware,
  useDisableMiddleware,
  useReloadMiddleware,
} from '@/hooks/use-middleware'
import { middlewareApi } from '@/lib/api'
import type { MiddlewareEntry } from '@/types/api'

type InstallSource = 'upload' | 'remote'

export function MiddlewarePage() {
  const { t } = useTranslation()
  const { data: config, isLoading } = useMiddleware()
  const updateMiddleware = useUpdateMiddleware()
  const enableMiddleware = useEnableMiddleware()
  const disableMiddleware = useDisableMiddleware()
  const reloadMiddleware = useReloadMiddleware()

  const [installDialogOpen, setInstallDialogOpen] = useState(false)
  const [installSource, setInstallSource] = useState<InstallSource>('upload')
  const [installName, setInstallName] = useState('')
  const [installUrl, setInstallUrl] = useState('')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [isInstalling, setIsInstalling] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleGlobalToggle = async (enabled: boolean) => {
    if (!config) return
    try {
      await updateMiddleware.mutateAsync({
        ...config,
        enabled,
      })
      toast.success(enabled ? t('middleware.enabled') : t('middleware.disabled'))
    } catch {
      toast.error(t('errors.unknown'))
    }
  }

  const handleMiddlewareToggle = async (middleware: MiddlewareEntry) => {
    try {
      if (middleware.enabled) {
        await disableMiddleware.mutateAsync(middleware.name)
        toast.success(t('middleware.middlewareDisabled', { name: middleware.name }))
      } else {
        await enableMiddleware.mutateAsync(middleware.name)
        toast.success(t('middleware.middlewareEnabled', { name: middleware.name }))
      }
    } catch {
      toast.error(t('errors.unknown'))
    }
  }

  const handleReload = async () => {
    try {
      await reloadMiddleware.mutateAsync()
      toast.success(t('middleware.reloaded'))
    } catch {
      toast.error(t('errors.unknown'))
    }
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      if (!file.name.endsWith('.so')) {
        toast.error(t('middleware.onlySoFiles'))
        return
      }
      setSelectedFile(file)
      // Auto-fill name from filename if empty
      if (!installName) {
        setInstallName(file.name.replace('.so', ''))
      }
    }
  }

  const handleInstall = async () => {
    if (!config) return
    if (!installName.trim()) {
      toast.error(t('middleware.nameRequired'))
      return
    }
    if (installSource === 'upload' && !selectedFile) {
      toast.error(t('middleware.fileRequired'))
      return
    }
    if (installSource === 'remote' && !installUrl.trim()) {
      toast.error(t('middleware.urlRequired'))
      return
    }

    // Check for duplicate name
    if (config.middlewares.some((m) => m.name === installName.trim())) {
      toast.error(t('middleware.duplicateName'))
      return
    }

    setIsInstalling(true)
    try {
      let pluginPath = ''

      if (installSource === 'upload' && selectedFile) {
        // Upload file first
        const uploadResult = await middlewareApi.upload(selectedFile, installName.trim())
        pluginPath = uploadResult.path
      }

      const newEntry: MiddlewareEntry = {
        name: installName.trim(),
        enabled: true,
        source: installSource === 'upload' ? 'local' : 'remote',
        ...(installSource === 'upload' ? { path: pluginPath } : { url: installUrl.trim() }),
      }

      await updateMiddleware.mutateAsync({
        ...config,
        middlewares: [...config.middlewares, newEntry],
      })

      // Reload to load the new plugin
      await reloadMiddleware.mutateAsync()

      toast.success(t('middleware.installSuccess', { name: installName }))
      setInstallDialogOpen(false)
      resetInstallForm()
    } catch {
      toast.error(t('middleware.installFailed'))
    } finally {
      setIsInstalling(false)
    }
  }

  const handleRemove = async (middleware: MiddlewareEntry) => {
    if (!config) return
    if (middleware.source === 'builtin') {
      toast.error(t('middleware.cannotRemoveBuiltin'))
      return
    }

    try {
      await updateMiddleware.mutateAsync({
        ...config,
        middlewares: config.middlewares.filter((m) => m.name !== middleware.name),
      })
      toast.success(t('middleware.removeSuccess', { name: middleware.name }))
    } catch {
      toast.error(t('errors.unknown'))
    }
  }

  const resetInstallForm = () => {
    setInstallName('')
    setInstallUrl('')
    setSelectedFile(null)
    setInstallSource('upload')
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  const middlewares = config?.middlewares || []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{t('middleware.title')}</h1>
          <p className="text-muted-foreground">{t('middleware.description')}</p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setInstallDialogOpen(true)}
          >
            <Plus className="mr-2 h-4 w-4" />
            {t('middleware.install')}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleReload}
            disabled={reloadMiddleware.isPending}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${reloadMiddleware.isPending ? 'animate-spin' : ''}`} />
            {t('middleware.reload')}
          </Button>
        </div>
      </div>

      {/* Global Enable/Disable */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings2 className="h-5 w-5" />
            {t('middleware.globalSettings')}
          </CardTitle>
          <CardDescription>{t('middleware.globalSettingsDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>{t('middleware.enablePipeline')}</Label>
              <p className="text-sm text-muted-foreground">
                {t('middleware.enablePipelineDesc')}
              </p>
            </div>
            <Switch
              checked={config?.enabled || false}
              onCheckedChange={handleGlobalToggle}
              disabled={updateMiddleware.isPending}
            />
          </div>
        </CardContent>
      </Card>

      {/* Middleware List */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Puzzle className="h-5 w-5" />
            {t('middleware.middlewares')}
          </CardTitle>
          <CardDescription>{t('middleware.middlewaresDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          {middlewares.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">
              {t('middleware.noMiddlewares')}
            </p>
          ) : (
            <div className="space-y-4">
              {middlewares.map((middleware) => (
                <div
                  key={middleware.name}
                  className="flex items-center justify-between rounded-lg border p-4"
                >
                  <div className="flex items-center gap-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                      {middleware.enabled ? (
                        <Power className="h-5 w-5 text-primary" />
                      ) : (
                        <PowerOff className="h-5 w-5 text-muted-foreground" />
                      )}
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{middleware.name}</span>
                        <Badge variant={middleware.source === 'builtin' ? 'secondary' : 'outline'}>
                          {middleware.source}
                        </Badge>
                        {middleware.version && (
                          <Badge variant="outline">v{middleware.version}</Badge>
                        )}
                        {middleware.priority !== undefined && (
                          <Badge variant="outline">
                            {t('middleware.priority')}: {middleware.priority}
                          </Badge>
                        )}
                      </div>
                      {middleware.description && (
                        <p className="text-sm text-muted-foreground mt-1">
                          {middleware.description}
                        </p>
                      )}
                      {middleware.path && (
                        <p className="text-xs text-muted-foreground mt-1 font-mono">
                          {middleware.path}
                        </p>
                      )}
                      {middleware.url && (
                        <p className="text-xs text-muted-foreground mt-1 font-mono">
                          {middleware.url}
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {middleware.source !== 'builtin' && (
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRemove(middleware)}
                        disabled={updateMiddleware.isPending}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    )}
                    <Switch
                      checked={middleware.enabled}
                      onCheckedChange={() => handleMiddlewareToggle(middleware)}
                      disabled={!config?.enabled || enableMiddleware.isPending || disableMiddleware.isPending}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Beta Notice */}
      <Card className="border-amber-500/50 bg-amber-500/5">
        <CardContent className="pt-6">
          <p className="text-sm text-amber-600 dark:text-amber-400">
            <strong>{t('middleware.betaNotice')}</strong> {t('middleware.betaNoticeDesc')}
          </p>
        </CardContent>
      </Card>

      {/* Install Dialog */}
      <Dialog open={installDialogOpen} onOpenChange={setInstallDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('middleware.installPlugin')}</DialogTitle>
            <DialogDescription>{t('middleware.installPluginDesc')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>{t('middleware.pluginName')}</Label>
              <Input
                value={installName}
                onChange={(e) => setInstallName(e.target.value)}
                placeholder="my-plugin"
              />
            </div>
            <div className="space-y-2">
              <Label>{t('middleware.source')}</Label>
              <Select value={installSource} onValueChange={(v) => setInstallSource(v as InstallSource)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="upload">
                    <div className="flex items-center gap-2">
                      <Upload className="h-4 w-4" />
                      {t('middleware.sourceUpload')}
                    </div>
                  </SelectItem>
                  <SelectItem value="remote">
                    <div className="flex items-center gap-2">
                      <Globe className="h-4 w-4" />
                      {t('middleware.sourceRemote')}
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            {installSource === 'upload' ? (
              <div className="space-y-2">
                <Label>{t('middleware.uploadFile')}</Label>
                <div className="flex gap-2">
                  <Input
                    type="file"
                    accept=".so"
                    ref={fileInputRef}
                    onChange={handleFileSelect}
                    className="cursor-pointer"
                  />
                </div>
                {selectedFile && (
                  <p className="text-xs text-muted-foreground">
                    {t('middleware.selectedFile')}: {selectedFile.name} ({(selectedFile.size / 1024).toFixed(1)} KB)
                  </p>
                )}
                <p className="text-xs text-muted-foreground">{t('middleware.uploadHint')}</p>
              </div>
            ) : (
              <div className="space-y-2">
                <Label>{t('middleware.remoteUrl')}</Label>
                <Input
                  value={installUrl}
                  onChange={(e) => setInstallUrl(e.target.value)}
                  placeholder="https://example.com/manifest.json"
                />
                <p className="text-xs text-muted-foreground">{t('middleware.remoteUrlHint')}</p>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setInstallDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleInstall} disabled={isInstalling}>
              {isInstalling ? t('middleware.installing') : t('middleware.install')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
