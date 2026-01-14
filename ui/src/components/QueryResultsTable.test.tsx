import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import QueryResultsTable from './QueryResultsTable';
import { EntityType, QueryResult } from '../types/query';

describe('QueryResultsTable', () => {
  const mockColumns = [
    { key: 'name', label: 'Name', sortable: true },
    { key: 'status', label: 'Status', sortable: false },
  ];

  const mockData: QueryResult = {
    data: [
      { id: 1, name: 'Asset 1', status: 'active' },
      { id: 2, name: 'Asset 2', status: 'inactive' },
    ],
    meta: {
      total_rows: 20,
      execution_time_ms: 10,
      has_more: true,
    },
  };

  const defaultProps = {
    entity: 'assets' as EntityType,
    result: mockData,
    columns: mockColumns,
    onSortChange: vi.fn(),
    onPageChange: vi.fn(),
    onRowClick: vi.fn(),
  };

  it('renders loading state', () => {
    render(<QueryResultsTable {...defaultProps} result={undefined} isLoading={true} />);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('renders error state', () => {
    const error = new Error('Failed to fetch');
    render(<QueryResultsTable {...defaultProps} result={undefined} error={error} />);
    expect(screen.getByText(/failed to fetch/i)).toBeInTheDocument();
  });

  it('renders empty state', () => {
    const emptyResult = { ...mockData, data: [], meta: { ...mockData.meta, total_rows: 0 } };
    render(<QueryResultsTable {...defaultProps} result={emptyResult} />);
    expect(screen.getByText(/no results found/i)).toBeInTheDocument();
  });

  it('renders data correctly', () => {
    render(<QueryResultsTable {...defaultProps} />);
    expect(screen.getByText('Asset 1')).toBeInTheDocument();
    expect(screen.getByText('Asset 2')).toBeInTheDocument();
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('handles sort clicks', () => {
    const { rerender } = render(<QueryResultsTable {...defaultProps} />);
    const nameHeader = screen.getByText('Name');
    fireEvent.click(nameHeader);
    expect(defaultProps.onSortChange).toHaveBeenCalledWith('name', 'asc');
    
    rerender(<QueryResultsTable {...defaultProps} sort={[{ field: 'name', order: 'asc' }]} />);
    fireEvent.click(nameHeader);
    expect(defaultProps.onSortChange).toHaveBeenCalledWith('name', 'desc');
  });

  it('does not sort unsortable columns', () => {
    render(<QueryResultsTable {...defaultProps} />);
    const statusHeader = screen.getByText('Status');
    fireEvent.click(statusHeader);
    expect(defaultProps.onSortChange).not.toHaveBeenCalledWith('status', expect.any(String));
  });

  it('handles row clicks', () => {
    render(<QueryResultsTable {...defaultProps} />);
    const row = screen.getByText('Asset 1').closest('tr');
    fireEvent.click(row!);
    expect(defaultProps.onRowClick).toHaveBeenCalledWith(mockData.data[0]);
  });

  it('renders pagination controls', () => {
    render(<QueryResultsTable {...defaultProps} />);
    expect(screen.getByRole('button', { name: /next/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /previous/i })).toBeInTheDocument();
  });

  it('handles page changes', () => {
    const paginationProps = {
      ...defaultProps,
      limit: 10,
      offset: 0,
      result: mockData,
    };
    
    render(<QueryResultsTable {...paginationProps} />);
    const nextBtn = screen.getByRole('button', { name: /next/i });
    expect(nextBtn).not.toBeDisabled();

    fireEvent.click(nextBtn);
    
    expect(defaultProps.onPageChange).toHaveBeenCalledWith(10, 10);
  });
});
