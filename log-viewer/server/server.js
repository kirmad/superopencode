import express from 'express'
import cors from 'cors'
import sqlite3 from 'sqlite3'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'
import { readFileSync, existsSync } from 'fs'
import { homedir } from 'os'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

const app = express()
const PORT = 3001

app.use(cors())
app.use(express.json())

const LOG_DIR = join(homedir(), '.opencode', 'detailed_logs')
const DB_PATH = join(LOG_DIR, 'sessions.db')

let db = null

function initDatabase() {
  if (!existsSync(DB_PATH)) {
    console.error(`Database not found at ${DB_PATH}`)
    console.error('Make sure you have run OpenCode with --detailed-logs flag to generate logs')
    return false
  }
  
  db = new sqlite3.Database(DB_PATH, (err) => {
    if (err) {
      console.error('Error opening database:', err.message)
      return false
    }
    console.log('Connected to SQLite database')
  })
  
  return true
}

app.get('/api/sessions', (req, res) => {
  if (!db) {
    return res.status(500).json({ error: 'Database not initialized' })
  }

  const {
    start_time,
    end_time,
    has_error,
    search,
    limit = 50,
    offset = 0
  } = req.query

  let query = `
    SELECT id, session_id, start_time, end_time, llm_call_count, 
           tool_call_count, http_call_count, total_tokens, total_cost,
           has_error, metadata, created_at
    FROM sessions
    WHERE 1=1
  `
  
  const params = []

  if (start_time) {
    query += ' AND start_time >= ?'
    params.push(start_time)
  }

  if (end_time) {
    query += ' AND start_time <= ?'
    params.push(end_time)
  }

  if (has_error !== undefined) {
    query += ' AND has_error = ?'
    params.push(has_error === 'true' ? 1 : 0)
  }

  if (search) {
    query += ' AND (session_id LIKE ? OR metadata LIKE ?)'
    params.push(`%${search}%`, `%${search}%`)
  }

  query += ' ORDER BY start_time DESC LIMIT ? OFFSET ?'
  params.push(parseInt(limit), parseInt(offset))

  db.all(query, params, (err, rows) => {
    if (err) {
      console.error('Database query error:', err)
      return res.status(500).json({ error: 'Database query failed' })
    }
    res.json(rows)
  })
})

app.get('/api/sessions/:sessionId', (req, res) => {
  const { sessionId } = req.params
  const sessionFile = join(LOG_DIR, `${sessionId}.json`)
  
  if (!existsSync(sessionFile)) {
    return res.status(404).json({ error: 'Session not found' })
  }
  
  try {
    const sessionData = readFileSync(sessionFile, 'utf8')
    const session = JSON.parse(sessionData)
    res.json(session)
  } catch (error) {
    console.error('Error reading session file:', error)
    res.status(500).json({ error: 'Failed to read session file' })
  }
})

app.get('/health', (req, res) => {
  res.json({ 
    status: 'ok', 
    database: db ? 'connected' : 'disconnected',
    logDir: LOG_DIR,
    dbPath: DB_PATH,
    dbExists: existsSync(DB_PATH)
  })
})

if (initDatabase()) {
  app.listen(PORT, () => {
    console.log(`Log viewer server running on http://localhost:${PORT}`)
    console.log(`Log directory: ${LOG_DIR}`)
    console.log(`Database: ${DB_PATH}`)
  })
} else {
  console.error('Failed to initialize database. Server not started.')
  process.exit(1)
}