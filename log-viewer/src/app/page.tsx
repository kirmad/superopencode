'use client'

import { useState } from 'react'
import { Layout } from '@/components/layout/Layout'
import { SessionList } from '@/components/sessions/SessionList'
import { RequestTimeline } from '@/components/requests/RequestTimeline'
import { RequestDetail } from '@/components/requests/RequestDetail'
import { ArrowLeft } from 'lucide-react'

export default function HomePage() {
  const [selectedSessionId, setSelectedSessionId] = useState<string>('')
  const [selectedRequestId, setSelectedRequestId] = useState<string>('')
  const [currentView, setCurrentView] = useState<'sessions' | 'timeline'>('sessions')
  
  // Handle session selection
  const handleSessionSelect = (sessionId: string) => {
    setSelectedSessionId(sessionId)
    setSelectedRequestId('')
    setCurrentView('timeline')
  }
  
  // Handle back to sessions
  const handleBackToSessions = () => {
    setCurrentView('sessions')
    setSelectedSessionId('')
    setSelectedRequestId('')
  }
  
  return (
    <Layout>
      <div className="h-screen flex overflow-hidden">
        {currentView === 'sessions' ? (
          // Sessions View - Full Screen
          <div className="flex-1 bg-white">
            <SessionList 
              onSessionSelect={handleSessionSelect}
              selectedSessionId={selectedSessionId}
            />
          </div>
        ) : (
          // Timeline View - Split Layout
          <>
            {/* Timeline Sidebar - Compact */}
            <div className="w-80 border-r bg-white flex-shrink-0">
              {/* Back Button Header */}
              <div className="p-3 border-b bg-gray-50">
                <button
                  onClick={handleBackToSessions}
                  className="flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900 transition-colors"
                >
                  <ArrowLeft className="h-4 w-4" />
                  <span>Back to Sessions</span>
                </button>
                <h2 className="text-sm font-medium text-gray-900 mt-1 truncate">
                  Session: {selectedSessionId}
                </h2>
              </div>
              
              {/* Timeline */}
              <RequestTimeline 
                sessionId={selectedSessionId}
                onRequestSelect={setSelectedRequestId}
                selectedRequestId={selectedRequestId}
              />
            </div>
            
            {/* Request Detail - Takes remaining space */}
            <div className="flex-1 bg-gray-50 min-w-0">
              {selectedRequestId ? (
                <RequestDetail 
                  sessionId={selectedSessionId}
                  requestId={selectedRequestId}
                />
              ) : (
                <div className="flex items-center justify-center h-full text-gray-500 p-4 text-center">
                  <div>
                    <div className="text-3xl mb-2">⏱️</div>
                    <p className="text-sm font-medium">Select a request from timeline</p>
                    <p className="text-xs text-gray-400">Click on any request to view its details</p>
                  </div>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </Layout>
  )
}