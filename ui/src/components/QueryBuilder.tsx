import React, { useCallback } from 'react';
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
}

export const QueryBuilder: React.FC<QueryBuilderProps> = ({ 
  entity, 
  query, 
  onChange 
}) => {
  const allowedFields = ALLOWED_FIELDS[entity];

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
    <div className="query-builder">
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
    </div>
  );
};
