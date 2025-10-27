import { useEffect, useCallback } from 'react';
import { Card, Spin } from 'antd';
import { LockOutlined } from '@ant-design/icons';

const Login = () => {
  const generateRandomString = (length: number): string => {
    const array = new Uint8Array(length);
    window.crypto.getRandomValues(array);
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
  };

  const initiateOAuthFlow = useCallback(() => {
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

  useEffect(() => {
    // Automatically redirect to OAuth authorization endpoint
    initiateOAuthFlow();
  }, [initiateOAuthFlow]);

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Card
        style={{
          width: 400,
          textAlign: 'center',
          borderRadius: 8,
          boxShadow: '0 10px 40px rgba(0,0,0,0.2)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: 80,
            height: 80,
            borderRadius: '50%',
            backgroundColor: '#1890ff',
            color: 'white',
            margin: '0 auto 24px',
          }}
        >
          <LockOutlined style={{ fontSize: 48 }} />
        </div>
        <h2 style={{ marginBottom: 8 }}>Admin Portal</h2>
        <p style={{ color: '#8c8c8c', marginBottom: 32 }}>
          OpenID Connect Server Administration
        </p>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 16 }}>
          <Spin size="large" />
          <p style={{ color: '#8c8c8c' }}>Redirecting to authentication...</p>
        </div>
        <div
          style={{
            marginTop: 32,
            paddingTop: 24,
            borderTop: '1px solid #f0f0f0',
          }}
        >
          <p style={{ fontSize: 12, color: '#8c8c8c' }}>
            Secure authentication powered by OpenID Connect
          </p>
        </div>
      </Card>
    </div>
  );
};

export default Login;
