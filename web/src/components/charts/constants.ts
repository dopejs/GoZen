// Chart colors - brand palette
export const CHART_COLORS = [
  '#5eead4', // teal
  '#c4b5fd', // lavender
  '#86efac', // sage
  '#fbbf24', // amber
  '#93c5fd', // blue
  '#fb7185', // red
  '#a78bfa', // purple
  '#34d399', // emerald
]

// Format helpers
export const formatCost = (cost: number) => `$${cost.toFixed(4)}`

export const formatTokens = (tokens: number) => {
  if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(2)}M`
  if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}K`
  return tokens.toString()
}
