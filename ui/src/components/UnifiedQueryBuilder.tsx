import React, { useCallback, useState } from 'react';
import {
  UnifiedQuery,
  Filter,
  ALLOWED_FIELDS,
  ALLOWED_OPERATORS,
} from '../types/query';
import './UnifiedQueryBuilder.css';

interface UnifiedQueryBuilderProps {
  query: UnifiedQuery;
  onChange: (query: UnifiedQuery) => void;
}

/**
 * Query builder for unified queries with dot-walking support
 * Allows building cross-entity correlation queries using simple dot notation
 */
export const UnifiedQueryBuilder: React.FC<UnifiedQueryBuilderProps> = ({
  query,
  onChange,
}) => {
  const [jsonQuery, setJsonQuery] = useState(JSON.stringify(query, null, 2));
  const [jsonError, setJsonError] = useState<string | null>(null);
  const [isAdvancedMode, setIsAdvancedMode] = useState(false);

  // Get all available fields with dot-walking support
  const getAvailableFields = () => {
    const fields: Array<{category: string; field: string; label: string}> = [];

    // Primary entity fields (assets)
    ALLOWED_FIELDS.assets.forEach(field => {
      fields.push({
        category: 'Primary Entity (Assets)',
        field: field,
        label: field,
      });
    });

    // Software fields (dot-walking)
    ALLOWED_FIELDS.software.forEach(field => {
      fields.push({
        category: 'Software (via software.field)',
        field: `software.${field}`,
        label: `software.${field}`,
      });
    });

    // Findings fields (dot-walking)
    ALLOWED_FIELDS.findings.forEach(field => {
      fields.push({
        category: 'Findings (via findings.field)',
        field: `findings.${field}`,
        label: `findings.${field}`,
      });
    });

    return fields;
  };

  // Update JSON when query changes in simple mode
  React.useEffect(() => {
    if (!isAdvancedMode) {
      setJsonQuery(JSON.stringify(query, null, 2));
      setJsonError(null);
    }
  }, [query, isAdvancedMode]);

  const handleAddFilter = useCallback(() => {
    const availableFields = getAvailableFields();
    const newFilter: Filter = {
      field: availableFields[0].field,
      operator: ALLOWED_OPERATORS[0],
      value: '',
    };

    onChange({
      ...query,
      filters: [...(query.filters || []), newFilter],
    });
  }, [query, onChange]);

  const handleRemoveFilter = useCallback((index: number) => {
    const newFilters = [...(query.filters || [])];
    newFilters.splice(index, 1);
    onChange({
      ...query,
      filters: newFilters,
    });
  }, [query, onChange]);

  const handleFilterChange = useCallback((index: number, changes: Partial<Filter>) => {
    const newFilters = [...(query.filters || [])];
    newFilters[index] = { ...newFilters[index], ...changes };
    onChange({
      ...query,
      filters: newFilters,
    });
  }, [query, onChange]);

  const handleJsonQueryChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newValue = e.target.value;
    setJsonQuery(newValue);

    try {
      const parsed = JSON.parse(newValue);
      // Validate the parsed query has the right structure
      if (typeof parsed === 'object' && parsed !== null) {
        const validatedQuery: UnifiedQuery = {
          filters: Array.isArray(parsed.filters) ? parsed.filters : [],
          aggregations: Array.isArray(parsed.aggregations) ? parsed.aggregations : undefined,
          sort: Array.isArray(parsed.sort) ? parsed.sort : undefined,
          limit: typeof parsed.limit === 'number' ? parsed.limit : 50,
          offset: typeof parsed.offset === 'number' ? parsed.offset : 0,
        };
        onChange(validatedQuery);
        setJsonError(null);
      }
    } catch (err) {
      // Don't update the query on invalid JSON, just show error
      setJsonError(err instanceof Error ? err.message : 'Invalid JSON');
    }
  }, [onChange]);

  const renderValueInput = (filter: Filter, index: number) => {
    if (['is_null', 'is_not_null'].includes(filter.operator)) {
      return null;
    }

    const handleChange = (val: any) => handleFilterChange(index, { value: val });

    // For multi-value operators
    if (['in', 'not_in'].includes(filter.operator)) {
      return (
        <input
          type="text"
          className="filter-input"
          value={filter.value}
          onChange={(e) => handleChange(e.target.value)}
          placeholder="Comma separated values..."
        />
      );
    }

    // Boolean fields
    if (['is_active', 'is_kev', 'has_cve'].includes(filter.field)) {
      return (
        <select
          className="filter-value-select"
          value={String(filter.value)}
          onChange={(e) => handleChange(e.target.value === 'true')}
        >
          <option value="">Select...</option>
          <option value="true">True</option>
          <option value="false">False</option>
        </select>
      );
    }

    // Default text input
    return (
      <input
        type="text"
        className="filter-input"
        value={filter.value}
        onChange={(e) => handleChange(e.target.value)}
        placeholder="Value..."
      />
    );
  };

  const availableFields = getAvailableFields();
  const fieldsByCategory = availableFields.reduce((acc, field) => {
    if (!acc[field.category]) {
      acc[field.category] = [];
    }
    acc[field.category].push(field);
    return acc;
  }, {} as Record<string, typeof availableFields>);

  return (
    <div className="unified-query-builder">
      <div className="builder-header">
        <div className="mode-toggles">
          <button
            onClick={() => setIsAdvancedMode(false)}
            className={`mode-toggle ${!isAdvancedMode ? 'active' : ''}`}
          >
            Simple Mode
          </button>
          <button
            onClick={() => setIsAdvancedMode(true)}
            className={`mode-toggle ${isAdvancedMode ? 'active' : ''}`}
          >
            Advanced Mode
          </button>
        </div>
        <span className="mode-hint">
          {isAdvancedMode
            ? 'Free-form JSON query editor'
            : 'Guided builder with dot-walking support'}
        </span>
      </div>

      {isAdvancedMode ? (
        <div className="advanced-mode">
          {jsonError && (
            <div className="json-error">
              <strong>JSON Error:</strong> {jsonError}
            </div>
          )}
          <textarea
            value={jsonQuery}
            onChange={handleJsonQueryChange}
            placeholder='{"filters": [{"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}]}'
            className="json-editor"
            data-testid="json-query-editor"
          />
          <div className="editor-hint">
            💡 Edit the JSON query directly. Use dot notation for related entities (e.g., <code>software.vendor</code>, <code>findings.severity</code>).
            Add <code>"negate": true</code> to filter for "NOT EXISTS" (e.g., assets without specific software).
          </div>
        </div>
      ) : (
        <div className="simple-mode">
          {/* Dot-Walking Help */}
          <div className="help-section">
            <h3>💡 Dot-Walking Syntax</h3>
            <p>
              Use dot notation to filter on related entities:
            </p>
            <ul>
              <li><code>software.vendor</code> - Filter by software vendor</li>
              <li><code>software.product_name</code> - Filter by software product</li>
              <li><code>findings.severity</code> - Filter by finding severity</li>
              <li><code>findings.cve</code> - Filter by CVE ID</li>
            </ul>
            <p>
              <strong>Anti-join:</strong> Check "Negate" to find assets <em>without</em> matching related records
              (e.g., <code>software.vendor = "CrowdStrike"</code> with Negate → assets without CrowdStrike).
            </p>
          </div>

          {/* Filters */}
          <div className="filters-section">
            <h3>Filters</h3>
            <div className="filters-list">
              {(query.filters || []).map((filter, index) => (
                <React.Fragment key={index}>
                  {index > 0 && (
                    <div className="filter-separator">
                      <span className="separator-text">AND</span>
                    </div>
                  )}
                  <div className="filter-row" data-testid="filter-row">
                    <select
                      className="filter-select"
                      value={filter.field}
                      onChange={(e) => handleFilterChange(index, { field: e.target.value })}
                    >
                      {Object.entries(fieldsByCategory).map(([category, fields]) => (
                        <optgroup key={category} label={category}>
                          {fields.map(field => (
                            <option key={field.field} value={field.field}>{field.label}</option>
                          ))}
                        </optgroup>
                      ))}
                    </select>

                    <select
                      className="filter-select"
                      value={filter.operator}
                      onChange={(e) => handleFilterChange(index, { operator: e.target.value })}
                    >
                      {ALLOWED_OPERATORS.map(op => (
                        <option key={op} value={op}>{op}</option>
                      ))}
                    </select>

                    {renderValueInput(filter, index)}

                    {/* Negate checkbox for anti-join patterns */}
                    {filter.field.includes('.') && (
                      <label className="negate-checkbox">
                        <input
                          type="checkbox"
                          checked={filter.negate || false}
                          onChange={(e) => handleFilterChange(index, { negate: e.target.checked })}
                        />
                        <span>Negate (NOT EXISTS)</span>
                      </label>
                    )}

                    <button
                      onClick={() => handleRemoveFilter(index)}
                      className="remove-btn"
                      aria-label="Remove filter"
                      title="Remove filter"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                      </svg>
                    </button>
                  </div>
                </React.Fragment>
              ))}
            </div>

            <button onClick={handleAddFilter} className="add-btn">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
              Add Filter
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default UnifiedQueryBuilder;
