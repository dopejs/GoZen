import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@/test/utils'
import { ProfileEditPage } from './edit'

// Mock react-router-dom params for "new" profile
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useParams: () => ({ name: 'new' }),
    useSearchParams: () => {
      const params = new URLSearchParams()
      return [params, vi.fn()] as const
    },
  }
})

describe('ProfileEditPage - New Profile', () => {
  it('renders the add profile form', async () => {
    render(<ProfileEditPage />)

    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument()
    })
  })

  it('renders name input field', async () => {
    render(<ProfileEditPage />)

    await waitFor(() => {
      expect(screen.getByLabelText(/name/i)).toBeInTheDocument()
    })
  })

  it('renders save and cancel buttons', async () => {
    render(<ProfileEditPage />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
    })
  })

  it('renders tabs for basic, providers, and routing', async () => {
    render(<ProfileEditPage />)

    await waitFor(() => {
      const tabs = screen.getAllByRole('tab')
      expect(tabs.length).toBeGreaterThanOrEqual(3)
    })
  })
})
