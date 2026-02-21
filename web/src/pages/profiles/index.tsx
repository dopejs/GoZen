import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Plus, Pencil, Trash2, Layers, Star } from 'lucide-react'
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
import { useProfiles, useDeleteProfile } from '@/hooks/use-profiles'

export function ProfilesPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: profiles, isLoading } = useProfiles()
  const deleteProfile = useDeleteProfile()

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deletingProfile, setDeletingProfile] = useState<string | null>(null)

  const handleOpenDelete = (name: string) => {
    setDeletingProfile(name)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!deletingProfile) return
    try {
      await deleteProfile.mutateAsync(deletingProfile)
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
          <h1 className="text-3xl font-bold">{t('profiles.title')}</h1>
          <p className="text-muted-foreground">{t('profiles.description')}</p>
        </div>
        <Button onClick={() => navigate('/profiles/new')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('profiles.addProfile')}
        </Button>
      </div>

      {profiles && profiles.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {profiles.map((profile) => (
            <Card key={profile.name}>
              <CardHeader className="flex flex-row items-start justify-between space-y-0">
                <div className="flex items-center gap-2">
                  <Layers className="h-5 w-5 text-muted-foreground" />
                  <CardTitle className="text-lg">{profile.name}</CardTitle>
                </div>
                {profile.is_default && (
                  <Badge variant="default" className="flex items-center gap-1">
                    <Star className="h-3 w-3" />
                    {t('profiles.default')}
                  </Badge>
                )}
              </CardHeader>
              <CardContent className="space-y-2">
                <CardDescription>
                  {t('profiles.providers')}: {profile.providers?.join(', ') || '-'}
                </CardDescription>
                {profile.strategy && (
                  <p className="text-sm text-muted-foreground">
                    {t('profiles.strategy')}: {profile.strategy}
                  </p>
                )}
                <div className="flex gap-2 pt-2">
                  <Button variant="outline" size="sm" onClick={() => navigate(`/profiles/${profile.name}`)}>
                    <Pencil className="mr-1 h-3 w-3" />
                    {t('common.edit')}
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => handleOpenDelete(profile.name)}>
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
            <Layers className="mb-4 h-12 w-12 text-muted-foreground" />
            <p className="text-muted-foreground">{t('profiles.noProfiles')}</p>
          </CardContent>
        </Card>
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('profiles.deleteProfile')}</DialogTitle>
            <DialogDescription>{t('profiles.deleteConfirm')}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteProfile.isPending}>
              {t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
