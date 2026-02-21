import { describe, it, expect } from 'vitest'
import { CHART_COLORS, formatCost, formatTokens } from './constants'

describe('CHART_COLORS', () => {
  it('has 8 colors', () => {
    expect(CHART_COLORS).toHaveLength(8)
  })

  it('contains valid hex colors', () => {
    CHART_COLORS.forEach((color) => {
      expect(color).toMatch(/^#[0-9a-f]{6}$/i)
    })
  })
})

describe('formatCost', () => {
  it('formats cost with 4 decimal places', () => {
    expect(formatCost(1.5)).toBe('$1.5000')
    expect(formatCost(0.0001)).toBe('$0.0001')
    expect(formatCost(100)).toBe('$100.0000')
  })

  it('handles zero', () => {
    expect(formatCost(0)).toBe('$0.0000')
  })
})

describe('formatTokens', () => {
  it('formats millions', () => {
    expect(formatTokens(1000000)).toBe('1.00M')
    expect(formatTokens(2500000)).toBe('2.50M')
    expect(formatTokens(10000000)).toBe('10.00M')
  })

  it('formats thousands', () => {
    expect(formatTokens(1000)).toBe('1.0K')
    expect(formatTokens(2500)).toBe('2.5K')
    expect(formatTokens(999999)).toBe('1000.0K')
  })

  it('formats small numbers as-is', () => {
    expect(formatTokens(0)).toBe('0')
    expect(formatTokens(100)).toBe('100')
    expect(formatTokens(999)).toBe('999')
  })
})
