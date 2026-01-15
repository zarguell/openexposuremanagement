// Types match backend query framework JSON schema
// See: docs/plans/2025-01-12-query-framework-design.md

export interface Filter {
  field: string;
  operator: string;
  value: any;
}

export interface Aggregation {
  type: 'count' | 'sum' | 'max' | 'min' | 'group_by';
  field?: string;
}

export interface Sort {
  field: string;
  order: 'asc' | 'desc';
}

export interface Query {
  filters: Filter[];
  aggregations?: Aggregation[];
  sort?: Sort[];
  limit?: number;
  offset?: number;
}

export interface QueryMeta {
  total_rows: number;
  execution_time_ms: number;
  has_more: boolean;
}

export interface QueryResult<T = any> {
  data: T[];
  meta: QueryMeta;
}

export type EntityType = 'findings' | 'assets';

// Unified Query Types for Cross-Entity Correlation
export interface JoinCondition {
  primary: string;
  joined: string;
}

export interface Join {
  entity: 'software_inventory' | 'findings';
  type: 'left';
  on: JoinCondition;
}

export interface UnifiedQuery {
  primary_entity: 'assets'; // Only assets for MVP
  join?: Join;
  filters?: Filter[];
  aggregations?: Aggregation[];
  sort?: Sort[];
  limit?: number;
  offset?: number;
}

export interface UnifiedQueryResult {
  data: any[];
  meta: QueryMeta;
}

export const ALLOWED_FIELDS = {
  findings: [
    'severity',
    'scanner_status',
    'effective_status',
    'cve',
    'source',
    'asset_name',
    'first_observed_at',
    'last_observed_at',
    'epss_score',
    'is_kev',
    'has_cve'
  ],
  assets: [
    'canonical_name',
    'hostname_norm',
    'shortname_norm',
    'ipv4',
    'first_seen_at',
    'last_seen_at',
    'is_active'
  ]
} as const;

export const ALLOWED_OPERATORS = [
  'eq',
  'neq',
  'in',
  'not_in',
  'like',
  'not_like',
  'gt',
  'gte',
  'lt',
  'lte',
  'between',
  'is_null',
  'is_not_null'
] as const;

export const SEVERITY_VALUES = [
  'critical',
  'high',
  'medium',
  'low'
] as const;

export const SCANNER_STATUS_VALUES = [
  'open',
  'resolved',
  'suppressed'
] as const;

export const EFFECTIVE_STATUS_VALUES = [
  'open',
  'suppressed'
] as const;
