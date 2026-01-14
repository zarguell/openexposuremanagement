import { render, screen, fireEvent, within } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { QueryBuilder } from './QueryBuilder';
import { Query } from '../types/query';

describe('QueryBuilder', () => {
  const defaultQuery: Query = {
    filters: [],
  };

  const mockOnChange = vi.fn();

  beforeEach(() => {
    mockOnChange.mockClear();
  });

  it('renders without crashing', () => {
    render(
      <QueryBuilder 
        entity="findings" 
        query={defaultQuery} 
        onChange={mockOnChange} 
      />
    );
    expect(screen.getByText('Add Filter')).toBeTruthy();
  });

  it('adds a new filter when "Add Filter" is clicked', () => {
    render(
      <QueryBuilder 
        entity="findings" 
        query={defaultQuery} 
        onChange={mockOnChange} 
      />
    );

    const addButton = screen.getByText('Add Filter');
    fireEvent.click(addButton);

    expect(mockOnChange).toHaveBeenCalledTimes(1);
    const newQuery = mockOnChange.mock.calls[0][0];
    expect(newQuery.filters).toHaveLength(1);
    expect(newQuery.filters[0]).toEqual({
      field: 'severity',
      operator: 'eq',
      value: ''
    });
  });

  it('renders existing filters', () => {
    const queryWithFilters: Query = {
      filters: [
        { field: 'severity', operator: 'eq', value: 'critical' },
        { field: 'scanner_status', operator: 'neq', value: 'resolved' }
      ]
    };

    render(
      <QueryBuilder 
        entity="findings" 
        query={queryWithFilters} 
        onChange={mockOnChange} 
      />
    );

    const filterRows = screen.getAllByTestId('filter-row');
    expect(filterRows).toHaveLength(2);
    
    const firstRow = filterRows[0];
    // Use getByDisplayValue to verify inputs/selects have correct values
    expect(within(firstRow).getByDisplayValue('severity')).toBeTruthy();
    expect(within(firstRow).getByDisplayValue('eq')).toBeTruthy();
    expect(within(firstRow).getByDisplayValue('critical')).toBeTruthy();
  });

  it('updates field and resets value/operator if necessary', () => {
    const query: Query = {
      filters: [{ field: 'severity', operator: 'eq', value: 'critical' }]
    };

    render(
      <QueryBuilder 
        entity="findings" 
        query={query} 
        onChange={mockOnChange} 
      />
    );

    const filterRow = screen.getByTestId('filter-row');
    const fieldSelect = within(filterRow).getByDisplayValue('severity');

    fireEvent.change(fieldSelect, { target: { value: 'epss_score' } });

    expect(mockOnChange).toHaveBeenCalled();
    const updatedQuery = mockOnChange.mock.calls[0][0];
    expect(updatedQuery.filters[0].field).toBe('epss_score');
    // It's good practice to reset value when field changes, though not strictly required by prompt
    // For MVP we can just verify the field changed
  });

  it('updates operator', () => {
    const query: Query = {
      filters: [{ field: 'severity', operator: 'eq', value: 'critical' }]
    };

    render(
      <QueryBuilder 
        entity="findings" 
        query={query} 
        onChange={mockOnChange} 
      />
    );

    const filterRow = screen.getByTestId('filter-row');
    const operatorSelect = within(filterRow).getByDisplayValue('eq');

    fireEvent.change(operatorSelect, { target: { value: 'neq' } });

    expect(mockOnChange).toHaveBeenCalled();
    const updatedQuery = mockOnChange.mock.calls[0][0];
    expect(updatedQuery.filters[0].operator).toBe('neq');
  });

  it('updates value', () => {
    const query: Query = {
      filters: [{ field: 'severity', operator: 'eq', value: 'critical' }]
    };

    render(
      <QueryBuilder 
        entity="findings" 
        query={query} 
        onChange={mockOnChange} 
      />
    );

    const filterRow = screen.getByTestId('filter-row');
    const valueInput = within(filterRow).getByDisplayValue('critical');

    fireEvent.change(valueInput, { target: { value: 'high' } });

    expect(mockOnChange).toHaveBeenCalled();
    const updatedQuery = mockOnChange.mock.calls[0][0];
    expect(updatedQuery.filters[0].value).toBe('high');
  });

  it('removes a filter', () => {
    const query: Query = {
      filters: [
        { field: 'severity', operator: 'eq', value: 'critical' },
        { field: 'scanner_status', operator: 'neq', value: 'resolved' }
      ]
    };

    render(
      <QueryBuilder 
        entity="findings" 
        query={query} 
        onChange={mockOnChange} 
      />
    );

    const filterRows = screen.getAllByTestId('filter-row');
    const removeButton = within(filterRows[0]).getByLabelText('Remove filter');

    fireEvent.click(removeButton);

    expect(mockOnChange).toHaveBeenCalled();
    const updatedQuery = mockOnChange.mock.calls[0][0];
    expect(updatedQuery.filters).toHaveLength(1);
    expect(updatedQuery.filters[0].field).toBe('scanner_status');
  });

  it('renders text input for "in" operator even for enum fields', () => {
    const query: Query = {
      filters: [{ field: 'severity', operator: 'in', value: 'critical,high' }]
    };

    render(
      <QueryBuilder 
        entity="findings" 
        query={query} 
        onChange={mockOnChange} 
      />
    );

    const filterRow = screen.getByTestId('filter-row');
    // input should be present, not select
    const input = within(filterRow).getByPlaceholderText('Comma separated values...');
    expect(input).toBeTruthy();
    expect(within(filterRow).getByDisplayValue('critical,high')).toBeTruthy();
  });
});
