import { useEffect } from 'react';
import './Login.css';

const Login = () => {
  useEffect(() => {
    // Automatically redirect to OAuth authorization endpoint
    initiateOAuthFlow();
  }, []);

  const generateRandomString = (length: number): string => {
    const array = new Uint8Array(length);
    window.crypto.getRandomValues(array);
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
  };

  const initiateOAuthFlow = () => {
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
  };

  return (
    <div className="login">
      <div className="login-container">
        <div className="login-header">
          <h1>🔐 Admin Portal</h1>
          <p>OpenID Connect Server Administration</p>
        </div>

        <div className="login-loading">
          <div className="spinner"></div>
          <p>Redirecting to authentication...</p>
        </div>

        <div className="login-footer">
          <p>Secure authentication powered by OpenID Connect</p>
        </div>
      </div>
    </div>
  );
};

export default Login;
