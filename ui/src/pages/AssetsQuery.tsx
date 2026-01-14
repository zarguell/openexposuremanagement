import { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { QueryBuilder } from '../components/QueryBuilder';
import QueryResultsTable from '../components/QueryResultsTable';
import { AssetDrawer, Asset } from '../components/AssetDrawer';
import { useQuery } from '../hooks/useQuery';
import { Query, Sort } from '../types/query';

const DEFAULT_QUERY: Query = {
  filters: [],
  limit: 50,
  offset: 0,
};

const ASSETS_COLUMNS = [
  {
    key: 'hostname',
    label: 'Hostname',
    sortable: true,
    render: (value: string, item: any) => (
      <div>
        <div style={{ fontWeight: '500' }}>{value}</div>
        {item.canonical_name && item.canonical_name !== value && (
          <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>
            {item.canonical_name}
          </div>
        )}
      </div>
    ),
  },
  {
    key: 'ip_address',
    label: 'IP Address',
    sortable: true,
    width: '150px',
    render: (value: string) => value || 'N/A',
  },
  {
    key: 'os',
    label: 'OS',
    sortable: true,
    width: '200px',
    render: (value: string) => value || 'Unknown',
  },
  {
    key: 'first_seen_at',
    label: 'First Seen',
    sortable: true,
    width: '180px',
    render: (value: string) => value ? new Date(value).toLocaleDateString() : 'Unknown',
  },
  {
    key: 'last_seen_at',
    label: 'Last Seen',
    sortable: true,
    width: '180px',
    render: (value: string) => value ? new Date(value).toLocaleDateString() : 'Unknown',
  },
  {
    key: 'is_active',
    label: 'Active',
    sortable: true,
    width: '100px',
    render: (value: boolean) => (value ? 'Yes' : 'No'),
  },
];

function AssetsQuery() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedAsset, setSelectedAsset] = useState<Asset | undefined>();
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
  const [sort, setSort] = useState<Sort[]>([]);
  const [limit, setLimit] = useState(50);
  const [offset, setOffset] = useState(0);

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

  const handleRowClick = useCallback((asset: Asset) => {
    setSelectedAsset(asset);
    setIsDrawerOpen(true);
  }, []);

  const { data, isLoading, error } = useQuery('assets', query);

  return (
    <div style={{ padding: '1.5rem 1rem' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2.25rem', fontWeight: 'bold', color: '#111827', marginBottom: '0.5rem' }}>
          Query Assets
        </h1>
        <p style={{ color: '#6b7280' }}>
          Build custom queries to search and filter assets
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
            entity="assets"
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
              {data.meta.total_rows} assets found
            </p>
          )}
        </div>

        <QueryResultsTable
          entity="assets"
          result={data}
          columns={ASSETS_COLUMNS}
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

      <AssetDrawer
        isOpen={isDrawerOpen}
        onClose={() => setIsDrawerOpen(false)}
        asset={selectedAsset}
      />
    </div>
  );
}

export default AssetsQuery;
