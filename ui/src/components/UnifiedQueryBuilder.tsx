import React, { useCallback, useState } from 'react';
import {
  UnifiedQuery,
  Join,
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
 * Query builder for unified queries with JOIN support
 * Allows building cross-entity correlation queries
 */
export const UnifiedQueryBuilder: React.FC<UnifiedQueryBuilderProps> = ({
  query,
  onChange,
}) => {
  const [jsonQuery, setJsonQuery] = useState(JSON.stringify(query, null, 2));
  const [jsonError, setJsonError] = useState<string | null>(null);
  const [isAdvancedMode, setIsAdvancedMode] = useState(false);

  // Get primary entity fields (assets)
  const primaryFields = ALLOWED_FIELDS.assets;

  // Get joined entity fields based on join type
  const getJoinedFields = () => {
    if (!query.join) return [];
    if (query.join.entity === 'software_inventory') {
      return ['vendor', 'product_name', 'version', 'cpe_string', 'title_formatted', 'first_seen_at', 'last_seen_at'];
    }
    if (query.join.entity === 'findings') {
      return ALLOWED_FIELDS.findings;
    }
    return [];
  };

  // Update JSON when query changes in simple mode
  React.useEffect(() => {
    if (!isAdvancedMode) {
      setJsonQuery(JSON.stringify(query, null, 2));
      setJsonError(null);
    }
  }, [query, isAdvancedMode]);

  const handleAddJoin = useCallback(() => {
    const newJoin: Join = {
      entity: 'software_inventory',
      type: 'left',
      on: {
        primary: 'id',
        joined: 'asset_id',
      },
    };

    onChange({
      ...query,
      join: newJoin,
    });
  }, [query, onChange]);

  const handleRemoveJoin = useCallback(() => {
    const { join, ...rest } = query;
    onChange(rest);
  }, [query, onChange]);

  const handleJoinChange = useCallback((changes: Partial<Join>) => {
    if (!query.join) return;
    onChange({
      ...query,
      join: { ...query.join, ...changes },
    });
  }, [query, onChange]);

  const handleAddFilter = useCallback(() => {
    const field = primaryFields[0];
    const newFilter: Filter = {
      field,
      operator: ALLOWED_OPERATORS[0],
      value: '',
    };

    onChange({
      ...query,
      filters: [...(query.filters || []), newFilter],
    });
  }, [query, onChange, primaryFields]);

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
          primary_entity: parsed.primary_entity || 'assets',
          join: parsed.join,
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
            : 'Guided builder for JOIN queries'}
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
            placeholder='{"primary_entity": "assets", "join": {...}, "filters": [...]}'
            className="json-editor"
            data-testid="json-query-editor"
          />
          <div className="editor-hint">
            💡 Edit the JSON query directly. Changes are applied automatically when valid.
          </div>
        </div>
      ) : (
        <div className="simple-mode">
          {/* Join Configuration */}
          <div className="join-section">
            <h3>Join Configuration</h3>
            {!query.join ? (
              <div className="join-empty">
                <p>No join configured. Query will return only asset data.</p>
                <button onClick={handleAddJoin} className="add-join-btn">
                  Add Join
                </button>
              </div>
            ) : (
              <div className="join-config">
                <div className="join-row">
                  <label>Joined Entity:</label>
                  <select
                    value={query.join.entity}
                    onChange={(e) => handleJoinChange({
                      entity: e.target.value as 'software_inventory' | 'findings'
                    })}
                    className="join-select"
                  >
                    <option value="software_inventory">Software Inventory</option>
                    <option value="findings">Findings</option>
                  </select>
                </div>

                <div className="join-info">
                  <strong>Join Type:</strong> LEFT JOIN
                </div>

                <div className="join-info">
                  <strong>On:</strong> assets.{query.join.on.primary} = {query.join.entity}.{query.join.on.joined}
                </div>

                <button
                  onClick={handleRemoveJoin}
                  className="remove-join-btn"
                >
                  Remove Join
                </button>
              </div>
            )}
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
                      <optgroup label="Primary Entity (Assets)">
                        {primaryFields.map(field => (
                          <option key={field} value={field}>{field}</option>
                        ))}
                      </optgroup>
                      {query.join && (
                        <optgroup label={`Joined Entity (${query.join.entity})`}>
                          {getJoinedFields().map(field => (
                            <option key={field} value={field}>{field}</option>
                          ))}
                        </optgroup>
                      )}
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
