import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Settings as SettingsIcon, FolderOpen, Lock, RefreshCw, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
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
          <TabsTrigger value="password">{t('settings.password')}</TabsTrigger>
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

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <SettingsIcon className="h-5 w-5" />
          {t('settings.general')}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-muted-foreground">
          General settings are configured via the CLI or config file.
        </p>
      </CardContent>
    </Card>
  )
}

function BindingsSettings() {
  const { t } = useTranslation()
  const { data: bindings, isLoading } = useBindings()
  const deleteBinding = useDeleteBinding()
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deletingPath, setDeletingPath] = useState<string | null>(null)

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

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FolderOpen className="h-5 w-5" />
            {t('settings.bindings')}
          </CardTitle>
          <CardDescription>Project-specific profile bindings</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center py-4">{t('common.loading')}</div>
          ) : bindings && bindings.length > 0 ? (
            <div className="space-y-2">
              {bindings.map((binding) => (
                <div key={binding.path} className="flex items-center justify-between rounded-lg border p-3">
                  <div>
                    <p className="font-medium">{binding.path}</p>
                    <p className="text-sm text-muted-foreground">Profile: {binding.profile}</p>
                  </div>
                  <Button variant="ghost" size="icon" onClick={() => handleOpenDelete(binding.path)}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground">{t('settings.noBindings')}</p>
          )}
        </CardContent>
      </Card>

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('common.delete')}</DialogTitle>
            <DialogDescription>Are you sure you want to delete this binding?</DialogDescription>
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

function SyncSettings() {
  const { t } = useTranslation()
  const { data: syncConfig } = useSyncConfig()
  const { data: syncStatus } = useSyncStatus()
  const syncPull = useSyncPull()
  const syncPush = useSyncPush()

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
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <Label>{t('settings.syncEnabled')}</Label>
          <Switch checked={syncConfig?.enabled ?? false} disabled />
        </div>

        {syncStatus?.last_sync && (
          <div className="text-sm text-muted-foreground">
            {t('settings.lastSync')}: {new Date(syncStatus.last_sync).toLocaleString()}
          </div>
        )}

        {syncStatus?.last_error && (
          <div className="text-sm text-destructive">{syncStatus.last_error}</div>
        )}

        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handlePull}
            disabled={!syncConfig?.enabled || syncPull.isPending}
          >
            Pull
          </Button>
          <Button
            variant="outline"
            onClick={handlePush}
            disabled={!syncConfig?.enabled || syncPush.isPending}
          >
            Push
          </Button>
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
          {t('settings.changePassword')}
        </CardTitle>
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
