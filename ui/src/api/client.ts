import { UserManager } from 'oidc-client-ts';
import type { UnifiedQuery } from '../types/query';

// API base URL
// Use /api prefix for nginx proxy in production
// Can be overridden with VITE_API_BASE_URL for development
const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL || '/api';

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
  const fullUrl = `${API_BASE_URL}${url}`;
  console.log('🌐 API Request:', {
    method: options.method || 'GET',
    url: fullUrl,
    baseUrl: API_BASE_URL,
    path: url,
    isDemoMode,
    timestamp: new Date().toISOString(),
  });

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

  console.log('📤 Request headers:', JSON.stringify(headers, null, 2));

  let response: Response;
  try {
    response = await fetch(fullUrl, {
      ...options,
      headers,
    });
    console.log('📥 API Response:', {
      url: fullUrl,
      status: response.status,
      statusText: response.statusText,
      ok: response.ok,
      headers: Object.fromEntries(response.headers.entries()),
    });
  } catch (error) {
    console.error('❌ API Request failed:', {
      url: fullUrl,
      error: error instanceof Error ? error.message : String(error),
      stack: error instanceof Error ? error.stack : undefined,
    });
    throw error;
  }

  // Log response body for debugging
  if (!response.ok) {
    response.clone().text().then(body => {
      console.error('❌ Error response body:', body);
    });
  } else {
    response.clone().json().then(body => {
      console.log('✅ Response data:', JSON.stringify(body, null, 2));
    }).catch(() => {
      console.log('✅ Response is not JSON');
    });
  }

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

    const response = await authenticatedFetch(`/v1/assets?${searchParams}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch assets: ${response.statusText}`);
    }
    return response.json();
  },

  async getAsset(id: number) {
    const response = await authenticatedFetch(`/v1/assets/${id}`);
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

    const response = await authenticatedFetch(`/v1/findings?${searchParams}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch findings: ${response.statusText}`);
    }
    return response.json();
  },

  // Dashboard API
  async getDashboard() {
    const response = await authenticatedFetch('/v1/dashboard');
    if (!response.ok) {
      throw new Error(`Failed to fetch dashboard: ${response.statusText}`);
    }
    return response.json();
  },

  // Intel API
  async getIntelStatus() {
    const response = await authenticatedFetch('/v1/intel/status');
    if (!response.ok) {
      throw new Error(`Failed to fetch intel status: ${response.statusText}`);
    }
    return response.json();
  },

  async refreshIntel() {
    const response = await authenticatedFetch('/v1/intel/refresh', { method: 'POST' });
    if (!response.ok) {
      throw new Error(`Failed to refresh intel: ${response.statusText}`);
    }
    return response.json();
  },

  // Query Framework API
  async queryExecute(entity: 'findings' | 'assets' | 'software_inventory', query: any) {
    const response = await authenticatedFetch(`/v1/query/${entity}`, {
      method: 'POST',
      body: JSON.stringify(query),
    });
    if (!response.ok) {
      throw new Error(`Failed to execute query: ${response.statusText}`);
    }
    return response.json();
  },

  async queryUnified(query: UnifiedQuery) {
    const response = await authenticatedFetch('/v1/query/unified', {
      method: 'POST',
      body: JSON.stringify(query),
    });
    if (!response.ok) {
      throw new Error(`Failed to execute unified query: ${response.statusText}`);
    }
    return response.json();
  },

  // Software API
  async getSoftware(params?: {
    vendor?: string;
    product?: string;
    version?: string;
    cpe?: string;
    limit?: number;
    offset?: number;
  }) {
    const searchParams = new URLSearchParams();
    if (params?.vendor) searchParams.set('vendor', params.vendor);
    if (params?.product) searchParams.set('product', params.product);
    if (params?.version) searchParams.set('version', params.version);
    if (params?.cpe) searchParams.set('cpe', params.cpe);
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const response = await authenticatedFetch(`/v1/software?${searchParams}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch software: ${response.statusText}`);
    }
    return response.json();
  },

  async getSoftwareById(id: number) {
    const response = await authenticatedFetch(`/v1/software/${id}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch software: ${response.statusText}`);
    }
    const json = await response.json();
    return json.data || json;
  },

  async getSoftwareForAsset(assetId: number) {
    const response = await authenticatedFetch(`/v1/assets/${assetId}/software`);
    if (!response.ok) {
      throw new Error(`Failed to fetch software for asset: ${response.statusText}`);
    }
    const json = await response.json();
    return json.data || json;
  },
};