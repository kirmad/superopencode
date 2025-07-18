'use client'

import { useState } from 'react'
import { ChevronDown, ChevronRight, Activity, Zap, CheckCircle, Clock, Hash } from 'lucide-react'
import { JsonRenderer } from './JsonRenderer'
import { parseSSEStream, type SSEEvent, getStreamSummary } from '@/lib/parsers/sse'

interface StreamEventRendererProps {
  responseBody: any
  title?: string
  maxHeight?: string
}

export function StreamEventRenderer({ responseBody, title = "Stream Events", maxHeight = "400px" }: StreamEventRendererProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [selectedEvent, setSelectedEvent] = useState<number | null>(null)
  
  if (typeof responseBody !== 'string' || !responseBody.includes('data:')) {
    return null
  }

  const parsedStream = parseSSEStream(responseBody)
  const { events, totalEvents, hasToolCalls, reconstructedMessage, metadata } = parsedStream

  if (totalEvents === 0) {
    return null
  }

  const summary = getStreamSummary(parsedStream)

  return (
    <div className="border border-gray-200 rounded">
      {/* Header */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-3 py-2 flex items-center justify-between text-left hover:bg-gray-50 transition-colors border-b border-gray-200"
      >
        <div className="flex items-center gap-2">
          {isExpanded ? (
            <ChevronDown className="h-3 w-3 text-gray-600" />
          ) : (
            <ChevronRight className="h-3 w-3 text-gray-600" />
          )}
          <Activity className="h-3 w-3 text-blue-600" />
          <span className="text-sm font-medium text-gray-900">{title}</span>
          <span className="px-2 py-0.5 bg-blue-100 text-blue-800 rounded text-xs font-medium">
            {totalEvents} events
          </span>
          {hasToolCalls && (
            <span className="px-2 py-0.5 bg-purple-100 text-purple-800 rounded text-xs font-medium">
              Tool Calls
            </span>
          )}
        </div>
        <div className="text-xs text-gray-500">
          {summary}
        </div>
      </button>

      {/* Stream Events */}
      {isExpanded && (
        <div className="p-0" style={{ maxHeight, overflow: 'auto' }}>
          {/* Stream Metadata */}
          {metadata && Object.keys(metadata).length > 0 && (
            <div className="bg-gray-50 px-3 py-2 border-b border-gray-200">
              <div className="text-xs font-medium text-gray-700 mb-1">Stream Metadata</div>
              <div className="grid grid-cols-2 gap-2 text-xs">
                {metadata.model && (
                  <div>
                    <span className="text-gray-500">Model:</span>
                    <span className="ml-1 font-mono text-gray-800">{metadata.model}</span>
                  </div>
                )}
                {metadata.finishReason && (
                  <div>
                    <span className="text-gray-500">Finish:</span>
                    <span className="ml-1 font-mono text-gray-800">{metadata.finishReason}</span>
                  </div>
                )}
                {metadata.usage?.total_tokens && (
                  <div>
                    <span className="text-gray-500">Tokens:</span>
                    <span className="ml-1 font-mono text-gray-800">{metadata.usage.total_tokens}</span>
                  </div>
                )}
                {metadata.streamDuration && metadata.streamDuration > 0 && (
                  <div>
                    <span className="text-gray-500">Duration:</span>
                    <span className="ml-1 font-mono text-gray-800">{metadata.streamDuration}ms</span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Reconstructed Message */}
          {reconstructedMessage && (
            <div className="bg-green-50 px-3 py-2 border-b border-gray-200">
              <div className="text-xs font-medium text-green-800 mb-1 flex items-center gap-1">
                <CheckCircle className="h-3 w-3" />
                Reconstructed Message
              </div>
              <JsonRenderer 
                data={reconstructedMessage}
                title="Final Message"
                maxHeight="150px"
              />
            </div>
          )}

          {/* Individual Events */}
          <div className="divide-y divide-gray-100">
            {events.map((event, index) => (
              <StreamEventItem
                key={index}
                event={event}
                index={index}
                isSelected={selectedEvent === index}
                onSelect={() => setSelectedEvent(selectedEvent === index ? null : index)}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

interface StreamEventItemProps {
  event: SSEEvent
  index: number
  isSelected: boolean
  onSelect: () => void
}

function StreamEventItem({ event, index, isSelected, onSelect }: StreamEventItemProps) {
  const getEventTypeInfo = () => {
    if (event.data === '[DONE]') {
      return { icon: CheckCircle, color: 'text-green-600', bg: 'bg-green-50', label: 'DONE' }
    }
    
    if (event.parsed) {
      const data = event.parsed
      
      // Check for different event types
      if (data.choices?.[0]?.finish_reason) {
        return { icon: CheckCircle, color: 'text-blue-600', bg: 'bg-blue-50', label: 'FINISH' }
      }
      if (data.choices?.[0]?.delta?.tool_calls) {
        return { icon: Zap, color: 'text-purple-600', bg: 'bg-purple-50', label: 'TOOL' }
      }
      if (data.choices?.[0]?.delta?.content) {
        return { icon: Hash, color: 'text-gray-600', bg: 'bg-gray-50', label: 'CONTENT' }
      }
      if (data.usage) {
        return { icon: Activity, color: 'text-orange-600', bg: 'bg-orange-50', label: 'USAGE' }
      }
      if (data.prompt_filter_results) {
        return { icon: Activity, color: 'text-yellow-600', bg: 'bg-yellow-50', label: 'FILTER' }
      }
    }
    
    return { icon: Activity, color: 'text-gray-600', bg: 'bg-gray-50', label: 'DATA' }
  }

  const typeInfo = getEventTypeInfo()
  const Icon = typeInfo.icon

  return (
    <div className={`p-2 ${isSelected ? 'bg-blue-50' : 'hover:bg-gray-50'} transition-colors`}>
      <button
        onClick={onSelect}
        className="w-full text-left"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className={`p-1 rounded ${typeInfo.bg}`}>
              <Icon className={`h-3 w-3 ${typeInfo.color}`} />
            </div>
            <span className="text-xs font-mono text-gray-700">#{index + 1}</span>
            <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${typeInfo.bg} ${typeInfo.color}`}>
              {typeInfo.label}
            </span>
            {event.timestamp !== undefined && (
              <div className="flex items-center gap-1 text-xs text-gray-500">
                <Clock className="h-2.5 w-2.5" />
                <span>{event.timestamp}</span>
              </div>
            )}
          </div>
          <div className="flex items-center gap-1">
            {isSelected ? (
              <ChevronDown className="h-3 w-3 text-gray-500" />
            ) : (
              <ChevronRight className="h-3 w-3 text-gray-500" />
            )}
          </div>
        </div>

        {/* Event preview */}
        {!isSelected && (
          <div className="mt-1 text-xs text-gray-600 truncate font-mono">
            {event.data === '[DONE]' ? 'Stream completed' : 
             event.parsed ? getEventPreview(event.parsed) : 
             event.data.substring(0, 80) + (event.data.length > 80 ? '...' : '')}
          </div>
        )}
      </button>

      {/* Event details */}
      {isSelected && (
        <div className="mt-2 space-y-2">
          {event.data === '[DONE]' ? (
            <div className="text-sm text-green-700 font-medium">Stream completed</div>
          ) : event.parsed ? (
            <JsonRenderer 
              data={event.parsed}
              title={`Event #${index + 1} Data`}
              maxHeight="200px"
            />
          ) : (
            <div className="p-2 bg-gray-100 rounded font-mono text-xs">
              {event.data}
            </div>
          )}
          
          {/* Event metadata */}
          {(event.id || event.event || event.retry !== undefined) && (
            <div className="text-xs text-gray-500 space-y-1">
              {event.id && <div><span className="font-medium">ID:</span> {event.id}</div>}
              {event.event && <div><span className="font-medium">Event:</span> {event.event}</div>}
              {event.retry !== undefined && <div><span className="font-medium">Retry:</span> {event.retry}ms</div>}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function getEventPreview(data: any): string {
  if (data.choices?.[0]?.finish_reason) {
    return `Finished: ${data.choices[0].finish_reason}`
  }
  if (data.choices?.[0]?.delta?.tool_calls?.[0]?.function?.name) {
    return `Tool: ${data.choices[0].delta.tool_calls[0].function.name}`
  }
  if (data.choices?.[0]?.delta?.tool_calls?.[0]?.function?.arguments) {
    const args = data.choices[0].delta.tool_calls[0].function.arguments
    return `Args: ${args.substring(0, 40)}${args.length > 40 ? '...' : ''}`
  }
  if (data.choices?.[0]?.delta?.content) {
    return `Content: ${data.choices[0].delta.content}`
  }
  if (data.usage) {
    return `Usage: ${data.usage.total_tokens || 0} tokens`
  }
  if (data.model) {
    return `Model: ${data.model}`
  }
  if (data.prompt_filter_results) {
    return 'Content filter results'
  }
  return 'Event data'
}