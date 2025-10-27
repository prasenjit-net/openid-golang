import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import AdminLayout from './components/layout/AdminLayout';
import Dashboard from './pages/Dashboard';
import Users from './pages/Users';
import Clients from './pages/Clients';
import Settings from './pages/Settings';
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
                <Route path="users" element={<Users />} />
                <Route path="clients" element={<Clients />} />
                <Route path="settings" element={<Settings />} />
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
