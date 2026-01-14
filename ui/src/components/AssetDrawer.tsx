import { DetailDrawer } from './DetailDrawer';
import './DrawerDetails.css';

export interface Asset {
  id: number;
  hostname: string;
  ip_address?: string;
  os?: string;
  last_observed_at?: string;
  findings_count?: number;
}

export interface AssetDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  asset?: Asset;
  assetId?: number;
}

export function AssetDrawer({ isOpen, onClose, asset }: AssetDrawerProps) {
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
    </DetailDrawer>
  );
}