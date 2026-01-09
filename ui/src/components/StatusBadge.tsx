import React from 'react';

interface StatusBadgeProps {
  status: string;
  variant?: 'severity' | 'status' | 'default';
  size?: 'sm' | 'md';
}

const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  variant = 'default',
  size = 'md'
}) => {
  const getStatusConfig = (status: string, variant: string) => {
    const statusLower = status?.toLowerCase();

    switch (variant) {
      case 'severity':
        switch (statusLower) {
          case 'critical': return { bg: '#dc2626', text: '#ffffff' };
          case 'high': return { bg: '#ea580c', text: '#ffffff' };
          case 'medium': return { bg: '#d97706', text: '#ffffff' };
          case 'low': return { bg: '#65a30d', text: '#ffffff' };
          case 'info': return { bg: '#0891b2', text: '#ffffff' };
          default: return { bg: '#6b7280', text: '#ffffff' };
        }
      case 'status':
        switch (statusLower) {
          case 'open': return { bg: '#dc2626', text: '#ffffff' };
          case 'fixed': return { bg: '#10b981', text: '#ffffff' };
          case 'active': return { bg: '#10b981', text: '#ffffff' };
          case 'inactive': return { bg: '#6b7280', text: '#ffffff' };
          default: return { bg: '#6b7280', text: '#ffffff' };
        }
      default:
        return { bg: '#e5e7eb', text: '#374151' };
    }
  };

  const config = getStatusConfig(status, variant);
  const sizeStyles = size === 'sm'
    ? { padding: '0.125rem 0.5rem', fontSize: '0.75rem' }
    : { padding: '0.25rem 0.75rem', fontSize: '0.875rem' };

  return (
    <span style={{
      ...sizeStyles,
      borderRadius: '9999px',
      fontWeight: '500',
      backgroundColor: config.bg,
      color: config.text,
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      textTransform: 'capitalize'
    }}>
      {status}
    </span>
  );
};

export default StatusBadge;