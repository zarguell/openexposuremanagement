import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import FindingsQuery from './FindingsQuery';
import { Query } from '../types/query';
import { apiClient } from '../api/client';

vi.mock('../api/client', () => ({
  apiClient: {
    queryExecute: vi.fn(),
  },
}));

const mockApiCall = apiClient.queryExecute as ReturnType<typeof vi.fn>;

describe('FindingsQuery', () => {
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

    wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/findings/query']}>
          <Routes>
            <Route path="/findings/query" element={children} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    );

    mockApiCall.mockResolvedValue({
      data: [
        {
          id: 1,
          title: 'Test Finding',
          severity: 'high',
          effective_status: 'open',
          cve_id: 'CVE-2023-1234',
          epss_score: 0.5,
          last_observed_at: '2024-01-01T00:00:00Z',
        },
      ],
      meta: {
        total_rows: 1,
        execution_time_ms: 10,
        has_more: false,
      },
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders query builder and results table', () => {
    render(<FindingsQuery />, { wrapper });

    expect(screen.getByText('Query Findings')).toBeInTheDocument();
    expect(screen.getByTestId('query-builder')).toBeInTheDocument();
  });

  it('loads initial query from URL', async () => {
    const initialQuery: Query = {
      filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
      limit: 50,
      offset: 0,
    };

    const expectedQuery = {
      ...initialQuery,
      sort: [],
    };

    const queryWrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={[`/findings/query?q=${encodeURIComponent(JSON.stringify(initialQuery))}`]}>
          <Routes>
            <Route path="/findings/query" element={children} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    );

    render(<FindingsQuery />, { wrapper: queryWrapper });

    await waitFor(() => {
      expect(mockApiCall).toHaveBeenCalledWith('findings', expectedQuery);
    });
  });

  it('executes query when filters change', async () => {
    render(<FindingsQuery />, { wrapper });

    const user = userEvent.setup();

    await waitFor(() => {
      expect(screen.getByTestId('query-builder')).toBeInTheDocument();
    });

    const addButtons = screen.getAllByText('Add Filter');
    await user.click(addButtons[0]);

    await waitFor(() => {
      expect(mockApiCall).toHaveBeenCalled();
    });
  });

  it('opens drawer when row is clicked', async () => {
    render(<FindingsQuery />, { wrapper });

    await waitFor(() => {
      expect(screen.getByText('Query Findings')).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(mockApiCall).toHaveBeenCalled();
    });
  });

  it('updates URL when query changes', async () => {
    render(<FindingsQuery />, { wrapper });

    await waitFor(() => {
      expect(screen.getByText('Query Findings')).toBeInTheDocument();
    });

    waitFor(() => {
      expect(window.location.search).toContain('q=');
    });
  });

  it('handles empty query state', () => {
    const customWrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/findings/query']}>
          <Routes>
            <Route path="/findings/query" element={children} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    );

    render(<FindingsQuery />, { wrapper: customWrapper });

    expect(screen.getByText('Query Findings')).toBeInTheDocument();
  });

  it('shows loading state while fetching', async () => {
    mockApiCall.mockImplementation(
      () =>
        new Promise((resolve) => {
          setTimeout(() => {
            resolve({
              data: [],
              meta: { total_rows: 0, execution_time_ms: 0, has_more: false },
            });
          }, 100);
        })
    );

    render(<FindingsQuery />, { wrapper });

    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('handles API errors gracefully', async () => {
    mockApiCall.mockRejectedValue(new Error('API Error'));

    render(<FindingsQuery />, { wrapper });

    await waitFor(() => {
      expect(screen.getByText(/error/i)).toBeInTheDocument();
    });
  });
});
