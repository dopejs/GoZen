import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { authApi } from '@/lib/api'

export function LoginPage() {
  const { t } = useTranslation()
  const [password, setPassword] = useState('')
  const queryClient = useQueryClient()

  const loginMutation = useMutation({
    mutationFn: (pwd: string) => authApi.login(pwd),
    onSuccess: (data) => {
      if (data.success) {
        queryClient.invalidateQueries({ queryKey: ['auth'] })
      } else {
        toast.error(t('auth.invalidPassword'))
      }
    },
    onError: () => {
      toast.error(t('auth.invalidPassword'))
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    loginMutation.mutate(password)
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center">
            <img src="/logo.svg" alt="GoZen" className="h-12 w-12" />
          </div>
          <CardTitle className="text-2xl text-teal">GoZen</CardTitle>
          <CardDescription>{t('auth.loginRequired')}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="password">{t('auth.password')}</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                autoFocus
              />
            </div>
            <Button type="submit" className="w-full" disabled={loginMutation.isPending}>
              {loginMutation.isPending ? t('common.loading') : t('auth.login')}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
