import { Link } from 'react-router-dom';
import { useDashboardQueries } from '../hooks/useDashboardQueries';
import { DEFAULT_DASHBOARD } from '../config/dashboard';
import LoadingSpinner from '../components/LoadingSpinner';

function Dashboard() {
  const { results, isLoading, isError, widgets } = useDashboardQueries(DEFAULT_DASHBOARD);

  if (isLoading) {
    return (
      <div style={{ padding: '1.5rem 1rem' }}>
        <LoadingSpinner message="Loading dashboard..." />
      </div>
    );
  }

  if (isError) {
    return (
      <div style={{ padding: '1.5rem 1rem' }}>
        <div style={{
          backgroundColor: '#fef2f2',
          border: '1px solid #fecaca',
          borderRadius: '0.5rem',
          padding: '1rem',
          color: '#dc2626'
        }}>
          <h3 style={{ fontSize: '1.125rem', fontWeight: '500', marginBottom: '0.5rem' }}>
            Error loading dashboard
          </h3>
          <p>Some dashboard widgets failed to load. Please try refreshing the page.</p>
        </div>
      </div>
    );
  }

  const renderMetricWidget = (widget: typeof widgets[number], result: typeof results[number]) => {
    const count = widget.aggregation === 'count'
      ? result.meta.total_rows
      : (result.data?.[0]?.[widget.displayField || 'count'] || result.meta.total_rows);

    const content = (
      <div style={{
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
        transition: 'transform 0.2s, box-shadow 0.2s',
        cursor: widget.linkTo ? 'pointer' : 'default',
        height: '100%',
      }}
      onMouseOver={(e) => {
        if (widget.linkTo) {
          e.currentTarget.style.transform = 'translateY(-2px)';
          e.currentTarget.style.boxShadow = '0 4px 6px 0 rgba(0, 0, 0, 0.1)';
        }
      }}
      onMouseOut={(e) => {
        e.currentTarget.style.transform = 'translateY(0)';
        e.currentTarget.style.boxShadow = '0 1px 3px 0 rgba(0, 0, 0, 0.1)';
      }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1rem' }}>
          <h3 style={{ fontSize: '1rem', fontWeight: '600', color: '#374151', margin: 0 }}>
            {widget.icon && <span style={{ marginRight: '0.5rem' }}>{widget.icon}</span>}
            {widget.title}
          </h3>
          {result.error && (
            <span title={result.error} style={{ color: '#ef4444', fontSize: '1.25rem' }}>⚠️</span>
          )}
        </div>
        <div style={{ fontSize: '2.5rem', fontWeight: 'bold', color: widget.color || '#3b82f6' }}>
          {count.toLocaleString()}
        </div>
      </div>
    );

    if (widget.linkTo) {
      return <Link to={widget.linkTo} style={{ textDecoration: 'none', display: 'block' }}>{content}</Link>;
    }
    return content;
  };

  const renderListWidget = (widget: typeof widgets[number], result: typeof results[number]) => {
    return (
      <div style={{
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', marginBottom: '1rem' }}>
          <h3 style={{ fontSize: '1rem', fontWeight: '600', color: '#374151', margin: 0 }}>
            {widget.icon && <span style={{ marginRight: '0.5rem' }}>{widget.icon}</span>}
            {widget.title}
          </h3>
          {result.error && (
            <span title={result.error} style={{ color: '#ef4444', fontSize: '1.25rem', marginLeft: '0.5rem' }}>⚠️</span>
          )}
        </div>
        {result.data && result.data.length > 0 ? (
          <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
            {result.data.slice(0, 5).map((item: any, idx: number) => (
              <li
                key={idx}
                style={{
                  padding: '0.5rem 0',
                  borderBottom: idx < result.data.length - 1 ? '1px solid #f3f4f6' : 'none',
                  fontSize: '0.875rem',
                  color: '#374151',
                }}
              >
                {item[widget.displayField || 'hostname_norm'] || item.hostname || item.canonical_name || 'Unknown'}
              </li>
            ))}
          </ul>
        ) : (
          <p style={{ fontSize: '0.875rem', color: '#6b7280', fontStyle: 'italic' }}>No data available</p>
        )}
      </div>
    );
  };

  const renderWidget = (widget: typeof widgets[number], result: typeof results[number]) => {
    switch (widget.type) {
      case 'metric':
        return renderMetricWidget(widget, result);
      case 'list':
        return renderListWidget(widget, result);
      default:
        return null;
    }
  };

  return (
    <div style={{ padding: '1.5rem 1rem' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2.25rem', fontWeight: 'bold', color: '#111827', marginBottom: '0.5rem' }}>
          {DEFAULT_DASHBOARD.title}
        </h1>
        {DEFAULT_DASHBOARD.description && (
          <p style={{ color: '#6b7280', marginBottom: '1rem' }}>
            {DEFAULT_DASHBOARD.description}
          </p>
        )}

        {/* Navigation Links */}
        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
          <Link
            to="/assets/query"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              padding: '0.625rem 1.25rem',
              backgroundColor: '#3b82f6',
              color: 'white',
              borderRadius: '0.375rem',
              textDecoration: 'none',
              fontSize: '0.875rem',
              fontWeight: '500',
              transition: 'background-color 0.2s',
            }}
            onMouseOver={(e) => { e.currentTarget.style.backgroundColor = '#2563eb'; }}
            onMouseOut={(e) => { e.currentTarget.style.backgroundColor = '#3b82f6'; }}
          >
            🔍 Query Assets
          </Link>
          <Link
            to="/findings/query"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              padding: '0.625rem 1.25rem',
              backgroundColor: '#8b5cf6',
              color: 'white',
              borderRadius: '0.375rem',
              textDecoration: 'none',
              fontSize: '0.875rem',
              fontWeight: '500',
              transition: 'background-color 0.2s',
            }}
            onMouseOver={(e) => { e.currentTarget.style.backgroundColor = '#7c3aed'; }}
            onMouseOut={(e) => { e.currentTarget.style.backgroundColor = '#8b5cf6'; }}
          >
            ⚠️ Query Findings
          </Link>
        </div>
      </div>

      {/* Assets Section */}
      <div style={{ marginBottom: '2rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: '600', color: '#111827', marginBottom: '1rem' }}>
          Assets Overview
        </h2>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
          gap: '1.5rem',
        }}>
          {widgets.filter(w => w.entity === 'assets').map((widget) => {
            const result = results.find(r => r.id === widget.id)!;
            return (
              <div key={widget.id}>
                {renderWidget(widget, result)}
              </div>
            );
          })}
        </div>
      </div>

      {/* Findings Section */}
      <div style={{ marginBottom: '2rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: '600', color: '#111827', marginBottom: '1rem' }}>
          Findings Overview
        </h2>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
          gap: '1.5rem',
        }}>
          {widgets.filter(w => w.entity === 'findings' && !w.id.includes('kev') && !w.id.includes('epss')).map((widget) => {
            const result = results.find(r => r.id === widget.id)!;
            return (
              <div key={widget.id}>
                {renderWidget(widget, result)}
              </div>
            );
          })}
        </div>
      </div>

      {/* Threat Intelligence Section */}
      <div style={{ marginBottom: '2rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: '600', color: '#111827', marginBottom: '1rem' }}>
          Threat Intelligence
        </h2>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
          gap: '1.5rem',
        }}>
          {widgets.filter(w => w.id.includes('kev') || w.id.includes('epss')).map((widget) => {
            const result = results.find(r => r.id === widget.id)!;
            return (
              <div key={widget.id}>
                {renderWidget(widget, result)}
              </div>
            );
          })}
        </div>
      </div>

      {/* Info Footer */}
      <div style={{
        backgroundColor: '#f9fafb',
        border: '1px solid #e5e7eb',
        borderRadius: '0.5rem',
        padding: '1rem',
        fontSize: '0.875rem',
        color: '#6b7280',
      }}>
        <p style={{ margin: 0 }}>
          💡 <strong>Tip:</strong> Click on any metric to view the detailed query results. All widgets automatically refresh every 30 seconds.
        </p>
      </div>
    </div>
  );
}

export default Dashboard;