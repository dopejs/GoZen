import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useSyncConfig, useSyncStatus, useSyncPull, useSyncPush } from '@/hooks/use-settings'

export function SyncSettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { data: syncConfig } = useSyncConfig()
  const { data: syncStatus } = useSyncStatus()
  const syncPull = useSyncPull()
  const syncPush = useSyncPush()

  const [enabled, setEnabled] = useState(false)
  const [backend, setBackend] = useState('')
  const [gistId, setGistId] = useState('')
  const [token, setToken] = useState('')
  const [repoOwner, setRepoOwner] = useState('')
  const [repoName, setRepoName] = useState('')
  const [repoPath, setRepoPath] = useState('')
  const [repoBranch, setRepoBranch] = useState('')
  const [s3Endpoint, setS3Endpoint] = useState('')
  const [s3Bucket, setS3Bucket] = useState('')
  const [s3Region, setS3Region] = useState('')
  const [s3AccessKey, setS3AccessKey] = useState('')
  const [s3SecretKey, setS3SecretKey] = useState('')
  const [webdavEndpoint, setWebdavEndpoint] = useState('')
  const [webdavUsername, setWebdavUsername] = useState('')
  const [webdavPassword, setWebdavPassword] = useState('')
  const [autoPull, setAutoPull] = useState(false)
  const [pullInterval, setPullInterval] = useState(300)

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
    mutationFn: async (data: Record<string, unknown>) => {
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
    const data: Record<string, unknown> = { enabled, backend: backend || undefined, auto_pull: autoPull, pull_interval: pullInterval }
    if (backend === 'gist') {
      data.gist_id = gistId || undefined
      data.token = token || undefined
    } else if (backend === 'repo') {
      Object.assign(data, { repo_owner: repoOwner || undefined, repo_name: repoName || undefined, repo_path: repoPath || undefined, repo_branch: repoBranch || undefined, token: token || undefined })
    } else if (backend === 's3') {
      Object.assign(data, { endpoint: s3Endpoint || undefined, bucket: s3Bucket || undefined, region: s3Region || undefined, access_key: s3AccessKey || undefined, secret_key: s3SecretKey || undefined })
    } else if (backend === 'webdav') {
      Object.assign(data, { endpoint: webdavEndpoint || undefined, username: webdavUsername || undefined, token: webdavPassword || undefined })
    }
    updateSync.mutate(data)
  }

  const handlePull = async () => {
    try { await syncPull.mutateAsync(); toast.success(t('common.success')) }
    catch (err) { toast.error(err instanceof Error ? err.message : t('common.error')) }
  }

  const handlePush = async () => {
    try { await syncPush.mutateAsync(); toast.success(t('common.success')) }
    catch (err) { toast.error(err instanceof Error ? err.message : t('common.error')) }
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
          <Switch id="sync-enabled" checked={enabled} onCheckedChange={setEnabled} />
        </div>

        {enabled && (
          <>
            <div className="grid gap-2">
              <Label>{t('settings.syncBackend')}</Label>
              <Select value={backend} onValueChange={setBackend}>
                <SelectTrigger><SelectValue placeholder={t('settings.selectBackend')} /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="gist">GitHub Gist</SelectItem>
                  <SelectItem value="repo">GitHub Repo</SelectItem>
                  <SelectItem value="s3">Amazon S3</SelectItem>
                  <SelectItem value="webdav">WebDAV</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {backend === 'gist' && <GistFields gistId={gistId} setGistId={setGistId} token={token} setToken={setToken} t={t} />}
            {backend === 'repo' && <RepoFields repoOwner={repoOwner} setRepoOwner={setRepoOwner} repoName={repoName} setRepoName={setRepoName} repoPath={repoPath} setRepoPath={setRepoPath} repoBranch={repoBranch} setRepoBranch={setRepoBranch} token={token} setToken={setToken} t={t} />}
            {backend === 's3' && <S3Fields s3Endpoint={s3Endpoint} setS3Endpoint={setS3Endpoint} s3Bucket={s3Bucket} setS3Bucket={setS3Bucket} s3Region={s3Region} setS3Region={setS3Region} s3AccessKey={s3AccessKey} setS3AccessKey={setS3AccessKey} s3SecretKey={s3SecretKey} setS3SecretKey={setS3SecretKey} t={t} />}
            {backend === 'webdav' && <WebDAVFields webdavEndpoint={webdavEndpoint} setWebdavEndpoint={setWebdavEndpoint} webdavUsername={webdavUsername} setWebdavUsername={setWebdavUsername} webdavPassword={webdavPassword} setWebdavPassword={setWebdavPassword} t={t} />}

            <div className="grid gap-2">
              <Label htmlFor="auto-pull">{t('settings.autoPull')}</Label>
              <Switch id="auto-pull" checked={autoPull} onCheckedChange={setAutoPull} />
            </div>
            {autoPull && (
              <div className="grid gap-2">
                <Label>{t('settings.pullInterval')}</Label>
                <Input type="number" value={pullInterval} onChange={(e) => setPullInterval(parseInt(e.target.value) || 300)} min={60} />
                <p className="text-xs text-muted-foreground">{t('settings.pullIntervalHint')}</p>
              </div>
            )}
          </>
        )}

        {syncStatus?.last_sync && <div className="text-sm text-muted-foreground">{t('settings.lastSync')}: {new Date(syncStatus.last_sync).toLocaleString()}</div>}
        {syncStatus?.last_error && <div className="text-sm text-destructive">{syncStatus.last_error}</div>}

        <div className="flex gap-2">
          <Button onClick={handleSave} disabled={updateSync.isPending}>{t('common.save')}</Button>
          {syncConfig?.configured && (
            <>
              <Button variant="outline" onClick={handlePull} disabled={syncPull.isPending}>Pull</Button>
              <Button variant="outline" onClick={handlePush} disabled={syncPush.isPending}>Push</Button>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

// Sub-components for backend fields
function GistFields({ gistId, setGistId, token, setToken, t }: { gistId: string; setGistId: (v: string) => void; token: string; setToken: (v: string) => void; t: (k: string) => string }) {
  return (
    <>
      <div className="grid gap-2">
        <Label>{t('settings.gistId')}</Label>
        <Input value={gistId} onChange={(e) => setGistId(e.target.value)} placeholder={t('settings.gistIdPlaceholder')} />
      </div>
      <div className="grid gap-2">
        <Label>{t('settings.githubToken')}</Label>
        <Input type="password" value={token} onChange={(e) => setToken(e.target.value)} placeholder="GitHub Personal Access Token" />
        <p className="text-xs text-muted-foreground">{t('settings.githubTokenHintGist')} <a href="https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">{t('settings.githubTokenLink')}</a></p>
      </div>
    </>
  )
}

function RepoFields({ repoOwner, setRepoOwner, repoName, setRepoName, repoPath, setRepoPath, repoBranch, setRepoBranch, token, setToken, t }: { repoOwner: string; setRepoOwner: (v: string) => void; repoName: string; setRepoName: (v: string) => void; repoPath: string; setRepoPath: (v: string) => void; repoBranch: string; setRepoBranch: (v: string) => void; token: string; setToken: (v: string) => void; t: (k: string) => string }) {
  return (
    <>
      <div className="grid gap-2"><Label>{t('settings.repoOwner')}</Label><Input value={repoOwner} onChange={(e) => setRepoOwner(e.target.value)} placeholder="username or org" /></div>
      <div className="grid gap-2"><Label>{t('settings.repoName')}</Label><Input value={repoName} onChange={(e) => setRepoName(e.target.value)} placeholder="repository-name" /></div>
      <div className="grid gap-2"><Label>{t('settings.repoPath')}</Label><Input value={repoPath} onChange={(e) => setRepoPath(e.target.value)} placeholder="zen-sync.json" /><p className="text-xs text-muted-foreground">{t('settings.repoPathHint')}</p></div>
      <div className="grid gap-2"><Label>{t('settings.repoBranch')}</Label><Input value={repoBranch} onChange={(e) => setRepoBranch(e.target.value)} placeholder="main" /></div>
      <div className="grid gap-2"><Label>{t('settings.githubToken')}</Label><Input type="password" value={token} onChange={(e) => setToken(e.target.value)} placeholder="GitHub Personal Access Token" /><p className="text-xs text-muted-foreground">{t('settings.githubTokenHintRepo')} <a href="https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">{t('settings.githubTokenLink')}</a></p></div>
    </>
  )
}

function S3Fields({ s3Endpoint, setS3Endpoint, s3Bucket, setS3Bucket, s3Region, setS3Region, s3AccessKey, setS3AccessKey, s3SecretKey, setS3SecretKey, t }: { s3Endpoint: string; setS3Endpoint: (v: string) => void; s3Bucket: string; setS3Bucket: (v: string) => void; s3Region: string; setS3Region: (v: string) => void; s3AccessKey: string; setS3AccessKey: (v: string) => void; s3SecretKey: string; setS3SecretKey: (v: string) => void; t: (k: string) => string }) {
  return (
    <>
      <div className="grid gap-2"><Label>{t('settings.s3Endpoint')}</Label><Input value={s3Endpoint} onChange={(e) => setS3Endpoint(e.target.value)} placeholder="https://s3.amazonaws.com" /><p className="text-xs text-muted-foreground">{t('settings.s3EndpointHint')}</p></div>
      <div className="grid gap-2"><Label>{t('settings.s3Bucket')}</Label><Input value={s3Bucket} onChange={(e) => setS3Bucket(e.target.value)} placeholder="my-bucket" /></div>
      <div className="grid gap-2"><Label>{t('settings.s3Region')}</Label><Input value={s3Region} onChange={(e) => setS3Region(e.target.value)} placeholder="us-east-1" /></div>
      <div className="grid gap-2"><Label>{t('settings.s3AccessKey')}</Label><Input value={s3AccessKey} onChange={(e) => setS3AccessKey(e.target.value)} placeholder="AKIAIOSFODNN7EXAMPLE" /></div>
      <div className="grid gap-2"><Label>{t('settings.s3SecretKey')}</Label><Input type="password" value={s3SecretKey} onChange={(e) => setS3SecretKey(e.target.value)} placeholder="Secret Access Key" /></div>
    </>
  )
}

function WebDAVFields({ webdavEndpoint, setWebdavEndpoint, webdavUsername, setWebdavUsername, webdavPassword, setWebdavPassword, t }: { webdavEndpoint: string; setWebdavEndpoint: (v: string) => void; webdavUsername: string; setWebdavUsername: (v: string) => void; webdavPassword: string; setWebdavPassword: (v: string) => void; t: (k: string) => string }) {
  return (
    <>
      <div className="grid gap-2"><Label>{t('settings.webdavEndpoint')}</Label><Input value={webdavEndpoint} onChange={(e) => setWebdavEndpoint(e.target.value)} placeholder="https://dav.example.com/path" /></div>
      <div className="grid gap-2"><Label>{t('settings.webdavUsername')}</Label><Input value={webdavUsername} onChange={(e) => setWebdavUsername(e.target.value)} placeholder="username" /></div>
      <div className="grid gap-2"><Label>{t('settings.webdavPassword')}</Label><Input type="password" value={webdavPassword} onChange={(e) => setWebdavPassword(e.target.value)} placeholder="password" /></div>
    </>
  )
}
