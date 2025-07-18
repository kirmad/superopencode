'use client'

interface ProviderBadgeProps {
  provider: string
  model?: string
  className?: string
}

export function ProviderBadge({ provider, model, className = '' }: ProviderBadgeProps) {
  const getProviderInfo = (provider: string) => {
    switch (provider.toLowerCase()) {
      case 'anthropic':
        return { 
          name: 'Anthropic', 
          icon: '🤖', 
          color: 'bg-orange-100 text-orange-800 border-orange-200',
          textColor: 'text-orange-700'
        }
      case 'openai':
        return { 
          name: 'OpenAI', 
          icon: '🟢', 
          color: 'bg-green-100 text-green-800 border-green-200',
          textColor: 'text-green-700'
        }
      case 'copilot':
        return { 
          name: 'GitHub Copilot', 
          icon: '🐙', 
          color: 'bg-purple-100 text-purple-800 border-purple-200',
          textColor: 'text-purple-700'
        }
      default:
        return { 
          name: provider, 
          icon: '🔷', 
          color: 'bg-gray-100 text-gray-800 border-gray-200',
          textColor: 'text-gray-700'
        }
    }
  }

  const info = getProviderInfo(provider)

  return (
    <div className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border font-medium text-xs ${info.color} ${className}`}>
      <span className="text-sm">{info.icon}</span>
      <span>{info.name}</span>
      {model && <span className={`font-mono text-xs ${info.textColor} opacity-75`}>({model})</span>}
    </div>
  )
}