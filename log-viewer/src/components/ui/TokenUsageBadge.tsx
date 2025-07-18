'use client'

interface TokenUsageBadgeProps {
  usage?: {
    prompt?: number
    completion?: number
    total?: number
    // Support different API response formats
    prompt_tokens?: number
    completion_tokens?: number
    total_tokens?: number
  }
  cost?: number
  className?: string
}

export function TokenUsageBadge({ usage, cost, className = '' }: TokenUsageBadgeProps) {
  if (!usage && !cost) return null

  // Normalize usage data to handle different API response formats
  const normalizedUsage = usage ? {
    prompt: usage.prompt ?? usage.prompt_tokens ?? 0,
    completion: usage.completion ?? usage.completion_tokens ?? 0,
    total: usage.total ?? usage.total_tokens ?? (usage.prompt ?? usage.prompt_tokens ?? 0) + (usage.completion ?? usage.completion_tokens ?? 0)
  } : null

  // Only show usage if we have valid data
  const hasValidUsage = normalizedUsage && (normalizedUsage.prompt > 0 || normalizedUsage.completion > 0 || normalizedUsage.total > 0)

  return (
    <div className={`inline-flex items-center gap-3 px-3 py-1.5 bg-blue-50 border border-blue-200 rounded-lg text-xs ${className}`}>
      {hasValidUsage && (
        <div className="flex items-center gap-2">
          <span className="text-blue-600 font-medium">Tokens:</span>
          <span className="text-blue-800 font-mono">
            {normalizedUsage.prompt.toLocaleString()} + {normalizedUsage.completion.toLocaleString()} = {normalizedUsage.total.toLocaleString()}
          </span>
        </div>
      )}
      {cost && cost > 0 && (
        <div className="flex items-center gap-1 text-green-700">
          <span className="font-medium">Cost:</span>
          <span className="font-mono">${cost.toFixed(4)}</span>
        </div>
      )}
    </div>
  )
}