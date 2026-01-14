import { render, screen, act, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { ToastProvider, useToast } from '../contexts/ToastContext';

// Test component that triggers toasts
const TestTrigger = ({ 
  message = "Test Message", 
  type = "info", 
  duration 
}: { 
  message?: string, 
  type?: 'success' | 'error' | 'warning' | 'info',
  duration?: number 
}) => {
  const { addToast, removeToast, clearAll } = useToast();

  return (
    <div>
      <button onClick={() => addToast(message, type, duration)}>Show Toast</button>
      <button onClick={() => clearAll()}>Clear All</button>
      <button onClick={() => removeToast('test-id')}>Remove Specific</button>
    </div>
  );
};

describe('Toast Notification System', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows a toast when addToast is called', async () => {
    render(
      <ToastProvider>
        <TestTrigger message="Hello World" type="success" />
      </ToastProvider>
    );

    const button = screen.getByText('Show Toast');
    fireEvent.click(button);

    const toast = await screen.findByText('Hello World');
    expect(toast).toBeInTheDocument();
  });

  it('displays correct styles for different toast types', async () => {
    render(
      <ToastProvider>
        <TestTrigger message="Error occurred" type="error" />
      </ToastProvider>
    );

    fireEvent.click(screen.getByText('Show Toast'));
    
    // We expect some visual indicator of error type, e.g., class or aria-label
    // Implementation detail: we'll look for the text for now, 
    // and ideally check for a class or style later once implementation is set.
    const toast = await screen.findByText('Error occurred');
    expect(toast).toBeInTheDocument();
  });

  it('auto-dismisses after duration', async () => {
    vi.useFakeTimers();
    render(
      <ToastProvider>
        <TestTrigger message="Auto dismiss" duration={2000} />
      </ToastProvider>
    );

    fireEvent.click(screen.getByText('Show Toast'));
    expect(screen.getByText('Auto dismiss')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(2500);
    });

    expect(screen.queryByText('Auto dismiss')).not.toBeInTheDocument();
  });

  it('can be manually dismissed', async () => {
    render(
      <ToastProvider>
        <TestTrigger message="Dismiss me" />
      </ToastProvider>
    );

    fireEvent.click(screen.getByText('Show Toast'));
    await screen.findByText('Dismiss me');
    
    // Find close button - assuming aria-label="Close" or similar
    const closeBtn = screen.getByLabelText(/close/i);
    fireEvent.click(closeBtn);

    await waitFor(() => {
      expect(screen.queryByText('Dismiss me')).not.toBeInTheDocument();
    });
  });

  it('stacks multiple toasts', async () => {
    const MultiTrigger = () => {
      const { addToast } = useToast();
      return (
        <button onClick={() => {
          addToast('Toast 1');
          addToast('Toast 2');
          addToast('Toast 3');
        }}>Show Multiple</button>
      );
    };

    render(
      <ToastProvider>
        <MultiTrigger />
      </ToastProvider>
    );

    fireEvent.click(screen.getByText('Show Multiple'));

    expect(await screen.findByText('Toast 1')).toBeInTheDocument();
    expect(screen.getByText('Toast 2')).toBeInTheDocument();
    expect(screen.getByText('Toast 3')).toBeInTheDocument();
  });

  it('limits the number of visible toasts', async () => {
     const ManyTrigger = () => {
      const { addToast } = useToast();
      return (
        <button onClick={() => {
          for(let i=0; i<7; i++) addToast(`Toast ${i}`);
        }}>Show Many</button>
      );
    };

    render(
      <ToastProvider>
        <ManyTrigger />
      </ToastProvider>
    );

    fireEvent.click(screen.getByText('Show Many'));
    
    // Assuming max 5
    // Wait for effects
    await waitFor(() => {
       const toasts = screen.getAllByText(/Toast \d/);
       expect(toasts.length).toBeLessThanOrEqual(5);
    });
  });

  it('clears all toasts', async () => {
    render(
      <ToastProvider>
        <TestTrigger message="To be cleared" />
      </ToastProvider>
    );

    fireEvent.click(screen.getByText('Show Toast'));
    expect(await screen.findByText('To be cleared')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Clear All'));
    
    await waitFor(() => {
      expect(screen.queryByText('To be cleared')).not.toBeInTheDocument();
    });
  });
});
