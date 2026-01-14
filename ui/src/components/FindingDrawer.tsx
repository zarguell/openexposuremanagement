import { DetailDrawer } from './DetailDrawer';
import StatusBadge from './StatusBadge';
import './DrawerDetails.css';

export interface Finding {
  id: number;
  title: string;
  severity: string;
  effective_status: string;
  cve_id?: string;
  description?: string;
  epss_score?: number;
  is_kev?: boolean;
  last_observed_at?: string;
}

export interface FindingDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  finding?: Finding;
  findingId?: number;
}

export function FindingDrawer({ isOpen, onClose, finding }: FindingDrawerProps) {
  if (!finding) {
    return (
      <DetailDrawer isOpen={isOpen} onClose={onClose} title="Finding Details">
        <div className="drawer-loading">Loading...</div>
      </DetailDrawer>
    );
  }

  return (
    <DetailDrawer isOpen={isOpen} onClose={onClose} title={`Finding: ${finding.title}`}>
      <div className="finding-details">
        <div className="detail-row">
          <div className="detail-label">Severity:</div>
          <div className="detail-value">
            <StatusBadge status={finding.severity || 'unknown'} variant="severity" />
          </div>
        </div>

        <div className="detail-row">
          <div className="detail-label">Status:</div>
          <div className="detail-value">
            <StatusBadge status={finding.effective_status || 'unknown'} variant="status" />
          </div>
        </div>

        {finding.cve_id && (
          <div className="detail-row">
            <div className="detail-label">CVE ID:</div>
            <div className="detail-value">
              <a
                href={`https://nvd.nist.gov/vuln/detail/${finding.cve_id}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                {finding.cve_id}
              </a>
            </div>
          </div>
        )}

        {finding.description && (
          <div className="detail-row">
            <div className="detail-label">Description:</div>
            <div className="detail-value">{finding.description}</div>
          </div>
        )}

        {finding.epss_score !== undefined && (
          <div className="detail-row">
            <div className="detail-label">EPSS Score:</div>
            <div className="detail-value">
              {finding.epss_score.toFixed(3)}
              {finding.epss_score > 0.9 && (
                <span className="epss-badge epss-high">High Exploitability</span>
              )}
            </div>
          </div>
        )}

        {finding.is_kev && (
          <div className="detail-row">
            <div className="detail-label">CISA KEV:</div>
            <div className="detail-value">
              <span className="kev-badge">
                ⚠️ Vulnerability included in CISA Known Exploited Vulnerabilities catalog
              </span>
            </div>
          </div>
        )}

        {finding.last_observed_at && (
          <div className="detail-row">
            <div className="detail-label">Last Observed:</div>
            <div className="detail-value">
              {new Date(finding.last_observed_at).toLocaleString()}
            </div>
          </div>
        )}
      </div>
    </DetailDrawer>
  );
}