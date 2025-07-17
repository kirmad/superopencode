import { ReactNode } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Database, Home, ArrowLeft } from 'lucide-react'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  const location = useLocation()
  const isDetailPage = location.pathname.includes('/sessions/')

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-card/50 backdrop-blur supports-[backdrop-filter]:bg-card/60">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Database className="h-6 w-6 text-primary" />
              <h1 className="text-xl font-semibold text-foreground">OpenCode Log Viewer</h1>
            </div>
            
            <nav className="flex items-center gap-2">
              {isDetailPage ? (
                <Link
                  to="/"
                  className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
                >
                  <ArrowLeft className="h-4 w-4" />
                  Back to Sessions
                </Link>
              ) : (
                <Link
                  to="/"
                  className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
                >
                  <Home className="h-4 w-4" />
                  Sessions
                </Link>
              )}
            </nav>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6">
        {children}
      </main>
    </div>
  )
}