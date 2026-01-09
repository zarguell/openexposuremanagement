import { UserManager } from 'oidc-client-ts';

// API base URL
// Use relative URL by default so nginx proxy works correctly
// Can be overridden with VITE_API_BASE_URL for development
const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL || '';

// Create a user manager instance for API client
const userManager = new UserManager({
  authority: (import.meta as any).env?.VITE_OIDC_ISSUER || 'https://dev-12345678.okta.com',
  client_id: (import.meta as any).env?.VITE_OIDC_CLIENT_ID || 'your-client-id',
  redirect_uri: `${window.location.origin}/auth/callback`,
  response_type: 'code',
  scope: 'openid profile email',
});

// Check if we're in demo mode
const isDemoMode = !((import.meta as any).env?.VITE_OIDC_ISSUER && (import.meta as any).env?.VITE_OIDC_CLIENT_ID);

// Custom fetch function that includes auth headers
async function authenticatedFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (!isDemoMode) {
    const user = await userManager.getUser();
    if (user?.access_token) {
      headers['Authorization'] = `Bearer ${user.access_token}`;
    }
  }

  const response = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });

  // Handle 401/403 errors by triggering re-authentication (only in OIDC mode)
  if (response.status === 401 || response.status === 403) {
    if (isDemoMode) {
      // In demo mode, ignore auth errors and continue
      console.warn('🔓 Demo mode: Ignoring authentication error for demo purposes');
    } else {
      // Clear user session and redirect to login
      await userManager.removeUser();
      userManager.signinRedirect();
      throw new Error('Authentication required');
    }
  }

  return response;
}

// API client functions
export const apiClient = {
  // Assets API
  async getAssets(params?: { query?: string; limit?: number; offset?: number }) {
    const searchParams = new URLSearchParams();
    if (params?.query) searchParams.set('query', params.query);
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const response = await authenticatedFetch(`/assets?${searchParams}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch assets: ${response.statusText}`);
    }
    return response.json();
  },

  async getAsset(id: number) {
    const response = await authenticatedFetch(`/assets/${id}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch asset: ${response.statusText}`);
    }
    return response.json();
  },

  // Findings API
  async getFindings(params?: {
    source?: string;
    severity?: string;
    effective_status?: string;
    cve?: string;
    asset?: string;
    include_intel?: boolean;
    limit?: number;
    offset?: number;
  }) {
    const searchParams = new URLSearchParams();
    if (params?.source) searchParams.set('source', params.source);
    if (params?.severity) searchParams.set('severity', params.severity);
    if (params?.effective_status) searchParams.set('effective_status', params.effective_status);
    if (params?.cve) searchParams.set('cve', params.cve);
    if (params?.asset) searchParams.set('asset', params.asset);
    if (params?.include_intel) searchParams.set('include_intel', 'true');
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const response = await authenticatedFetch(`/findings?${searchParams}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch findings: ${response.statusText}`);
    }
    return response.json();
  },

  // Dashboard API
  async getDashboard() {
    const response = await authenticatedFetch('/dashboard');
    if (!response.ok) {
      throw new Error(`Failed to fetch dashboard: ${response.statusText}`);
    }
    return response.json();
  },

  // Intel API
  async getIntelStatus() {
    const response = await authenticatedFetch('/intel/status');
    if (!response.ok) {
      throw new Error(`Failed to fetch intel status: ${response.statusText}`);
    }
    return response.json();
  },

  async refreshIntel() {
    const response = await authenticatedFetch('/intel/refresh', { method: 'POST' });
    if (!response.ok) {
      throw new Error(`Failed to refresh intel: ${response.statusText}`);
    }
    return response.json();
  },
};