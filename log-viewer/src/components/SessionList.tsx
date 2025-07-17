import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { format } from 'date-fns'
import { 
  Search, 
  Filter, 
  AlertTriangle, 
  CheckCircle,
  Clock,
  Activity,
  DollarSign,
  Zap,
  Globe
} from 'lucide-react'
import { apiService } from '@/services/api'
import { SessionFilters } from '@/types'

export function SessionList() {
  const [filters, setFilters] = useState<SessionFilters>({
    limit: 50,
    offset: 0
  })

  const { data: sessions, isLoading, error } = useQuery({
    queryKey: ['sessions', filters],
    queryFn: () => apiService.getSessions(filters),
  })

  const handleSearch = (search: string) => {
    setFilters(prev => ({ ...prev, search, offset: 0 }))
  }

  const handleErrorFilter = (hasError?: boolean) => {
    setFilters(prev => ({ ...prev, has_error: hasError, offset: 0 }))
  }

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
          <AlertTriangle className="h-12 w-12 text-destructive mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-foreground mb-2">Failed to load sessions</h3>
          <p className="text-muted-foreground">Please check if the database is accessible</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <input
            type="text"
            placeholder="Search sessions..."
            className="w-full pl-10 pr-4 py-2 border border-input bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
            onChange={(e) => handleSearch(e.target.value)}
          />
        </div>
        
        <div className="flex gap-2">
          <button
            onClick={() => handleErrorFilter(undefined)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              filters.has_error === undefined 
                ? 'bg-primary text-primary-foreground' 
                : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            }`}
          >
            All
          </button>
          <button
            onClick={() => handleErrorFilter(false)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              filters.has_error === false 
                ? 'bg-primary text-primary-foreground' 
                : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            }`}
          >
            <CheckCircle className="h-4 w-4 inline mr-1" />
            Success
          </button>
          <button
            onClick={() => handleErrorFilter(true)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              filters.has_error === true 
                ? 'bg-primary text-primary-foreground' 
                : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            }`}
          >
            <AlertTriangle className="h-4 w-4 inline mr-1" />
            Errors
          </button>
        </div>
      </div>

      <div className="grid gap-4">
        {sessions?.map((session) => (
          <Link
            key={session.id}
            to={`/sessions/${session.session_id}`}
            className="block p-6 bg-card rounded-lg border border-border hover:bg-accent transition-colors"
          >
            <div className="flex items-start justify-between mb-4">
              <div>
                <h3 className="text-lg font-semibold text-foreground mb-1">
                  Session {session.session_id}
                </h3>
                <p className="text-sm text-muted-foreground">
                  {format(new Date(session.start_time), 'PPpp')}
                  {session.end_time && (
                    <span className="ml-2">
                      • Duration: {Math.round((new Date(session.end_time).getTime() - new Date(session.start_time).getTime()) / 1000)}s
                    </span>
                  )}
                </p>
              </div>
              
              <div className="flex items-center gap-2">
                {session.has_error ? (
                  <div className="flex items-center gap-1 text-destructive">
                    <AlertTriangle className="h-4 w-4" />
                    <span className="text-xs font-medium">Error</span>
                  </div>
                ) : (
                  <div className="flex items-center gap-1 text-green-600">
                    <CheckCircle className="h-4 w-4" />
                    <span className="text-xs font-medium">Success</span>
                  </div>
                )}
              </div>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
              <div className="flex items-center gap-2">
                <Activity className="h-4 w-4 text-muted-foreground" />
                <span className="text-muted-foreground">LLM:</span>
                <span className="font-medium">{session.llm_call_count}</span>
              </div>
              
              <div className="flex items-center gap-2">
                <Zap className="h-4 w-4 text-muted-foreground" />
                <span className="text-muted-foreground">Tools:</span>
                <span className="font-medium">{session.tool_call_count}</span>
              </div>
              
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4 text-muted-foreground" />
                <span className="text-muted-foreground">HTTP:</span>
                <span className="font-medium">{session.http_call_count}</span>
              </div>
              
              <div className="flex items-center gap-2">
                <DollarSign className="h-4 w-4 text-muted-foreground" />
                <span className="text-muted-foreground">Cost:</span>
                <span className="font-medium">${session.total_cost.toFixed(4)}</span>
              </div>
            </div>
          </Link>
        ))}
      </div>

      {sessions?.length === 0 && (
        <div className="text-center py-12">
          <Clock className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-foreground mb-2">No sessions found</h3>
          <p className="text-muted-foreground">Try adjusting your search or filters</p>
        </div>
      )}
    </div>
  )
}