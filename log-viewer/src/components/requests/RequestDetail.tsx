'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { LLMMessageCard } from '../ui/LLMMessageCard'
import { ProviderBadge } from '../ui/ProviderBadge'
import { TokenUsageBadge } from '../ui/TokenUsageBadge'
import { JsonRenderer } from '../ui/JsonRenderer'
import { StreamEventRenderer } from '../ui/StreamEventRenderer'
import { AgentActionsRenderer } from '../ui/AgentActionsRenderer'
import { Clock, Server, ExternalLink, Activity, CheckCircle, XCircle, ChevronDown, ChevronRight, Send } from 'lucide-react'
import { isSSEResponse } from '@/lib/parsers/sse'

interface RequestDetailProps {
  sessionId: string
  requestId: string
}

export function RequestDetail({ sessionId, requestId }: RequestDetailProps) {
  const [isConversationOpen, setIsConversationOpen] = useState(true)
  const [isTechnicalOpen, setIsTechnicalOpen] = useState(false)
  
  const { data: request, isLoading } = useQuery({
    queryKey: ['request', requestId, sessionId],
    queryFn: async () => {
      const response = await fetch(`/api/requests/${requestId}?sessionId=${sessionId}`)
      if (!response.ok) {
        throw new Error('Failed to fetch request')
      }
      return response.json()
    },
    enabled: !!sessionId && !!requestId
  })
  
  if (isLoading) {
    return (
      <div className="h-full bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading request details...</p>
        </div>
      </div>
    )
  }
  
  if (!request) {
    return (
      <div className="h-full bg-gray-50 flex items-center justify-center">
        <div className="text-center text-gray-500">
          <Server className="h-16 w-16 mx-auto mb-4 opacity-50" />
          <p className="text-lg font-medium">Select a request to view details</p>
          <p className="text-sm">Choose an LLM request from the list to analyze the conversation</p>
        </div>
      </div>
    )
  }

  // Extract provider from URL or headers
  const getProviderInfo = () => {
    const url = request.url?.toLowerCase() || ''
    if (url.includes('anthropic.com')) return { provider: 'anthropic', name: 'Anthropic' }
    if (url.includes('openai.com')) return { provider: 'openai', name: 'OpenAI' }
    if (url.includes('github.com') || url.includes('copilot')) return { provider: 'copilot', name: 'GitHub Copilot' }
    return { provider: 'unknown', name: 'Unknown' }
  }

  const providerInfo = getProviderInfo()
  const isSuccess = request.statusCode >= 200 && request.statusCode < 300
  const isError = request.statusCode >= 400 || request.error

  return (
    <div className="h-full flex flex-col bg-slate-50">
      {/* Compact Header */}
      <div className="bg-white border-b border-slate-200 shadow-sm">
        <div className="p-3">
          {/* Title and Status Row */}
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <h2 className="text-sm font-semibold text-slate-900">Request Analysis</h2>
              {isSuccess ? (
                <CheckCircle className="h-4 w-4 text-emerald-600" />
              ) : isError ? (
                <XCircle className="h-4 w-4 text-red-600" />
              ) : (
                <Activity className="h-4 w-4 text-blue-600" />
              )}
            </div>
            <div className="flex items-center gap-2">
              <ProviderBadge 
                provider={providerInfo.provider}
                model={request.body?.model || request.responseBody?.model || request.response?.model}
              />
              <div className="flex items-center gap-1 px-2 py-1 bg-slate-100 rounded text-xs">
                <Clock className="h-3 w-3 text-slate-600" />
                <span className="font-mono text-slate-700">{request.durationMs}ms</span>
              </div>
            </div>
          </div>
          
          {/* Compact Summary Grid */}
          <div className="grid grid-cols-4 gap-1.5 mb-2">
            <div className="bg-slate-50 px-1.5 py-1 rounded">
              <div className="text-xs text-slate-500">METHOD</div>
              <div className="font-mono text-xs font-medium text-slate-900">{request.method}</div>
            </div>
            <div className="bg-slate-50 px-1.5 py-1 rounded">
              <div className="text-xs text-slate-500">STATUS</div>
              <div className={`font-mono text-xs font-medium ${
                isSuccess ? 'text-emerald-700' : isError ? 'text-red-700' : 'text-blue-700'
              }`}>
                {request.statusCode || '—'}
              </div>
            </div>
            <div className="bg-slate-50 px-1.5 py-1 rounded col-span-2">
              <div className="text-xs text-slate-500">ENDPOINT</div>
              <div className="font-mono text-xs truncate text-slate-900" title={request.url}>
                {new URL(request.url).pathname}
              </div>
            </div>
          </div>
          
          {/* Token Usage - Inline */}
          {(request.responseBody?.usage || request.response?.usage || request.tokensUsed) && (
            <div className="flex justify-end">
              <TokenUsageBadge 
                usage={request.responseBody?.usage || request.response?.usage || request.tokensUsed}
                cost={request.cost}
              />
            </div>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-y-auto">
        <div className="p-3 space-y-3">
          {/* Conversation Section */}
          {request.messages && request.messages.length > 0 ? (
            <div className="bg-white rounded-lg border border-slate-200">
              <button
                onClick={() => setIsConversationOpen(!isConversationOpen)}
                className="w-full px-3 py-1.5 flex items-center justify-between text-left hover:bg-slate-50 transition-colors border-b border-slate-200"
              >
                <div className="flex items-center gap-2">
                  <Send className="h-3 w-3 text-blue-600" />
                  <h3 className="text-sm font-medium text-slate-900">Conversation</h3>
                  <span className="px-1.5 py-0.5 bg-blue-100 text-blue-800 rounded text-xs font-medium">
                    {request.messages.length}
                  </span>
                </div>
                {isConversationOpen ? (
                  <ChevronDown className="h-3 w-3 text-slate-600" />
                ) : (
                  <ChevronRight className="h-3 w-3 text-slate-600" />
                )}
              </button>
              
              {isConversationOpen && (
                <div className="p-3">
                  {/* All Messages in a single flow */}
                  <div className="space-y-2">
                    {request.messages.map((message: unknown, index: number) => (
                      <LLMMessageCard
                        key={`msg-${index}`}
                        message={message as any}
                        index={index}
                        defaultOpen={index === 0 || (message as any).type === 'assistant'}
                      />
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="bg-white rounded-lg border border-slate-200 p-8 text-center">
              <Server className="h-12 w-12 mx-auto mb-3 text-slate-400" />
              <p className="text-slate-600">No conversation data available for this request</p>
            </div>
          )}
          
          {/* Agent Actions Section */}
          <AgentActionsRenderer 
            messages={request.messages || []}
            responseBody={request.responseBody || request.response}
          />
          
          {/* Technical Details */}
          <div className="bg-white rounded-lg border border-slate-200">
            <button
              onClick={() => setIsTechnicalOpen(!isTechnicalOpen)}
              className="w-full px-3 py-1.5 flex items-center justify-between text-left hover:bg-slate-50 transition-colors border-b border-slate-200"
            >
              <div className="flex items-center gap-2">
                <Server className="h-3 w-3 text-slate-600" />
                <h3 className="text-sm font-medium text-slate-900">Technical Details</h3>
              </div>
              {isTechnicalOpen ? (
                <ChevronDown className="h-3 w-3 text-slate-600" />
              ) : (
                <ChevronRight className="h-3 w-3 text-slate-600" />
              )}
            </button>
            
            {isTechnicalOpen && (
              <div className="p-3 space-y-3">
                {/* Full URL - Compact */}
                <div>
                  <label className="text-xs font-medium text-slate-600 block mb-1">Request URL</label>
                  <div className="flex items-center gap-2">
                    <div className="flex-1 p-2 bg-slate-50 rounded font-mono text-xs break-all text-slate-800">
                      {request.url}
                    </div>
                    <a 
                      href={request.url} 
                      target="_blank" 
                      rel="noopener noreferrer"
                      className="p-1 hover:bg-slate-100 rounded transition-colors flex-shrink-0"
                      title="Open in new tab"
                    >
                      <ExternalLink className="h-3 w-3 text-slate-600" />
                    </a>
                  </div>
                </div>
                
                {/* Request Payload */}
                {request.body && (
                  <div>
                    <label className="text-xs font-medium text-slate-600 block mb-1">Request Payload</label>
                    <JsonRenderer 
                      data={request.body} 
                      title="Request Body"
                      maxHeight="250px"
                    />
                  </div>
                )}
                
                {/* Response Payload */}
                {(request.responseBody || request.response) && (
                  <div>
                    <label className="text-xs font-medium text-slate-600 block mb-1">Response Payload</label>
                    {isSSEResponse(request.responseBody || request.response) ? (
                      <StreamEventRenderer 
                        responseBody={request.responseBody || request.response}
                        title="Stream Events"
                        maxHeight="400px"
                      />
                    ) : (
                      <JsonRenderer 
                        data={request.responseBody || request.response} 
                        title="Response Body"
                        maxHeight="250px"
                      />
                    )}
                  </div>
                )}
                
                {/* Headers in side-by-side layout for better space usage */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
                  {/* Request Headers */}
                  {request.headers && Object.keys(request.headers).length > 0 && (
                    <div>
                      <label className="text-xs font-medium text-slate-600 block mb-1">Request Headers</label>
                      <JsonRenderer 
                        data={request.headers} 
                        title="Request Headers"
                        maxHeight="200px"
                      />
                    </div>
                  )}
                  
                  {/* Response Headers */}
                  {request.responseHeaders && Object.keys(request.responseHeaders).length > 0 && (
                    <div>
                      <label className="text-xs font-medium text-slate-600 block mb-1">Response Headers</label>
                      <JsonRenderer 
                        data={request.responseHeaders} 
                        title="Response Headers"
                        maxHeight="200px"
                      />
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
          
          {/* Error Display - Compact */}
          {request.error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-1">
                <XCircle className="h-4 w-4 text-red-600" />
                <h3 className="font-medium text-red-800 text-sm">Error</h3>
              </div>
              <p className="text-red-700 font-mono text-xs">{request.error}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}