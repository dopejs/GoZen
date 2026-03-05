import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StatCard } from './StatCard'

describe('StatCard', () => {
  it('renders title and value', () => {
    render(
      <StatCard
        title="Total Requests"
        value={100}
        icon={<span data-testid="icon">📊</span>}
      />
    )

    expect(screen.getByText('Total Requests')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()
    expect(screen.getByTestId('icon')).toBeInTheDocument()
  })

  it('renders string value', () => {
    render(
      <StatCard
        title="Cost"
        value="$1.50"
        icon={<span>💰</span>}
      />
    )

    expect(screen.getByText('Cost')).toBeInTheDocument()
    expect(screen.getByText('$1.50')).toBeInTheDocument()
  })
})
