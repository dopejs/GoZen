import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { RefreshCw, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useSyncConfig, useSyncStatus, useSyncPull, useSyncPush } from '@/hooks/use-settings'
import type { SyncConfig } from '@/types/api'

export function SyncSettings() {
  const { data: syncConfig, isLoading } = useSyncConfig()

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  // Key forces remount (and fresh useState) when config changes after save
  const key = syncConfig ? JSON.stringify(syncConfig) : 'empty'
  return <SyncForm key={key} initialConfig={syncConfig} />
}

function SyncForm({ initialConfig }: { initialConfig?: SyncConfig }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { data: syncStatus } = useSyncStatus()
  const pullMutation = useSyncPull()
  const pushMutation = useSyncPush()

  const [enabled, setEnabled] = useState(initialConfig?.enabled ?? false)
  const [backend, setBackend] = useState(initialConfig?.backend ?? '')
  const [gistId, setGistId] = useState(initialConfig?.gist_id ?? '')
  const [token, setToken] = useState(initialConfig?.token ?? '')
  const [repoOwner, setRepoOwner] = useState(initialConfig?.repo_owner ?? '')
  const [repoName, setRepoName] = useState(initialConfig?.repo_name ?? '')
  const [repoPath, setRepoPath] = useState(initialConfig?.repo_path ?? '')
  const [repoBranch, setRepoBranch] = useState(initialConfig?.repo_branch ?? '')
  const [s3Endpoint, setS3Endpoint] = useState(initialConfig?.endpoint ?? '')
  const [s3Bucket, setS3Bucket] = useState(initialConfig?.bucket ?? '')
  const [s3Region, setS3Region] = useState(initialConfig?.region ?? '')
  const [s3AccessKey, setS3AccessKey] = useState(initialConfig?.access_key ?? '')
  const [s3SecretKey, setS3SecretKey] = useState(initialConfig?.secret_key ?? '')
  const [webdavEndpoint, setWebdavEndpoint] = useState(initialConfig?.endpoint ?? '')
  const [webdavUsername, setWebdavUsername] = useState(initialConfig?.username ?? '')
  const [webdavPassword, setWebdavPassword] = useState(initialConfig?.token ?? '')
  const [passphrase, setPassphrase] = useState(initialConfig?.passphrase ?? '')
  const [autoPull, setAutoPull] = useState(initialConfig?.auto_pull ?? false)
  const [pullInterval, setPullInterval] = useState(initialConfig?.pull_interval ?? 300)

  const saveMutation = useMutation({
    mutationFn: async (config: Partial<SyncConfig>) => {
      const res = await fetch('/api/v1/sync/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
        credentials: 'include',
      })
      if (!res.ok) throw new Error(await res.text())
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sync'] })
      toast.success(t('settings.sync.saved', 'Sync configuration saved'))
    },
    onError: (err: Error) => {
      toast.error(err.message)
    },
  })

  const handleSave = () => {
    const config: Record<string, unknown> = {
      enabled,
      backend,
      auto_pull: autoPull,
      pull_interval: pullInterval,
      passphrase,
    }
    if (backend === 'gist') {
      config.gist_id = gistId
      config.token = token
    } else if (backend === 'repo') {
      config.repo_owner = repoOwner
      config.repo_name = repoName
      config.repo_path = repoPath
      config.repo_branch = repoBranch
      config.token = token
    } else if (backend === 's3') {
      config.endpoint = s3Endpoint
      config.bucket = s3Bucket
      config.region = s3Region
      config.access_key = s3AccessKey
      config.secret_key = s3SecretKey
    } else if (backend === 'webdav') {
      config.endpoint = webdavEndpoint
      config.username = webdavUsername
      config.token = webdavPassword
    }
    saveMutation.mutate(config as Partial<SyncConfig>)
  }

  const handlePull = () => {
    pullMutation.mutate(undefined, {
      onSuccess: () => toast.success(t('settings.sync.pullSuccess', 'Pull completed')),
      onError: (err: Error) => toast.error(err.message),
    })
  }

  const handlePush = () => {
    pushMutation.mutate(undefined, {
      onSuccess: () => toast.success(t('settings.sync.pushSuccess', 'Push completed')),
      onError: (err: Error) => toast.error(err.message),
    })
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{t('settings.sync.title', 'Config Sync')}</CardTitle>
              <CardDescription>{t('settings.sync.description', 'Sync your configuration across devices')}</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Label htmlFor="sync-enabled">{t('settings.sync.enabled', 'Enabled')}</Label>
              <Switch id="sync-enabled" checked={enabled} onCheckedChange={setEnabled} />
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label>{t('settings.sync.backend', 'Backend')}</Label>
            <Select value={backend} onValueChange={setBackend}>
              <SelectTrigger><SelectValue placeholder={t('settings.sync.selectBackend', 'Select backend')} /></SelectTrigger>
              <SelectContent>
                <SelectItem value="gist">GitHub Gist</SelectItem>
                <SelectItem value="repo">GitHub Repo</SelectItem>
                <SelectItem value="s3">S3</SelectItem>
                <SelectItem value="webdav">WebDAV</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {backend === 'gist' && <GistFields gistId={gistId} setGistId={setGistId} token={token} setToken={setToken} />}
          {backend === 'repo' && <RepoFields repoOwner={repoOwner} setRepoOwner={setRepoOwner} repoName={repoName} setRepoName={setRepoName} repoPath={repoPath} setRepoPath={setRepoPath} repoBranch={repoBranch} setRepoBranch={setRepoBranch} token={token} setToken={setToken} />}
          {backend === 's3' && <S3Fields endpoint={s3Endpoint} setEndpoint={setS3Endpoint} bucket={s3Bucket} setBucket={setS3Bucket} region={s3Region} setRegion={setS3Region} accessKey={s3AccessKey} setAccessKey={setS3AccessKey} secretKey={s3SecretKey} setSecretKey={setS3SecretKey} />}
          {backend === 'webdav' && <WebDAVFields endpoint={webdavEndpoint} setEndpoint={setWebdavEndpoint} username={webdavUsername} setUsername={setWebdavUsername} password={webdavPassword} setPassword={setWebdavPassword} />}

          {backend && (
            <>
              <div className="space-y-2">
                <Label>{t('settings.sync.passphrase', 'Encryption Passphrase')}</Label>
                <Input type="password" value={passphrase} onChange={e => setPassphrase(e.target.value)} placeholder={t('settings.sync.passphrasePlaceholder', 'Optional encryption passphrase')} />
              </div>
              <div className="flex items-center gap-2">
                <Switch id="auto-pull" checked={autoPull} onCheckedChange={setAutoPull} />
                <Label htmlFor="auto-pull">{t('settings.sync.autoPull', 'Auto Pull')}</Label>
              </div>
              {autoPull && (
                <div className="space-y-2">
                  <Label>{t('settings.sync.pullInterval', 'Pull Interval (seconds)')}</Label>
                  <Input type="number" value={pullInterval} onChange={e => setPullInterval(Number(e.target.value))} min={60} />
                </div>
              )}
            </>
          )}

          <div className="flex gap-2 pt-2">
            <Button onClick={handleSave} disabled={saveMutation.isPending || !backend}>
              {saveMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              {t('common.save', 'Save')}
            </Button>
            {initialConfig?.configured && (
              <>
                <Button variant="outline" onClick={handlePull} disabled={pullMutation.isPending}>
                  {pullMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCw className="mr-2 h-4 w-4" />}
                  {t('settings.sync.pull', 'Pull')}
                </Button>
                <Button variant="outline" onClick={handlePush} disabled={pushMutation.isPending}>
                  {pushMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCw className="mr-2 h-4 w-4" />}
                  {t('settings.sync.push', 'Push')}
                </Button>
              </>
            )}
          </div>

          {syncStatus && initialConfig?.configured && (
            <div className="text-sm text-muted-foreground pt-2">
              {syncStatus.last_sync && <p>{t('settings.sync.lastSync', 'Last sync')}: {new Date(syncStatus.last_sync).toLocaleString()}</p>}
              {syncStatus.last_error && <p className="text-destructive">{t('settings.sync.lastError', 'Last error')}: {syncStatus.last_error}</p>}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function GistFields({ gistId, setGistId, token, setToken }: {
  gistId: string; setGistId: (v: string) => void
  token: string; setToken: (v: string) => void
}) {
  const { t } = useTranslation()
  return (
    <>
      <div className="space-y-2">
        <Label>{t('settings.sync.gistId', 'Gist ID')}</Label>
        <Input value={gistId} onChange={e => setGistId(e.target.value)} placeholder="abc123def456..." />
      </div>
      <div className="space-y-2">
        <Label>{t('settings.sync.token', 'GitHub Token')}</Label>
        <Input value={token} onChange={e => setToken(e.target.value)} placeholder="ghp_..." />
      </div>
    </>
  )
}

function RepoFields({ repoOwner, setRepoOwner, repoName, setRepoName, repoPath, setRepoPath, repoBranch, setRepoBranch, token, setToken }: {
  repoOwner: string; setRepoOwner: (v: string) => void
  repoName: string; setRepoName: (v: string) => void
  repoPath: string; setRepoPath: (v: string) => void
  repoBranch: string; setRepoBranch: (v: string) => void
  token: string; setToken: (v: string) => void
}) {
  const { t } = useTranslation()
  return (
    <>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>{t('settings.sync.repoOwner', 'Owner')}</Label>
          <Input value={repoOwner} onChange={e => setRepoOwner(e.target.value)} placeholder="username" />
        </div>
        <div className="space-y-2">
          <Label>{t('settings.sync.repoName', 'Repository')}</Label>
          <Input value={repoName} onChange={e => setRepoName(e.target.value)} placeholder="my-config" />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>{t('settings.sync.repoPath', 'File Path')}</Label>
          <Input value={repoPath} onChange={e => setRepoPath(e.target.value)} placeholder="zen-sync.json" />
        </div>
        <div className="space-y-2">
          <Label>{t('settings.sync.repoBranch', 'Branch')}</Label>
          <Input value={repoBranch} onChange={e => setRepoBranch(e.target.value)} placeholder="main" />
        </div>
      </div>
      <div className="space-y-2">
        <Label>{t('settings.sync.token', 'GitHub Token')}</Label>
        <Input value={token} onChange={e => setToken(e.target.value)} placeholder="ghp_..." />
      </div>
    </>
  )
}

function S3Fields({ endpoint, setEndpoint, bucket, setBucket, region, setRegion, accessKey, setAccessKey, secretKey, setSecretKey }: {
  endpoint: string; setEndpoint: (v: string) => void
  bucket: string; setBucket: (v: string) => void
  region: string; setRegion: (v: string) => void
  accessKey: string; setAccessKey: (v: string) => void
  secretKey: string; setSecretKey: (v: string) => void
}) {
  const { t } = useTranslation()
  return (
    <>
      <div className="space-y-2">
        <Label>{t('settings.sync.s3Endpoint', 'Endpoint')}</Label>
        <Input value={endpoint} onChange={e => setEndpoint(e.target.value)} placeholder="https://s3.amazonaws.com" />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>{t('settings.sync.s3Bucket', 'Bucket')}</Label>
          <Input value={bucket} onChange={e => setBucket(e.target.value)} placeholder="my-bucket" />
        </div>
        <div className="space-y-2">
          <Label>{t('settings.sync.s3Region', 'Region')}</Label>
          <Input value={region} onChange={e => setRegion(e.target.value)} placeholder="us-east-1" />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>{t('settings.sync.s3AccessKey', 'Access Key')}</Label>
          <Input value={accessKey} onChange={e => setAccessKey(e.target.value)} placeholder="AKIA..." />
        </div>
        <div className="space-y-2">
          <Label>{t('settings.sync.s3SecretKey', 'Secret Key')}</Label>
          <Input value={secretKey} onChange={e => setSecretKey(e.target.value)} placeholder="..." />
        </div>
      </div>
    </>
  )
}

function WebDAVFields({ endpoint, setEndpoint, username, setUsername, password, setPassword }: {
  endpoint: string; setEndpoint: (v: string) => void
  username: string; setUsername: (v: string) => void
  password: string; setPassword: (v: string) => void
}) {
  const { t } = useTranslation()
  return (
    <>
      <div className="space-y-2">
        <Label>{t('settings.sync.webdavEndpoint', 'WebDAV URL')}</Label>
        <Input value={endpoint} onChange={e => setEndpoint(e.target.value)} placeholder="https://dav.example.com/zen-sync.json" />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>{t('settings.sync.webdavUsername', 'Username')}</Label>
          <Input value={username} onChange={e => setUsername(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>{t('settings.sync.webdavPassword', 'Password')}</Label>
          <Input value={password} onChange={e => setPassword(e.target.value)} />
        </div>
      </div>
    </>
  )
}
