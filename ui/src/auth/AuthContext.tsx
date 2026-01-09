import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { User } from 'oidc-client-ts';
import { userManager } from './authConfig';

// Check if OIDC is configured
const isOIDCConfigured = () => {
  return !!((import.meta as any).env?.VITE_OIDC_ISSUER && (import.meta as any).env?.VITE_OIDC_CLIENT_ID);
};

// Mock user for demo mode
const createDemoUser = (): User => ({
  id_token: 'demo-token',
  session_state: null,
  access_token: 'demo-access-token',
  refresh_token: undefined,
  token_type: 'Bearer',
  scope: 'openid profile email',
  profile: {
    sub: 'demo-user',
    name: 'Demo User',
    email: 'demo@example.com',
    email_verified: true,
    iss: 'demo-issuer',
    aud: 'demo-client',
    exp: Math.floor(Date.now() / 1000) + 3600,
    iat: Math.floor(Date.now() / 1000),
  },
  expires_at: Date.now() / 1000 + 3600, // 1 hour from now
  expires_in: 3600,
  expired: false,
  scopes: ['openid', 'profile', 'email'],
  state: undefined,
  toStorageString: () => 'demo-user-data',
});

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isDemoMode: boolean;
  login: () => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isDemoMode] = useState(!isOIDCConfigured());

  useEffect(() => {
    if (isDemoMode) {
      // Demo mode - auto-login with mock user
      console.warn('🔓 DEMO MODE: Authentication is disabled. This is NOT secure for production use!');
      console.warn('🔓 To enable authentication, set VITE_OIDC_ISSUER and VITE_OIDC_CLIENT_ID environment variables.');
      setUser(createDemoUser());
      setIsLoading(false);
    } else {
      // OIDC mode - check if user is already logged in
      userManager.getUser().then((user) => {
        setUser(user);
        setIsLoading(false);
      }).catch(() => {
        setIsLoading(false);
      });

      // Listen for user loaded events
      const userLoaded = (user: User) => {
        setUser(user);
        setIsLoading(false);
      };

      const userUnloaded = () => {
        setUser(null);
        setIsLoading(false);
      };

      const silentRenewError = () => {
        // Silent renew failed, user needs to login again
        userManager.removeUser();
        setUser(null);
        setIsLoading(false);
      };

      userManager.events.addUserLoaded(userLoaded);
      userManager.events.addUserUnloaded(userUnloaded);
      userManager.events.addSilentRenewError(silentRenewError);

      return () => {
        userManager.events.removeUserLoaded(userLoaded);
        userManager.events.removeUserUnloaded(userUnloaded);
        userManager.events.removeSilentRenewError(silentRenewError);
      };
    }
  }, [isDemoMode]);

  const login = () => {
    if (isDemoMode) {
      // Already logged in in demo mode
      return;
    }
    userManager.signinRedirect();
  };

  const logout = () => {
    if (isDemoMode) {
      // Can't really logout in demo mode, but clear the user
      setUser(null);
      return;
    }
    userManager.signoutRedirect();
  };

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated: !!user,
    isDemoMode,
    login,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};