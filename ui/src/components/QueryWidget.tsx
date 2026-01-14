import { useState, useCallback } from 'react';
import { EntityType, Query, Sort } from '../types/query';
import { useQuery } from '../hooks/useQuery';
import { QueryBuilder } from './QueryBuilder';
import QueryResultsTable, { Column } from './QueryResultsTable';
import './QueryWidget.css';

export interface QueryWidgetProps<T> {
  entity: EntityType;
  title: string;
  initialQuery?: Query;
  columns: Column<T>[];
  maxHeight?: string;
  onRowClick?: (item: T) => void;
}

const DEFAULT_QUERY: Query = {
  filters: [],
  limit: 50,
  offset: 0,
  sort: []
};

function QueryWidget<T = any>({
  entity,
  title,
  initialQuery,
  columns,
  maxHeight,
  onRowClick
}: QueryWidgetProps<T>) {
  const [query, setQuery] = useState<Query>(initialQuery || DEFAULT_QUERY);
  const [showFilters, setShowFilters] = useState(true);

  const { data: result, isLoading, isError, error, refetch } = useQuery<T>(entity, query);

  const handleQueryChange = useCallback((newQuery: Query) => {
    setQuery({ ...newQuery, offset: 0 });
  }, []);

  const handleSortChange = useCallback((field: string, direction: 'asc' | 'desc') => {
    const newSort: Sort[] = [{ field, order: direction }];
    setQuery(prev => ({ ...prev, sort: newSort }));
  }, []);

  const handlePageChange = useCallback((limit: number, offset: number) => {
    setQuery(prev => ({ ...prev, limit, offset }));
  }, []);

  if (isError) {
    return (
      <div className="query-widget error">
        <div className="widget-header">
          <h3>{title}</h3>
        </div>
        <div className="widget-error-content">
          <p>Error: {error instanceof Error ? error.message : 'Unknown error'}</p>
          <button onClick={() => refetch()} className="retry-btn">
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="query-widget">
      <div className="widget-header">
        <h3>{title}</h3>
        <button 
          className="toggle-filters-btn"
          onClick={() => setShowFilters(!showFilters)}
          aria-label={showFilters ? "Hide filters" : "Toggle filters"}
        >
          <svg 
            xmlns="http://www.w3.org/2000/svg" 
            width="16" 
            height="16" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2" 
            strokeLinecap="round" 
            strokeLinejoin="round"
            className={showFilters ? 'rotated' : ''}
          >
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
      </div>

      {showFilters && (
        <div className="widget-filters">
          <QueryBuilder 
            entity={entity} 
            query={query} 
            onChange={handleQueryChange} 
          />
        </div>
      )}

      <div className="widget-content" style={{ maxHeight }}>
        <QueryResultsTable
          entity={entity}
          result={result}
          columns={columns}
          isLoading={isLoading}
          error={error}
          limit={query.limit}
          offset={query.offset}
          sort={query.sort}
          onSortChange={handleSortChange}
          onPageChange={handlePageChange}
          onRowClick={onRowClick}
        />
      </div>
    </div>
  );
}

export default QueryWidget;
