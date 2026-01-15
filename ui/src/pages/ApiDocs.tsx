import React, { useState } from 'react';
import './ApiDocs.css';

interface ApiEndpoint {
  method: string;
  path: string;
  description: string;
  auth: boolean;
  roles?: string[];
  parameters?: Array<{
    name: string;
    type: string;
    required: boolean;
    description: string;
    location: 'query' | 'path' | 'body';
  }>;
  requestBody?: {
    contentType: string;
    schema: any;
    example: any;
  };
  responses: Array<{
    code: number;
    description: string;
    example?: any;
  }>;
}

const API_ENDPOINTS: ApiEndpoint[] = [
  // Health endpoints
  {
    method: 'GET',
    path: '/healthz',
    description: 'Health check endpoint',
    auth: false,
    responses: [
      { code: 200, description: 'Service is healthy', example: { status: 'healthy' } },
    ],
  },
  {
    method: 'GET',
    path: '/healthz/live',
    description: 'Liveness probe',
    auth: false,
    responses: [
      { code: 200, description: 'Service is alive', example: { status: 'alive' } },
    ],
  },
  {
    method: 'GET',
    path: '/healthz/ready',
    description: 'Readiness probe (checks database)',
    auth: false,
    responses: [
      { code: 200, description: 'Service is ready', example: { status: 'ready' } },
      { code: 503, description: 'Service not ready (e.g., database unavailable)' },
    ],
  },

  // User endpoints
  {
    method: 'GET',
    path: '/me',
    description: 'Get current user information',
    auth: true,
    responses: [
      {
        code: 200,
        description: 'User information',
        example: {
          id: '123',
          email: 'user@example.com',
          name: 'John Doe',
          roles: ['analyst'],
          tenant_id: 'abc-123',
        },
      },
    ],
  },

  // Ingestion endpoints
  {
    method: 'POST',
    path: '/ingest/vm/findings',
    description: 'Ingest vulnerability findings from VM scanners (Tenable, Qualys, etc.)',
    auth: true,
    parameters: [
      {
        name: 'source',
        type: 'string',
        required: true,
        description: 'Scanner source (e.g., "tenable", "qualys", "rapid7")',
        location: 'body',
      },
    ],
    requestBody: {
      contentType: 'application/json',
      schema: {
        type: 'object',
        properties: {
          source: { type: 'string', enum: ['tenable', 'qualys', 'rapid7'] },
          scan_info: {
            type: 'object',
            properties: {
              scan_id: { type: 'string' },
              scan_start: { type: 'string', format: 'date-time' },
              scan_end: { type: 'string', format: 'date-time' },
            },
          },
          findings: {
            type: 'array',
            items: {
              type: 'object',
              properties: {
                asset: {
                  type: 'object',
                  properties: {
                    hostname: { type: 'string' },
                    ip_address: { type: 'string' },
                    mac_address: { type: 'string' },
                    external_id: { type: 'string' },
                  },
                },
                finding: {
                  type: 'object',
                  properties: {
                    source_definition_id: { type: 'string' },
                    title: { type: 'string' },
                    severity: { type: 'string', enum: ['critical', 'high', 'medium', 'low'] },
                    scanner_status: { type: 'string', enum: ['open', 'fixed', 'fixed_by_verification'] },
                    first_observed_at: { type: 'string', format: 'date-time' },
                    last_observed_at: { type: 'string', format: 'date-time' },
                  },
                },
                cve: { type: 'array', items: { type: 'string' } },
                software: {
                  type: 'array',
                  items: {
                    type: 'object',
                    properties: {
                      cpe_string: { type: 'string' },
                      vendor: { type: 'string' },
                      product_name: { type: 'string' },
                      version: { type: 'string' },
                    },
                  },
                },
              },
            },
          },
        },
        required: ['source', 'findings'],
      },
      example: {
        source: 'tenable',
        scan_info: {
          scan_id: 'abc123',
          scan_start: '2025-01-14T10:00:00Z',
          scan_end: '2025-01-14T11:00:00Z',
        },
        findings: [
          {
            asset: {
              hostname: 'webserver01',
              ip_address: '192.168.1.100',
            },
            finding: {
              source_definition_id: '12345',
              title: 'Apache Log4j Remote Code Execution Vulnerability',
              severity: 'critical',
              scanner_status: 'open',
              first_observed_at: '2025-01-14T10:05:00Z',
              last_observed_at: '2025-01-14T10:05:00Z',
            },
            cve: ['CVE-2021-44228'],
            software: [
              {
                cpe_string: 'cpe:2.3:a:apache:log4j:2.14.1:*:*:*:*:*:*:*',
                vendor: 'Apache',
                product_name: 'Log4j',
                version: '2.14.1',
              },
            ],
          },
        ],
      },
    },
    responses: [
      {
        code: 200,
        description: 'Findings ingested successfully',
        example: {
          message: 'Findings ingested successfully',
          stats: {
            assets_created: 1,
            assets_updated: 0,
            findings_created: 5,
            findings_updated: 2,
            software_upserted: 10,
          },
        },
      },
      { code: 400, description: 'Invalid request payload' },
      { code: 401, description: 'Unauthorized' },
      { code: 403, description: 'Forbidden (API key source mismatch)' },
    ],
  },

  // Asset endpoints
  {
    method: 'GET',
    path: '/assets',
    description: 'List assets with optional filtering',
    auth: true,
    parameters: [
      { name: 'query', type: 'string', required: false, description: 'Search query', location: 'query' },
      { name: 'limit', type: 'integer', required: false, description: 'Max results (default: 50)', location: 'query' },
      { name: 'offset', type: 'integer', required: false, description: 'Results offset (default: 0)', location: 'query' },
    ],
    responses: [
      {
        code: 200,
        description: 'List of assets',
        example: {
          data: [
            {
              id: 1,
              canonical_name: 'webserver01.example.com',
              first_seen_at: '2025-01-01T00:00:00Z',
              last_seen_at: '2025-01-14T10:00:00Z',
              is_active: true,
            },
          ],
          meta: { total_rows: 100, execution_time_ms: 15, has_more: true },
        },
      },
    ],
  },
  {
    method: 'GET',
    path: '/assets/{id}',
    description: 'Get asset details by ID',
    auth: true,
    parameters: [
      { name: 'id', type: 'integer', required: true, description: 'Asset ID', location: 'path' },
    ],
    responses: [
      {
        code: 200,
        description: 'Asset details',
        example: {
          id: 1,
          canonical_name: 'webserver01.example.com',
          first_seen_at: '2025-01-01T00:00:00Z',
          last_seen_at: '2025-01-14T10:00:00Z',
          is_active: true,
          software: [
            {
              id: 1,
              cpe_string: 'cpe:2.3:a:apache:log4j:2.14.1:*:*:*:*:*:*:*',
              vendor: 'Apache',
              product_name: 'Log4j',
              version: '2.14.1',
              title_formatted: 'Apache Log4j 2.14.1',
              first_seen_at: '2025-01-01T00:00:00Z',
              last_seen_at: '2025-01-14T10:00:00Z',
            },
          ],
        },
      },
      { code: 404, description: 'Asset not found' },
    ],
  },
  {
    method: 'GET',
    path: '/assets/{id}/software',
    description: 'Get software installed on an asset',
    auth: true,
    parameters: [
      { name: 'id', type: 'integer', required: true, description: 'Asset ID', location: 'path' },
    ],
    responses: [
      {
        code: 200,
        description: 'List of software',
        example: {
          data: [
            {
              id: 1,
              cpe_string: 'cpe:2.3:a:apache:log4j:2.14.1:*:*:*:*:*:*:*',
              vendor: 'Apache',
              product_name: 'Log4j',
              version: '2.14.1',
              title_formatted: 'Apache Log4j 2.14.1',
            },
          ],
          meta: { total_rows: 25, execution_time_ms: 10 },
        },
      },
    ],
  },

  // Findings endpoints
  {
    method: 'GET',
    path: '/findings',
    description: 'List findings with optional filtering',
    auth: true,
    parameters: [
      { name: 'source', type: 'string', required: false, description: 'Filter by source', location: 'query' },
      { name: 'severity', type: 'string', required: false, description: 'Filter by severity', location: 'query' },
      { name: 'effective_status', type: 'string', required: false, description: 'Filter by effective status', location: 'query' },
      { name: 'cve', type: 'string', required: false, description: 'Filter by CVE', location: 'query' },
      { name: 'asset', type: 'string', required: false, description: 'Filter by asset name', location: 'query' },
      { name: 'include_intel', type: 'boolean', required: false, description: 'Include threat intel data', location: 'query' },
      { name: 'limit', type: 'integer', required: false, description: 'Max results (default: 50)', location: 'query' },
      { name: 'offset', type: 'integer', required: false, description: 'Results offset (default: 0)', location: 'query' },
    ],
    responses: [
      {
        code: 200,
        description: 'List of findings',
        example: {
          data: [
            {
              id: 1,
              asset_name: 'webserver01.example.com',
              severity: 'critical',
              scanner_status: 'open',
              effective_status: 'open',
              source: 'tenable',
              title: 'Apache Log4j Remote Code Execution Vulnerability',
              cve: 'CVE-2021-44228',
              epss_score: 0.92,
              is_kev: true,
              first_observed_at: '2025-01-14T10:00:00Z',
              last_observed_at: '2025-01-14T10:00:00Z',
            },
          ],
          meta: { total_rows: 500, execution_time_ms: 20, has_more: true },
        },
      },
    ],
  },

  // Software endpoints
  {
    method: 'GET',
    path: '/software',
    description: 'Browse software catalog',
    auth: true,
    parameters: [
      { name: 'vendor', type: 'string', required: false, description: 'Filter by vendor', location: 'query' },
      { name: 'product', type: 'string', required: false, description: 'Filter by product name', location: 'query' },
      { name: 'version', type: 'string', required: false, description: 'Filter by version', location: 'query' },
      { name: 'cpe', type: 'string', required: false, description: 'Filter by CPE string', location: 'query' },
      { name: 'limit', type: 'integer', required: false, description: 'Max results (default: 50)', location: 'query' },
      { name: 'offset', type: 'integer', required: false, description: 'Results offset (default: 0)', location: 'query' },
    ],
    responses: [
      {
        code: 200,
        description: 'Software catalog',
        example: {
          data: [
            {
              id: 1,
              cpe_string: 'cpe:2.3:a:apache:log4j:2.14.1:*:*:*:*:*:*:*',
              vendor: 'Apache',
              product_name: 'Log4j',
              version: '2.14.1',
              title_formatted: 'Apache Log4j 2.14.1',
              assets_count: 15,
            },
          ],
          meta: { total_rows: 1000, execution_time_ms: 25 },
        },
      },
    ],
  },
  {
    method: 'GET',
    path: '/software/{id}',
    description: 'Get software details with affected assets and findings',
    auth: true,
    parameters: [
      { name: 'id', type: 'integer', required: true, description: 'Software ID', location: 'path' },
    ],
    responses: [
      {
        code: 200,
        description: 'Software details',
        example: {
          id: 1,
          cpe_string: 'cpe:2.3:a:apache:log4j:2.14.1:*:*:*:*:*:*:*',
          vendor: 'Apache',
          product_name: 'Log4j',
          version: '2.14.1',
          title_formatted: 'Apache Log4j 2.14.1',
          assets: [
            { id: 1, canonical_name: 'webserver01.example.com' },
            { id: 2, canonical_name: 'appserver01.example.com' },
          ],
          findings: [
            {
              id: 1,
              title: 'Apache Log4j Remote Code Execution Vulnerability',
              severity: 'critical',
              cve: 'CVE-2021-44228',
            },
          ],
        },
      },
    ],
  },

  // Query endpoints
  {
    method: 'POST',
    path: '/query/findings',
    description: 'Execute a query on findings',
    auth: true,
    requestBody: {
      contentType: 'application/json',
      schema: {
        type: 'object',
        properties: {
          filters: {
            type: 'array',
            items: {
              type: 'object',
              properties: {
                field: { type: 'string' },
                operator: { type: 'string' },
                value: {},
              },
            },
          },
          sort: { type: 'array' },
          limit: { type: 'integer' },
          offset: { type: 'integer' },
        },
      },
      example: {
        filters: [
          { field: 'severity', operator: 'eq', value: 'critical' },
          { field: 'effective_status', operator: 'eq', value: 'open' },
        ],
        sort: [{ field: 'last_observed_at', order: 'desc' }],
        limit: 50,
        offset: 0,
      },
    },
    responses: [
      {
        code: 200,
        description: 'Query results',
        example: {
          data: [
            {
              id: 1,
              severity: 'critical',
              effective_status: 'open',
              cve: 'CVE-2021-44228',
            },
          ],
          meta: { total_rows: 10, execution_time_ms: 15 },
        },
      },
    ],
  },
  {
    method: 'POST',
    path: '/query/assets',
    description: 'Execute a query on assets',
    auth: true,
    requestBody: {
      contentType: 'application/json',
      example: {
        filters: [{ field: 'is_active', operator: 'eq', value: true }],
        limit: 50,
      },
    },
    responses: [{ code: 200, description: 'Query results' }],
  },
  {
    method: 'POST',
    path: '/query/software_inventory',
    description: 'Execute a query on software inventory',
    auth: true,
    requestBody: {
      contentType: 'application/json',
      example: {
        filters: [{ field: 'vendor', operator: 'eq', value: 'Microsoft' }],
        limit: 50,
      },
    },
    responses: [{ code: 200, description: 'Query results' }],
  },
  {
    method: 'POST',
    path: '/query/unified',
    description: 'Execute a unified query with JOINs (cross-entity correlation)',
    auth: true,
    requestBody: {
      contentType: 'application/json',
      example: {
        primary_entity: 'assets',
        join: {
          entity: 'software_inventory',
          type: 'left',
          on: { primary: 'id', joined: 'asset_id' },
        },
        filters: [{ field: 'vendor', operator: 'neq', value: 'CrowdStrike' }],
        limit: 50,
      },
    },
    responses: [
      {
        code: 200,
        description: 'Unified query results with joined data',
        example: {
          data: [
            {
              assets_id: 1,
              assets_canonical_name: 'webserver01.example.com',
              software_inventory_vendor: 'Apache',
              software_inventory_product_name: 'Log4j',
            },
          ],
          meta: { total_rows: 100, execution_time_ms: 45 },
        },
      },
    ],
  },

  // Dashboard endpoints
  {
    method: 'GET',
    path: '/dashboard',
    description: 'Get dashboard statistics',
    auth: true,
    responses: [
      {
        code: 200,
        description: 'Dashboard data',
        example: {
          total_assets: 100,
          active_assets: 95,
          total_findings: 5000,
          open_findings: 250,
          critical_findings: 50,
          high_findings: 200,
        },
      },
    ],
  },
  {
    method: 'POST',
    path: '/dashboard/refresh',
    description: 'Refresh dashboard materialized views',
    auth: true,
    roles: ['admin'],
    responses: [
      { code: 200, description: 'Views refreshed successfully', example: { message: 'Dashboard views refreshed' } },
    ],
  },

  // Intel endpoints
  {
    method: 'GET',
    path: '/intel/status',
    description: 'Get threat intel sync status',
    auth: true,
    responses: [
      {
        code: 200,
        description: 'Intel sync status',
        example: {
          last_sync: '2025-01-14T09:00:00Z',
          status: 'success',
          nvd_count: 150000,
          epss_count: 150000,
          kev_count: 1000,
        },
      },
    ],
  },
  {
    method: 'POST',
    path: '/intel/refresh',
    description: 'Trigger threat intel sync (admin only)',
    auth: true,
    roles: ['admin'],
    responses: [
      { code: 200, description: 'Sync started', example: { message: 'Intel sync started' } },
      { code: 403, description: 'Forbidden (admin only)' },
    ],
  },
];

function ApiDocs() {
  const [selectedEndpoint, setSelectedEndpoint] = useState<ApiEndpoint | null>(null);
  const [filterText, setFilterText] = useState('');
  const [selectedMethod, setSelectedMethod] = useState<string>('ALL');

  const filteredEndpoints = API_ENDPOINTS.filter(endpoint => {
    const matchesFilter =
      filterText === '' ||
      endpoint.path.toLowerCase().includes(filterText.toLowerCase()) ||
      endpoint.description.toLowerCase().includes(filterText.toLowerCase());
    const matchesMethod = selectedMethod === 'ALL' || endpoint.method === selectedMethod;
    return matchesFilter && matchesMethod;
  });

  const getMethodColor = (method: string) => {
    const colors = {
      GET: '#61affe',
      POST: '#49cc90',
      PUT: '#fca130',
      DELETE: '#f93e3e',
      PATCH: '#50e3c2',
    };
    return colors[method as keyof typeof colors] || '#999';
  };

  return (
    <div style={{ padding: '2rem', maxWidth: '1400px', margin: '0 auto' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2.25rem', fontWeight: 'bold', color: '#111827', marginBottom: '0.5rem' }}>
          API Documentation
        </h1>
        <p style={{ color: '#6b7280', fontSize: '1rem', margin: 0 }}>
          Complete reference for the Open Exposure Management API
        </p>
      </div>

      {/* Filters */}
      <div style={{ marginBottom: '2rem', display: 'flex', gap: '1rem', alignItems: 'center' }}>
        <input
          type="text"
          placeholder="Search endpoints..."
          value={filterText}
          onChange={(e) => setFilterText(e.target.value)}
          style={{
            flex: 1,
            padding: '0.75rem',
            border: '1px solid #d1d5db',
            borderRadius: '0.5rem',
            fontSize: '0.875rem',
          }}
        />
        <select
          value={selectedMethod}
          onChange={(e) => setSelectedMethod(e.target.value)}
          style={{
            padding: '0.75rem',
            border: '1px solid #d1d5db',
            borderRadius: '0.5rem',
            fontSize: '0.875rem',
            cursor: 'pointer',
          }}
        >
          <option value="ALL">All Methods</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
        </select>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '350px 1fr', gap: '2rem' }}>
        {/* Endpoint List */}
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '1rem' }}>
            Endpoints ({filteredEndpoints.length})
          </h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            {filteredEndpoints.map((endpoint, index) => (
              <div
                key={index}
                onClick={() => setSelectedEndpoint(endpoint)}
                style={{
                  padding: '0.75rem',
                  background: selectedEndpoint === endpoint ? '#eff6ff' : 'white',
                  border: '1px solid #e5e7eb',
                  borderRadius: '0.5rem',
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                }}
                onMouseEnter={(e) => {
                  if (selectedEndpoint !== endpoint) {
                    e.currentTarget.style.background = '#f9fafb';
                    e.currentTarget.style.borderColor = '#3b82f6';
                  }
                }}
                onMouseOut={(e) => {
                  if (selectedEndpoint !== endpoint) {
                    e.currentTarget.style.background = 'white';
                    e.currentTarget.style.borderColor = '#e5e7eb';
                  }
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
                  <span
                    style={{
                      padding: '0.125rem 0.375rem',
                      borderRadius: '0.25rem',
                      fontSize: '0.75rem',
                      fontWeight: '600',
                      background: getMethodColor(endpoint.method),
                      color: 'white',
                      minWidth: '50px',
                      textAlign: 'center',
                    }}
                  >
                    {endpoint.method}
                  </span>
                  <span
                    style={{
                      fontSize: '0.8125rem',
                      color: '#1f2937',
                      fontFamily: 'monospace',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {endpoint.path}
                  </span>
                </div>
                <div style={{ fontSize: '0.75rem', color: '#6b7280', paddingLeft: '3.25rem' }}>
                  {endpoint.description}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Endpoint Details */}
        <div>
          {selectedEndpoint ? (
            <div style={{ background: 'white', borderRadius: '0.5rem', padding: '1.5rem', border: '1px solid #e5e7eb' }}>
              <div style={{ marginBottom: '1.5rem', paddingBottom: '1rem', borderBottom: '1px solid #e5e7eb' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.75rem' }}>
                  <span
                    style={{
                      padding: '0.25rem 0.5rem',
                      borderRadius: '0.375rem',
                      fontSize: '0.875rem',
                      fontWeight: '600',
                      background: getMethodColor(selectedEndpoint.method),
                      color: 'white',
                      minWidth: '60px',
                      textAlign: 'center',
                    }}
                  >
                    {selectedEndpoint.method}
                  </span>
                  <span style={{ fontSize: '1.125rem', fontFamily: 'monospace', color: '#1f2937' }}>
                    {selectedEndpoint.path}
                  </span>
                </div>
                <p style={{ margin: 0, color: '#6b7280', fontSize: '0.875rem' }}>
                  {selectedEndpoint.description}
                </p>
                <div style={{ marginTop: '0.75rem', display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                  {selectedEndpoint.auth ? (
                    <span style={{ padding: '0.125rem 0.5rem', background: '#dbeafe', color: '#1e40af', borderRadius: '0.25rem', fontSize: '0.75rem' }}>
                      🔒 Auth Required
                    </span>
                  ) : (
                    <span style={{ padding: '0.125rem 0.5rem', background: '#d1fae5', color: '#065f46', borderRadius: '0.25rem', fontSize: '0.75rem' }}>
                      ✓ Public
                    </span>
                  )}
                  {selectedEndpoint.roles && (
                    <span style={{ padding: '0.125rem 0.5rem', background: '#fef3c7', color: '#92400e', borderRadius: '0.25rem', fontSize: '0.75rem' }}>
                      Roles: {selectedEndpoint.roles.join(', ')}
                    </span>
                  )}
                </div>
              </div>

              {/* Parameters */}
              {selectedEndpoint.parameters && selectedEndpoint.parameters.length > 0 && (
                <div style={{ marginBottom: '1.5rem' }}>
                  <h3 style={{ fontSize: '1rem', fontWeight: '600', marginBottom: '0.75rem', color: '#1f2937' }}>
                    Parameters
                  </h3>
                  <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.875rem' }}>
                    <thead>
                      <tr style={{ borderBottom: '2px solid #e5e7eb' }}>
                        <th style={{ padding: '0.5rem', textAlign: 'left', color: '#6b7280' }}>Name</th>
                        <th style={{ padding: '0.5rem', textAlign: 'left', color: '#6b7280' }}>Type</th>
                        <th style={{ padding: '0.5rem', textAlign: 'left', color: '#6b7280' }}>Location</th>
                        <th style={{ padding: '0.5rem', textAlign: 'left', color: '#6b7280' }}>Required</th>
                        <th style={{ padding: '0.5rem', textAlign: 'left', color: '#6b7280' }}>Description</th>
                      </tr>
                    </thead>
                    <tbody>
                      {selectedEndpoint.parameters.map((param, idx) => (
                        <tr key={idx} style={{ borderBottom: '1px solid #f3f4f6' }}>
                          <td style={{ padding: '0.5rem', fontFamily: 'monospace', color: '#1f2937' }}>{param.name}</td>
                          <td style={{ padding: '0.5rem', color: '#6b7280' }}>{param.type}</td>
                          <td style={{ padding: '0.5rem', color: '#6b7280' }}>{param.location}</td>
                          <td style={{ padding: '0.5rem', color: param.required ? '#dc2626' : '#6b7280' }}>
                            {param.required ? 'Required' : 'Optional'}
                          </td>
                          <td style={{ padding: '0.5rem', color: '#374151' }}>{param.description}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Request Body */}
              {selectedEndpoint.requestBody && (
                <div style={{ marginBottom: '1.5rem' }}>
                  <h3 style={{ fontSize: '1rem', fontWeight: '600', marginBottom: '0.75rem', color: '#1f2937' }}>
                    Request Body
                  </h3>
                  <div style={{ marginBottom: '0.75rem' }}>
                    <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                      Content-Type: {selectedEndpoint.requestBody.contentType}
                    </span>
                  </div>
                  {selectedEndpoint.requestBody.example && (
                    <div>
                      <div style={{ fontSize: '0.875rem', fontWeight: '500', marginBottom: '0.5rem', color: '#1f2937' }}>
                        Example:
                      </div>
                      <pre
                        style={{
                          background: '#1f2937',
                          color: '#f9fafb',
                          padding: '1rem',
                          borderRadius: '0.375rem',
                          overflow: 'auto',
                          fontSize: '0.8125rem',
                          margin: 0,
                        }}
                      >
                        {JSON.stringify(selectedEndpoint.requestBody.example, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              )}

              {/* Responses */}
              <div>
                <h3 style={{ fontSize: '1rem', fontWeight: '600', marginBottom: '0.75rem', color: '#1f2937' }}>
                  Responses
                </h3>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  {selectedEndpoint.responses.map((response, idx) => (
                    <div key={idx} style={{ border: '1px solid #e5e7eb', borderRadius: '0.5rem', overflow: 'hidden' }}>
                      <div
                        style={{
                          padding: '0.5rem 0.75rem',
                          background: parseInt(String(response.code)) >= 400 ? '#fef2f2' : '#f0fdf4',
                          borderBottom: '1px solid #e5e7eb',
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.5rem',
                        }}
                      >
                        <span
                          style={{
                            padding: '0.125rem 0.375rem',
                            borderRadius: '0.25rem',
                            fontSize: '0.75rem',
                            fontWeight: '600',
                            background: parseInt(String(response.code)) >= 400 ? '#dc2626' : '#16a34a',
                            color: 'white',
                          }}
                        >
                          {response.code}
                        </span>
                        <span style={{ fontSize: '0.875rem', color: '#374151' }}>{response.description}</span>
                      </div>
                      {response.example && (
                        <pre
                          style={{
                            margin: 0,
                            padding: '1rem',
                            background: '#1f2937',
                            color: '#f9fafb',
                            overflow: 'auto',
                            fontSize: '0.8125rem',
                          }}
                        >
                          {JSON.stringify(response.example, null, 2)}
                        </pre>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ) : (
            <div
              style={{
                background: 'white',
                borderRadius: '0.5rem',
                padding: '3rem',
                border: '1px solid #e5e7eb',
                textAlign: 'center',
                color: '#6b7280',
              }}
            >
              <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>📚</div>
              <div style={{ fontSize: '1.125rem', fontWeight: '500', marginBottom: '0.5rem' }}>
                Select an Endpoint
              </div>
              <div style={{ fontSize: '0.875rem' }}>
                Click on any endpoint from the list to view its documentation
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default ApiDocs;
