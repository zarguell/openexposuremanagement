import React, { useCallback, useState } from 'react';
import {
  Query,
  EntityType,
  Filter,
  ALLOWED_FIELDS,
  ALLOWED_OPERATORS,
  SEVERITY_VALUES,
  SCANNER_STATUS_VALUES,
  EFFECTIVE_STATUS_VALUES
} from '../types/query';
import './QueryBuilder.css';

interface QueryBuilderProps {
  entity: EntityType;
  query: Query;
  onChange: (query: Query) => void;
  showAdvancedToggle?: boolean;
}

export const QueryBuilder: React.FC<QueryBuilderProps> = ({
  entity,
  query,
  onChange,
  showAdvancedToggle = true,
}) => {
  const allowedFields = ALLOWED_FIELDS[entity];
  const [isAdvancedMode, setIsAdvancedMode] = useState(false);
  const [jsonQuery, setJsonQuery] = useState(JSON.stringify(query, null, 2));
  const [jsonError, setJsonError] = useState<string | null>(null);

  // Update JSON when query changes in simple mode
  React.useEffect(() => {
    if (!isAdvancedMode) {
      setJsonQuery(JSON.stringify(query, null, 2));
      setJsonError(null);
    }
  }, [query, isAdvancedMode]);

  const handleAddFilter = useCallback(() => {
    const newFilter: Filter = {
      field: allowedFields[0],
      operator: ALLOWED_OPERATORS[0],
      value: ''
    };

    onChange({
      ...query,
      filters: [...query.filters, newFilter]
    });
  }, [query, onChange, allowedFields]);

  const handleRemoveFilter = useCallback((index: number) => {
    const newFilters = [...query.filters];
    newFilters.splice(index, 1);
    onChange({
      ...query,
      filters: newFilters
    });
  }, [query, onChange]);

  const handleFilterChange = useCallback((index: number, changes: Partial<Filter>) => {
    const newFilters = [...query.filters];
    newFilters[index] = { ...newFilters[index], ...changes };
    onChange({
      ...query,
      filters: newFilters
    });
  }, [query, onChange]);

  const handleJsonQueryChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newValue = e.target.value;
    setJsonQuery(newValue);

    try {
      const parsed = JSON.parse(newValue);
      // Validate the parsed query has the right structure
      if (typeof parsed === 'object' && parsed !== null) {
        const validatedQuery: Query = {
          filters: Array.isArray(parsed.filters) ? parsed.filters : [],
          sort: Array.isArray(parsed.sort) ? parsed.sort : [],
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

    // For multi-value operators, always use text input for MVP (comma separated)
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

    // In simple mode, use dropdowns for known fields
    if (!isAdvancedMode) {
      // Severity dropdown
      if (filter.field === 'severity') {
        return (
          <select
            className="filter-value-select"
            value={filter.value}
            onChange={(e) => handleChange(e.target.value)}
          >
            <option value="">Select Severity...</option>
            {SEVERITY_VALUES.map(val => (
              <option key={val} value={val}>{val}</option>
            ))}
          </select>
        );
      }

      // Scanner status dropdown
      if (filter.field === 'scanner_status') {
        return (
          <select
            className="filter-value-select"
            value={filter.value}
            onChange={(e) => handleChange(e.target.value)}
          >
            <option value="">Select Status...</option>
            {SCANNER_STATUS_VALUES.map(val => (
              <option key={val} value={val}>{val}</option>
            ))}
          </select>
        );
      }

      // Effective status dropdown
      if (filter.field === 'effective_status') {
        return (
          <select
            className="filter-value-select"
            value={filter.value}
            onChange={(e) => handleChange(e.target.value)}
          >
            <option value="">Select Status...</option>
            {EFFECTIVE_STATUS_VALUES.map(val => (
              <option key={val} value={val}>{val}</option>
            ))}
          </select>
        );
      }

      // Boolean fields dropdown
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

      // Source field dropdown
      if (filter.field === 'source') {
        return (
          <select
            className="filter-value-select"
            value={filter.value}
            onChange={(e) => handleChange(e.target.value)}
          >
            <option value="">Select Source...</option>
            <option value="tenable">Tenable</option>
            <option value="qualys">Qualys</option>
            <option value="rapid7">Rapid7</option>
          </select>
        );
      }
    }

    if (filter.field === 'severity') {
      return (
        <select
          className="filter-value-select"
          value={filter.value}
          onChange={(e) => handleChange(e.target.value)}
        >
          <option value="">Select Severity...</option>
          {SEVERITY_VALUES.map(val => (
            <option key={val} value={val}>{val}</option>
          ))}
        </select>
      );
    }

    if (filter.field === 'scanner_status') {
      return (
        <select
          className="filter-value-select"
          value={filter.value}
          onChange={(e) => handleChange(e.target.value)}
        >
          <option value="">Select Status...</option>
          {SCANNER_STATUS_VALUES.map(val => (
            <option key={val} value={val}>{val}</option>
          ))}
        </select>
      );
    }

    if (filter.field === 'effective_status') {
      return (
        <select
          className="filter-value-select"
          value={filter.value}
          onChange={(e) => handleChange(e.target.value)}
        >
          <option value="">Select Status...</option>
          {EFFECTIVE_STATUS_VALUES.map(val => (
            <option key={val} value={val}>{val}</option>
          ))}
        </select>
      );
    }

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

    // Default text input for advanced mode or unknown fields
    return (
      <input
        type="text"
        className="filter-input"
        value={filter.value}
        onChange={(e) => handleChange(e.target.value)}
        placeholder={isAdvancedMode ? "Enter value..." : "Value..."}
      />
    );
  };

  return (
    <div className="query-builder">
      {showAdvancedToggle && (
        <div style={{ marginBottom: '1rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
            <button
              onClick={() => setIsAdvancedMode(false)}
              className={`mode-toggle ${!isAdvancedMode ? 'active' : ''}`}
              style={{
                padding: '0.5rem 1rem',
                fontSize: '0.875rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                backgroundColor: !isAdvancedMode ? '#3b82f6' : 'white',
                color: !isAdvancedMode ? 'white' : '#374151',
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
              onMouseOver={(e) => {
                if (!isAdvancedMode) return;
                e.currentTarget.style.backgroundColor = '#f9fafb';
              }}
              onMouseOut={(e) => {
                if (!isAdvancedMode) return;
                e.currentTarget.style.backgroundColor = 'white';
              }}
            >
              Simple Mode
            </button>
            <button
              onClick={() => setIsAdvancedMode(true)}
              className={`mode-toggle ${isAdvancedMode ? 'active' : ''}`}
              style={{
                padding: '0.5rem 1rem',
                fontSize: '0.875rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                backgroundColor: isAdvancedMode ? '#3b82f6' : 'white',
                color: isAdvancedMode ? 'white' : '#374151',
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
              onMouseOver={(e) => {
                if (isAdvancedMode) return;
                e.currentTarget.style.backgroundColor = '#f9fafb';
              }}
              onMouseOut={(e) => {
                if (isAdvancedMode) return;
                e.currentTarget.style.backgroundColor = 'white';
              }}
            >
              Advanced Mode
            </button>
          </div>
          <span style={{ fontSize: '0.75rem', color: '#6b7280', fontStyle: 'italic' }}>
            {isAdvancedMode ? 'Free-form JSON query editor' : 'Guided dropdowns for common fields'}
          </span>
        </div>
      )}

      {isAdvancedMode ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          {jsonError && (
            <div style={{
              padding: '0.75rem',
              backgroundColor: '#fef2f2',
              border: '1px solid #fecaca',
              borderRadius: '0.375rem',
              color: '#991b1b',
              fontSize: '0.875rem',
            }}>
              <strong>JSON Error:</strong> {jsonError}
            </div>
          )}
          <textarea
            value={jsonQuery}
            onChange={handleJsonQueryChange}
            placeholder='{"filters": [...], "limit": 50, "offset": 0}'
            style={{
              width: '100%',
              minHeight: '200px',
              padding: '0.75rem',
              fontFamily: 'monospace',
              fontSize: '0.875rem',
              border: `1px solid ${jsonError ? '#ef4444' : '#d1d5db'}`,
              borderRadius: '0.375rem',
              backgroundColor: jsonError ? '#fef2f2' : 'white',
              resize: 'vertical',
            }}
            data-testid="json-query-editor"
          />
          <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>
            💡 Edit the JSON query directly. Changes are applied automatically when valid.
          </div>
        </div>
      ) : (
        <>
          <div className="filters-list">
            {query.filters.map((filter, index) => (
              <React.Fragment key={index}>
                {index > 0 && (
                  <div className="filter-separator">
                    <span className="separator-text">AND</span>
                  </div>
                )}
                <div
                  data-testid="filter-row"
                  className="filter-row"
                >
                <select
                  className="filter-select"
                  value={filter.field}
                  onChange={(e) => handleFilterChange(index, { field: e.target.value })}
                >
                  {allowedFields.map(field => (
                    <option key={field} value={field}>
                      {field}
                    </option>
                  ))}
                </select>

                <select
                  className="filter-select"
                  value={filter.operator}
                  onChange={(e) => handleFilterChange(index, { operator: e.target.value })}
                >
                  {ALLOWED_OPERATORS.map(op => (
                    <option key={op} value={op}>
                      {op}
                    </option>
                  ))}
                </select>

                {renderValueInput(filter, index)}

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

          <button
            onClick={handleAddFilter}
            className="add-btn"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="12" y1="5" x2="12" y2="19"></line>
              <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
            Add Filter
          </button>
        </>
      )}
    </div>
  );
};
