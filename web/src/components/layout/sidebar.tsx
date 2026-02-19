import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Server,
  Layers,
  ScrollText,
  BarChart3,
  Settings,
  LayoutDashboard,
} from 'lucide-react'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/', icon: LayoutDashboard, labelKey: 'nav.dashboard' },
  { to: '/providers', icon: Server, labelKey: 'nav.providers' },
  { to: '/profiles', icon: Layers, labelKey: 'nav.profiles' },
  { to: '/logs', icon: ScrollText, labelKey: 'nav.logs' },
  { to: '/usage', icon: BarChart3, labelKey: 'nav.usage' },
  { to: '/settings', icon: Settings, labelKey: 'nav.settings' },
]

export function Sidebar() {
  const { t } = useTranslation()

  return (
    <aside className="fixed left-0 top-0 z-40 h-screen w-64 border-r bg-card">
      <div className="flex h-16 items-center gap-2 border-b px-6">
        <img src="/logo.svg" alt="GoZen" className="h-8 w-8" />
        <span className="text-xl font-semibold text-teal">GoZen</span>
      </div>
      <nav className="space-y-1 p-4">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-primary/10 text-primary'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
              )
            }
          >
            <item.icon className="h-5 w-5" />
            {t(item.labelKey)}
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}
