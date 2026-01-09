import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import LoadingSpinner from '../components/LoadingSpinner';
import StatusBadge from '../components/StatusBadge';
import Pagination from '../components/Pagination';

function Findings() {
  const [filters, setFilters] = useState({
    source: '',
    severity: '',
    effective_status: '',
    cve: '',
    asset: '',
    include_intel: true,
  });
  const [currentPage, setCurrentPage] = useState(0);
  const pageSize = 20;

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['findings', filters, currentPage],
    queryFn: () => apiClient.getFindings({
      ...filters,
      limit: pageSize,
      offset: currentPage * pageSize,
    }),
  });

  const handleFilterChange = (key: string, value: string | boolean) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setCurrentPage(0);
  };

  const handleSearch = () => {
    refetch();
  };

  if (isLoading) {
    return (
      <div style={{ padding: '1.5rem 1rem' }}>
        <LoadingSpinner message="Loading findings..." />
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
            Error loading findings
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
          Findings
        </h1>
        <p style={{ color: '#6b7280' }}>
          Browse and filter vulnerability findings across your assets
        </p>
      </div>

      {/* Filters */}
      <div style={{
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
        marginBottom: '1.5rem'
      }}>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
          gap: '1rem',
          marginBottom: '1rem'
        }}>
          <div>
            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
              Source
            </label>
            <select
              value={filters.source}
              onChange={(e) => handleFilterChange('source', e.target.value)}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem'
              }}
            >
              <option value="">All Sources</option>
              <option value="tenable">Tenable</option>
              <option value="qualys">Qualys</option>
              <option value="rapid7">Rapid7</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
              Severity
            </label>
            <select
              value={filters.severity}
              onChange={(e) => handleFilterChange('severity', e.target.value)}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem'
              }}
            >
              <option value="">All Severities</option>
              <option value="Critical">Critical</option>
              <option value="High">High</option>
              <option value="Medium">Medium</option>
              <option value="Low">Low</option>
              <option value="Info">Info</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
              Status
            </label>
            <select
              value={filters.effective_status}
              onChange={(e) => handleFilterChange('effective_status', e.target.value)}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem'
              }}
            >
              <option value="">All Statuses</option>
              <option value="open">Open</option>
              <option value="fixed">Fixed</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.5rem' }}>
              CVE
            </label>
            <input
              type="text"
              placeholder="CVE-2023-1234"
              value={filters.cve}
              onChange={(e) => handleFilterChange('cve', e.target.value)}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem'
              }}
            />
          </div>
        </div>

        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.875rem', color: '#374151' }}>
            <input
              type="checkbox"
              checked={filters.include_intel}
              onChange={(e) => handleFilterChange('include_intel', e.target.checked)}
            />
            Include threat intelligence data
          </label>

          <button
            onClick={handleSearch}
            style={{
              backgroundColor: '#3b82f6',
              color: 'white',
              padding: '0.5rem 1rem',
              borderRadius: '0.375rem',
              border: 'none',
              fontSize: '0.875rem',
              cursor: 'pointer'
            }}
          >
            Apply Filters
          </button>
        </div>
      </div>

      {/* Results */}
      <div style={{
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
        overflow: 'hidden'
      }}>
        {data?.findings?.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: '#6b7280' }}>
            No findings found. Try adjusting your filters.
          </div>
        ) : (
          <>
            <div style={{
              borderBottom: '1px solid #e5e7eb',
              padding: '1rem 1.5rem',
              backgroundColor: '#f9fafb'
            }}>
              <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                {data?.total || 0} findings found
              </span>
            </div>

            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead style={{ backgroundColor: '#f9fafb' }}>
                  <tr>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Asset
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Finding
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Severity
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Status
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      CVEs
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Last Seen
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {data?.findings?.map((finding: any) => (
                    <tr key={`${finding.asset_id}-${finding.definition_uid}`} style={{ borderBottom: '1px solid #e5e7eb' }}>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#111827' }}>
                        <div style={{ fontWeight: '500' }}>{finding.asset_canonical_name}</div>
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#111827' }}>
                        <div style={{ fontWeight: '500' }}>{finding.definition_title}</div>
                        <div style={{ color: '#6b7280', fontSize: '0.75rem' }}>
                          {finding.source} - {finding.definition_uid}
                        </div>
                      </td>
                      <td style={{ padding: '1rem 1.5rem' }}>
                        <StatusBadge
                          status={finding.definition_severity}
                          variant="severity"
                        />
                      </td>
                      <td style={{ padding: '1rem 1.5rem' }}>
                        <StatusBadge
                          status={finding.effective_status}
                          variant="status"
                        />
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#6b7280' }}>
                        {finding.cve_aliases?.join(', ') || 'None'}
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#6b7280' }}>
                        {finding.last_observed_at ? new Date(finding.last_observed_at).toLocaleDateString() : 'Unknown'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {data?.total > pageSize && (
              <Pagination
                currentPage={currentPage}
                totalItems={data.total}
                itemsPerPage={pageSize}
                onPageChange={setCurrentPage}
              />
            )}
          </>
        )}
      </div>
    </div>
  );
}

export default Findings