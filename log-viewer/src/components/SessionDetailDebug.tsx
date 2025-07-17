import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { apiService } from '@/services/api'

export function SessionDetailDebug() {
  const { sessionId } = useParams<{ sessionId: string }>()
  
  console.log('SessionDetailDebug rendered with sessionId:', sessionId)
  
  const { data: session, isLoading, error } = useQuery({
    queryKey: ['session', sessionId],
    queryFn: () => {
      console.log('Fetching session:', sessionId)
      return apiService.getSession(sessionId!)
    },
    enabled: !!sessionId,
  })
  
  console.log('SessionDetailDebug state:', { isLoading, error, hasData: !!session })
  
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <div className="text-center">
          <h3 className="text-lg font-semibold text-foreground mb-2">Error loading session</h3>
          <p className="text-muted-foreground">{error.toString()}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold">Session Detail (Debug)</h1>
      <p>Session ID: {sessionId}</p>
      <p>Session loaded: {session ? 'Yes' : 'No'}</p>
      {session && (
        <div>
          <p>Start time: {session.start_time}</p>
          <p>LLM calls: {session.llm_calls.length}</p>
          <p>Tool calls: {session.tool_calls.length}</p>
          <p>HTTP calls: {session.http_calls.length}</p>
        </div>
      )}
    </div>
  )
}