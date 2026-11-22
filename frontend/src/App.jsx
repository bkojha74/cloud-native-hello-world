import { useState, useEffect, useCallback } from 'react'
import axios from 'axios'
import './App.css'

function App() {
  const [message, setMessage] = useState('')
  const [source, setSource] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [useCache, setUseCache] = useState(true)
  const [lastFetchedAt, setLastFetchedAt] = useState(null)
  const [latency, setLatency] = useState(null)

  const fetchMessage = useCallback(async () => {
    setLoading(true)
    setError(null)
    const startTime = performance.now()

    try {
      const cacheParam = useCache ? '1' : '0'
      const response = await axios.get(`/api/message?fromdb=1&cache=${cacheParam}&sql=0`)
      const endTime = performance.now()

      setMessage(response.data.message)
      setSource(response.data.source)
      setLatency(Math.round(endTime - startTime))
      setLastFetchedAt(new Date().toLocaleTimeString())
    } catch (err) {
      console.error('API Error:', err)
      setError('Unable to load message.')
      setMessage('')
      setSource('')
      setLatency(null)
    } finally {
      setLoading(false)
    }
  }, [useCache])

  useEffect(() => {
    fetchMessage()
  }, [fetchMessage])

  return (
    <div className="container">
      <header className="header">
        <h1>Cloud-Native Hello World</h1>
        <p className="subtitle">React + Go (Gin) + Redis + MongoDB + Prometheus</p>
      </header>

      <main className="card">
        {loading && <div className="loader">Loading message...</div>}

        {error && (
          <div className="error-banner" role="alert">
            <span className="error-icon">⚠️</span> {error}
          </div>
        )}

        {!loading && !error && message && (
          <div className="message-content">
            <h2 className="greeting">{message}</h2>
            <div className="metadata">
              <span className={`badge badge-${source}`}>
                Source: <strong>{source}</strong>
              </span>
              {latency !== null && (
                <span className="badge badge-latency">
                  Latency: <strong>{latency} ms</strong>
                </span>
              )}
            </div>
          </div>
        )}

        <div className="controls">
          <label className="toggle-label">
            <input
              type="checkbox"
              checked={useCache}
              onChange={(e) => setUseCache(e.target.checked)}
            />
            <span>Enable Redis Cache (<code>cache={useCache ? '1' : '0'}</code>)</span>
          </label>

          <button
            className="btn-fetch"
            onClick={fetchMessage}
            disabled={loading}
          >
            {loading ? 'Fetching...' : 'Fetch Message Again'}
          </button>
        </div>

        {lastFetchedAt && (
          <footer className="footer-meta">
            Last checked at: {lastFetchedAt}
          </footer>
        )}
      </main>
    </div>
  )
}

export default App
