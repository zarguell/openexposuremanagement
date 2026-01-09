import React, { useState } from 'react';

interface SearchInputProps {
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  onSubmit?: () => void;
  width?: string;
}

const SearchInput: React.FC<SearchInputProps> = ({
  placeholder = 'Search...',
  value,
  onChange,
  onSubmit,
  width = '100%'
}) => {
  const [localValue, setLocalValue] = useState(value);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onChange(localValue);
    onSubmit?.();
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setLocalValue(e.target.value);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSubmit(e);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ width }}>
      <div style={{ position: 'relative' }}>
        <input
          type="text"
          placeholder={placeholder}
          value={localValue}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          style={{
            width: '100%',
            padding: '0.75rem 3rem 0.75rem 0.75rem',
            border: '1px solid #d1d5db',
            borderRadius: '0.375rem',
            fontSize: '1rem',
            backgroundColor: 'white'
          }}
        />
        <button
          type="submit"
          style={{
            position: 'absolute',
            right: '0.5rem',
            top: '50%',
            transform: 'translateY(-50%)',
            backgroundColor: 'transparent',
            border: 'none',
            cursor: 'pointer',
            padding: '0.5rem',
            color: '#6b7280'
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.color = '#374151';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.color = '#6b7280';
          }}
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
              clipRule="evenodd"
            />
          </svg>
        </button>
      </div>
    </form>
  );
};

export default SearchInput;