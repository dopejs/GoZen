import { describe, it, expect } from 'vitest'
import { render, screen, waitFor } from '@/test/utils'
import userEvent from '@testing-library/user-event'
import { PasswordSettings } from './PasswordSettings'

describe('PasswordSettings', () => {
  it('renders the password settings card', () => {
    render(<PasswordSettings />)
    expect(screen.getByRole('heading')).toBeInTheDocument()
  })

  it('renders password form fields', () => {
    render(<PasswordSettings />)

    expect(screen.getByLabelText(/current.?password/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/new.?password/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/confirm.?password/i)).toBeInTheDocument()
  })

  it('renders submit button', () => {
    render(<PasswordSettings />)
    expect(screen.getByRole('button', { name: /change|save|submit/i })).toBeInTheDocument()
  })

  it('allows typing in password fields', async () => {
    const user = userEvent.setup()
    render(<PasswordSettings />)

    const currentPw = screen.getByLabelText(/current.?password/i) as HTMLInputElement
    const newPw = screen.getByLabelText(/new.?password/i) as HTMLInputElement

    await user.type(currentPw, 'oldpass')
    await user.type(newPw, 'newpass')

    expect(currentPw.value).toBe('oldpass')
    expect(newPw.value).toBe('newpass')
  })

  it('submits form successfully when passwords match', async () => {
    const user = userEvent.setup()
    render(<PasswordSettings />)

    const currentPw = screen.getByLabelText(/current.?password/i)
    const newPw = screen.getByLabelText(/new.?password/i)
    const confirmPw = screen.getByLabelText(/confirm.?password/i)
    const submitBtn = screen.getByRole('button', { name: /change|save|submit/i })

    await user.type(currentPw, 'oldpass')
    await user.type(newPw, 'newpass123')
    await user.type(confirmPw, 'newpass123')
    await user.click(submitBtn)

    // After successful submit, fields should be cleared
    await waitFor(() => {
      expect((currentPw as HTMLInputElement).value).toBe('')
    })
  })

  it('shows error when passwords do not match', async () => {
    const user = userEvent.setup()
    render(<PasswordSettings />)

    const currentPw = screen.getByLabelText(/current.?password/i)
    const newPw = screen.getByLabelText(/new.?password/i)
    const confirmPw = screen.getByLabelText(/confirm.?password/i)
    const submitBtn = screen.getByRole('button', { name: /change|save|submit/i })

    await user.type(currentPw, 'oldpass')
    await user.type(newPw, 'newpass123')
    await user.type(confirmPw, 'differentpass')
    await user.click(submitBtn)

    // Fields should NOT be cleared since it's a mismatch error
    expect((newPw as HTMLInputElement).value).toBe('newpass123')
  })
})
