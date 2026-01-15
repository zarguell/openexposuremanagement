import { useState, useCallback } from 'react';
import { UnifiedQueryBuilder } from '../components/UnifiedQueryBuilder';
import UnifiedQueryResults from '../components/UnifiedQueryResults';
import { apiClient } from '../api/client';
import { UnifiedQuery, UnifiedQueryResult } from '../types/query';

const DEFAULT_QUERY: UnifiedQuery = {
  primary_entity: 'assets',
  filters: [],
  limit: 50,
  offset: 0,
};

// Query templates for common use cases
const QUERY_TEMPLATES: Array<{
  id: string;
  name: string;
  description: string;
  query: UnifiedQuery;
}> = [
  {
    id: 'missing-crowdstrike',
    name: 'Assets Missing CrowdStrike',
    description: 'Find all assets that do not have CrowdStrike installed',
    query: {
      primary_entity: 'assets',
      join: {
        entity: 'software_inventory',
        type: 'left',
        on: {
          primary: 'id',
          joined: 'asset_id',
        },
      },
      filters: [
        {
          field: 'vendor',
          operator: 'neq',
          value: 'CrowdStrike',
        },
      ],
      limit: 50,
    },
  },
  {
    id: 'exploitable-cves',
    name: 'Assets with Exploitable CVEs',
    description: 'Find assets with critical or high CVEs that are known exploited (KEV)',
    query: {
      primary_entity: 'assets',
      join: {
        entity: 'findings',
        type: 'left',
        on: {
          primary: 'id',
          joined: 'asset_id',
        },
      },
      filters: [
        {
          field: 'is_kev',
          operator: 'eq',
          value: true,
        },
      ],
      limit: 50,
    },
  },
  {
    id: 'critical-vulns-software',
    name: 'Critical Vulnerabilities by Software',
    description: 'View software installed on assets along with critical vulnerabilities',
    query: {
      primary_entity: 'assets',
      join: {
        entity: 'findings',
        type: 'left',
        on: {
          primary: 'id',
          joined: 'asset_id',
        },
      },
      filters: [
        {
          field: 'severity',
          operator: 'eq',
          value: 'critical',
        },
      ],
      limit: 50,
    },
  },
];

function UnifiedQueries() {
  const [query, setQuery] = useState<UnifiedQuery>(DEFAULT_QUERY);
  const [result, setResult] = useState<UnifiedQueryResult | undefined>();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [selectedTemplateId, setSelectedTemplateId] = useState<string | null>(null);

  const handleExecuteQuery = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setResult(undefined);

    try {
      const response = await apiClient.queryUnified(query);
      setResult(response);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to execute query'));
    } finally {
      setIsLoading(false);
    }
  }, [query]);

  const handleLoadTemplate = useCallback((templateId: string) => {
    const template = QUERY_TEMPLATES.find(t => t.id === templateId);
    if (template) {
      setQuery(template.query);
      setSelectedTemplateId(templateId);
      setResult(undefined);
      setError(null);
    }
  }, []);

  const handleResetQuery = useCallback(() => {
    setQuery(DEFAULT_QUERY);
    setSelectedTemplateId(null);
    setResult(undefined);
    setError(null);
  }, []);

  return (
    <div style={{ padding: '2rem', maxWidth: '1400px', margin: '0 auto' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '1.875rem', fontWeight: '700', color: '#1f2937', margin: '0 0 0.5rem 0' }}>
          Unified Queries
        </h1>
        <p style={{ margin: 0, color: '#6b7280', fontSize: '1rem' }}>
          Build cross-entity correlation queries to join assets with software inventory and findings
        </p>
      </div>

      {/* Template Gallery */}
      <div style={{ marginBottom: '2rem' }}>
        <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#1f2937', marginBottom: '1rem' }}>
          Query Templates
        </h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '1rem' }}>
          {QUERY_TEMPLATES.map(template => (
            <div
              key={template.id}
              onClick={() => handleLoadTemplate(template.id)}
              style={{
                padding: '1rem',
                border: '2px solid #e5e7eb',
                borderRadius: '0.5rem',
                cursor: 'pointer',
                transition: 'all 0.2s',
                background: selectedTemplateId === template.id ? '#eff6ff' : 'white',
                borderColor: selectedTemplateId === template.id ? '#3b82f6' : '#e5e7eb',
              }}
              onMouseEnter={(e) => {
                if (selectedTemplateId !== template.id) {
                  e.currentTarget.style.borderColor = '#3b82f6';
                  e.currentTarget.style.boxShadow = '0 4px 6px -1px rgba(0, 0, 0, 0.1)';
                }
              }}
              onMouseLeave={(e) => {
                if (selectedTemplateId !== template.id) {
                  e.currentTarget.style.borderColor = '#e5e7eb';
                  e.currentTarget.style.boxShadow = 'none';
                }
              }}
            >
              <h3 style={{ fontSize: '1rem', fontWeight: '600', color: '#1f2937', margin: '0 0 0.5rem 0' }}>
                {template.name}
              </h3>
              <p style={{ fontSize: '0.875rem', color: '#6b7280', margin: 0, lineHeight: '1.4' }}>
                {template.description}
              </p>
            </div>
          ))}
        </div>
      </div>

      {/* Query Builder */}
      <div style={{ marginBottom: '2rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#1f2937', margin: 0 }}>
            Query Builder
          </h2>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button
              onClick={handleResetQuery}
              disabled={selectedTemplateId === null && JSON.stringify(query) === JSON.stringify(DEFAULT_QUERY)}
              style={{
                padding: '0.5rem 1rem',
                background: '#6b7280',
                color: 'white',
                border: 'none',
                borderRadius: '0.375rem',
                cursor: selectedTemplateId === null && JSON.stringify(query) === JSON.stringify(DEFAULT_QUERY) ? 'not-allowed' : 'pointer',
                fontSize: '0.875rem',
                fontWeight: '500',
                opacity: selectedTemplateId === null && JSON.stringify(query) === JSON.stringify(DEFAULT_QUERY) ? 0.5 : 1,
              }}
            >
              Reset
            </button>
            <button
              onClick={handleExecuteQuery}
              disabled={isLoading}
              style={{
                padding: '0.5rem 1rem',
                background: isLoading ? '#9ca3af' : '#3b82f6',
                color: 'white',
                border: 'none',
                borderRadius: '0.375rem',
                cursor: isLoading ? 'not-allowed' : 'pointer',
                fontSize: '0.875rem',
                fontWeight: '500',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
              }}
            >
              {isLoading ? (
                <>
                  <span style={{ display: 'inline-block', animation: 'spin 1s linear infinite' }}>⚙</span>
                  Executing...
                </>
              ) : (
                <>
                  ▶ Run Query
                </>
              )}
            </button>
          </div>
        </div>

        <UnifiedQueryBuilder query={query} onChange={setQuery} />
      </div>

      {/* Query Results */}
      {result || error ? (
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: '600', color: '#1f2937', marginBottom: '1rem' }}>
            Results
          </h2>
          <UnifiedQueryResults
            result={result}
            isLoading={isLoading}
            error={error}
            query={query}
          />
        </div>
      ) : null}
    </div>
  );
}

export default UnifiedQueries;
