import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import { SoftwareSummary, SoftwareListResponse } from '../types/software';
import LoadingSpinner from '../components/LoadingSpinner';
import SearchInput from '../components/SearchInput';
import Pagination from '../components/Pagination';
import SoftwareDrawer from '../components/SoftwareDrawer';

function Software() {
  const [searchQuery, setSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(0);
  const [selectedSoftware, setSelectedSoftware] = useState<SoftwareSummary | null>(null);
  const pageSize = 50;

  const { data, isLoading, error, refetch } = useQuery<SoftwareListResponse>({
    queryKey: ['software', searchQuery, currentPage],
    queryFn: () => apiClient.getSoftware({
      product: searchQuery || undefined,
      limit: pageSize,
      offset: currentPage * pageSize,
    }),
  });

  const handleSearch = () => {
    setCurrentPage(0);
    refetch();
  };

  const handleRowClick = (software: SoftwareSummary) => {
    setSelectedSoftware(software);
  };

  const handleCloseDrawer = () => {
    setSelectedSoftware(null);
  };

  if (isLoading) {
    return (
      <div style={{ padding: '1.5rem 1rem' }}>
        <LoadingSpinner message="Loading software inventory..." />
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
            Error loading software
          </h3>
          <p>{(error as Error).message}</p>
        </div>
      </div>
    );
  }

  return (
    <div style={{ padding: '1.5rem 1rem' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2.25rem', fontWeight: 'bold', color: '#111827', marginBottom: '0.5rem' }}>
          Software Inventory
        </h1>
        <p style={{ color: '#6b7280' }}>
          Browse and search software installed across your assets
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
          placeholder="Search software by product name or vendor..."
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
        {data?.software?.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: '#6b7280' }}>
            No software found. Try adjusting your search criteria.
          </div>
        ) : (
          <>
            <div style={{
              borderBottom: '1px solid #e5e7eb',
              padding: '1rem 1.5rem',
              backgroundColor: '#f9fafb'
            }}>
              <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                {data?.pagination?.total || 0} software products found
              </span>
            </div>

            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead style={{ backgroundColor: '#f9fafb' }}>
                  <tr>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Software
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Version
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Vendor
                    </th>
                    <th style={{ padding: '0.75rem 1.5rem', textAlign: 'left', fontSize: '0.875rem', fontWeight: '600', color: '#374151', borderBottom: '1px solid #e5e7eb' }}>
                      Installations
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {data?.software?.map((software) => (
                    <tr
                      key={software.software_id}
                      onClick={() => handleRowClick(software)}
                      style={{ borderBottom: '1px solid #e5e7eb', cursor: 'pointer' }}
                      onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f9fafb'}
                      onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                    >
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#111827' }}>
                        <div style={{ fontWeight: '500' }}>{software.title_formatted}</div>
                        <div style={{ color: '#6b7280', fontSize: '0.75rem', marginTop: '0.25rem' }}>
                          {software.product_name}
                        </div>
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#6b7280' }}>
                        {software.version || '-'}
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#6b7280' }}>
                        {software.vendor}
                      </td>
                      <td style={{ padding: '1rem 1.5rem', fontSize: '0.875rem', color: '#111827' }}>
                        <span style={{
                          backgroundColor: '#dbeafe',
                          color: '#1e40af',
                          padding: '0.25rem 0.75rem',
                          borderRadius: '9999px',
                          fontSize: '0.75rem',
                          fontWeight: '500'
                        }}>
                          {software.install_count}
                        </span>
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

      {/* Software Details Drawer */}
      {selectedSoftware && (
        <SoftwareDrawer
          softwareId={selectedSoftware.software_id}
          onClose={handleCloseDrawer}
        />
      )}
    </div>
  );
}

export default Software;
