import React from 'react';
import { useAuth } from './AuthContext';

const Login: React.FC = () => {
  const { login, isDemoMode } = useAuth();

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      backgroundColor: '#f9fafb',
      padding: '1rem'
    }}>
      <div style={{
        backgroundColor: 'white',
        padding: '2rem',
        borderRadius: '0.5rem',
        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
        maxWidth: '32rem',
        width: '100%'
      }}>
        <div style={{ textAlign: 'center' }}>
          <h1 style={{
            fontSize: '2.25rem',
            fontWeight: 'bold',
            color: '#111827',
            marginBottom: '0.5rem'
          }}>
            Open Exposure Management
          </h1>
          <p style={{
            color: '#6b7280',
            marginBottom: '2rem'
          }}>
            {isDemoMode
              ? 'Demo mode: No authentication required'
              : 'Secure access to vulnerability management and asset inventory'
            }
          </p>

          {isDemoMode && (
            <div style={{
              backgroundColor: '#fef3c7',
              border: '1px solid #f59e0b',
              borderRadius: '0.375rem',
              padding: '1rem',
              marginBottom: '1.5rem',
              textAlign: 'left'
            }}>
              <div style={{
                fontSize: '0.875rem',
                fontWeight: '600',
                color: '#92400e',
                marginBottom: '0.5rem'
              }}>
                ⚠️ Demo Mode Warning
              </div>
              <p style={{
                fontSize: '0.875rem',
                color: '#92400e',
                margin: 0
              }}>
                Authentication is disabled for demonstration purposes. This is NOT secure for production use.
                Configure OIDC environment variables to enable proper authentication.
              </p>
            </div>
          )}

          <button
            onClick={login}
            style={{
              width: '100%',
              backgroundColor: isDemoMode ? '#10b981' : '#3b82f6',
              color: 'white',
              padding: '0.75rem 1rem',
              borderRadius: '0.375rem',
              border: 'none',
              fontSize: '1rem',
              fontWeight: '500',
              cursor: 'pointer',
              transition: 'background-color 0.2s'
            }}
            onMouseOver={(e) => {
              e.currentTarget.style.backgroundColor = isDemoMode ? '#059669' : '#2563eb';
            }}
            onMouseOut={(e) => {
              e.currentTarget.style.backgroundColor = isDemoMode ? '#10b981' : '#3b82f6';
            }}
          >
            {isDemoMode ? 'Continue to Demo' : 'Sign In'}
          </button>

          <div style={{
            marginTop: '1.5rem',
            paddingTop: '1rem',
            borderTop: '1px solid #e5e7eb'
          }}>
            <p style={{
              fontSize: '0.875rem',
              color: '#6b7280'
            }}>
              {isDemoMode
                ? 'Demo mode allows full access to all features without authentication.'
                : 'Configure your identity provider in the environment variables to enable authentication.'
              }
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;