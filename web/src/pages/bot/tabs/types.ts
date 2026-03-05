import type { BotConfig } from '@/types/api'

export interface TabProps {
  config: Partial<BotConfig>
  setConfig: React.Dispatch<React.SetStateAction<Partial<BotConfig>>>
}
