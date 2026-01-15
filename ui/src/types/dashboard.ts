import { Query, UnifiedQuery } from './query';

export interface DashboardWidget {
  id: string;
  title: string;
  type: 'metric' | 'chart' | 'table' | 'list';
  entity: 'assets' | 'findings';
  query: Query;
  displayField?: string;
  aggregation?: 'count' | 'sum' | 'avg' | 'max' | 'min';
  color?: string;
  icon?: string;
  linkTo?: string;
}

// Unified query widget type for cross-entity correlation queries
// NOTE: Disabled by default for performance reasons
export interface DashboardUnifiedWidget {
  id: string;
  title: string;
  type: 'metric';
  query: UnifiedQuery;
  color?: string;
  icon?: string;
  linkTo?: string;
  disabled?: boolean; // When true, widget is not rendered
}

export interface DashboardConfig {
  id: string;
  title: string;
  description?: string;
  widgets: DashboardWidget[];
}

export interface DashboardWidgetResult {
  id: string;
  title: string;
  type: string;
  data: any;
  meta: {
    total_rows: number;
    execution_time_ms: number;
  };
  error?: string;
}
