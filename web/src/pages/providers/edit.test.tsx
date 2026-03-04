import { describe, it, expect } from 'vitest'
import { render, screen, waitFor } from '@/test/utils'
import { ProviderEditPage } from './edit'

// Mock react-router-dom params for "new" provider
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

describe('ProviderEditPage - New Provider', () => {
  it('renders the add provider form', async () => {
    render(<ProviderEditPage />)

    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument()
    })
  })

  it('renders form fields for basic settings', async () => {
    render(<ProviderEditPage />)

    await waitFor(() => {
      expect(screen.getByLabelText(/name/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/base.?url/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/auth.?token/i)).toBeInTheDocument()
    })
  })

  it('renders save and cancel buttons', async () => {
    render(<ProviderEditPage />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
    })
  })

  it('name field is enabled for new provider', async () => {
    render(<ProviderEditPage />)

    await waitFor(() => {
      const nameInput = screen.getByLabelText(/name/i) as HTMLInputElement
      expect(nameInput).not.toBeDisabled()
    })
  })
})
