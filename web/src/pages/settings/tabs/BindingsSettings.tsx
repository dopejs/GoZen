import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { FolderOpen, Trash2, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useBindings, useDeleteBinding } from '@/hooks/use-settings'
import { bindingsApi } from '@/lib/api'

export function BindingsSettings() {
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

      <Dialog open={addDialogOpen} onOpenChange={setAddDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('settings.addBinding')}</DialogTitle>
            <DialogDescription>{t('settings.addBindingDesc')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="grid gap-2">
              <Label>{t('settings.projectPath')}</Label>
              <Input value={newPath} onChange={(e) => setNewPath(e.target.value)} placeholder="/path/to/project" />
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
            <Button variant="outline" onClick={() => setAddDialogOpen(false)}>{t('common.cancel')}</Button>
            <Button onClick={handleAdd} disabled={addBinding.isPending}>{t('common.add')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('common.delete')}</DialogTitle>
            <DialogDescription>{t('settings.deleteBindingConfirm')}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>{t('common.cancel')}</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteBinding.isPending}>{t('common.delete')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
