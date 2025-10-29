import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Spin, Alert, Card } from 'antd';
import { LoadingOutlined, WarningOutlined } from '@ant-design/icons';
import { useAuth } from '../context/AuthContext';

const OAuthCallback = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [error, setError] = useState('');

  const handleCallback = useCallback(() => {
    try {
      // Extract token from URL fragment
      const hash = window.location.hash.substring(1);
      const params = new URLSearchParams(hash);
      
      const idToken = params.get('id_token');
      const state = params.get('state');
      const error = params.get('error');
      const errorDescription = params.get('error_description');

      if (error) {
        setError(errorDescription || error);
        setTimeout(() => navigate('/login'), 3000);
        return;
      }

      if (!idToken) {
        setError('No ID token received');
        setTimeout(() => navigate('/login'), 3000);
        return;
      }

      // Verify state (should match what we sent)
      const savedState = sessionStorage.getItem('oauth_state');
      if (state !== savedState) {
        setError('Invalid state parameter');
        setTimeout(() => navigate('/login'), 3000);
        return;
      }

      // Decode JWT to extract user info and check role
      const payload = JSON.parse(atob(idToken.split('.')[1]));
      
      // Store the token and user info
      localStorage.setItem('user_info', JSON.stringify(payload));
      login(idToken);  // This will update auth context
      console.log('OAuthCallback: idToken set as admin_token', idToken);
      // Clean up
      sessionStorage.removeItem('oauth_state');
      sessionStorage.removeItem('oauth_nonce');

      // Navigate to dashboard using React Router
      navigate('/dashboard', { replace: true });
    } catch (err) {
      console.error('OAuth callback error:', err);
      setError('Failed to process authentication');
      setTimeout(() => navigate('/login'), 3000);
    }
  }, [navigate, login]);

  useEffect(() => {
    handleCallback();
  }, [handleCallback]);

  if (error) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100vh',
          backgroundColor: '#f0f2f5',
        }}
      >
        <Card style={{ width: 400, textAlign: 'center' }}>
          <WarningOutlined style={{ fontSize: 48, color: '#ff4d4f', marginBottom: 16 }} />
          <Alert
            message="Authentication Error"
            description={error}
            type="error"
            showIcon={false}
            style={{ marginBottom: 16 }}
          />
          <p style={{ color: '#8c8c8c' }}>Redirecting to login...</p>
        </Card>
      </div>
    );
  }

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        backgroundColor: '#f0f2f5',
      }}
    >
      <Card style={{ width: 400, textAlign: 'center', padding: '24px' }}>
        <LoadingOutlined style={{ fontSize: 48, color: '#1890ff', marginBottom: 16 }} />
        <h2 style={{ marginBottom: 8 }}>Authenticating...</h2>
        <p style={{ color: '#8c8c8c', marginBottom: 24 }}>Processing your login...</p>
        <Spin size="large" />
      </Card>
    </div>
  );
};

export default OAuthCallback;
