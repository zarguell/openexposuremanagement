import React from 'react';

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{ error?: Error; resetError: () => void }>;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error boundary caught an error:', error, errorInfo);
  }

  resetError = () => {
    this.setState({ hasError: false, error: undefined });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        const FallbackComponent = this.props.fallback;
        return <FallbackComponent error={this.state.error} resetError={this.resetError} />;
      }

      return (
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '400px',
          padding: '2rem',
          backgroundColor: '#fef2f2',
          borderRadius: '0.5rem',
          border: '1px solid #fecaca'
        }}>
          <h2 style={{
            fontSize: '1.5rem',
            fontWeight: 'bold',
            color: '#dc2626',
            marginBottom: '1rem'
          }}>
            Something went wrong
          </h2>
          <p style={{
            color: '#7f1d1d',
            marginBottom: '1.5rem',
            textAlign: 'center'
          }}>
            {this.state.error?.message || 'An unexpected error occurred. Please try refreshing the page.'}
          </p>
          <button
            onClick={this.resetError}
            style={{
              backgroundColor: '#dc2626',
              color: 'white',
              padding: '0.5rem 1rem',
              borderRadius: '0.375rem',
              border: 'none',
              fontSize: '0.875rem',
              cursor: 'pointer'
            }}
          >
            Try Again
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;