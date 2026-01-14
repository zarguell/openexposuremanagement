import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useQuery } from './useQuery';
import { apiClient } from '../api/client';

// Mock the API client
vi.mock('../api/client', () => ({
  apiClient: {
    queryExecute: vi.fn(),
  },
}));

describe('useQuery', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });
    vi.clearAllMocks();
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  it('should execute query on findings entity', async () => {
    const mockData = {
      data: [
        { id: 1, title: 'Test Finding 1', severity: 'critical' },
        { id: 2, title: 'Test Finding 2', severity: 'high' },
      ],
      meta: { total: 2, limit: 50, offset: 0 },
    };

    vi.mocked(apiClient.queryExecute).mockResolvedValue(mockData);

    const { result } = renderHook(
      () =>
        useQuery('findings', {
          filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
          limit: 50,
        }),
      { wrapper }
    );

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    // Wait for the query to complete
    await waitFor(() => expect(result.current.isLoading).toBe(false));

    // Verify data is loaded
    expect(result.current.data).toEqual(mockData);
    expect(result.current.isLoading).toBe(false);
    expect(result.current.isError).toBe(false);
    expect(apiClient.queryExecute).toHaveBeenCalledTimes(1);
    expect(apiClient.queryExecute).toHaveBeenCalledWith('findings', {
      filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
      limit: 50,
    });
  });

  it('should execute query on assets entity', async () => {
    const mockData = {
      data: [
        { id: 1, hostname: 'server01.example.com', ip_address: '192.168.1.10' },
      ],
      meta: { total: 1, limit: 50, offset: 0 },
    };

    vi.mocked(apiClient.queryExecute).mockResolvedValue(mockData);

    const { result } = renderHook(
      () =>
        useQuery('assets', {
          filters: [{ field: 'hostname', operator: 'contains', value: 'server' }],
        }),
      { wrapper }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data).toEqual(mockData);
    expect(apiClient.queryExecute).toHaveBeenCalledWith('assets', {
      filters: [{ field: 'hostname', operator: 'contains', value: 'server' }],
    });
  });

  it('should handle query errors', async () => {
    const mockError = new Error('Failed to execute query: Internal Server Error');
    vi.mocked(apiClient.queryExecute).mockRejectedValue(mockError);

    const { result } = renderHook(
      () =>
        useQuery('findings', {
          filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
        }),
      { wrapper }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toEqual(mockError);
    expect(result.current.isLoading).toBe(false);
    expect(result.current.isError).toBe(true);
  });

  it('should respect enabled option', async () => {
    const mockData = {
      data: [{ id: 1, title: 'Test Finding' }],
      meta: { total: 1, limit: 50, offset: 0 },
    };

    vi.mocked(apiClient.queryExecute).mockResolvedValue(mockData);

    const { result, rerender } = renderHook(
      ({ enabled }) =>
        useQuery(
          'findings',
          {
            filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
          },
          { enabled }
        ),
      {
        wrapper,
        initialProps: { enabled: false },
      }
    );

    // Query should not execute when disabled
    expect(apiClient.queryExecute).not.toHaveBeenCalled();
    expect(result.current.fetchStatus).toBe('idle');

    // Enable the query
    rerender({ enabled: true });

    // Now it should execute
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(apiClient.queryExecute).toHaveBeenCalledTimes(1);
  });

  it('should generate unique query keys based on query parameters', async () => {
    const mockData = {
      data: [{ id: 1, title: 'Test Finding' }],
      meta: { total: 1, limit: 50, offset: 0 },
    };

    vi.mocked(apiClient.queryExecute).mockResolvedValue(mockData);

    const query1 = {
      filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
    };

    const query2 = {
      filters: [{ field: 'severity', operator: 'eq', value: 'high' }],
    };

    const { result: result1 } = renderHook(() => useQuery('findings', query1), { wrapper });
    const { result: result2 } = renderHook(() => useQuery('findings', query2), { wrapper });

    await waitFor(() => expect(result1.current.isLoading).toBe(false));
    await waitFor(() => expect(result2.current.isLoading).toBe(false));

    // Both queries should have been executed
    expect(apiClient.queryExecute).toHaveBeenCalledTimes(2);
  });

  it('should support refetchInterval option', async () => {
    const mockData = {
      data: [{ id: 1, title: 'Test Finding' }],
      meta: { total: 1, limit: 50, offset: 0 },
    };

    vi.mocked(apiClient.queryExecute).mockResolvedValue(mockData);

    const { result } = renderHook(
      () =>
        useQuery(
          'findings',
          {
            filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
          },
          { refetchInterval: 5000 }
        ),
      { wrapper }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data).toEqual(mockData);
    expect(apiClient.queryExecute).toHaveBeenCalledTimes(1);
  });
});
