import { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';

interface AuthContextType {
  isAuthenticated: boolean;
  isSetupComplete: boolean;
  loading: boolean;
  login: (token: string) => void;
  logout: () => void;
  checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isSetupComplete, setIsSetupComplete] = useState(true);
  const [loading, setLoading] = useState(true);

  const checkAuth = async () => {
    try {
      // First check if we have a token in localStorage
      const token = localStorage.getItem('admin_token');
      if (token) {
        // Validate token is not expired
        try {
          const payload = JSON.parse(atob(token.split('.')[1]));
          const exp = payload.exp * 1000; // Convert to milliseconds
          if (exp > Date.now()) {
            setIsAuthenticated(true);
            setIsSetupComplete(true);
            setLoading(false);
            return;
          } else {
            // Token expired, remove it
            localStorage.removeItem('admin_token');
            localStorage.removeItem('user_info');
          }
        } catch (e) {
          // Invalid token, remove it
          localStorage.removeItem('admin_token');
          localStorage.removeItem('user_info');
        }
      }

      // No valid token, check setup status
      const response = await fetch('/api/admin/setup/status');
      const data = await response.json();
      setIsSetupComplete(data.setupComplete);
      setIsAuthenticated(false);
    } catch (error) {
      console.error('Failed to check authentication:', error);
      setIsAuthenticated(false);
    } finally {
      setLoading(false);
    }
  };

  const login = (token: string) => {
    localStorage.setItem('admin_token', token);
    setIsAuthenticated(true);
  };

  const logout = () => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('user_info');
    sessionStorage.clear();
    setIsAuthenticated(false);
  };

  useEffect(() => {
    checkAuth();
  }, []);

  return (
    <AuthContext.Provider value={{ isAuthenticated, isSetupComplete, loading, login, logout, checkAuth }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
