import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Plus, Pencil, Trash2, Server, Ban, Check } from 'lucide-react'
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useProviders, useDeleteProvider, useDisableProvider, useEnableProvider } from '@/hooks/use-providers'

export function ProvidersPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: providers, isLoading } = useProviders()
  const deleteProvider = useDeleteProvider()
  const disableProvider = useDisableProvider()
  const enableProvider = useEnableProvider()

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deletingProvider, setDeletingProvider] = useState<string | null>(null)

  const handleOpenDelete = (name: string) => {
    setDeletingProvider(name)
    setDeleteDialogOpen(true)
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

  const handleDisable = async (name: string, type: 'today' | 'month' | 'permanent') => {
    try {
      await disableProvider.mutateAsync({ name, type })
      toast.success(t('common.success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const handleEnable = async (name: string) => {
    try {
      await enableProvider.mutateAsync(name)
      toast.success(t('common.success'))
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
        <Button onClick={() => navigate('/providers/new')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('providers.addProvider')}
        </Button>
      </div>

      {providers && providers.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {providers.map((provider) => (
            <Card key={provider.name} className={provider.disabled ? 'opacity-60' : ''}>
              <CardHeader className="flex flex-row items-start justify-between space-y-0">
                <div className="flex items-center gap-2">
                  <Server className="h-5 w-5 text-muted-foreground" />
                  <CardTitle className="text-lg">{provider.name}</CardTitle>
                </div>
                <div className="flex items-center gap-1">
                  {provider.disabled && (
                    <Badge variant="destructive">
                      <Ban className="mr-1 h-3 w-3" />
                      {provider.disabled.type === 'permanent'
                        ? t('providers.disabledPermanent', 'Disabled')
                        : provider.disabled.type === 'month'
                          ? t('providers.disabledMonth', 'Disabled (month)')
                          : t('providers.disabledToday', 'Disabled (today)')}
                    </Badge>
                  )}
                  {provider.type && !provider.disabled && (
                    <Badge variant="outline">{provider.type}</Badge>
                  )}
                </div>
              </CardHeader>
              <CardContent className="space-y-2">
                <CardDescription className="truncate">{provider.base_url}</CardDescription>
                {provider.model && (
                  <p className="text-sm text-muted-foreground">
                    {t('providers.model')}: {provider.model}
                  </p>
                )}
                <div className="flex gap-2 pt-2">
                  <Button variant="outline" size="sm" onClick={() => navigate(`/providers/${provider.name}`)}>
                    <Pencil className="mr-1 h-3 w-3" />
                    {t('common.edit')}
                  </Button>
                  {provider.disabled ? (
                    <Button variant="outline" size="sm" onClick={() => handleEnable(provider.name)}>
                      <Check className="mr-1 h-3 w-3" />
                      {t('providers.enable', 'Enable')}
                    </Button>
                  ) : (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="outline" size="sm">
                          <Ban className="mr-1 h-3 w-3" />
                          {t('providers.disable', 'Disable')}
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent>
                        <DropdownMenuItem onClick={() => handleDisable(provider.name, 'today')}>
                          {t('providers.disableToday', 'Today only')}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => handleDisable(provider.name, 'month')}>
                          {t('providers.disableMonth', 'This month')}
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={() => handleDisable(provider.name, 'permanent')}>
                          {t('providers.disablePermanent', 'Permanently')}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
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
