import React, { useEffect, useRef } from 'react';
import './DetailDrawer.css';

export interface DetailDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  width?: string;
}

/**
 * Generic slide-out drawer component for displaying detailed information.
 * Features:
 * - Slide-in from right side
 * - Backdrop overlay (click to close)
 * - Close button
 * - ESC key to close
 * - Focus trap for accessibility
 *
 * @example
 * <DetailDrawer isOpen={isOpen} onClose={closeDrawer} title="Details">
 *   <div>Drawer content</div>
 * </DetailDrawer>
 */
export function DetailDrawer({ isOpen, onClose, title, children, width = '600px' }: DetailDrawerProps) {
  const drawerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    const trapFocus = (event: KeyboardEvent) => {
      if (!drawerRef.current) return;

      const focusableElements = drawerRef.current.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      const firstElement = focusableElements[0] as HTMLElement;
      const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;

      if (event.key === 'Tab') {
        if (event.shiftKey) {
          if (document.activeElement === firstElement) {
            event.preventDefault();
            lastElement?.focus();
          }
        } else {
          if (document.activeElement === lastElement) {
            event.preventDefault();
            firstElement?.focus();
          }
        }
      }
    };

    if (isOpen) {
      document.addEventListener('keyup', handleEscape);
      document.addEventListener('keydown', trapFocus);
    }

    return () => {
      document.removeEventListener('keyup', handleEscape);
      document.removeEventListener('keydown', trapFocus);
    };
  }, [isOpen, onClose]);

  useEffect(() => {
    if (isOpen && drawerRef.current) {
      const focusableElement = drawerRef.current.querySelector(
        'button, [href], input, select, textarea'
      ) as HTMLElement;
      focusableElement?.focus();
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <>
      <div
        className="drawer-backdrop"
        data-testid="drawer-backdrop"
        onClick={onClose}
        aria-hidden="true"
      />
      <div className="drawer-container" style={{ width }}>
        <div className="drawer-header">
          <h2 className="drawer-title">{title}</h2>
          <button
            className="drawer-close-button"
            onClick={onClose}
            aria-label="Close drawer"
            type="button"
          >
            ×
          </button>
        </div>
        <div className="drawer-content" ref={drawerRef}>
          {children}
        </div>
      </div>
    </>
  );
}