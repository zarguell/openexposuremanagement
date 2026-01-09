import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import LoadingSpinner from '../components/LoadingSpinner';
import StatusBadge from '../components/StatusBadge';

function Dashboard() {
  const { data: dashboard, isLoading, error } = useQuery({
    queryKey: ['dashboard'],
    queryFn: apiClient.getDashboard,
    refetchInterval: 30000, // Refetch every 30 seconds
  });

  const { data: intelStatus } = useQuery({
    queryKey: ['intel-status'],
    queryFn: apiClient.getIntelStatus,
    refetchInterval: 60000, // Refetch every minute
  });

  if (isLoading) {
    return (
      <div style={{ padding: '1.5rem 1rem' }}>
        <LoadingSpinner message="Loading dashboard..." />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '1.5rem 1rem' }}>
        <div style={{
          backgroundColor: '#fef2f2',
          border: '1px solid #fecaca',
          borderRadius: '0.5rem',
          padding: '1rem',
          color: '#dc2626'
        }}>
          <h3 style={{ fontSize: '1.125rem', fontWeight: '500', marginBottom: '0.5rem' }}>
            Error loading dashboard
          </h3>
          <p>{error.message}</p>
        </div>
      </div>
    );
  }

  return (
    <div style={{ padding: '1.5rem 1rem' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2.25rem', fontWeight: 'bold', color: '#111827', marginBottom: '0.5rem' }}>
          Dashboard
        </h1>
        <p style={{ color: '#6b7280' }}>
          Overview of your vulnerability management and asset inventory
        </p>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
        gap: '1.5rem',
        marginBottom: '2rem'
      }}>
        {/* Asset Counts */}
        <div style={{
          backgroundColor: 'white',
          borderRadius: '0.5rem',
          padding: '1.5rem',
          boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)'
        }}>
          <h3 style={{ fontSize: '1.125rem', fontWeight: '600', color: '#111827', marginBottom: '1rem' }}>
            Assets
          </h3>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#3b82f6' }}>
                {dashboard?.asset_counts?.total || 0}
              </div>
              <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>Total</div>
            </div>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#10b981' }}>
                {dashboard?.asset_counts?.active || 0}
              </div>
              <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>Active</div>
            </div>
          </div>
        </div>

        {/* Finding Counts */}
        <div style={{
          backgroundColor: 'white',
          borderRadius: '0.5rem',
          padding: '1.5rem',
          boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)'
        }}>
          <h3 style={{ fontSize: '1.125rem', fontWeight: '600', color: '#111827', marginBottom: '1rem' }}>
            Findings
          </h3>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#ef4444' }}>
                {dashboard?.finding_counts?.open || 0}
              </div>
              <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>Open</div>
            </div>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#10b981' }}>
                {dashboard?.finding_counts?.fixed || 0}
              </div>
              <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>Fixed</div>
            </div>
          </div>
        </div>

        {/* Threat Intel Status */}
        <div style={{
          backgroundColor: 'white',
          borderRadius: '0.5rem',
          padding: '1.5rem',
          boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
          gridColumn: '1 / -1'
        }}>
          <h3 style={{ fontSize: '1.125rem', fontWeight: '600', color: '#111827', marginBottom: '1rem' }}>
            Threat Intelligence Status
          </h3>
          <div style={{ display: 'flex', gap: '2rem', flexWrap: 'wrap' }}>
            {intelStatus?.sources && Object.entries(intelStatus.sources).map(([source, status]: [string, any]) => (
              <div key={source} style={{ minWidth: '200px' }}>
                <div style={{
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  color: '#374151',
                  marginBottom: '0.5rem',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em'
                }}>
                  {source.toUpperCase()}
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <StatusBadge
                    status={status.last_sync_at ? 'Active' : 'Inactive'}
                    variant="status"
                    size="sm"
                  />
                  <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                    {status.last_sync_at ?
                      new Date(status.last_sync_at).toLocaleDateString() :
                      'Never synced'
                    }
                  </span>
                </div>
                {status.error && (
                  <div style={{ fontSize: '0.75rem', color: '#dc2626', marginTop: '0.25rem' }}>
                    Error: {status.error}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Recent Activity Placeholder */}
      <div style={{
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)'
      }}>
        <h3 style={{ fontSize: '1.125rem', fontWeight: '600', color: '#111827', marginBottom: '1rem' }}>
          Recent Activity
        </h3>
        <p style={{ color: '#6b7280' }}>
          Recent asset discoveries and finding updates will appear here.
        </p>
      </div>
    </div>
  );
}

export default Dashboard