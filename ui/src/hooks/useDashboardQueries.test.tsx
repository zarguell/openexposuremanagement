import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useDashboardQueries } from './useDashboardQueries';
import { DashboardConfig } from '../types/dashboard';
import { apiClient } from '../api/client';

vi.mock('../api/client', () => ({
  apiClient: {
    queryExecute: vi.fn(),
  },
}));

const mockQueryExecute = apiClient.queryExecute as ReturnType<typeof vi.fn>;

describe('useDashboardQueries', () => {
  let queryClient: QueryClient;
  let wrapper: ({ children }: { children: React.ReactNode }) => JSX.Element;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          staleTime: 0,
        },
      },
    });

    wrapper = ({ children }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    vi.clearAllMocks();
  });

  it('executes queries for all dashboard widgets', async () => {
    const mockConfig: DashboardConfig = {
      id: 'test-dashboard',
      title: 'Test Dashboard',
      widgets: [
        {
          id: 'widget1',
          title: 'Widget 1',
          type: 'metric',
          entity: 'assets',
          query: { filters: [], limit: 10, offset: 0 },
          aggregation: 'count',
        },
        {
          id: 'widget2',
          title: 'Widget 2',
          type: 'metric',
          entity: 'findings',
          query: { filters: [{ field: 'severity', operator: 'eq', value: 'critical' }], limit: 10, offset: 0 },
          aggregation: 'count',
        },
      ],
    };

    mockQueryExecute.mockResolvedValue({
      data: [],
      meta: { total_rows: 42, execution_time_ms: 5, has_more: false },
    });

    renderHook(() => useDashboardQueries(mockConfig), { wrapper });

    expect(mockQueryExecute).toHaveBeenCalledTimes(2);
    expect(mockQueryExecute).toHaveBeenCalledWith('assets', { filters: [], limit: 10, offset: 0 });
    expect(mockQueryExecute).toHaveBeenCalledWith('findings', { filters: [{ field: 'severity', operator: 'eq', value: 'critical' }], limit: 10, offset: 0 });
  });

  it('returns results with correct structure', async () => {
    const mockConfig: DashboardConfig = {
      id: 'test-dashboard',
      title: 'Test Dashboard',
      widgets: [
        {
          id: 'widget1',
          title: 'Widget 1',
          type: 'metric',
          entity: 'assets',
          query: { filters: [], limit: 10, offset: 0 },
          aggregation: 'count',
        },
      ],
    };

    mockQueryExecute.mockResolvedValue({
      data: [{ count: 100 }],
      meta: { total_rows: 100, execution_time_ms: 5, has_more: false },
    });

    const { result } = renderHook(() => useDashboardQueries(mockConfig), { wrapper });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.results).toHaveLength(1);
    expect(result.current.results[0]).toMatchObject({
      id: 'widget1',
      title: 'Widget 1',
      type: 'metric',
      data: [{ count: 100 }],
      meta: {
        total_rows: 100,
        execution_time_ms: 5,
      },
    });
  });

  it('handles errors from individual widgets', async () => {
    const mockConfig: DashboardConfig = {
      id: 'test-dashboard',
      title: 'Test Dashboard',
      widgets: [
        {
          id: 'widget1',
          title: 'Widget 1',
          type: 'metric',
          entity: 'assets',
          query: { filters: [], limit: 10, offset: 0 },
          aggregation: 'count',
        },
      ],
    };

    mockQueryExecute.mockRejectedValue(new Error('API Error'));

    const { result } = renderHook(() => useDashboardQueries(mockConfig), { wrapper });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.isError).toBe(true);
    expect(result.current.results[0].error).toBe('API Error');
  });

  it('provides widgets array from config', () => {
    const mockConfig: DashboardConfig = {
      id: 'test-dashboard',
      title: 'Test Dashboard',
      widgets: [
        {
          id: 'widget1',
          title: 'Widget 1',
          type: 'metric',
          entity: 'assets',
          query: { filters: [], limit: 10, offset: 0 },
        },
      ],
    };

    mockQueryExecute.mockResolvedValue({
      data: [],
      meta: { total_rows: 0, execution_time_ms: 0, has_more: false },
    });

    const { result: hookResult } = renderHook(() => useDashboardQueries(mockConfig), { wrapper });

    expect(hookResult.current.widgets).toBe(mockConfig.widgets);
  });
});
