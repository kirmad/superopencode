import { Routes, Route } from 'react-router-dom'
import { SessionList } from '@/components/SessionList'
import { SessionDetailSimplified as SessionDetail } from '@/components/SessionDetailSimplified'
import { Layout } from '@/components/Layout'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<SessionList />} />
        <Route path="/sessions/:sessionId" element={<SessionDetail />} />
      </Routes>
    </Layout>
  )
}

export default App