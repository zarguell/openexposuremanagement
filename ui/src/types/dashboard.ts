import { Query } from './query';

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
