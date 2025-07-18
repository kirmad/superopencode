'use client'

import { useState } from 'react'
import { ChevronRight, Copy } from 'lucide-react'
import { MessagePart } from '@/lib/types'

interface CollapsibleMessageViewerProps {
  message: MessagePart
  defaultOpen?: boolean
}

export function CollapsibleMessageViewer({ 
  message, 
  defaultOpen = false 
}: CollapsibleMessageViewerProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen)
  
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(message.content)
      // TODO: Show toast notification
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }
  
  return (
    <div className={`border rounded-lg mb-3 ${message.colorClass} border-gray-200`}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full p-3 text-left flex justify-between items-center hover:bg-opacity-80 transition-colors"
      >
        <div className="flex items-center gap-2">
          <span className="text-lg">{message.icon}</span>
          <span className="font-medium text-gray-900 capitalize">
            {message.type.replace('_', ' ')}
          </span>
          {message.metadata?.toolName && (
            <span className="text-sm text-gray-600">
              ({message.metadata.toolName})
            </span>
          )}
        </div>
        
        <div className="flex items-center gap-2">
          <button
            onClick={(e) => {
              e.stopPropagation()
              handleCopy()
            }}
            className="p-1 hover:bg-gray-200 rounded"
            title="Copy content"
          >
            <Copy className="h-4 w-4" />
          </button>
          
          <ChevronRight 
            className={`h-4 w-4 transform transition-transform ${
              isOpen ? 'rotate-90' : ''
            }`}
          />
        </div>
      </button>
      
      {isOpen && (
        <div className="p-2 pt-0">
          <ContentRenderer message={message} />
        </div>
      )}
    </div>
  )
}

function ContentRenderer({ message }: { message: MessagePart }) {
  if (message.type === 'tool_call' && message.metadata?.arguments) {
    return (
      <div>
        <div className="font-mono text-xs mb-1 text-gray-700">
          Function: {message.metadata.toolName}
        </div>
        <pre className="bg-gray-100 p-2 rounded text-xs overflow-x-auto font-mono leading-tight">
          {JSON.stringify(message.metadata.arguments, null, 2)}
        </pre>
      </div>
    )
  }
  
  if (message.type === 'system' || message.type === 'user') {
    return (
      <div className="whitespace-pre-wrap text-xs leading-snug">
        {message.content}
      </div>
    )
  }
  
  return (
    <div className="prose prose-xs max-w-none">
      <div className="whitespace-pre-wrap text-xs leading-snug">
        {message.content}
      </div>
    </div>
  )
}