import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Settings as SettingsIcon, FolderOpen, Lock, RefreshCw, Trash2, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  useBindings,
  useDeleteBinding,
  useChangePassword,
  useSyncConfig,
  useSyncStatus,
  useSyncPull,
  useSyncPush,
} from '@/hooks/use-settings'
import { settingsApi, bindingsApi } from '@/lib/api'

export function SettingsPage() {
  const { t } = useTranslation()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">{t('settings.title')}</h1>
        <p className="text-muted-foreground">{t('settings.description')}</p>
      </div>

      <Tabs defaultValue="general">
        <TabsList>
          <TabsTrigger value="general">{t('settings.general')}</TabsTrigger>
          <TabsTrigger value="bindings">{t('settings.bindings')}</TabsTrigger>
          <TabsTrigger value="sync">{t('settings.sync')}</TabsTrigger>
          <TabsTrigger value="password">{t('settings.webPassword')}</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-4">
          <GeneralSettings />
        </TabsContent>

        <TabsContent value="bindings" className="mt-4">
          <BindingsSettings />
        </TabsContent>

        <TabsContent value="sync" className="mt-4">
          <SyncSettings />
        </TabsContent>

        <TabsContent value="password" className="mt-4">
          <PasswordSettings />
        </TabsContent>
      </Tabs>
    </div>
  )
}

function GeneralSettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const { data: settings, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: settingsApi.get,
  })

  const updateSettings = useMutation({
    mutationFn: settingsApi.update,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
      toast.success(t('common.success'))
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    },
  })

  const [defaultProfile, setDefaultProfile] = useState('')
  const [defaultClient, setDefaultClient] = useState('')

  // Initialize form when data loads
  useState(() => {
    if (settings) {
      setDefaultProfile(settings.default_profile || '')
      setDefaultClient(settings.default_client || '')
    }
  })

  if (isLoading) {
    return <div className="flex justify-center py-4">{t('common.loading')}</div>
  }

  const handleSave = () => {
    updateSettings.mutate({
      default_profile: defaultProfile || settings?.default_profile,
      default_client: defaultClient || settings?.default_client,
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <SettingsIcon className="h-5 w-5" />
          {t('settings.general')}
        </CardTitle>
        <CardDescription>{t('settings.generalDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-2">
          <Label>{t('settings.defaultProfile')}</Label>
          <Select
            value={defaultProfile || settings?.default_profile || ''}
            onValueChange={setDefaultProfile}
          >
            <SelectTrigger>
              <SelectValue placeholder={t('settings.selectProfile')} />
            </SelectTrigger>
            <SelectContent>
              {settings?.profiles?.map((p) => (
                <SelectItem key={p} value={p}>{p}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="grid gap-2">
          <Label>{t('settings.defaultClient')}</Label>
          <Select
            value={defaultClient || settings?.default_client || ''}
            onValueChange={setDefaultClient}
          >
            <SelectTrigger>
              <SelectValue placeholder={t('settings.selectClient')} />
            </SelectTrigger>
            <SelectContent>
              {settings?.clients?.map((c) => (
                <SelectItem key={c} value={c}>{c}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="grid gap-2">
          <Label>{t('settings.webPort')}</Label>
          <Input value={settings?.web_port || ''} disabled />
          <p className="text-xs text-muted-foreground">{t('settings.webPortHint')}</p>
        </div>

        <Button onClick={handleSave} disabled={updateSettings.isPending}>
          {t('common.save')}
        </Button>
      </CardContent>
    </Card>
  )
}

function BindingsSettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { data: bindingsData, isLoading } = useBindings()
  const deleteBinding = useDeleteBinding()

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deletingPath, setDeletingPath] = useState<string | null>(null)
  const [addDialogOpen, setAddDialogOpen] = useState(false)
  const [newPath, setNewPath] = useState('')
  const [newProfile, setNewProfile] = useState('')
  const [newClient, setNewClient] = useState('')

  const addBinding = useMutation({
    mutationFn: (data: { path: string; profile?: string; cli?: string }) =>
      bindingsApi.create(data.path, data.profile, data.cli),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bindings'] })
      toast.success(t('common.success'))
      setAddDialogOpen(false)
      setNewPath('')
      setNewProfile('')
      setNewClient('')
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    },
  })

  const handleOpenDelete = (path: string) => {
    setDeletingPath(path)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!deletingPath) return
    try {
      await deleteBinding.mutateAsync(deletingPath)
      toast.success(t('common.success'))
      setDeleteDialogOpen(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const handleAdd = () => {
    if (!newPath) {
      toast.error(t('settings.pathRequired'))
      return
    }
    addBinding.mutate({
      path: newPath,
      profile: newProfile || undefined,
      cli: newClient || undefined,
    })
  }

  const bindings = bindingsData?.bindings || []
  const profiles = bindingsData?.profiles || []
  const clients = bindingsData?.clients || []

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <FolderOpen className="h-5 w-5" />
                {t('settings.bindings')}
              </CardTitle>
              <CardDescription className="mt-1.5">{t('settings.bindingsDesc')}</CardDescription>
            </div>
            <Button onClick={() => setAddDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              {t('settings.addBinding')}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center py-4">{t('common.loading')}</div>
          ) : bindings.length > 0 ? (
            <div className="space-y-2">
              {bindings.map((binding) => (
                <div key={binding.path} className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <p className="font-mono text-sm">{binding.path}</p>
                    <p className="text-sm text-muted-foreground">
                      {binding.profile && `Profile: ${binding.profile}`}
                      {binding.profile && binding.cli && ' · '}
                      {binding.cli && `CLI: ${binding.cli}`}
                    </p>
                  </div>
                  <Button variant="ghost" size="icon" onClick={() => handleOpenDelete(binding.path)}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-center text-muted-foreground py-4">{t('settings.noBindings')}</p>
          )}
        </CardContent>
      </Card>

      {/* Add Binding Dialog */}
      <Dialog open={addDialogOpen} onOpenChange={setAddDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('settings.addBinding')}</DialogTitle>
            <DialogDescription>{t('settings.addBindingDesc')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="grid gap-2">
              <Label>{t('settings.projectPath')}</Label>
              <Input
                value={newPath}
                onChange={(e) => setNewPath(e.target.value)}
                placeholder="/path/to/project"
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('settings.profile')}</Label>
              <Select value={newProfile} onValueChange={setNewProfile}>
                <SelectTrigger>
                  <SelectValue placeholder={t('settings.selectProfile')} />
                </SelectTrigger>
                <SelectContent>
                  {profiles.map((p) => (
                    <SelectItem key={p} value={p}>{p}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label>{t('settings.client')}</Label>
              <Select value={newClient} onValueChange={setNewClient}>
                <SelectTrigger>
                  <SelectValue placeholder={t('settings.selectClient')} />
                </SelectTrigger>
                <SelectContent>
                  {clients.map((c) => (
                    <SelectItem key={c} value={c}>{c}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setAddDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleAdd} disabled={addBinding.isPending}>
              {t('common.add')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Binding Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('common.delete')}</DialogTitle>
            <DialogDescription>{t('settings.deleteBindingConfirm')}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteBinding.isPending}>
              {t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

// __CONTINUE_SYNC__

function SyncSettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { data: syncConfig } = useSyncConfig()
  const { data: syncStatus } = useSyncStatus()
  const syncPull = useSyncPull()
  const syncPush = useSyncPush()

  const [enabled, setEnabled] = useState(false)
  const [backend, setBackend] = useState('')
  // GitHub Gist & Repo
  const [gistId, setGistId] = useState('')
  const [token, setToken] = useState('')
  const [repoOwner, setRepoOwner] = useState('')
  const [repoName, setRepoName] = useState('')
  const [repoPath, setRepoPath] = useState('')
  const [repoBranch, setRepoBranch] = useState('')
  // S3
  const [s3Endpoint, setS3Endpoint] = useState('')
  const [s3Bucket, setS3Bucket] = useState('')
  const [s3Region, setS3Region] = useState('')
  const [s3AccessKey, setS3AccessKey] = useState('')
  const [s3SecretKey, setS3SecretKey] = useState('')
  // WebDAV
  const [webdavEndpoint, setWebdavEndpoint] = useState('')
  const [webdavUsername, setWebdavUsername] = useState('')
  const [webdavPassword, setWebdavPassword] = useState('')
  // Common
  const [autoPull, setAutoPull] = useState(false)
  const [pullInterval, setPullInterval] = useState(300)

  // Initialize form when data loads
  useState(() => {
    if (syncConfig) {
      setEnabled(syncConfig.enabled || false)
      setBackend(syncConfig.backend || '')
      setGistId(syncConfig.gist_id || '')
      setRepoOwner(syncConfig.repo_owner || '')
      setRepoName(syncConfig.repo_name || '')
      setRepoPath(syncConfig.repo_path || '')
      setRepoBranch(syncConfig.repo_branch || '')
      setS3Endpoint(syncConfig.endpoint || '')
      setS3Bucket(syncConfig.bucket || '')
      setS3Region(syncConfig.region || '')
      setS3AccessKey(syncConfig.access_key || '')
      setWebdavEndpoint(syncConfig.endpoint || '')
      setWebdavUsername(syncConfig.username || '')
      setAutoPull(syncConfig.auto_pull || false)
      setPullInterval(syncConfig.pull_interval || 300)
    }
  })

  const updateSync = useMutation({
    mutationFn: async (data: any) => {
      const res = await fetch('/api/v1/sync/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      })
      if (!res.ok) throw new Error('Failed to update sync config')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sync', 'config'] })
      toast.success(t('common.success'))
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    },
  })

  const handleSave = () => {
    const data: any = {
      enabled,
      backend: backend || undefined,
      auto_pull: autoPull,
      pull_interval: pullInterval,
    }
    if (backend === 'gist') {
      data.gist_id = gistId || undefined
      data.token = token || undefined
    } else if (backend === 'repo') {
      data.repo_owner = repoOwner || undefined
      data.repo_name = repoName || undefined
      data.repo_path = repoPath || undefined
      data.repo_branch = repoBranch || undefined
      data.token = token || undefined
    } else if (backend === 's3') {
      data.endpoint = s3Endpoint || undefined
      data.bucket = s3Bucket || undefined
      data.region = s3Region || undefined
      data.access_key = s3AccessKey || undefined
      data.secret_key = s3SecretKey || undefined
    } else if (backend === 'webdav') {
      data.endpoint = webdavEndpoint || undefined
      data.username = webdavUsername || undefined
      data.token = webdavPassword || undefined
    }
    updateSync.mutate(data)
  }

  const handlePull = async () => {
    try {
      await syncPull.mutateAsync()
      toast.success(t('common.success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const handlePush = async () => {
    try {
      await syncPush.mutateAsync()
      toast.success(t('common.success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <RefreshCw className="h-5 w-5" />
          {t('settings.sync')}
        </CardTitle>
        <CardDescription>{t('settings.syncDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-2">
          <Label htmlFor="sync-enabled">{t('settings.syncEnabled')}</Label>
          <Switch
            id="sync-enabled"
            checked={enabled}
            onCheckedChange={setEnabled}
          />
        </div>

        {enabled && (
          <>
            <div className="grid gap-2">
              <Label>{t('settings.syncBackend')}</Label>
              <Select value={backend} onValueChange={setBackend}>
                <SelectTrigger>
                  <SelectValue placeholder={t('settings.selectBackend')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="gist">GitHub Gist</SelectItem>
                  <SelectItem value="repo">GitHub Repo</SelectItem>
                  <SelectItem value="s3">Amazon S3</SelectItem>
                  <SelectItem value="webdav">WebDAV</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {backend === 'gist' && (
              <>
                <div className="grid gap-2">
                  <Label>{t('settings.gistId')}</Label>
                  <Input
                    value={gistId}
                    onChange={(e) => setGistId(e.target.value)}
                    placeholder={t('settings.gistIdPlaceholder')}
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.githubToken')}</Label>
                  <Input
                    type="password"
                    value={token}
                    onChange={(e) => setToken(e.target.value)}
                    placeholder="GitHub Personal Access Token"
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('settings.githubTokenHintGist')}{' '}
                    <a
                      href="https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      {t('settings.githubTokenLink')}
                    </a>
                  </p>
                </div>
              </>
            )}

            {backend === 'repo' && (
              <>
                <div className="grid gap-2">
                  <Label>{t('settings.repoOwner')}</Label>
                  <Input
                    value={repoOwner}
                    onChange={(e) => setRepoOwner(e.target.value)}
                    placeholder="username or org"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.repoName')}</Label>
                  <Input
                    value={repoName}
                    onChange={(e) => setRepoName(e.target.value)}
                    placeholder="repository-name"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.repoPath')}</Label>
                  <Input
                    value={repoPath}
                    onChange={(e) => setRepoPath(e.target.value)}
                    placeholder="zen-sync.json"
                  />
                  <p className="text-xs text-muted-foreground">{t('settings.repoPathHint')}</p>
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.repoBranch')}</Label>
                  <Input
                    value={repoBranch}
                    onChange={(e) => setRepoBranch(e.target.value)}
                    placeholder="main"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.githubToken')}</Label>
                  <Input
                    type="password"
                    value={token}
                    onChange={(e) => setToken(e.target.value)}
                    placeholder="GitHub Personal Access Token"
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('settings.githubTokenHintRepo')}{' '}
                    <a
                      href="https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      {t('settings.githubTokenLink')}
                    </a>
                  </p>
                </div>
              </>
            )}

            {backend === 's3' && (
              <>
                <div className="grid gap-2">
                  <Label>{t('settings.s3Endpoint')}</Label>
                  <Input
                    value={s3Endpoint}
                    onChange={(e) => setS3Endpoint(e.target.value)}
                    placeholder="https://s3.amazonaws.com"
                  />
                  <p className="text-xs text-muted-foreground">{t('settings.s3EndpointHint')}</p>
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.s3Bucket')}</Label>
                  <Input
                    value={s3Bucket}
                    onChange={(e) => setS3Bucket(e.target.value)}
                    placeholder="my-bucket"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.s3Region')}</Label>
                  <Input
                    value={s3Region}
                    onChange={(e) => setS3Region(e.target.value)}
                    placeholder="us-east-1"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.s3AccessKey')}</Label>
                  <Input
                    value={s3AccessKey}
                    onChange={(e) => setS3AccessKey(e.target.value)}
                    placeholder="AKIAIOSFODNN7EXAMPLE"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.s3SecretKey')}</Label>
                  <Input
                    type="password"
                    value={s3SecretKey}
                    onChange={(e) => setS3SecretKey(e.target.value)}
                    placeholder="Secret Access Key"
                  />
                </div>
              </>
            )}

            {backend === 'webdav' && (
              <>
                <div className="grid gap-2">
                  <Label>{t('settings.webdavEndpoint')}</Label>
                  <Input
                    value={webdavEndpoint}
                    onChange={(e) => setWebdavEndpoint(e.target.value)}
                    placeholder="https://dav.example.com/path"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.webdavUsername')}</Label>
                  <Input
                    value={webdavUsername}
                    onChange={(e) => setWebdavUsername(e.target.value)}
                    placeholder="username"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('settings.webdavPassword')}</Label>
                  <Input
                    type="password"
                    value={webdavPassword}
                    onChange={(e) => setWebdavPassword(e.target.value)}
                    placeholder="password"
                  />
                </div>
              </>
            )}

            <div className="grid gap-2">
              <Label htmlFor="auto-pull">{t('settings.autoPull')}</Label>
              <Switch
                id="auto-pull"
                checked={autoPull}
                onCheckedChange={setAutoPull}
              />
            </div>

            {autoPull && (
              <div className="grid gap-2">
                <Label>{t('settings.pullInterval')}</Label>
                <Input
                  type="number"
                  value={pullInterval}
                  onChange={(e) => setPullInterval(parseInt(e.target.value) || 300)}
                  min={60}
                />
                <p className="text-xs text-muted-foreground">{t('settings.pullIntervalHint')}</p>
              </div>
            )}
          </>
        )}

        {syncStatus?.last_sync && (
          <div className="text-sm text-muted-foreground">
            {t('settings.lastSync')}: {new Date(syncStatus.last_sync).toLocaleString()}
          </div>
        )}

        {syncStatus?.last_error && (
          <div className="text-sm text-destructive">{syncStatus.last_error}</div>
        )}

        <div className="flex gap-2">
          <Button onClick={handleSave} disabled={updateSync.isPending}>
            {t('common.save')}
          </Button>
          {syncConfig?.configured && (
            <>
              <Button
                variant="outline"
                onClick={handlePull}
                disabled={syncPull.isPending}
              >
                Pull
              </Button>
              <Button
                variant="outline"
                onClick={handlePush}
                disabled={syncPush.isPending}
              >
                Push
              </Button>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

function PasswordSettings() {
  const { t } = useTranslation()
  const changePassword = useChangePassword()
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (newPassword !== confirmPassword) {
      toast.error(t('settings.passwordMismatch'))
      return
    }
    try {
      await changePassword.mutateAsync({ currentPassword, newPassword })
      toast.success(t('settings.passwordChanged'))
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lock className="h-5 w-5" />
          {t('settings.webPassword')}
        </CardTitle>
        <CardDescription>{t('settings.webPasswordDesc')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-2">
            <Label htmlFor="current-password">{t('settings.currentPassword')}</Label>
            <Input
              id="current-password"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="new-password">{t('settings.newPassword')}</Label>
            <Input
              id="new-password"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="confirm-password">{t('settings.confirmPassword')}</Label>
            <Input
              id="confirm-password"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
            />
          </div>
          <Button type="submit" disabled={changePassword.isPending}>
            {t('settings.changePassword')}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
