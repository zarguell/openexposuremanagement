import React from 'react';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  color?: string;
  message?: string;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'md',
  color = '#3b82f6',
  message
}) => {
  const sizeStyles = {
    sm: { width: '20px', height: '20px' },
    md: { width: '40px', height: '40px' },
    lg: { width: '60px', height: '60px' }
  };

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '1rem'
    }}>
      <div
        style={{
          ...sizeStyles[size],
          border: `4px solid #e5e7eb`,
          borderTop: `4px solid ${color}`,
          borderRadius: '50%',
          animation: 'spin 1s linear infinite'
        }}
      />
      {message && (
        <p style={{
          color: '#6b7280',
          fontSize: '0.875rem',
          textAlign: 'center',
          margin: 0
        }}>
          {message}
        </p>
      )}
      <style>{`
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
};

export default LoadingSpinner;