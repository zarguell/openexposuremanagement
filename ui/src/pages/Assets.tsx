import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import LoadingSpinner from '../components/LoadingSpinner';
import SearchInput from '../components/SearchInput';
import StatusBadge from '../components/StatusBadge';
import Pagination from '../components/Pagination';

function Assets() {
  const [searchQuery, setSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(0);
  const pageSize = 20;

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['assets', searchQuery, currentPage],
    queryFn: () => apiClient.getAssets({
      query: searchQuery || undefined,
      limit: pageSize,
      offset: currentPage * pageSize,
    }),
  });

  const handleSearch = () => {
    setCurrentPage(0);
    refetch();
  };

  if (isLoading) {
    return (
      <div style={{ padding: '1.5rem 1rem' }}>
        <LoadingSpinner message="Loading assets..." />
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
            Error loading assets
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
          Assets
        </h1>
        <p style={{ color: '#6b7280' }}>
          Search and browse your asset inventory
        </p>
      </div>

      {/* Search */}
      <div style={{
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
        marginBottom: '1.5rem'
      }}>
        <SearchInput
          placeholder="Search assets by hostname or canonical name..."
          value={searchQuery}
          onChange={setSearchQuery}
          onSubmit={handleSearch}
        />
      </div>

      {/* Results */}
      <div style={{
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
        overflow: 'hidden'
      }}>
        {data?.assets?.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: '#6b7280' }}>
            No assets found. Try adjusting your search criteria.
          </div>
        ) : (
          <>
            <div style={{
              borderBottom: '1px solid #e5e7eb',
              padding: '1rem 1.5rem',
              backgroundColor: '#f9fafb'
            }}>
              <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                {data?.pagination?.total || 0} assets found
              </span>
            </div>

            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead style={{ backgroundColor: '#f9fafb' }}>
                  <tr>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Canonical Name
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      First Seen
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Last Seen
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {data?.assets?.map((asset: any) => (
                    <tr key={asset.id} style={{ borderBottom: '1px solid #e5e7eb' }}>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#111827' }}>
                        <div style={{ fontWeight: '500' }}>{asset.canonical_name}</div>
                        {asset.hostname && asset.hostname !== asset.canonical_name && (
                          <div style={{ color: '#6b7280', fontSize: '0.75rem' }}>{asset.hostname}</div>
                        )}
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#6b7280' }}>
                        {asset.first_seen_at ? new Date(asset.first_seen_at).toLocaleDateString() : 'Unknown'}
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#6b7280' }}>
                        {asset.last_seen_at ? new Date(asset.last_seen_at).toLocaleDateString() : 'Unknown'}
                      </td>
                      <td style={{ padding: '1rem 1.5rem' }}>
                        <StatusBadge
                          status={asset.is_active ? 'Active' : 'Inactive'}
                          variant="status"
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {data?.pagination && data.pagination.total > pageSize && (
              <Pagination
                currentPage={currentPage}
                totalItems={data.pagination.total}
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

export default Assets;