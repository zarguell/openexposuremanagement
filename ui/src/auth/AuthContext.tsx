import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { User } from 'oidc-client-ts';
import { userManager } from './authConfig';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
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

  useEffect(() => {
    // Check if user is already logged in
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
  }, []);

  const login = () => {
    userManager.signinRedirect();
  };

  const logout = () => {
    userManager.signoutRedirect();
  };

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated: !!user,
    login,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};