'use client'

import { useState } from 'react'
import { ChevronDown, ChevronRight, Copy, CheckCircle } from 'lucide-react'

interface JsonRendererProps {
  data: any
  title?: string
  maxHeight?: string
}

export function JsonRenderer({ data, title, maxHeight = "300px" }: JsonRendererProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [copied, setCopied] = useState(false)
  
  const jsonString = JSON.stringify(data, null, 2)
  
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(jsonString)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }
  
  if (!data || (typeof data === 'object' && Object.keys(data).length === 0)) {
    return (
      <div className="text-xs text-gray-500 italic p-2 bg-gray-50 rounded">
        No data available
      </div>
    )
  }
  
  return (
    <div className="border border-gray-200 rounded">
      {/* Header with expand/collapse and copy */}
      <div className="flex items-center justify-between p-2 bg-gray-50 border-b border-gray-200">
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className="flex items-center gap-1 text-xs font-medium text-gray-700 hover:text-gray-900"
        >
          {isExpanded ? (
            <ChevronDown className="h-3 w-3" />
          ) : (
            <ChevronRight className="h-3 w-3" />
          )}
          {title || 'JSON Data'}
        </button>
        
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 px-2 py-1 text-xs text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded transition-colors"
          title="Copy JSON"
        >
          {copied ? (
            <>
              <CheckCircle className="h-3 w-3 text-green-600" />
              <span className="text-green-600">Copied</span>
            </>
          ) : (
            <>
              <Copy className="h-3 w-3" />
              <span>Copy</span>
            </>
          )}
        </button>
      </div>
      
      {/* JSON Content */}
      {isExpanded && (
        <div className="p-2">
          <JsonObject 
            data={data} 
            level={0} 
            maxHeight={maxHeight}
          />
        </div>
      )}
    </div>
  )
}

interface JsonObjectProps {
  data: any
  level: number
  maxHeight?: string
}

function JsonObject({ data, level, maxHeight }: JsonObjectProps) {
  const [expandedKeys, setExpandedKeys] = useState<Set<string>>(new Set())
  
  const toggleKey = (key: string) => {
    const newExpanded = new Set(expandedKeys)
    if (newExpanded.has(key)) {
      newExpanded.delete(key)
    } else {
      newExpanded.add(key)
    }
    setExpandedKeys(newExpanded)
  }
  
  const indent = level * 8
  
  if (data === null) {
    return <span className="text-gray-500 italic text-xs">null</span>
  }
  
  if (typeof data === 'string') {
    return <span className="text-green-700 text-xs">&quot;{data}&quot;</span>
  }
  
  if (typeof data === 'number') {
    return <span className="text-blue-700 text-xs">{data}</span>
  }
  
  if (typeof data === 'boolean') {
    return <span className="text-purple-700 text-xs">{data.toString()}</span>
  }
  
  if (Array.isArray(data)) {
    if (data.length === 0) {
      return <span className="text-gray-600 text-xs">[]</span>
    }
    
    return (
      <div className={level === 0 ? `overflow-auto text-xs leading-relaxed` : 'text-xs leading-relaxed'} style={level === 0 ? { maxHeight } : {}}>
        <div className="text-gray-600 text-xs">[</div>
        {data.map((item, index) => (
          <div key={index} style={{ paddingLeft: `${indent + 8}px` }}>
            <JsonObject 
              data={item} 
              level={level + 1}
            />
            {index < data.length - 1 && <span className="text-gray-600 text-xs">,</span>}
          </div>
        ))}
        <div style={{ paddingLeft: `${indent}px` }} className="text-gray-600 text-xs">]</div>
      </div>
    )
  }
  
  if (typeof data === 'object' && data !== null) {
    const keys = Object.keys(data)
    if (keys.length === 0) {
      return <span className="text-gray-600 text-xs">{"{}"}</span>
    }
    
    return (
      <div className={level === 0 ? `overflow-auto text-xs leading-relaxed` : 'text-xs leading-relaxed'} style={level === 0 ? { maxHeight } : {}}>
        <div className="text-gray-600 text-xs">{"{"}</div>
        {keys.map((key, index) => {
          const value = data[key]
          const isExpandable = typeof value === 'object' && value !== null && 
            (Array.isArray(value) ? value.length > 0 : Object.keys(value).length > 0)
          const isExpanded = expandedKeys.has(key)
          const isLast = index === keys.length - 1
          
          return (
            <div key={key} style={{ paddingLeft: `${indent + 8}px` }}>
              <div className="flex items-start gap-1">
                {isExpandable && (
                  <button
                    onClick={() => toggleKey(key)}
                    className="mt-0.5 hover:bg-gray-100 rounded p-0.5"
                  >
                    {isExpanded ? (
                      <ChevronDown className="h-2.5 w-2.5 text-gray-500" />
                    ) : (
                      <ChevronRight className="h-2.5 w-2.5 text-gray-500" />
                    )}
                  </button>
                )}
                {!isExpandable && <div className="w-3" />}
                
                <div className="flex-1 min-w-0">
                  <span className="text-blue-800 font-medium text-xs">&quot;{key}&quot;</span>
                  <span className="text-gray-600 mx-1 text-xs">:</span>
                  
                  {isExpandable ? (
                    <div>
                      {!isExpanded ? (
                        <span className="text-gray-500 text-xs">
                          {Array.isArray(value) 
                            ? `[${value.length}]` 
                            : `{${Object.keys(value).length}}`
                          }
                        </span>
                      ) : (
                        <div className="mt-1">
                          <JsonObject 
                            data={value} 
                            level={level + 1}
                          />
                        </div>
                      )}
                    </div>
                  ) : (
                    <JsonObject 
                      data={value} 
                      level={level + 1}
                    />
                  )}
                  
                  {!isLast && <span className="text-gray-600 text-xs">,</span>}
                </div>
              </div>
            </div>
          )
        })}
        <div style={{ paddingLeft: `${indent}px` }} className="text-gray-600 text-xs">{"}"}</div>
      </div>
    )
  }
  
  return <span className="text-xs">{String(data)}</span>
}