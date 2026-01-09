import { useState, useEffect } from 'react';

interface LogEntry {
  timestamp: string;
  message: string;
  data?: any;
}

// Capture console.log output
export const DebugPanel: React.FC<{ visible?: boolean }> = ({ visible = true }) => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isMinimized, setIsMinimized] = useState(true);

  useEffect(() => {
    if (!visible) return;

    // Capture console.log
    const originalLog = console.log;
    const originalError = console.error;
    const originalWarn = console.warn;

    const addLog = (message: string, data?: any) => {
      setLogs(prev => [...prev.slice(-49), {
        timestamp: new Date().toLocaleTimeString(),
        message,
        data
      }]);
    };

    console.log = (...args) => {
      originalLog.apply(console, args);
      const message = args.map(arg =>
        typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg)
      ).join(' ');
      addLog(message);
    };

    console.error = (...args) => {
      originalError.apply(console, args);
      const message = '❌ ' + args.map(arg =>
        typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg)
      ).join(' ');
      addLog(message);
    };

    console.warn = (...args) => {
      originalWarn.apply(console, args);
      const message = '⚠️ ' + args.map(arg =>
        typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg)
      ).join(' ');
      addLog(message);
    };

    return () => {
      console.log = originalLog;
      console.error = originalError;
      console.warn = originalWarn;
    };
  }, [visible]);

  if (!visible) return null;

  return (
    <div style={{
      position: 'fixed',
      bottom: 0,
      left: 0,
      right: 0,
      backgroundColor: '#1f2937',
      color: '#f9fafb',
      fontFamily: 'monospace',
      fontSize: '0.75rem',
      zIndex: 9999,
      borderTop: '2px solid #374151',
      maxHeight: isMinimized ? '40px' : '300px',
      overflow: 'auto',
    }}>
      <div
        onClick={() => setIsMinimized(!isMinimized)}
        style={{
          padding: '0.5rem 1rem',
          backgroundColor: '#111827',
          cursor: 'pointer',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          position: 'sticky',
          top: 0,
        }}
      >
        <span style={{ fontWeight: 'bold' }}>
          🐛 Debug Console ({logs.length} logs)
        </span>
        <span>{isMinimized ? '▲' : '▼'}</span>
      </div>
      {!isMinimized && (
        <div style={{ padding: '0.5rem' }}>
          {logs.length === 0 ? (
            <div style={{ color: '#9ca3af', padding: '0.5rem' }}>
              No logs yet...
            </div>
          ) : (
            logs.map((log, i) => (
              <div
                key={i}
                style={{
                  padding: '0.25rem 0.5rem',
                  borderBottom: '1px solid #374151',
                  wordBreak: 'break-all',
                }}
              >
                <span style={{ color: '#9ca3af', marginRight: '0.5rem' }}>
                  [{log.timestamp}]
                </span>
                <span style={{
                  whiteSpace: 'pre-wrap',
                  color: log.message.includes('❌') ? '#fca5a5' :
                         log.message.includes('✅') ? '#86efac' :
                         log.message.includes('⚠️') ? '#fcd34d' :
                         '#f9fafb'
                }}>
                  {log.message}
                </span>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};

export default DebugPanel;
