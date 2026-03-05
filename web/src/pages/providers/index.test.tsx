import { describe, it, expect } from 'vitest'
import { render, screen, waitFor } from '@/test/utils'
import { ProvidersPage } from './index'

describe('ProvidersPage', () => {
  it('renders the providers page title', async () => {
    render(<ProvidersPage />)

    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument()
    })
  })

  it('renders provider list from MSW mock data', async () => {
    render(<ProvidersPage />)

    await waitFor(() => {
      expect(screen.getByText('anthropic')).toBeInTheDocument()
      expect(screen.getByText('openai')).toBeInTheDocument()
    })
  })

  it('renders add provider button', async () => {
    render(<ProvidersPage />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /add/i })).toBeInTheDocument()
    })
  })

  it('renders edit and delete buttons for each provider', async () => {
    render(<ProvidersPage />)

    await waitFor(() => {
      expect(screen.getByText('anthropic')).toBeInTheDocument()
    })

    const editButtons = screen.getAllByRole('button', { name: /edit/i })
    const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
    expect(editButtons.length).toBeGreaterThanOrEqual(2)
    expect(deleteButtons.length).toBeGreaterThanOrEqual(2)
  })

  it('shows provider base URL', async () => {
    render(<ProvidersPage />)

    await waitFor(() => {
      expect(screen.getByText('https://api.anthropic.com')).toBeInTheDocument()
    })
  })
})
