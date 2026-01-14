import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './client';

describe('apiClient.queryExecute', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    // Mock window object for browser APIs
    globalThis.window = {
      location: { origin: 'http://localhost:3000' }
    } as any;
    globalThis.fetch = mockFetch;
    mockFetch.mockClear();
  });

  it('should execute query for findings entity', async () => {
    const query = {
      filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
      limit: 10
    };

    const mockResponse = {
      ok: true,
      status: 200,
      statusText: 'OK',
      headers: {
        entries: () => []
      },
      json: async () => ({ data: [], meta: { total_rows: 0, execution_time_ms: 5, has_more: false } }),
      clone: () => mockResponse
    } as any as Response;

    mockFetch.mockResolvedValueOnce(mockResponse);

    const result = await apiClient.queryExecute('findings', query);

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/query/findings'),
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json'
        })
      })
    );

    expect(result).toEqual({
      data: [],
      meta: { total_rows: 0, execution_time_ms: 5, has_more: false }
    });
  });

  it('should execute query for assets entity', async () => {
    const query = {
      filters: [{ field: 'is_active', operator: 'eq', value: true }]
    };

    const mockResponse = {
      ok: true,
      status: 200,
      statusText: 'OK',
      headers: {
        entries: () => []
      },
      json: async () => ({ data: [], meta: { total_rows: 0, execution_time_ms: 5, has_more: false } }),
      clone: () => mockResponse
    } as any as Response;

    mockFetch.mockResolvedValueOnce(mockResponse);

    await apiClient.queryExecute('assets', query);

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/query/assets'),
      expect.any(Object)
    );
  });

  it('should throw error on HTTP error', async () => {
    const query = {
      filters: [{ field: 'severity', operator: 'eq', value: 'critical' }]
    };

    const mockResponse = {
      ok: false,
      statusText: 'Bad Request',
      status: 400,
      headers: {
        entries: () => []
      },
      clone: () => ({
        text: async () => JSON.stringify({ error: 'Test error body' }),
        json: async () => ({ data: [], meta: { total_rows: 0, execution_time_ms: 5, has_more: false } })
      })
    } as any as Response;

    mockFetch.mockResolvedValueOnce(mockResponse);

    await expect(apiClient.queryExecute('findings', query)).rejects.toThrow('Failed to execute query: Bad Request');
  });
});
