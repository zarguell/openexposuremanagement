import { Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from './auth/AuthContext'
import AuthCallback from './auth/AuthCallback'
import ProtectedRoute from './auth/ProtectedRoute'
import Login from './auth/Login'
import ErrorBoundary from './components/ErrorBoundary'
import Dashboard from './pages/Dashboard'
import Assets from './pages/Assets'
import Findings from './pages/Findings'

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
  const { isAuthenticated, logout, user } = useAuth();

  const handleLogout = () => {
    logout();
  };

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
            {isAuthenticated && (
              <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                  Welcome, {user?.profile?.name || user?.profile?.email || 'User'}
                </span>
                <button
                  onClick={handleLogout}
                  style={{
                    backgroundColor: '#ef4444',
                    color: 'white',
                    padding: '0.5rem 1rem',
                    borderRadius: '0.375rem',
                    border: 'none',
                    fontSize: '0.875rem',
                    cursor: 'pointer',
                    transition: 'background-color 0.2s'
                  }}
                  onMouseOver={(e) => {
                    e.currentTarget.style.backgroundColor = '#dc2626';
                  }}
                  onMouseOut={(e) => {
                    e.currentTarget.style.backgroundColor = '#ef4444';
                  }}
                >
                  Sign Out
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
            <Route path="/findings" element={
              <ProtectedRoute>
                <Findings />
              </ProtectedRoute>
            } />
          </Routes>
        </ErrorBoundary>
      </main>
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