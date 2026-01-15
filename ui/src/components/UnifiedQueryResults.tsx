import React, { useState } from 'react';
import { UnifiedQueryResult, Sort } from '../types/query';
import QueryResultsTable, { Column } from './QueryResultsTable';
import './UnifiedQueryResults.css';

export interface JoinedData {
  [key: string]: any;
}

interface UnifiedQueryResultsProps {
  result?: UnifiedQueryResult;
  isLoading?: boolean;
  error?: Error | null;
  query?: any;
  onExport?: () => void;
}

/**
 * Component for displaying unified query results with cross-entity joins
 * Supports drilling down to individual entities and exporting results
 */
function UnifiedQueryResults({
  result,
  isLoading = false,
  error = null,
  query,
  onExport,
}: UnifiedQueryResultsProps) {
  const [sort, setSort] = useState<Sort[]>([]);

  if (isLoading) {
    return (
      <div className="unified-query-results">
        <div className="unified-query-loading">Loading unified query results...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="unified-query-results">
        <div className="unified-query-error">
          <strong>Error executing query:</strong> {error.message}
        </div>
      </div>
    );
  }

  if (!result || !result.data || result.data.length === 0) {
    return (
      <div className="unified-query-results">
        <div className="unified-query-empty">No results found</div>
      </div>
    );
  }

  // Dynamically generate columns from the first result row
  const firstRow = result.data[0];
  const columns: Column<JoinedData>[] = Object.keys(firstRow).map(key => ({
    key: key,
    label: formatColumnName(key),
    sortable: true,
    render: (value: any) => formatCellValue(key, value),
  }));

  const handleSortChange = (field: string, direction: 'asc' | 'desc') => {
    setSort([{ field, order: direction }]);
  };

  const handleExportToCsv = () => {
    if (!result || !result.data) return;

    // Get headers from columns
    const headers = columns.map(col => col.label);

    // Convert data to CSV
    const csvRows = [
      headers.join(','),
      ...result.data.map(row =>
        columns.map(col => {
          const value = (row as any)[col.key];
          return formatCsvValue(value);
        }).join(',')
      ),
    ];

    const csv = csvRows.join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `unified-query-results-${Date.now()}.csv`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    onExport?.();
  };

  return (
    <div className="unified-query-results">
      <div className="unified-query-header">
        <div className="unified-query-meta">
          <span className="meta-item">
            <strong>Total Rows:</strong> {result.meta.total_rows.toLocaleString()}
          </span>
          <span className="meta-item">
            <strong>Execution Time:</strong> {result.meta.execution_time_ms}ms
          </span>
          {query?.join && (
            <span className="meta-item join-badge">
              Joined with {query.join.entity}
            </span>
          )}
        </div>
        <button
          className="unified-query-export-button"
          onClick={handleExportToCsv}
        >
          Export to CSV
        </button>
      </div>

      <QueryResultsTable<JoinedData>
        entity="assets"
        result={result}
        columns={columns}
        sort={sort}
        onSortChange={handleSortChange}
      />

      {result.meta.has_more && (
        <div className="unified-query-warning">
          ⚠️ Results limited to {result.data.length} rows. Use filters to narrow your query.
        </div>
      )}
    </div>
  );
}

/**
 * Format column name for display
 * Converts snake_case to Title Case and handles special prefixes
 */
function formatColumnName(key: string): string {
  // Remove table prefixes if present (e.g., "assets_canonical_name" -> "canonical_name")
  const cleanKey = key.replace(/^(assets|software_inventory|findings)_/, '');

  // Convert snake_case to Title Case
  return cleanKey
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * Format cell value for display
 * Handles special formatting for certain fields
 */
function formatCellValue(key: string, value: any): React.ReactNode {
  // Handle NULL values from LEFT JOIN
  if (value === null || value === undefined) {
    return <span className="null-value">—</span>;
  }

  // Handle boolean values
  if (typeof value === 'boolean') {
    return value ? '✓' : '✗';
  }

  // Handle dates
  if (key.toLowerCase().includes('_at') || key.toLowerCase().includes('_date')) {
    try {
      const date = new Date(value);
      if (!isNaN(date.getTime())) {
        return date.toLocaleString();
      }
    } catch {
      // Fall through to default rendering
    }
  }

  // Handle severity special formatting
  if (key === 'severity' && typeof value === 'string') {
    const severityClass = `severity-${value.toLowerCase()}`;
    return <span className={`severity-badge ${severityClass}`}>{value}</span>;
  }

  // Default rendering
  return String(value);
}

/**
 * Format value for CSV export
 * Handles escaping and special values
 */
function formatCsvValue(value: any): string {
  if (value === null || value === undefined) {
    return '';
  }

  const stringValue = String(value);

  // Escape quotes and wrap in quotes if contains comma, quote, or newline
  if (stringValue.includes(',') || stringValue.includes('"') || stringValue.includes('\n')) {
    return `"${stringValue.replace(/"/g, '""')}"`;
  }

  return stringValue;
}

export default UnifiedQueryResults;
