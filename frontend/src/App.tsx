import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, theme as antTheme } from 'antd';
import AdminLayout from './components/layout/AdminLayout';
import { AuthProvider, useAuth } from './context/AuthContext';
import { QueryProvider } from './providers/QueryProvider';
import { ThemeProvider, useTheme } from './context/ThemeContext';

import Dashboard from './pages/Dashboard';
import UserSearch from './pages/users/UserSearch';
import UserDetail from './pages/users/UserDetail';
import UserEdit from './pages/users/UserEdit';
import UserCreate from './pages/users/UserCreate';
import ClientSearch from './pages/clients/ClientSearch';
import ClientDetail from './pages/clients/ClientDetail';
import ClientEdit from './pages/clients/ClientEdit';
import ClientCreate from './pages/clients/ClientCreate';
import KeyManagement from './pages/KeyManagement';
import Tokens from './pages/Tokens';
import SettingsDetail from './pages/settings/SettingsDetail';
import SettingsEdit from './pages/settings/SettingsEdit';
import Profile from './pages/Profile';
import Logout from './pages/Logout';
import Setup from './pages/Setup';
import SignIn from './pages/SignIn';
import OAuthCallback from './pages/OAuthCallback';
import AuditLog from './pages/AuditLog';

// Component to initiate OAuth flow for unauthenticated users
function OAuthRedirect() {
  useEffect(() => {
    const generateRandomString = (length: number): string => {
      const array = new Uint8Array(length);
      window.crypto.getRandomValues(array);
      return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
    };

    // Generate state and nonce for security
    const state = generateRandomString(16);
    const nonce = generateRandomString(16);

    // Save state and nonce in session storage for verification
    sessionStorage.setItem('oauth_state', state);
    sessionStorage.setItem('oauth_nonce', nonce);

    // Build authorization URL
    const authParams = new URLSearchParams({
      client_id: 'admin-ui',
      redirect_uri: `${window.location.origin}/admin/callback`,
      response_type: 'id_token',
      scope: 'openid profile email',
      state: state,
      nonce: nonce,
    });

    // Redirect to authorization endpoint
    window.location.href = `/authorize?${authParams.toString()}`;
  }, []);

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
      }}
    >
      <p>Redirecting to login...</p>
    </div>
  );
}

function AppContent() {
  const { isAuthenticated, isSetupComplete, loading } = useAuth();
  const { isDark } = useTheme();

  const configTheme = {
    algorithm: isDark ? antTheme.darkAlgorithm : antTheme.defaultAlgorithm,
    token: {
      colorPrimary: '#0D9488',
      colorSuccess: '#10B981',
      colorWarning: '#F59E0B',
      colorError: '#EF4444',
      colorInfo: '#0EA5E9',
      borderRadius: 8,
      fontFamily: "'Inter', system-ui, -apple-system, sans-serif",
      fontSize: 14,
      colorBgBase: isDark ? '#1E293B' : '#FFFFFF',
      colorBgContainer: isDark ? '#1E293B' : '#FFFFFF',
      colorBgLayout: isDark ? '#0B1120' : '#F1F5F9',
      colorBgElevated: isDark ? '#253347' : '#FFFFFF',
      colorBorder: isDark ? '#334155' : '#E2E8F0',
      colorBorderSecondary: isDark ? '#1E293B' : '#F1F5F9',
      colorText: isDark ? '#F1F5F9' : '#0F172A',
      colorTextSecondary: isDark ? '#94A3B8' : '#475569',
      colorTextTertiary: isDark ? '#64748B' : '#94A3B8',
      colorTextQuaternary: isDark ? '#475569' : '#CBD5E1',
      boxShadow: '0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04)',
      boxShadowSecondary: '0 4px 6px -1px rgba(0,0,0,0.08), 0 2px 4px -2px rgba(0,0,0,0.04)',
    },
  };

  if (loading) {
    return (
      <ConfigProvider theme={configTheme}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: '100vh',
          }}
        >
          <div style={{ textAlign: 'center' }}>
            <div className="spinner"></div>
            <p>Loading...</p>
          </div>
        </div>
      </ConfigProvider>
    );
  }

  return (
    <ConfigProvider theme={configTheme}>
      <BrowserRouter>
        <Routes>
            {/* OAuth callback route - always accessible */}
            <Route path="/admin/callback" element={<OAuthCallback />} />
            {/* Logout route - always accessible */}
            <Route path="/logout" element={<Logout />} />
            
            {!isSetupComplete ? (
              <>
                <Route path="/setup" element={<Setup />} />
                <Route path="*" element={<Navigate to="/setup" replace />} />
              </>
            ) : !isAuthenticated ? (
              <>
                {/* Optional signin page - mainly redirects to OAuth */}
                <Route path="/signin" element={<SignIn />} />
                {/* Redirect unauthenticated users to OAuth authorize endpoint */}
                <Route path="*" element={<OAuthRedirect />} />
              </>
            ) : (
              <>
                <Route path="/" element={<AdminLayout />}>
                  <Route index element={<Navigate to="/dashboard" replace />} />
                  <Route path="dashboard" element={<Dashboard />} />
                  <Route path="users" element={<UserSearch />} />
                <Route path="users/new" element={<UserCreate />} />
                <Route path="users/:id" element={<UserDetail />} />
                <Route path="users/:id/edit" element={<UserEdit />} />
                <Route path="clients" element={<ClientSearch />} />
                <Route path="clients/new" element={<ClientCreate />} />
                <Route path="clients/:id" element={<ClientDetail />} />
                <Route path="clients/:id/edit" element={<ClientEdit />} />
                <Route path="keys" element={<KeyManagement />} />
                <Route path="tokens" element={<Tokens />} />
                <Route path="settings" element={<SettingsDetail />} />
                <Route path="settings/edit" element={<SettingsEdit />} />
                <Route path="profile" element={<Profile />} />
                <Route path="audit" element={<AuditLog />} />
              </Route>
              <Route path="*" element={<Navigate to="/dashboard" replace />} />
            </>
          )}
          </Routes>
      </BrowserRouter>
    </ConfigProvider>
  );
}

function App() {
  return (
    <ThemeProvider>
      <QueryProvider>
        <AuthProvider>
          <AppContent />
        </AuthProvider>
      </QueryProvider>
    </ThemeProvider>
  );
}

export default App;
