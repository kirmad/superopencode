'use client'

import { ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes (was cacheTime)
    },
  },
})

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white shadow-sm border-b">
          <div className="max-w-7xl mx-auto px-3 py-3">
            <div className="flex items-center space-x-2">
              <div className="w-7 h-7 bg-blue-600 rounded-lg flex items-center justify-center flex-shrink-0">
                <span className="text-white text-base">🔍</span>
              </div>
              <h1 className="text-lg md:text-xl lg:text-2xl font-bold text-gray-900 truncate">
                HTTP Request Log Viewer
              </h1>
            </div>
          </div>
        </header>
        
        <main className="max-w-7xl mx-auto h-full">
          {children}
        </main>
      </div>
      
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  )
}