import { describe, it, expect } from 'vitest'
import { render, screen, waitFor } from '@/test/utils'
import { PermissionSettings } from './PermissionSettings'

describe('PermissionSettings', () => {
  it('renders the permission settings card', async () => {
    render(<PermissionSettings />)

    await waitFor(() => {
      expect(screen.getByRole('heading')).toBeInTheDocument()
    })
  })

  it('renders client sections for Claude, Codex, and OpenCode', async () => {
    render(<PermissionSettings />)

    await waitFor(() => {
      expect(screen.getByText('Claude Code')).toBeInTheDocument()
      expect(screen.getByText('Codex CLI')).toBeInTheDocument()
      expect(screen.getByText('OpenCode')).toBeInTheDocument()
    })
  })

  it('renders toggle switches for each client', async () => {
    render(<PermissionSettings />)

    await waitFor(() => {
      const switches = screen.getAllByRole('switch')
      expect(switches.length).toBe(3)
    })
  })
})
