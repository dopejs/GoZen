import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useBot, useUpdateBot } from '@/hooks/use-bot'
import type { TabProps } from './types'

export function AliasesTab({ config, setConfig }: TabProps) {
  const { t } = useTranslation()
  const { data: botData } = useBot()
  const updateBot = useUpdateBot()
  const [newAlias, setNewAlias] = useState('')
  const [newPath, setNewPath] = useState('')

  const handleSave = async () => {
    try {
      await updateBot.mutateAsync({ aliases: config.aliases })
      toast.success(t('common.success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const addAlias = () => {
    if (!newAlias || !newPath) return
    setConfig((c) => ({
      ...c,
      aliases: { ...c.aliases, [newAlias]: newPath },
    }))
    setNewAlias('')
    setNewPath('')
  }

  const removeAlias = (alias: string) => {
    setConfig((c) => {
      const newAliases = { ...c.aliases }
      delete newAliases[alias]
      return { ...c, aliases: newAliases }
    })
  }

  const selectRecentPath = (path: string) => {
    setNewPath(path)
  }

  const aliases = Object.entries(config.aliases || {})
  const recentPaths = botData?.recent_paths || []

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('bot.aliases')}</CardTitle>
        <CardDescription>{t('bot.aliasesDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {recentPaths.length > 0 && (
          <div className="space-y-2">
            <Label>{t('bot.recentPaths')}</Label>
            <div className="flex flex-wrap gap-2">
              {recentPaths.map((path) => (
                <button
                  key={path}
                  type="button"
                  onClick={() => selectRecentPath(path)}
                  className="inline-flex items-center rounded-md bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground hover:bg-muted/80 hover:text-foreground transition-colors"
                >
                  {path.length > 40 ? `...${path.slice(-37)}` : path}
                </button>
              ))}
            </div>
          </div>
        )}

        <div className="rounded-lg border">
          <div className="grid grid-cols-[1fr_2fr_auto] gap-2 border-b bg-muted/50 p-3 text-sm font-medium">
            <div>{t('bot.alias')}</div>
            <div>{t('bot.projectPath')}</div>
            <div></div>
          </div>
          {aliases.length === 0 ? (
            <div className="p-4 text-center text-muted-foreground">{t('bot.noAliases')}</div>
          ) : (
            aliases.map(([alias, path]) => (
              <div key={alias} className="grid grid-cols-[1fr_2fr_auto] items-center gap-2 border-b p-3 last:border-0">
                <div className="font-mono text-sm">{alias}</div>
                <div className="truncate text-sm text-muted-foreground">{path}</div>
                <Button variant="ghost" size="icon" onClick={() => removeAlias(alias)}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))
          )}
        </div>

        <div className="flex gap-2">
          <Input
            placeholder={t('bot.alias')}
            value={newAlias}
            onChange={(e) => setNewAlias(e.target.value)}
            className="flex-1"
          />
          <Input
            placeholder={t('bot.projectPath')}
            value={newPath}
            onChange={(e) => setNewPath(e.target.value)}
            className="flex-[2]"
          />
          <Button variant="outline" onClick={addAlias}>
            <Plus className="h-4 w-4" />
          </Button>
        </div>

        <Button onClick={handleSave} disabled={updateBot.isPending}>
          {t('common.save')}
        </Button>
      </CardContent>
    </Card>
  )
}
