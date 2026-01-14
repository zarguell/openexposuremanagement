import { useQueries } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import { DashboardConfig, DashboardWidgetResult } from '../types/dashboard';

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
      error: query.error?.message,
    };
  });

  const isLoading = queries.some((q) => q.isLoading);
  const isError = queries.some((q) => q.isError);

  return {
    results,
    isLoading,
    isError,
    widgets: config.widgets,
  };
}
