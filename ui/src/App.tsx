import { Routes, Route } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Assets from './pages/Assets'
import Findings from './pages/Findings'

function App() {
  return (
    <div style={{ minHeight: '100vh', backgroundColor: '#f9fafb' }}>
      <header style={{ backgroundColor: 'white', boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)' }}>
        <div style={{ maxWidth: '80rem', margin: '0 auto', padding: '0 1rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', height: '4rem', alignItems: 'center' }}>
            <div style={{ display: 'flex' }}>
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <h1 style={{ fontSize: '1.25rem', fontWeight: 'bold', color: '#111827' }}>Open Exposure Management</h1>
              </div>
            </div>
          </div>
        </div>
      </header>

      <main style={{ maxWidth: '80rem', margin: '0 auto', padding: '1.5rem 1rem' }}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/assets" element={<Assets />} />
          <Route path="/findings" element={<Findings />} />
        </Routes>
      </main>
    </div>
  )
}

export default App