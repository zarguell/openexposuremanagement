import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import { SoftwareDetails } from '../types/software';
import LoadingSpinner from './LoadingSpinner';
import StatusBadge from './StatusBadge';

interface SoftwareDrawerProps {
  softwareId: number;
  onClose: () => void;
}

function SoftwareDrawer({ softwareId, onClose }: SoftwareDrawerProps) {
  const { data, isLoading, error } = useQuery<SoftwareDetails>({
    queryKey: ['software', softwareId],
    queryFn: () => apiClient.getSoftwareById(softwareId),
  });

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      right: 0,
      width: '600px',
      height: '100vh',
      backgroundColor: 'white',
      boxShadow: '-4px 0 24px rgba(0, 0, 0, 0.15)',
      zIndex: 50,
      display: 'flex',
      flexDirection: 'column',
    }}>
      {/* Header */}
      <div style={{
        padding: '1.5rem',
        borderBottom: '1px solid #e5e7eb',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
      }}>
        <div style={{ flex: 1 }}>
          <h2 style={{
            fontSize: '1.5rem',
            fontWeight: 'bold',
            color: '#111827',
            marginBottom: '0.5rem',
          }}>
            {data?.software.title_formatted || 'Software Details'}
          </h2>
          {data?.software && (
            <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
              {data.software.vendor} • {data.software.product_name}
            </div>
          )}
        </div>
        <button
          onClick={onClose}
          style={{
            background: 'none',
            border: 'none',
            fontSize: '1.5rem',
            color: '#6b7280',
            cursor: 'pointer',
            padding: '0.25rem',
            lineHeight: 1,
          }}
          aria-label="Close"
        >
          ×
        </button>
      </div>

      {/* Content */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '1.5rem' }}>
        {isLoading ? (
          <LoadingSpinner message="Loading software details..." />
        ) : error ? (
          <div style={{
            backgroundColor: '#fef2f2',
            border: '1px solid #fecaca',
            borderRadius: '0.5rem',
            padding: '1rem',
            color: '#dc2626'
          }}>
            Error loading software details: {(error as Error).message}
          </div>
        ) : data ? (
          <div>
            {/* Software Information */}
            <div style={{ marginBottom: '2rem' }}>
              <h3 style={{
                fontSize: '1.125rem',
                fontWeight: '600',
                color: '#111827',
                marginBottom: '1rem',
              }}>
                Software Information
              </h3>
              <div style={{
                backgroundColor: '#f9fafb',
                borderRadius: '0.5rem',
                padding: '1rem',
              }}>
                <InfoRow label="CPE" value={data.software.cpe_string} monospace />
                <InfoRow label="Vendor" value={data.software.vendor} />
                <InfoRow label="Product" value={data.software.product_name} />
                <InfoRow label="Version" value={data.software.version || '-'} />
                <InfoRow label="Edition" value={data.software.edition || '-'} />
                <InfoRow label="Target Hardware" value={data.software.target_hw || '-'} />
                <InfoRow label="Language" value={data.software.lang || '-'} />
              </div>
            </div>

            {/* Affected Findings */}
            <div style={{ marginBottom: '2rem' }}>
              <h3 style={{
                fontSize: '1.125rem',
                fontWeight: '600',
                color: '#111827',
                marginBottom: '1rem',
              }}>
                Affected Findings
              </h3>
              <div style={{
                backgroundColor: '#f9fafb',
                borderRadius: '0.5rem',
                padding: '1rem',
              }}>
                <div style={{ display: 'flex', gap: '2rem', flexWrap: 'wrap' }}>
                  <StatItem
                    label="Total"
                    value={data.affected_findings.total_findings}
                    color="#6b7280"
                  />
                  <StatItem
                    label="Critical"
                    value={data.affected_findings.critical_count}
                    color="#dc2626"
                  />
                  <StatItem
                    label="High"
                    value={data.affected_findings.high_count}
                    color="#ea580c"
                  />
                  <StatItem
                    label="Medium"
                    value={data.affected_findings.medium_count}
                    color="#ca8a04"
                  />
                  <StatItem
                    label="Low"
                    value={data.affected_findings.low_count}
                    color="#65a30d"
                  />
                </div>
              </div>
            </div>

            {/* Affected Assets */}
            <div>
              <h3 style={{
                fontSize: '1.125rem',
                fontWeight: '600',
                color: '#111827',
                marginBottom: '1rem',
              }}>
                Affected Assets ({data.total_assets})
              </h3>
              {data.affected_assets.length === 0 ? (
                <div style={{
                  backgroundColor: '#f9fafb',
                  borderRadius: '0.5rem',
                  padding: '2rem',
                  textAlign: 'center',
                  color: '#6b7280',
                }}>
                  No affected assets
                </div>
              ) : (
                <div style={{
                  backgroundColor: '#f9fafb',
                  borderRadius: '0.5rem',
                  overflow: 'hidden',
                }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                    <thead>
                      <tr style={{ borderBottom: '1px solid #e5e7eb' }}>
                        <th style={{ padding: '0.75rem 1rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151' }}>
                          Asset
                        </th>
                        <th style={{ padding: '0.75rem 1rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151' }}>
                          Status
                        </th>
                        <th style={{ padding: '0.75rem 1rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151' }}>
                          Last Seen
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {data.affected_assets.map((asset) => (
                        <tr key={asset.asset_id} style={{ borderBottom: '1px solid #e5e7eb' }}>
                          <td style={{ padding: '0.75rem 1rem', fontSize: '0.875rem' }}>
                            <div style={{ fontWeight: '500', color: '#111827' }}>
                              {asset.canonical_name}
                            </div>
                            {asset.install_path && (
                              <div style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem' }}>
                                {asset.install_path}
                              </div>
                            )}
                          </td>
                          <td style={{ padding: '0.75rem 1rem' }}>
                            <StatusBadge
                              status={asset.is_active ? 'Active' : 'Inactive'}
                              variant="status"
                            />
                          </td>
                          <td style={{ padding: '0.75rem 1rem', fontSize: '0.875rem', color: '#6b7280' }}>
                            {new Date(asset.last_seen_at).toLocaleDateString()}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}

function InfoRow({ label, value, monospace = false }: { label: string; value: string; monospace?: boolean }) {
  return (
    <div style={{ display: 'flex', marginBottom: '0.75rem', fontSize: '0.875rem' }}>
      <div style={{ width: '140px', color: '#6b7280', flexShrink: 0 }}>
        {label}:
      </div>
      <div style={{
        flex: 1,
        color: '#111827',
        wordBreak: 'break-all',
        fontFamily: monospace ? 'monospace' : 'inherit',
        fontSize: monospace ? '0.75rem' : 'inherit',
      }}>
        {value}
      </div>
    </div>
  );
}

function StatItem({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div>
      <div style={{ fontSize: '0.75rem', color: '#6b7280', marginBottom: '0.25rem' }}>
        {label}
      </div>
      <div style={{ fontSize: '1.5rem', fontWeight: 'bold', color }}>
        {value}
      </div>
    </div>
  );
}

export default SoftwareDrawer;
