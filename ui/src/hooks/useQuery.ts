import { useQuery as useReactQuery, UseQueryResult } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import type { Query, QueryResult } from '../types/query';

export type EntityType = 'findings' | 'assets';

interface UseQueryOptions {
  enabled?: boolean;
  refetchInterval?: number;
}

/**
 * Custom hook that wraps React Query with our query framework types.
 * Executes queries against findings or assets with automatic caching and refetching.
 *
 * @param entity - The entity type to query ('findings' | 'assets')
 * @param query - The query object with filters, aggregations, sort, pagination
 * @param options - React Query options (enabled, refetchInterval)
 * @returns React Query result with data, isLoading, isError, error, etc.
 *
 * @example
 * const { data, isLoading, error } = useQuery('findings', {
 *   filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
 *   limit: 50
 * });
 */
export function useQuery<T = any>(
  entity: EntityType,
  query: Query,
  options?: UseQueryOptions
): UseQueryResult<QueryResult<T>, Error> {
  return useReactQuery({
    queryKey: ['query', entity, query] as const,
    queryFn: () => apiClient.queryExecute(entity, query),
    ...options,
  });
}
