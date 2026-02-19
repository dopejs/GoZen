import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Plus, Pencil, Trash2, Server } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { useProviders, useCreateProvider, useUpdateProvider, useDeleteProvider } from '@/hooks/use-providers'
import type { Provider } from '@/types/api'

export function ProvidersPage() {
  const { t } = useTranslation()
  const { data: providers, isLoading } = useProviders()
  const createProvider = useCreateProvider()
  const updateProvider = useUpdateProvider()
  const deleteProvider = useDeleteProvider()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingProvider, setEditingProvider] = useState<Provider | null>(null)
  const [deletingProvider, setDeletingProvider] = useState<string | null>(null)
  const [formData, setFormData] = useState<Partial<Provider>>({
    name: '',
    base_url: 'https://api.anthropic.com',
    api_key: '',
    api_key_env: '',
    model: '',
    priority: 1,
    weight: 1,
    enabled: true,
  })

  const handleOpenCreate = () => {
    setEditingProvider(null)
    setFormData({
      name: '',
      base_url: 'https://api.anthropic.com',
      api_key: '',
      api_key_env: '',
      model: '',
      priority: 1,
      weight: 1,
      enabled: true,
    })
    setDialogOpen(true)
  }

  const handleOpenEdit = (provider: Provider) => {
    setEditingProvider(provider)
    setFormData({ ...provider })
    setDialogOpen(true)
  }

  const handleOpenDelete = (name: string) => {
    setDeletingProvider(name)
    setDeleteDialogOpen(true)
  }

  const handleSubmit = async () => {
    try {
      if (editingProvider) {
        await updateProvider.mutateAsync({ name: editingProvider.name, provider: formData })
        toast.success(t('common.success'))
      } else {
        await createProvider.mutateAsync(formData as Provider)
        toast.success(t('common.success'))
      }
      setDialogOpen(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const handleDelete = async () => {
    if (!deletingProvider) return
    try {
      await deleteProvider.mutateAsync(deletingProvider)
      toast.success(t('common.success'))
      setDeleteDialogOpen(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  if (isLoading) {
    return <div className="flex justify-center p-8">{t('common.loading')}</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{t('providers.title')}</h1>
          <p className="text-muted-foreground">{t('providers.description')}</p>
        </div>
        <Button onClick={handleOpenCreate}>
          <Plus className="mr-2 h-4 w-4" />
          {t('providers.addProvider')}
        </Button>
      </div>

      {providers && providers.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {providers.map((provider) => (
            <Card key={provider.name}>
              <CardHeader className="flex flex-row items-start justify-between space-y-0">
                <div className="flex items-center gap-2">
                  <Server className="h-5 w-5 text-muted-foreground" />
                  <CardTitle className="text-lg">{provider.name}</CardTitle>
                </div>
                <Badge variant={provider.enabled !== false ? 'success' : 'secondary'}>
                  {provider.enabled !== false ? t('common.enabled') : t('common.disabled')}
                </Badge>
              </CardHeader>
              <CardContent className="space-y-2">
                <CardDescription className="truncate">{provider.base_url}</CardDescription>
                {provider.model && (
                  <p className="text-sm text-muted-foreground">
                    {t('providers.model')}: {provider.model}
                  </p>
                )}
                <div className="flex gap-2 pt-2">
                  <Button variant="outline" size="sm" onClick={() => handleOpenEdit(provider)}>
                    <Pencil className="mr-1 h-3 w-3" />
                    {t('common.edit')}
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => handleOpenDelete(provider.name)}>
                    <Trash2 className="mr-1 h-3 w-3" />
                    {t('common.delete')}
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Server className="mb-4 h-12 w-12 text-muted-foreground" />
            <p className="text-muted-foreground">{t('providers.noProviders')}</p>
          </CardContent>
        </Card>
      )}

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingProvider ? t('providers.editProvider') : t('providers.addProvider')}
            </DialogTitle>
            <DialogDescription>
              {editingProvider ? t('providers.editProvider') : t('providers.addProvider')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">{t('providers.name')}</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                disabled={!!editingProvider}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="base_url">{t('providers.baseUrl')}</Label>
              <Input
                id="base_url"
                value={formData.base_url}
                onChange={(e) => setFormData({ ...formData, base_url: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="api_key">{t('providers.apiKey')}</Label>
              <Input
                id="api_key"
                type="password"
                value={formData.api_key}
                onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
                placeholder={editingProvider ? '••••••••' : ''}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="api_key_env">{t('providers.envVar')}</Label>
              <Input
                id="api_key_env"
                value={formData.api_key_env}
                onChange={(e) => setFormData({ ...formData, api_key_env: e.target.value })}
                placeholder="ANTHROPIC_API_KEY"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="model">{t('providers.model')}</Label>
              <Input
                id="model"
                value={formData.model}
                onChange={(e) => setFormData({ ...formData, model: e.target.value })}
                placeholder="claude-sonnet-4-20250514"
              />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="enabled">{t('common.enabled')}</Label>
              <Switch
                id="enabled"
                checked={formData.enabled !== false}
                onCheckedChange={(checked) => setFormData({ ...formData, enabled: checked })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSubmit} disabled={createProvider.isPending || updateProvider.isPending}>
              {t('common.save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('providers.deleteProvider')}</DialogTitle>
            <DialogDescription>{t('providers.deleteConfirm')}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteProvider.isPending}>
              {t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
