import React from 'react';

interface PaginationProps {
  currentPage: number;
  totalItems: number;
  itemsPerPage: number;
  onPageChange: (page: number) => void;
}

const Pagination: React.FC<PaginationProps> = ({
  currentPage,
  totalItems,
  itemsPerPage,
  onPageChange
}) => {
  const totalPages = Math.ceil(totalItems / itemsPerPage);
  const startItem = currentPage * itemsPerPage + 1;
  const endItem = Math.min((currentPage + 1) * itemsPerPage, totalItems);

  if (totalPages <= 1) return null;

  const getVisiblePages = () => {
    const pages: number[] = [];
    const maxVisible = 5;

    if (totalPages <= maxVisible) {
      for (let i = 0; i < totalPages; i++) {
        pages.push(i);
      }
    } else {
      pages.push(0);

      if (currentPage > 2) {
        pages.push(-1); // ellipsis
      }

      const start = Math.max(1, currentPage - 1);
      const end = Math.min(totalPages - 2, currentPage + 1);

      for (let i = start; i <= end; i++) {
        pages.push(i);
      }

      if (currentPage < totalPages - 3) {
        pages.push(-1); // ellipsis
      }

      if (totalPages > 1) {
        pages.push(totalPages - 1);
      }
    }

    return pages;
  };

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'between',
      alignItems: 'center',
      padding: '1rem 1.5rem',
      borderTop: '1px solid #e5e7eb',
      backgroundColor: '#f9fafb'
    }}>
      <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>
        Showing {startItem} to {endItem} of {totalItems} results
      </div>

      <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
        <button
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 0}
          style={{
            padding: '0.5rem 1rem',
            border: '1px solid #d1d5db',
            borderRadius: '0.375rem',
            backgroundColor: currentPage === 0 ? '#f9fafb' : 'white',
            color: currentPage === 0 ? '#9ca3af' : '#374151',
            cursor: currentPage === 0 ? 'not-allowed' : 'pointer',
            fontSize: '0.875rem'
          }}
        >
          Previous
        </button>

        <div style={{ display: 'flex', gap: '0.25rem' }}>
          {getVisiblePages().map((page, index) => (
            page === -1 ? (
              <span
                key={`ellipsis-${index}`}
                style={{
                  padding: '0.5rem 0.75rem',
                  color: '#9ca3af',
                  fontSize: '0.875rem'
                }}
              >
                ...
              </span>
            ) : (
              <button
                key={page}
                onClick={() => onPageChange(page)}
                style={{
                  padding: '0.5rem 0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  backgroundColor: currentPage === page ? '#3b82f6' : 'white',
                  color: currentPage === page ? 'white' : '#374151',
                  cursor: 'pointer',
                  fontSize: '0.875rem',
                  minWidth: '2.5rem'
                }}
              >
                {page + 1}
              </button>
            )
          ))}
        </div>

        <button
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage >= totalPages - 1}
          style={{
            padding: '0.5rem 1rem',
            border: '1px solid #d1d5db',
            borderRadius: '0.375rem',
            backgroundColor: currentPage >= totalPages - 1 ? '#f9fafb' : 'white',
            color: currentPage >= totalPages - 1 ? '#9ca3af' : '#374151',
            cursor: currentPage >= totalPages - 1 ? 'not-allowed' : 'pointer',
            fontSize: '0.875rem'
          }}
        >
          Next
        </button>
      </div>
    </div>
  );
};

export default Pagination;