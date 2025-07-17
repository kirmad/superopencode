import { useParams } from 'react-router-dom'

export function SessionDetailSimple() {
  const { sessionId } = useParams<{ sessionId: string }>()
  
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold">Session Detail</h1>
      <p>Session ID: {sessionId}</p>
      <p>This is a simple test component</p>
    </div>
  )
}