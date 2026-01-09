import React from 'react';

interface Column<T> {
  key: keyof T | string;
  header: string;
  render?: (value: any, item: T) => React.ReactNode;
  sortable?: boolean;
  width?: string;
}

interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  loading?: boolean;
  emptyMessage?: string;
  onRowClick?: (item: T) => void;
}

function DataTable<T extends Record<string, any>>({
  data,
  columns,
  loading = false,
  emptyMessage = 'No data available',
  onRowClick
}: DataTableProps<T>) {
  if (loading) {
    return (
      <div style={{
        padding: '3rem',
        textAlign: 'center',
        color: '#6b7280'
      }}>
        <div style={{
          width: '40px',
          height: '40px',
          border: '4px solid #e5e7eb',
          borderTop: '4px solid #3b82f6',
          borderRadius: '50%',
          animation: 'spin 1s linear infinite',
          margin: '0 auto 1rem'
        }}></div>
        Loading...
        <style>{`
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
        `}</style>
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div style={{
        padding: '3rem',
        textAlign: 'center',
        color: '#6b7280'
      }}>
        {emptyMessage}
      </div>
    );
  }

  return (
    <div style={{ overflowX: 'auto' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead style={{ backgroundColor: '#f9fafb' }}>
          <tr>
            {columns.map((column, index) => (
              <th
                key={index}
                style={{
                  padding: '0.75rem 1.5rem',
                  textAlign: 'left',
                  fontSize: '0.875rem',
                  fontWeight: '600',
                  color: '#374151',
                  borderBottom: '1px solid #e5e7eb',
                  width: column.width
                }}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((item, rowIndex) => (
            <tr
              key={rowIndex}
              onClick={() => onRowClick?.(item)}
              style={{
                borderBottom: '1px solid #e5e7eb',
                cursor: onRowClick ? 'pointer' : 'default',
                backgroundColor: 'white'
              }}
              onMouseEnter={(e) => {
                if (onRowClick) {
                  e.currentTarget.style.backgroundColor = '#f9fafb';
                }
              }}
              onMouseLeave={(e) => {
                if (onRowClick) {
                  e.currentTarget.style.backgroundColor = 'white';
                }
              }}
            >
              {columns.map((column, colIndex) => {
                const value = item[column.key as keyof T];
                const renderedValue = column.render
                  ? column.render(value, item)
                  : value;

                return (
                  <td
                    key={colIndex}
                    style={{
                      padding: '1rem 1.5rem',
                      fontSize: '0.875rem',
                      color: '#111827'
                    }}
                  >
                    {renderedValue}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default DataTable;