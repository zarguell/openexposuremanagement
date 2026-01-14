import { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { QueryBuilder } from '../components/QueryBuilder';
import QueryResultsTable from '../components/QueryResultsTable';
import { FindingDrawer, Finding } from '../components/FindingDrawer';
import { useQuery } from '../hooks/useQuery';
import { Query, Sort } from '../types/query';
import StatusBadge from '../components/StatusBadge';

const DEFAULT_QUERY: Query = {
  filters: [],
  limit: 50,
  offset: 0,
};

const FINDINGS_COLUMNS = [
  {
    key: 'title',
    label: 'Finding',
    sortable: true,
    render: (value: string, item: any) => (
      <div>
        <div style={{ fontWeight: '500' }}>{value}</div>
        {item.definition_uid && (
          <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>
            {item.source} - {item.definition_uid}
          </div>
        )}
      </div>
    ),
  },
  {
    key: 'severity',
    label: 'Severity',
    sortable: true,
    width: '120px',
    render: (value: string) => <StatusBadge status={value} variant="severity" />,
  },
  {
    key: 'effective_status',
    label: 'Status',
    sortable: true,
    width: '120px',
    render: (value: string) => <StatusBadge status={value} variant="status" />,
  },
  {
    key: 'cve_id',
    label: 'CVE',
    sortable: true,
    width: '150px',
    render: (value: string) => value || 'None',
  },
  {
    key: 'epss_score',
    label: 'EPSS',
    sortable: true,
    width: '100px',
    render: (value: number) => value != null ? value.toFixed(3) : 'N/A',
  },
  {
    key: 'last_observed_at',
    label: 'Last Seen',
    sortable: true,
    width: '180px',
    render: (value: string) => value ? new Date(value).toLocaleDateString() : 'Unknown',
  },
];

function FindingsQuery() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedFinding, setSelectedFinding] = useState<Finding | undefined>();
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);

  const parseQueryFromUrl = useCallback((): Query => {
    const queryParam = searchParams.get('q');
    if (!queryParam) return DEFAULT_QUERY;

    try {
      const parsed = JSON.parse(queryParam);
      return {
        filters: parsed.filters || [],
        sort: parsed.sort || [],
        limit: parsed.limit || 50,
        offset: parsed.offset || 0,
      };
    } catch {
      return DEFAULT_QUERY;
    }
  }, [searchParams]);

  const updateQueryInUrl = useCallback((query: Query) => {
    const params = new URLSearchParams();
    params.set('q', JSON.stringify(query));
    setSearchParams(params, { replace: true });
  }, [setSearchParams]);

  const [query, setQuery] = useState<Query>(parseQueryFromUrl);
  const [sort, setSort] = useState<Sort[]>(query.sort || []);
  const [limit, setLimit] = useState(query.limit || 50);
  const [offset, setOffset] = useState(query.offset || 0);

  useEffect(() => {
    const parsedQuery = parseQueryFromUrl();
    setQuery(parsedQuery);
    setSort(parsedQuery.sort || []);
    setLimit(parsedQuery.limit || 50);
    setOffset(parsedQuery.offset || 0);
  }, [searchParams, parseQueryFromUrl]);

  const handleQueryChange = useCallback((newQuery: Query) => {
    setQuery(newQuery);
    setOffset(0);
    updateQueryInUrl({ ...newQuery, offset: 0 });
  }, [updateQueryInUrl]);

  const handleSortChange = useCallback((field: string, direction: 'asc' | 'desc') => {
    const newSort = [{ field, order: direction }];
    setSort(newSort);
    updateQueryInUrl({ ...query, sort: newSort });
  }, [query, updateQueryInUrl]);

  const handlePageChange = useCallback((newLimit: number, newOffset: number) => {
    setLimit(newLimit);
    setOffset(newOffset);
    updateQueryInUrl({ ...query, limit: newLimit, offset: newOffset });
  }, [query, updateQueryInUrl]);

  const handleRowClick = useCallback((finding: Finding) => {
    setSelectedFinding(finding);
    setIsDrawerOpen(true);
  }, []);

  const { data, isLoading, error } = useQuery('findings', query);

  return (
    <div style={{ padding: '1.5rem 1rem' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2.25rem', fontWeight: 'bold', color: '#111827', marginBottom: '0.5rem' }}>
          Query Findings
        </h1>
        <p style={{ color: '#6b7280' }}>
          Build custom queries to search and filter vulnerability findings
        </p>
      </div>

      <div style={{
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
        marginBottom: '1.5rem'
      }}>
        <div style={{ marginBottom: '1rem' }}>
          <h2 style={{ fontSize: '1.125rem', fontWeight: '600', color: '#374151', marginBottom: '0.5rem' }}>
            Filters
          </h2>
          <p style={{ fontSize: '0.875rem', color: '#6b7280' }}>
            Add filters to narrow down results
          </p>
        </div>
        <div data-testid="query-builder">
          <QueryBuilder
            entity="findings"
            query={query}
            onChange={handleQueryChange}
          />
        </div>
      </div>

      <div style={{
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
        padding: '1.5rem'
      }}>
        <div style={{ marginBottom: '1rem' }}>
          <h2 style={{ fontSize: '1.125rem', fontWeight: '600', color: '#374151', marginBottom: '0.5rem' }}>
            Results
          </h2>
          {data && (
            <p style={{ fontSize: '0.875rem', color: '#6b7280' }}>
              {data.meta.total_rows} findings found
            </p>
          )}
        </div>

        <QueryResultsTable
          entity="findings"
          result={data}
          columns={FINDINGS_COLUMNS}
          isLoading={isLoading}
          error={error}
          limit={limit}
          offset={offset}
          sort={sort}
          onSortChange={handleSortChange}
          onPageChange={handlePageChange}
          onRowClick={handleRowClick}
        />
      </div>

      <FindingDrawer
        isOpen={isDrawerOpen}
        onClose={() => setIsDrawerOpen(false)}
        finding={selectedFinding}
      />
    </div>
  );
}

export default FindingsQuery;
