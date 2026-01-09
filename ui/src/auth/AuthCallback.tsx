import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { signinCallback } from './authConfig';

const AuthCallback: React.FC = () => {
  const navigate = useNavigate();

  useEffect(() => {
    signinCallback()
      .then(() => {
        // Redirect to dashboard after successful login
        navigate('/', { replace: true });
      })
      .catch((error) => {
        console.error('Login callback error:', error);
        // Redirect to home page on error
        navigate('/', { replace: true });
      });
  }, [navigate]);

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      height: '100vh',
      flexDirection: 'column'
    }}>
      <div style={{ textAlign: 'center' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: '600', color: '#111827', marginBottom: '1rem' }}>
          Completing login...
        </h2>
        <div style={{
          width: '40px',
          height: '40px',
          border: '4px solid #e5e7eb',
          borderTop: '4px solid #3b82f6',
          borderRadius: '50%',
          animation: 'spin 1s linear infinite',
          margin: '0 auto'
        }}></div>
      </div>
      <style>{`
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
};

export default AuthCallback;