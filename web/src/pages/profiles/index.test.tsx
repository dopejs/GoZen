import { describe, it, expect } from 'vitest'
import { render, screen, waitFor } from '@/test/utils'
import { ProfilesPage } from './index'

describe('ProfilesPage', () => {
  it('renders the profiles page title', async () => {
    render(<ProfilesPage />)

    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument()
    })
  })

  it('renders profile list from MSW mock data', async () => {
    render(<ProfilesPage />)

    await waitFor(() => {
      expect(screen.getByText('default')).toBeInTheDocument()
      expect(screen.getByText('work')).toBeInTheDocument()
    })
  })

  it('renders add profile button', async () => {
    render(<ProfilesPage />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /add/i })).toBeInTheDocument()
    })
  })

  it('shows provider names in each profile card', async () => {
    render(<ProfilesPage />)

    await waitFor(() => {
      expect(screen.getByText(/anthropic, openai/i)).toBeInTheDocument()
    })
  })

  it('renders edit and delete buttons for each profile', async () => {
    render(<ProfilesPage />)

    await waitFor(() => {
      expect(screen.getByText('default')).toBeInTheDocument()
    })

    const editButtons = screen.getAllByRole('button', { name: /edit/i })
    const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
    expect(editButtons.length).toBeGreaterThanOrEqual(2)
    expect(deleteButtons.length).toBeGreaterThanOrEqual(2)
  })
})
