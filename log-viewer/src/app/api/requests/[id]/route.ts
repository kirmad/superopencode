import { NextResponse } from 'next/server'
import { getHTTPRequest } from '@/lib/database'
import { parseHTTPRequest } from '@/lib/parsers'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { searchParams } = new URL(request.url)
    const sessionId = searchParams.get('sessionId')
    const resolvedParams = await params
    
    if (!sessionId) {
      return NextResponse.json(
        { error: 'Session ID required' },
        { status: 400 }
      )
    }
    
    const httpRequest = getHTTPRequest(sessionId, resolvedParams.id)
    
    if (!httpRequest) {
      return NextResponse.json(
        { error: 'Request not found' },
        { status: 404 }
      )
    }
    
    // Parse the request into human-readable messages
    const messages = parseHTTPRequest(httpRequest)
    
    return NextResponse.json({
      ...httpRequest,
      messages
    })
  } catch (error) {
    console.error('Error fetching request:', error)
    return NextResponse.json(
      { error: 'Failed to fetch request' },
      { status: 500 }
    )
  }
}