import { useEffect, useCallback } from 'react';
import { Spin } from 'antd';
import { LockOutlined } from '@ant-design/icons';

const SignIn = () => {
  const generateRandomString = (length: number): string => {
    const array = new Uint8Array(length);
    window.crypto.getRandomValues(array);
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
  };

  const initiateOAuthFlow = useCallback(() => {
    const state = generateRandomString(16);
    const nonce = generateRandomString(16);
    sessionStorage.setItem('oauth_state', state);
    sessionStorage.setItem('oauth_nonce', nonce);
    const authParams = new URLSearchParams({
      client_id: 'admin-ui',
      redirect_uri: `${window.location.origin}/admin/callback`,
      response_type: 'id_token',
      scope: 'openid profile email',
      state,
      nonce,
    });
    window.location.href = `/authorize?${authParams.toString()}`;
  }, []);

  useEffect(() => {
    initiateOAuthFlow();
  }, [initiateOAuthFlow]);

  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #0B1120 0%, #0F2027 50%, #162032 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div
        style={{
          width: 400,
          background: 'var(--surface, #fff)',
          borderRadius: 16,
          padding: 24,
          boxShadow: 'var(--shadow-lg, 0 20px 60px rgba(0,0,0,0.4))',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 12, marginBottom: 20 }}>
          <div
            style={{
              width: 48,
              height: 48,
              borderRadius: '50%',
              background: '#0D9488',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
            }}
          >
            <LockOutlined style={{ fontSize: 22, color: '#fff' }} />
          </div>
          <span style={{ fontSize: 18, fontWeight: 700, color: 'var(--text-primary, #111)' }}>OpenID Admin</span>
        </div>

        <div style={{ height: 1, background: 'var(--border, #e5e7eb)', marginBottom: 20 }} />

        <div style={{ textAlign: 'center' }}>
          <p style={{ fontSize: 16, color: 'var(--text-primary, #111)', marginBottom: 20 }}>
            Redirecting to authentication...
          </p>
          <Spin size="large" style={{ color: '#0D9488' }} />
          <p style={{ fontSize: 12, color: 'var(--text-muted, #6b7280)', marginTop: 24, marginBottom: 0 }}>
            Secure authentication powered by OpenID Connect
          </p>
        </div>
      </div>
    </div>
  );
};

export default SignIn;
