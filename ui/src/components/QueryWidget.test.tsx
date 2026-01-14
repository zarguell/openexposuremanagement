import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import QueryWidget from './QueryWidget';
import { EntityType, Query } from '../types/query';
import { useQuery } from '../hooks/useQuery';

vi.mock('../hooks/useQuery', () => ({
  useQuery: vi.fn(),
}));

vi.mock('./QueryBuilder', () => ({
  QueryBuilder: ({ query, onChange, entity }: any) => (
    <div data-testid="mock-query-builder">
      <span>Query Builder: {entity}</span>
      <button 
        data-testid="update-query-btn" 
        onClick={() => onChange({ ...query, filters: [...query.filters, { field: 'test', operator: 'eq', value: 'test' }] })}
      >
        Update Query
      </button>
    </div>
  )
}));

vi.mock('./QueryResultsTable', () => ({
  default: ({ onSortChange, onPageChange, sort, limit, offset, result, isLoading, error }: any) => (
    <div data-testid="mock-results-table">
      <span>Results Table</span>
      {isLoading && <span>Loading...</span>}
      {error && <span>Error: {error.message}</span>}
      {result && <span>Rows: {result.meta.total_rows}</span>}
      <div data-testid="sort-info">Sort: {JSON.stringify(sort)}</div>
      <div data-testid="page-info">Limit: {limit}, Offset: {offset}</div>
      <button 
        data-testid="sort-change-btn" 
        onClick={() => onSortChange('severity', 'desc')}
      >
        Sort Change
      </button>
      <button 
        data-testid="page-change-btn" 
        onClick={() => onPageChange(limit, offset + limit)}
      >
        Page Change
      </button>
    </div>
  )
}));

describe('QueryWidget', () => {
  const mockColumns = [{ key: 'id', label: 'ID' }];
  const defaultProps = {
    entity: 'findings' as EntityType,
    title: 'Test Widget',
    columns: mockColumns,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (useQuery as any).mockReturnValue({
      data: { data: [], meta: { total_rows: 0, execution_time_ms: 0, has_more: false } },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    });
  });

  it('renders with correct title and initial state', () => {
    render(<QueryWidget {...defaultProps} />);
    
    expect(screen.getByText('Test Widget')).toBeTruthy();
    expect(screen.getByTestId('mock-query-builder')).toBeTruthy();
    expect(screen.getByTestId('mock-results-table')).toBeTruthy();
    expect(screen.getByText('Query Builder: findings')).toBeTruthy();
  });

  it('updates query when QueryBuilder changes', async () => {
    render(<QueryWidget {...defaultProps} />);
    
    const updateBtn = screen.getByTestId('update-query-btn');
    fireEvent.click(updateBtn);
    
    expect(useQuery).toHaveBeenLastCalledWith(
      'findings',
      expect.objectContaining({
        filters: expect.arrayContaining([
          expect.objectContaining({ field: 'test', value: 'test' })
        ])
      })
    );
  });

  it('handles sort changes from results table', () => {
    render(<QueryWidget {...defaultProps} />);
    
    const sortBtn = screen.getByTestId('sort-change-btn');
    fireEvent.click(sortBtn);
    
    expect(useQuery).toHaveBeenLastCalledWith(
      'findings',
      expect.objectContaining({
        sort: [{ field: 'severity', order: 'desc' }]
      })
    );
  });

  it('handles pagination changes from results table', () => {
    render(<QueryWidget {...defaultProps} />);
    
    const pageBtn = screen.getByTestId('page-change-btn');
    fireEvent.click(pageBtn);
    
    expect(useQuery).toHaveBeenLastCalledWith(
      'findings',
      expect.objectContaining({
        offset: 50
      })
    );
  });

  it('displays loading state correctly', () => {
    (useQuery as any).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    });

    render(<QueryWidget {...defaultProps} />);
    
    expect(screen.getByText('Loading...')).toBeTruthy();
  });

  it('displays error state and retry button', () => {
    const mockRefetch = vi.fn();
    (useQuery as any).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: new Error('Network error'),
      refetch: mockRefetch,
    });

    render(<QueryWidget {...defaultProps} />);
    
    expect(screen.getByText(/Error: Network error/)).toBeTruthy();
    
    const retryBtn = screen.getByText('Retry');
    fireEvent.click(retryBtn);
    
    expect(mockRefetch).toHaveBeenCalled();
  });

  it('uses initialQuery if provided', () => {
    const initialQuery: Query = {
      filters: [{ field: 'severity', operator: 'eq', value: 'critical' }],
      limit: 10,
      offset: 0
    };

    render(<QueryWidget {...defaultProps} initialQuery={initialQuery} />);
    
    expect(useQuery).toHaveBeenCalledWith(
      'findings',
      expect.objectContaining({
        filters: initialQuery.filters,
        limit: 10
      })
    );
  });
  
  it('toggles query builder visibility', () => {
    render(<QueryWidget {...defaultProps} />);
    
    const toggleBtn = screen.getByLabelText(/Hide filters/i); 
    
    expect(screen.getByTestId('mock-query-builder')).toBeTruthy();
    
    fireEvent.click(toggleBtn);
    
    expect(screen.queryByTestId('mock-query-builder')).toBeNull();
  });
});
