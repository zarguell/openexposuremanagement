import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { DetailDrawer } from './DetailDrawer';
import { apiClient } from '../api/client';
import LoadingSpinner from './LoadingSpinner';
import './DrawerDetails.css';

export interface Asset {
  id: number;
  hostname: string;
  ip_address?: string;
  os?: string;
  last_observed_at?: string;
  findings_count?: number;
}

export interface AssetSoftware {
  id: number;
  cpe_string: string;
  vendor: string;
  product_name: string;
  version?: string;
  install_path?: string;
  last_seen_at: string;
}

export interface AssetDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  asset?: Asset;
  assetId?: number;
}

export function AssetDrawer({ isOpen, onClose, asset }: AssetDrawerProps) {
  const [showSoftware, setShowSoftware] = useState(false);

  const { data: softwareData, isLoading: softwareLoading } = useQuery({
    queryKey: ['asset-software', asset?.id],
    queryFn: () => apiClient.getSoftwareForAsset(asset!.id),
    enabled: showSoftware && !!asset?.id,
  });

  if (!asset) {
    return (
      <DetailDrawer isOpen={isOpen} onClose={onClose} title="Asset Details">
        <div className="drawer-loading">Loading...</div>
      </DetailDrawer>
    );
  }

  return (
    <DetailDrawer isOpen={isOpen} onClose={onClose} title={`Asset: ${asset.hostname}`}>
      <div className="asset-details">
        <div className="detail-row">
          <div className="detail-label">Hostname:</div>
          <div className="detail-value">{asset.hostname}</div>
        </div>

        {asset.ip_address && (
          <div className="detail-row">
            <div className="detail-label">IP Address:</div>
            <div className="detail-value">{asset.ip_address}</div>
          </div>
        )}

        {asset.os && (
          <div className="detail-row">
            <div className="detail-label">Operating System:</div>
            <div className="detail-value">{asset.os}</div>
          </div>
        )}

        {asset.findings_count !== undefined && (
          <div className="detail-row">
            <div className="detail-label">Findings:</div>
            <div className="detail-value">
              <span className={`findings-count ${asset.findings_count > 0 ? 'has-findings' : ''}`}>
                {asset.findings_count} {asset.findings_count === 1 ? 'finding' : 'findings'}
              </span>
            </div>
          </div>
        )}

        {asset.last_observed_at && (
          <div className="detail-row">
            <div className="detail-label">Last Observed:</div>
            <div className="detail-value">
              {new Date(asset.last_observed_at).toLocaleString()}
            </div>
          </div>
        )}
      </div>

      <div style={{ marginTop: '2rem' }}>
        <button
          onClick={() => setShowSoftware(!showSoftware)}
          style={{
            width: '100%',
            padding: '0.75rem 1rem',
            backgroundColor: '#f9fafb',
            border: '1px solid #e5e7eb',
            borderRadius: '0.5rem',
            fontSize: '0.875rem',
            fontWeight: '600',
            color: '#374151',
            cursor: 'pointer',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <span>Installed Software</span>
          <span style={{ fontSize: '1.25rem' }}>{showSoftware ? '−' : '+'}</span>
        </button>

        {showSoftware && (
          <div style={{ marginTop: '1rem' }}>
            {softwareLoading ? (
              <LoadingSpinner message="Loading software..." />
            ) : softwareData && softwareData.length > 0 ? (
              <div style={{
                backgroundColor: '#f9fafb',
                borderRadius: '0.5rem',
                overflow: 'hidden',
              }}>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid #e5e7eb' }}>
                      <th style={{ padding: '0.75rem 1rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151' }}>
                        Software
                      </th>
                      <th style={{ padding: '0.75rem 1rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151' }}>
                        Version
                      </th>
                      <th style={{ padding: '0.75rem 1rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151' }}>
                        Last Seen
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {softwareData.map((sw: AssetSoftware) => (
                      <tr key={sw.id} style={{ borderBottom: '1px solid #e5e7eb' }}>
                        <td style={{ padding: '0.75rem 1rem', fontSize: '0.875rem' }}>
                          <div style={{ fontWeight: '500', color: '#111827' }}>
                            {sw.vendor} {sw.product_name}
                          </div>
                          {sw.install_path && (
                            <div style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.25rem', fontFamily: 'monospace' }}>
                              {sw.install_path}
                            </div>
                          )}
                        </td>
                        <td style={{ padding: '0.75rem 1rem', fontSize: '0.875rem', color: '#6b7280' }}>
                          {sw.version || '-'}
                        </td>
                        <td style={{ padding: '0.75rem 1rem', fontSize: '0.875rem', color: '#6b7280' }}>
                          {new Date(sw.last_seen_at).toLocaleDateString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div style={{
                backgroundColor: '#f9fafb',
                borderRadius: '0.5rem',
                padding: '2rem',
                textAlign: 'center',
                color: '#6b7280',
                fontSize: '0.875rem',
              }}>
                No software installed on this asset
              </div>
            )}
          </div>
        )}
      </div>
    </DetailDrawer>
  );
}