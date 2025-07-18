import { NextResponse } from 'next/server'
import { getSessionMetadata } from '@/lib/database'

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)
    
    const filters = {
      startTime: searchParams.get('startTime') || undefined,
      endTime: searchParams.get('endTime') || undefined,
      hasError: searchParams.get('hasError') ? 
        searchParams.get('hasError') === 'true' : undefined,
      limit: searchParams.get('limit') ? 
        parseInt(searchParams.get('limit')!) : 100
    }
    
    const sessions = getSessionMetadata(filters)
    
    return NextResponse.json(sessions)
  } catch (error) {
    console.error('Error fetching sessions:', error)
    return NextResponse.json(
      { error: 'Failed to fetch sessions' },
      { status: 500 }
    )
  }
}