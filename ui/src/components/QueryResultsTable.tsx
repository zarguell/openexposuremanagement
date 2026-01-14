import React from 'react';
import { EntityType, QueryResult, Sort } from '../types/query';
import './QueryResultsTable.css';

export interface Column<T> {
  key: keyof T | string;
  label: string;
  sortable?: boolean;
  width?: string;
  render?: (value: any, item: T) => React.ReactNode;
}

interface QueryResultsTableProps<T> {
  entity: EntityType;
  result?: QueryResult<T>;
  columns: Column<T>[];
  isLoading?: boolean;
  error?: Error | null;
  limit?: number;
  offset?: number;
  sort?: Sort[];
  onSortChange?: (field: string, direction: 'asc' | 'desc') => void;
  onPageChange?: (limit: number, offset: number) => void;
  onRowClick?: (item: T) => void;
}

function QueryResultsTable<T = any>({
  result,
  columns,
  isLoading = false,
  error = null,
  limit = 50,
  offset = 0,
  sort = [],
  onSortChange,
  onPageChange,
  onRowClick,
}: QueryResultsTableProps<T>) {
  if (isLoading) {
    return (
      <div className="query-results-table-container">
        <div className="query-results-loading">
          Loading...
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="query-results-table-container">
        <div className="query-results-error">
          Error: {error.message}
        </div>
      </div>
    );
  }

  if (!result || !result.data || result.data.length === 0) {
    return (
      <div className="query-results-table-container">
        <div className="query-results-empty">
          No results found
        </div>
      </div>
    );
  }

  const handleHeaderClick = (column: Column<T>) => {
    if (!column.sortable || !onSortChange) return;

    const key = column.key.toString();
    const currentSort = sort.find(s => s.field === key);
    
    let direction: 'asc' | 'desc' = 'asc';
    if (currentSort && currentSort.order === 'asc') {
      direction = 'desc';
    }

    onSortChange(key, direction);
  };

  const totalPages = Math.ceil(result.meta.total_rows / limit);
  const currentPage = Math.floor(offset / limit);

  const handlePageChange = (newPage: number) => {
    if (!onPageChange) return;
    onPageChange(limit, newPage * limit);
  };

  return (
    <div className="query-results-table-container">
      <div className="query-results-table-scroll">
        <table className="query-results-table">
          <thead>
            <tr>
              {columns.map((column, index) => {
                const sortState = sort.find(s => s.field === column.key.toString());
                return (
                  <th
                    key={index}
                    className={column.sortable ? 'sortable' : ''}
                    onClick={() => handleHeaderClick(column)}
                    style={column.width ? { width: column.width } : undefined}
                  >
                    {column.label}
                    {sortState && (
                      <span className="sort-icon">
                        {sortState.order === 'asc' ? '↑' : '↓'}
                      </span>
                    )}
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {result.data.map((item, rowIndex) => (
              <tr
                key={rowIndex}
                onClick={() => onRowClick?.(item)}
                className={onRowClick ? 'clickable' : ''}
              >
                {columns.map((column, colIndex) => {
                  const value = (item as any)[column.key];
                  return (
                    <td key={colIndex}>
                      {column.render ? column.render(value, item) : value}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="query-results-pagination">
        <div className="pagination-info">
          Showing {offset + 1} to {Math.min(offset + limit, result.meta.total_rows)} of {result.meta.total_rows} results
        </div>
        <div className="pagination-controls">
          <button
            className="pagination-button"
            disabled={currentPage === 0}
            onClick={() => handlePageChange(currentPage - 1)}
          >
            Previous
          </button>
          
          <button className="pagination-button active">
            {currentPage + 1}
          </button>

          <button
            className="pagination-button"
            disabled={currentPage >= totalPages - 1}
            onClick={() => handlePageChange(currentPage + 1)}
          >
            Next
          </button>
        </div>
      </div>
    </div>
  );
}

export default QueryResultsTable;
