import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import AdminLayout from './components/layout/AdminLayout';
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
import SettingsDetail from './pages/settings/SettingsDetail';
import SettingsEdit from './pages/settings/SettingsEdit';
import Profile from './pages/Profile';
import Logout from './pages/Logout';
import Setup from './pages/Setup';
import SignIn from './pages/SignIn';
import OAuthCallback from './pages/OAuthCallback';
import { AuthProvider, useAuth } from './context/AuthContext';
import { QueryProvider } from './providers/QueryProvider';
import { useEffect } from 'react';

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

  if (loading) {
    return (
      <ConfigProvider theme={{ algorithm: theme.defaultAlgorithm }}>
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
    <ConfigProvider theme={{ algorithm: theme.defaultAlgorithm }}>
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
                <Route path="settings" element={<SettingsDetail />} />
                <Route path="settings/edit" element={<SettingsEdit />} />
                <Route path="profile" element={<Profile />} />
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
    <QueryProvider>
      <AuthProvider>
        <AppContent />
      </AuthProvider>
    </QueryProvider>
  );
}

export default App;
