import { useQueries } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import { DashboardConfig, DashboardWidgetResult } from '../types/dashboard';

function extractErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    // Try to parse API error responses
    const message = error.message;
    try {
      const parsed = JSON.parse(message);
      if (parsed.error?.details?.error) {
        return parsed.error.details.error;
      }
      if (parsed.error?.message) {
        return parsed.error.message;
      }
    } catch {
      // Not JSON, return original message
    }
    return message;
  }
  return String(error);
}

function getErrorType(error: string): 'database' | 'connection' | 'validation' | 'unknown' {
  const lowerError = error.toLowerCase();

  if (lowerError.includes('relation') && lowerError.includes('does not exist')) {
    return 'database';
  }
  if (lowerError.includes('column') && lowerError.includes('does not exist')) {
    return 'database';
  }
  if (lowerError.includes('database') || lowerError.includes('pq:') || lowerError.includes('sql')) {
    return 'database';
  }
  if (lowerError.includes('connection') || lowerError.includes('connect')) {
    return 'connection';
  }
  if (lowerError.includes('validation') || lowerError.includes('missing_filters') || lowerError.includes('not allowed')) {
    return 'validation';
  }
  return 'unknown';
}

export function useDashboardQueries(config: DashboardConfig) {
  const queries = useQueries({
    queries: config.widgets.map((widget) => ({
      queryKey: ['dashboard-widget', widget.id, widget.query],
      queryFn: async () => {
        const result = await apiClient.queryExecute(widget.entity, widget.query);
        return {
          id: widget.id,
          title: widget.title,
          type: widget.type,
          data: result.data,
          meta: result.meta,
        };
      },
      refetchInterval: 30000, // Refetch every 30 seconds
      staleTime: 15000, // Consider data fresh for 15 seconds
      retry: 1, // Only retry once to avoid infinite loops on persistent errors
    })),
  });

  const results: DashboardWidgetResult[] = queries.map((query, index) => {
    const widget = config.widgets[index];
    return {
      id: widget.id,
      title: widget.title,
      type: widget.type,
      data: query.data?.data || [],
      meta: query.data?.meta || { total_rows: 0, execution_time_ms: 0 },
      error: query.error ? extractErrorMessage(query.error) : undefined,
    };
  });

  const isLoading = queries.some((q) => q.isLoading);
  const isError = queries.some((q) => q.isError);

  // Collect all unique errors for better debugging
  const errors = queries
    .filter((q) => q.error)
    .map((q, index) => ({
      widget: config.widgets[index].title,
      error: extractErrorMessage(q.error!),
      errorType: getErrorType(extractErrorMessage(q.error!)),
    }));

  return {
    results,
    isLoading,
    isError,
    widgets: config.widgets,
    errors,
  };
}
