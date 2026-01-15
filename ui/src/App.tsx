import { Routes, Route, Link } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from './auth/AuthContext'
import AuthCallback from './auth/AuthCallback'
import ProtectedRoute from './auth/ProtectedRoute'
import Login from './auth/Login'
import ErrorBoundary from './components/ErrorBoundary'
import DebugPanel from './components/DebugPanel'
import Dashboard from './pages/Dashboard'
import Assets from './pages/Assets'
import AssetsQuery from './pages/AssetsQuery'
import Findings from './pages/Findings'
import FindingsQuery from './pages/FindingsQuery'
import Software from './pages/Software'
import UnifiedQueries from './pages/UnifiedQueries'

// Create a client for React Query
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error: any) => {
        // Don't retry on 401/403 errors
        if (error?.status === 401 || error?.status === 403) {
          return false;
        }
        return failureCount < 3;
      },
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
});

function AppContent() {
  const { isAuthenticated, logout, user, isDemoMode } = useAuth();

  const handleLogout = () => {
    logout();
  };

  return (
    <div style={{ minHeight: '100vh', backgroundColor: '#f9fafb' }}>
      {isDemoMode && (
        <div style={{
          backgroundColor: '#fef3c7',
          borderBottom: '1px solid #f59e0b',
          padding: '0.5rem 1rem',
          textAlign: 'center'
        }}>
          <span style={{
            fontSize: '0.875rem',
            fontWeight: '500',
            color: '#92400e'
          }}>
            🔓 Demo Mode: Authentication disabled for demonstration purposes. Not secure for production use.
          </span>
        </div>
      )}
      <header style={{ backgroundColor: 'white', boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)' }}>
        <div style={{ maxWidth: '80rem', margin: '0 auto', padding: '0 1rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', height: '4rem', alignItems: 'center' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
              <Link to="/" style={{ textDecoration: 'none' }}>
                <h1 style={{ fontSize: '1.25rem', fontWeight: 'bold', color: '#111827' }}>Open Exposure Management</h1>
              </Link>
              {isAuthenticated && (
                <nav style={{ display: 'flex', gap: '1rem' }}>
                  <Link
                    to="/"
                    style={{
                      fontSize: '0.875rem',
                      color: '#6b7280',
                      textDecoration: 'none',
                      padding: '0.5rem',
                      borderRadius: '0.375rem',
                      transition: 'all 0.2s',
                    }}
                    onMouseOver={(e) => {
                      e.currentTarget.style.backgroundColor = '#f3f4f6';
                      e.currentTarget.style.color = '#111827';
                    }}
                    onMouseOut={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                      e.currentTarget.style.color = '#6b7280';
                    }}
                  >
                    Home
                  </Link>
                  <Link
                    to="/findings/query"
                    style={{
                      fontSize: '0.875rem',
                      color: '#6b7280',
                      textDecoration: 'none',
                      padding: '0.5rem',
                      borderRadius: '0.375rem',
                      transition: 'all 0.2s',
                    }}
                    onMouseOver={(e) => {
                      e.currentTarget.style.backgroundColor = '#f3f4f6';
                      e.currentTarget.style.color = '#111827';
                    }}
                    onMouseOut={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                      e.currentTarget.style.color = '#6b7280';
                    }}
                  >
                    Findings
                  </Link>
                  <Link
                    to="/assets/query"
                    style={{
                      fontSize: '0.875rem',
                      color: '#6b7280',
                      textDecoration: 'none',
                      padding: '0.5rem',
                      borderRadius: '0.375rem',
                      transition: 'all 0.2s',
                    }}
                    onMouseOver={(e) => {
                      e.currentTarget.style.backgroundColor = '#f3f4f6';
                      e.currentTarget.style.color = '#111827';
                    }}
                    onMouseOut={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                      e.currentTarget.style.color = '#6b7280';
                    }}
                  >
                    Assets
                  </Link>
                  <Link
                    to="/software"
                    style={{
                      fontSize: '0.875rem',
                      color: '#6b7280',
                      textDecoration: 'none',
                      padding: '0.5rem',
                      borderRadius: '0.375rem',
                      transition: 'all 0.2s',
                    }}
                    onMouseOver={(e) => {
                      e.currentTarget.style.backgroundColor = '#f3f4f6';
                      e.currentTarget.style.color = '#111827';
                    }}
                    onMouseOut={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                      e.currentTarget.style.color = '#6b7280';
                    }}
                  >
                    Software
                  </Link>
                  <Link
                    to="/unified-queries"
                    style={{
                      fontSize: '0.875rem',
                      color: '#6b7280',
                      textDecoration: 'none',
                      padding: '0.5rem',
                      borderRadius: '0.375rem',
                      transition: 'all 0.2s',
                    }}
                    onMouseOver={(e) => {
                      e.currentTarget.style.backgroundColor = '#f3f4f6';
                      e.currentTarget.style.color = '#111827';
                    }}
                    onMouseOut={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                      e.currentTarget.style.color = '#6b7280';
                    }}
                  >
                    Unified Queries
                  </Link>
                </nav>
              )}
            </div>
            {isAuthenticated && (
              <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                  Welcome, {user?.profile?.name || user?.profile?.email || 'User'}
                  {isDemoMode && (
                    <span style={{
                      marginLeft: '0.5rem',
                      padding: '0.125rem 0.5rem',
                      backgroundColor: '#fef3c7',
                      color: '#92400e',
                      borderRadius: '9999px',
                      fontSize: '0.75rem',
                      fontWeight: '500'
                    }}>
                      DEMO
                    </span>
                  )}
                </span>
                <button
                  onClick={handleLogout}
                  style={{
                    backgroundColor: isDemoMode ? '#6b7280' : '#ef4444',
                    color: 'white',
                    padding: '0.5rem 1rem',
                    borderRadius: '0.375rem',
                    border: 'none',
                    fontSize: '0.875rem',
                    cursor: 'pointer',
                    transition: 'background-color 0.2s'
                  }}
                  onMouseOver={(e) => {
                    e.currentTarget.style.backgroundColor = isDemoMode ? '#4b5563' : '#dc2626';
                  }}
                  onMouseOut={(e) => {
                    e.currentTarget.style.backgroundColor = isDemoMode ? '#6b7280' : '#ef4444';
                  }}
                >
                  {isDemoMode ? 'Exit Demo' : 'Sign Out'}
                </button>
              </div>
            )}
          </div>
        </div>
      </header>

      <main style={{ maxWidth: '80rem', margin: '0 auto', padding: '1.5rem 1rem' }}>
        <ErrorBoundary>
          <Routes>
            <Route path="/auth/callback" element={<AuthCallback />} />
            <Route path="/" element={
              isAuthenticated ? (
                <ProtectedRoute>
                  <Dashboard />
                </ProtectedRoute>
              ) : (
                <Login />
              )
            } />
            <Route path="/assets" element={
              <ProtectedRoute>
                <Assets />
              </ProtectedRoute>
            } />
            <Route path="/assets/query" element={
              <ProtectedRoute>
                <AssetsQuery />
              </ProtectedRoute>
            } />
            <Route path="/findings" element={
              <ProtectedRoute>
                <Findings />
              </ProtectedRoute>
            } />
            <Route path="/findings/query" element={
              <ProtectedRoute>
                <FindingsQuery />
              </ProtectedRoute>
            } />
            <Route path="/software" element={
              <ProtectedRoute>
                <Software />
              </ProtectedRoute>
            } />
            <Route path="/unified-queries" element={
              <ProtectedRoute>
                <UnifiedQueries />
              </ProtectedRoute>
            } />
          </Routes>
        </ErrorBoundary>
      </main>
      <DebugPanel visible={isDemoMode} />
    </div>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <AppContent />
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App