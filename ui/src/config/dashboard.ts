import { DashboardConfig } from '../types/dashboard';
import { Query, Sort } from '../types/query';

// Helper to create queries with proper typing
const createQuery = (filters: Query['filters'], sort?: Sort[], limit?: number): Query => ({
  filters,
  sort,
  limit,
  offset: 0,
});

export const DEFAULT_DASHBOARD: DashboardConfig = {
  id: 'main',
  title: 'Vulnerability Management Dashboard',
  description: 'Overview of assets, findings, and threat intelligence',
  widgets: [
    // Assets Section
    {
      id: 'total-assets',
      title: 'Total Assets',
      type: 'metric',
      entity: 'assets',
      query: createQuery([{ field: 'hostname_norm', operator: 'is_not_null', value: null }]),
      aggregation: 'count',
      icon: '🖥️',
      color: '#3b82f6',
      linkTo: '/assets/query',
    },
    {
      id: 'active-assets',
      title: 'Active Assets',
      type: 'metric',
      entity: 'assets',
      query: createQuery([{ field: 'is_active', operator: 'eq', value: true }]),
      aggregation: 'count',
      icon: '✅',
      color: '#10b981',
      linkTo: '/assets/query?q=' + encodeURIComponent(JSON.stringify({
        filters: [{ field: 'is_active', operator: 'eq', value: true }],
        limit: 50,
        offset: 0,
      })),
    },
    {
      id: 'recent-assets',
      title: 'Recently Seen Assets',
      type: 'list',
      entity: 'assets',
      query: createQuery([{ field: 'hostname_norm', operator: 'is_not_null', value: null }], [{ field: 'last_seen_at', order: 'desc' }], 5),
      displayField: 'hostname_norm',
      icon: '🕐',
      linkTo: '/assets/query',
    },

    // Findings Section
    {
      id: 'total-findings',
      title: 'Total Findings',
      type: 'metric',
      entity: 'findings',
      query: createQuery([{ field: 'effective_status', operator: 'is_not_null', value: null }]),
      aggregation: 'count',
      icon: '🔍',
      color: '#8b5cf6',
      linkTo: '/findings/query',
    },
    {
      id: 'open-findings',
      title: 'Open Findings',
      type: 'metric',
      entity: 'findings',
      query: createQuery([{ field: 'effective_status', operator: 'eq', value: 'open' }]),
      aggregation: 'count',
      icon: '⚠️',
      color: '#ef4444',
      linkTo: '/findings/query?q=' + encodeURIComponent(JSON.stringify({
        filters: [{ field: 'effective_status', operator: 'eq', value: 'open' }],
        limit: 50,
        offset: 0,
      })),
    },
    {
      id: 'critical-findings',
      title: 'Critical Findings',
      type: 'metric',
      entity: 'findings',
      query: createQuery([
        { field: 'severity', operator: 'eq', value: 'critical' },
        { field: 'effective_status', operator: 'eq', value: 'open' },
      ]),
      aggregation: 'count',
      icon: '🚨',
      color: '#dc2626',
      linkTo: '/findings/query?q=' + encodeURIComponent(JSON.stringify({
        filters: [
          { field: 'severity', operator: 'eq', value: 'critical' },
          { field: 'effective_status', operator: 'eq', value: 'open' },
        ],
        limit: 50,
        offset: 0,
      })),
    },
    {
      id: 'high-findings',
      title: 'High Severity Findings',
      type: 'metric',
      entity: 'findings',
      query: createQuery([
        { field: 'severity', operator: 'eq', value: 'high' },
        { field: 'effective_status', operator: 'eq', value: 'open' },
      ]),
      aggregation: 'count',
      icon: '🔴',
      color: '#f97316',
      linkTo: '/findings/query?q=' + encodeURIComponent(JSON.stringify({
        filters: [
          { field: 'severity', operator: 'eq', value: 'high' },
          { field: 'effective_status', operator: 'eq', value: 'open' },
        ],
        limit: 50,
        offset: 0,
      })),
    },

    // Threat Intelligence Section
    {
      id: 'kev-findings',
      title: 'Known Exploited Vulnerabilities',
      type: 'metric',
      entity: 'findings',
      query: createQuery([
        { field: 'is_kev', operator: 'eq', value: true },
        { field: 'effective_status', operator: 'eq', value: 'open' },
      ]),
      aggregation: 'count',
      icon: '⚡',
      color: '#fbbf24',
      linkTo: '/findings/query?q=' + encodeURIComponent(JSON.stringify({
        filters: [
          { field: 'is_kev', operator: 'eq', value: true },
          { field: 'effective_status', operator: 'eq', value: 'open' },
        ],
        limit: 50,
        offset: 0,
      })),
    },
    {
      id: 'high-epss',
      title: 'High EPSS (>0.9)',
      type: 'metric',
      entity: 'findings',
      query: createQuery([
        { field: 'epss_score', operator: 'gt', value: 0.9 },
        { field: 'effective_status', operator: 'eq', value: 'open' },
      ]),
      aggregation: 'count',
      icon: '📊',
      color: '#f59e0b',
      linkTo: '/findings/query?q=' + encodeURIComponent(JSON.stringify({
        filters: [
          { field: 'epss_score', operator: 'gt', value: 0.9 },
          { field: 'effective_status', operator: 'eq', value: 'open' },
        ],
        limit: 50,
        offset: 0,
      })),
    },
  ],
};
