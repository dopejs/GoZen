import { describe, it, expect } from 'vitest'
import { render, screen, waitFor } from '@/test/utils'
import { GeneralSettings } from './GeneralSettings'

describe('GeneralSettings', () => {
  it('renders the general settings card', async () => {
    render(<GeneralSettings />)

    await waitFor(() => {
      expect(screen.getByRole('heading')).toBeInTheDocument()
    })
  })

  it('shows proxy port as read-only', async () => {
    render(<GeneralSettings />)

    await waitFor(() => {
      // The proxy port input should be disabled
      const inputs = screen.getAllByRole('textbox') as HTMLInputElement[]
      const disabledInputs = inputs.filter((input) => input.disabled)
      expect(disabledInputs.length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders save button', async () => {
    render(<GeneralSettings />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument()
    })
  })

  it('renders default profile and client selectors', async () => {
    render(<GeneralSettings />)

    await waitFor(() => {
      // There should be combo/select controls for profile and client
      const comboboxes = screen.getAllByRole('combobox')
      expect(comboboxes.length).toBeGreaterThanOrEqual(2)
    })
  })
})
