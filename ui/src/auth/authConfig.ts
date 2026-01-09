import { UserManager, UserManagerSettings } from 'oidc-client-ts';

// Auth configuration for OIDC
export const authConfig: UserManagerSettings = {
  authority: (import.meta as any).env?.VITE_OIDC_ISSUER || 'https://dev-12345678.okta.com',
  client_id: (import.meta as any).env?.VITE_OIDC_CLIENT_ID || 'your-client-id',
  redirect_uri: `${window.location.origin}/auth/callback`,
  post_logout_redirect_uri: `${window.location.origin}`,
  response_type: 'code',
  scope: 'openid profile email',
  automaticSilentRenew: true,
  loadUserInfo: true,
  revokeTokensOnSignout: true,
};

// Create and export the user manager instance
export const userManager = new UserManager(authConfig);

// Helper functions
export const signinRedirect = () => userManager.signinRedirect();
export const signoutRedirect = () => userManager.signoutRedirect();
export const signinSilent = () => userManager.signinSilent();
export const signinCallback = () => userManager.signinCallback();
export const getUser = () => userManager.getUser();
export const removeUser = () => userManager.removeUser();