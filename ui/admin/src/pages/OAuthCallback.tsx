import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import './Login.css';

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
      
      // Clean up
      sessionStorage.removeItem('oauth_state');
      sessionStorage.removeItem('oauth_nonce');

      // Force a full page reload to ensure auth state is picked up
      window.location.href = '/dashboard';
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
      <div className="login">
        <div className="login-container">
          <div className="login-header">
            <h1>⚠️ Authentication Error</h1>
          </div>
          <div className="error-message">{error}</div>
          <p>Redirecting to login...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="login">
      <div className="login-container">
        <div className="login-header">
          <h1>🔐 Authenticating...</h1>
        </div>
        <div className="spinner"></div>
        <p>Processing your login...</p>
      </div>
    </div>
  );
};

export default OAuthCallback;
