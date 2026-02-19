import { Routes, Route, Navigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Toaster } from 'sonner'
import { AppShell } from '@/components/layout/app-shell'
import { LoginPage } from '@/pages/login'
import { DashboardPage } from '@/pages/dashboard'
import { ProvidersPage } from '@/pages/providers'
import { ProfilesPage } from '@/pages/profiles'
import { BotPage } from '@/pages/bot'
import { LogsPage } from '@/pages/logs'
import { UsagePage } from '@/pages/usage'
import { SettingsPage } from '@/pages/settings'
import { authApi } from '@/lib/api'

function App() {
  const { data: authStatus, isLoading } = useQuery({
    queryKey: ['auth', 'check'],
    queryFn: authApi.check,
    retry: false,
  })

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  const needsAuth = authStatus?.password_set && !authStatus?.authenticated

  if (needsAuth) {
    return (
      <>
        <LoginPage />
        <Toaster position="top-right" />
      </>
    )
  }

  return (
    <>
      <Routes>
        <Route element={<AppShell />}>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/providers" element={<ProvidersPage />} />
          <Route path="/profiles" element={<ProfilesPage />} />
          <Route path="/bot" element={<BotPage />} />
          <Route path="/logs" element={<LogsPage />} />
          <Route path="/usage" element={<UsagePage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
      <Toaster position="top-right" />
    </>
  )
}

export default App
