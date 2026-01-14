import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DetailDrawer } from './DetailDrawer';
import { FindingDrawer } from './FindingDrawer';
import { AssetDrawer } from './AssetDrawer';

// Mock the API client
vi.mock('../api/client', () => ({
  apiClient: {
    getAsset: vi.fn(),
    getFindings: vi.fn(),
  },
}));

describe('DetailDrawer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };

  describe('DetailDrawer (base component)', () => {
    it('should not render when isOpen is false', () => {
      render(
        <DetailDrawer isOpen={false} onClose={vi.fn()} title="Test Drawer">
          <div>Drawer content</div>
        </DetailDrawer>,
        { wrapper }
      );

      expect(screen.queryByText('Test Drawer')).not.toBeInTheDocument();
      expect(screen.queryByText('Drawer content')).not.toBeInTheDocument();
    });

    it('should render when isOpen is true', () => {
      render(
        <DetailDrawer isOpen={true} onClose={vi.fn()} title="Test Drawer">
          <div>Drawer content</div>
        </DetailDrawer>,
        { wrapper }
      );

      expect(screen.getByText('Test Drawer')).toBeInTheDocument();
      expect(screen.getByText('Drawer content')).toBeInTheDocument();
    });

    it('should call onClose when close button is clicked', async () => {
      const onClose = vi.fn();

      render(
        <DetailDrawer isOpen={true} onClose={onClose} title="Test Drawer">
          <div>Drawer content</div>
        </DetailDrawer>,
        { wrapper }
      );

      const closeButton = screen.getByRole('button', { name: /close drawer/i });
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(onClose).toHaveBeenCalledTimes(1);
      });
    });

    it('should call onClose when backdrop is clicked', async () => {
      const onClose = vi.fn();

      render(
        <DetailDrawer isOpen={true} onClose={onClose} title="Test Drawer">
          <div>Drawer content</div>
        </DetailDrawer>,
        { wrapper }
      );

      const backdrop = document.querySelector('[data-testid="drawer-backdrop"]');
      expect(backdrop).toBeInTheDocument();

      if (backdrop) {
        fireEvent.click(backdrop);
        await waitFor(() => {
          expect(onClose).toHaveBeenCalledTimes(1);
        });
      }
});

    it('should not call onClose when clicking inside drawer content', async () => {
      const onClose = vi.fn();

      render(
        <DetailDrawer isOpen={true} onClose={onClose} title="Test Drawer">
          <div>Drawer content</div>
        </DetailDrawer>,
        { wrapper }
      );

      const content = screen.getByText('Drawer content');
      fireEvent.click(content);

      await waitFor(() => {
        expect(onClose).not.toHaveBeenCalled();
      });
    });
  });

  describe('FindingDrawer', () => {
    it('should render finding details', () => {
      const mockFinding = {
        id: 1,
        title: 'CVE-2023-1234',
        severity: 'critical',
        effective_status: 'active',
        cve_id: 'CVE-2023-1234',
        description: 'Test vulnerability description',
        epss_score: 0.95,
        is_kev: true,
        last_observed_at: '2024-01-15T10:00:00Z',
      };

      render(<FindingDrawer isOpen={true} onClose={vi.fn()} finding={mockFinding} />, { wrapper });

      expect(screen.getByText('CVE-2023-1234')).toBeInTheDocument();
      expect(screen.getByText(/Test vulnerability description/)).toBeInTheDocument();
      expect(screen.getByText(/0.950/)).toBeInTheDocument();
      expect(screen.getByText(/High Exploitability/)).toBeInTheDocument();
      expect(screen.getByText(/CISA KEV/)).toBeInTheDocument();
    });

    it('should render loading state when no finding data provided', () => {
      render(<FindingDrawer isOpen={true} onClose={vi.fn()} findingId={1} finding={undefined} />, {
        wrapper,
      });

      expect(screen.getByText(/loading/i)).toBeInTheDocument();
    });
  });

  describe('AssetDrawer', () => {
    it('should render asset details', () => {
      const mockAsset = {
        id: 1,
        hostname: 'server01.example.com',
        ip_address: '192.168.1.10',
        os: 'Ubuntu 22.04',
        last_observed_at: '2024-01-15T10:00:00Z',
        findings_count: 5,
      };

      render(<AssetDrawer isOpen={true} onClose={vi.fn()} asset={mockAsset} />, { wrapper });

      expect(screen.getByText('server01.example.com')).toBeInTheDocument();
      expect(screen.getByText('192.168.1.10')).toBeInTheDocument();
      expect(screen.getByText('Ubuntu 22.04')).toBeInTheDocument();
      expect(screen.getByText(/5 findings/)).toBeInTheDocument();
    });

    it('should render loading state when no asset data provided', () => {
      render(<AssetDrawer isOpen={true} onClose={vi.fn()} assetId={1} asset={undefined} />, {
        wrapper,
      });

      expect(screen.getByText(/loading/i)).toBeInTheDocument();
    });
  });
});
