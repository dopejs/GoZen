import { useState } from 'react'
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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useProfiles, useCreateProfile, useUpdateProfile, useDeleteProfile } from '@/hooks/use-profiles'
import { useProviders } from '@/hooks/use-providers'
import type { Profile } from '@/types/api'

export function ProfilesPage() {
  const { t } = useTranslation()
  const { data: profiles, isLoading } = useProfiles()
  const { data: providers } = useProviders()
  const createProfile = useCreateProfile()
  const updateProfile = useUpdateProfile()
  const deleteProfile = useDeleteProfile()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingProfile, setEditingProfile] = useState<Profile | null>(null)
  const [deletingProfile, setDeletingProfile] = useState<string | null>(null)
  const [formData, setFormData] = useState<Partial<Profile>>({
    name: '',
    providers: [],
    fallback: [],
    is_default: false,
  })

  const handleOpenCreate = () => {
    setEditingProfile(null)
    setFormData({
      name: '',
      providers: [],
      fallback: [],
      is_default: false,
    })
    setDialogOpen(true)
  }

  const handleOpenEdit = (profile: Profile) => {
    setEditingProfile(profile)
    setFormData({ ...profile })
    setDialogOpen(true)
  }

  const handleOpenDelete = (name: string) => {
    setDeletingProfile(name)
    setDeleteDialogOpen(true)
  }

  const handleSubmit = async () => {
    try {
      if (editingProfile) {
        await updateProfile.mutateAsync({ name: editingProfile.name, profile: formData })
        toast.success(t('common.success'))
      } else {
        await createProfile.mutateAsync(formData as Profile)
        toast.success(t('common.success'))
      }
      setDialogOpen(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
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

  const toggleProvider = (providerName: string) => {
    const current = formData.providers || []
    if (current.includes(providerName)) {
      setFormData({ ...formData, providers: current.filter((p) => p !== providerName) })
    } else {
      setFormData({ ...formData, providers: [...current, providerName] })
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
        <Button onClick={handleOpenCreate}>
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
                {profile.fallback && profile.fallback.length > 0 && (
                  <p className="text-sm text-muted-foreground">
                    {t('profiles.fallback')}: {profile.fallback.join(', ')}
                  </p>
                )}
                <div className="flex gap-2 pt-2">
                  <Button variant="outline" size="sm" onClick={() => handleOpenEdit(profile)}>
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

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingProfile ? t('profiles.editProfile') : t('profiles.addProfile')}
            </DialogTitle>
            <DialogDescription>
              {editingProfile ? t('profiles.editProfile') : t('profiles.addProfile')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">{t('profiles.name')}</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                disabled={!!editingProfile}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('profiles.providers')}</Label>
              <div className="flex flex-wrap gap-2">
                {providers?.map((provider) => (
                  <Badge
                    key={provider.name}
                    variant={formData.providers?.includes(provider.name) ? 'default' : 'outline'}
                    className="cursor-pointer"
                    onClick={() => toggleProvider(provider.name)}
                  >
                    {provider.name}
                  </Badge>
                ))}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSubmit} disabled={createProfile.isPending || updateProfile.isPending}>
              {t('common.save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

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
