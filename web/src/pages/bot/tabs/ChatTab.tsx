import { useState, useRef, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Send, Trash2, MessageSquare } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

interface Message {
  role: 'user' | 'assistant'
  content: string
}

const STORAGE_KEY = 'bot-chat-state'

function loadChatState(): { messages: Message[]; sessionId: string | null } {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (raw) {
      const state = JSON.parse(raw)
      return { messages: state.messages || [], sessionId: state.sessionId || null }
    }
  } catch {
    // ignore
  }
  return { messages: [], sessionId: null }
}

function saveChatState(messages: Message[], sessionId: string | null) {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify({ messages, sessionId }))
  } catch {
    // ignore
  }
}

export function ChatTab() {
  const { t } = useTranslation()
  const initial = useRef(loadChatState())
  const [messages, setMessages] = useState<Message[]>(initial.current.messages)
  const [input, setInput] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [sessionId, setSessionId] = useState<string | null>(initial.current.sessionId)
  const [streamingContent, setStreamingContent] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  // Persist chat state on changes
  useEffect(() => {
    saveChatState(messages, sessionId)
  }, [messages, sessionId])

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages, streamingContent])

  const handleSend = async () => {
    if (!input.trim() || isLoading) return

    const userMessage = input.trim()
    setInput('')
    setMessages((prev) => [...prev, { role: 'user', content: userMessage }])
    setIsLoading(true)
    setStreamingContent('')

    try {
      const response = await fetch('/api/v1/bot/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          message: userMessage,
          session_id: sessionId,
        }),
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Failed to send message')
      }

      const reader = response.body?.getReader()
      if (!reader) throw new Error('No response body')

      const decoder = new TextDecoder()
      let fullContent = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (line.startsWith('event: ')) {
            const event = line.slice(7)
            const dataLine = lines[lines.indexOf(line) + 1]
            if (dataLine?.startsWith('data: ')) {
              const data = JSON.parse(dataLine.slice(6))

              if (event === 'session') {
                setSessionId(data.session_id)
              } else if (event === 'delta') {
                fullContent += data.content
                setStreamingContent(fullContent)
              } else if (event === 'done') {
                setMessages((prev) => [...prev, { role: 'assistant', content: data.content }])
                setStreamingContent('')
              } else if (event === 'error') {
                throw new Error(data.error)
              }
            }
          }
        }
      }
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: `Error: ${err instanceof Error ? err.message : 'Unknown error'}` },
      ])
      setStreamingContent('')
    } finally {
      setIsLoading(false)
    }
  }

  const handleClear = async () => {
    if (sessionId) {
      try {
        await fetch('/api/v1/bot/chat', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ session_id: sessionId, clear: true }),
        })
      } catch {
        // Ignore errors
      }
    }
    setMessages([])
    setSessionId(null)
    setStreamingContent('')
    sessionStorage.removeItem(STORAGE_KEY)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <Card className="flex flex-col h-[600px]">
      <CardHeader className="flex-shrink-0">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <MessageSquare className="h-5 w-5" />
              {t('bot.chat', 'Chat')}
            </CardTitle>
            <CardDescription>{t('bot.chatDesc', 'Test the bot with a live conversation')}</CardDescription>
          </div>
          <Button variant="outline" size="sm" onClick={handleClear} disabled={messages.length === 0 && !streamingContent}>
            <Trash2 className="h-4 w-4 mr-2" />
            {t('bot.clearChat', 'Clear')}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col min-h-0">
        <div className="flex-1 overflow-y-auto space-y-4 mb-4 p-4 bg-muted/30 rounded-lg">
          {messages.length === 0 && !streamingContent && (
            <div className="text-center text-muted-foreground py-8">
              {t('bot.chatEmpty', 'Send a message to start chatting with the bot')}
            </div>
          )}
          {messages.map((msg, i) => (
            <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
              <div
                className={`max-w-[80%] rounded-lg px-4 py-2 ${
                  msg.role === 'user' ? 'bg-primary text-primary-foreground' : 'bg-muted'
                }`}
              >
                <pre className="whitespace-pre-wrap font-sans text-sm">{msg.content}</pre>
              </div>
            </div>
          ))}
          {streamingContent && (
            <div className="flex justify-start">
              <div className="max-w-[80%] rounded-lg px-4 py-2 bg-muted">
                <pre className="whitespace-pre-wrap font-sans text-sm">{streamingContent}</pre>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>
        <div className="flex gap-2">
          <Input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={t('bot.chatPlaceholder', 'Type a message...')}
            disabled={isLoading}
          />
          <Button onClick={handleSend} disabled={!input.trim() || isLoading}>
            <Send className="h-4 w-4" />
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
