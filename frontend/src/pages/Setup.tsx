import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSetup } from '../hooks/useApi';
import { LockOutlined, SecurityScanOutlined, KeyOutlined, TeamOutlined, SafetyOutlined, CheckOutlined } from '@ant-design/icons';

const inputStyle: React.CSSProperties = {
  width: '100%',
  border: '1px solid var(--border, #e2e8f0)',
  borderRadius: 8,
  padding: '10px 12px',
  fontSize: 14,
  color: 'var(--text-primary, #111)',
  background: 'var(--surface, #fff)',
  outline: 'none',
  boxSizing: 'border-box',
};

const labelStyle: React.CSSProperties = {
  display: 'block',
  fontSize: 13,
  fontWeight: 500,
  color: 'var(--text-secondary, #374151)',
  marginBottom: 6,
};

const btnPrimary: React.CSSProperties = {
  background: '#0D9488',
  color: '#fff',
  border: 'none',
  borderRadius: 8,
  padding: '10px 24px',
  fontSize: 14,
  fontWeight: 600,
  cursor: 'pointer',
};

const btnSecondary: React.CSSProperties = {
  background: 'var(--border, #e2e8f0)',
  color: 'var(--text-primary, #111)',
  border: 'none',
  borderRadius: 8,
  padding: '10px 24px',
  fontSize: 14,
  fontWeight: 600,
  cursor: 'pointer',
};

const features = [
  { icon: <SecurityScanOutlined />, text: 'OpenID Connect 1.0 compliant' },
  { icon: <KeyOutlined />, text: 'OAuth 2.0 authorization server' },
  { icon: <TeamOutlined />, text: 'User and client management' },
  { icon: <SafetyOutlined />, text: 'JWT token signing with RSA keys' },
];

const Setup = () => {
  const navigate = useNavigate();
  const setupMutation = useSetup();
  const [step, setStep] = useState(1);
  const [formData, setFormData] = useState({
    issuer: window.location.origin,
    adminUsername: '',
    adminPassword: '',
    adminEmail: '',
    adminName: '',
  });

  const handleNext = () => {
    if (step < 3) setStep(step + 1);
  };

  const handleBack = () => {
    if (step > 1) setStep(step - 1);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await setupMutation.mutateAsync({
        username: formData.adminUsername,
        email: formData.adminEmail,
        password: formData.adminPassword,
      });
      navigate('/admin/login');
    } catch (error) {
      console.error('Setup failed:', error);
      alert('Setup failed. Please try again.');
    }
  };

  const StepIndicator = () => (
    <div style={{ display: 'flex', alignItems: 'center', marginBottom: 36 }}>
      {[1, 2, 3].map((s, i) => (
        <div key={s} style={{ display: 'flex', alignItems: 'center', flex: i < 2 ? 1 : undefined }}>
          <div
            style={{
              width: 32,
              height: 32,
              borderRadius: '50%',
              background: step > s ? '#0D9488' : step === s ? '#0D9488' : '#E2E8F0',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
              transition: 'background 0.2s',
            }}
          >
            {step > s ? (
              <CheckOutlined style={{ fontSize: 14, color: '#fff' }} />
            ) : (
              <span style={{ fontSize: 13, fontWeight: 700, color: step === s ? '#fff' : '#94A3B8' }}>{s}</span>
            )}
          </div>
          {i < 2 && (
            <div
              style={{
                flex: 1,
                height: 2,
                background: step > s ? '#0D9488' : '#E2E8F0',
                margin: '0 8px',
                transition: 'background 0.2s',
              }}
            />
          )}
        </div>
      ))}
    </div>
  );

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      {/* Left panel */}
      <div
        style={{
          width: 380,
          flexShrink: 0,
          background: '#0B1120',
          padding: 48,
          display: 'flex',
          flexDirection: 'column',
        }}
        className="setup-left-panel"
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 40 }}>
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: '50%',
              background: '#0D9488',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <LockOutlined style={{ fontSize: 20, color: '#fff' }} />
          </div>
          <span style={{ fontSize: 16, fontWeight: 700, color: '#fff' }}>OpenID Connect</span>
        </div>

        <div style={{ fontSize: 20, fontWeight: 600, color: '#CBD5E1', marginBottom: 12 }}>
          Secure Identity Management
        </div>
        <p style={{ fontSize: 14, color: '#94A3B8', lineHeight: 1.7, marginBottom: 40 }}>
          Set up your own OpenID Connect identity provider in minutes. Manage users, clients, and tokens with ease.
        </p>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {features.map((f, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <span style={{ color: '#0D9488', fontSize: 16 }}>{f.icon}</span>
              <span style={{ fontSize: 14, color: '#CBD5E1' }}>{f.text}</span>
            </div>
          ))}
        </div>

        <div style={{ marginTop: 'auto', fontSize: 12, color: '#475569' }}>
          © {new Date().getFullYear()} OpenID Connect Server
        </div>
      </div>

      {/* Right panel */}
      <div
        style={{
          flex: 1,
          background: 'var(--bg, #f8fafc)',
          overflowY: 'auto',
          display: 'flex',
          alignItems: 'flex-start',
          justifyContent: 'center',
          padding: '48px 24px',
        }}
      >
        <div style={{ width: '100%', maxWidth: 480 }}>
          <StepIndicator />

          <form onSubmit={handleSubmit}>
            {step === 1 && (
              <div>
                <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary, #111)', marginBottom: 8 }}>
                  Welcome
                </div>
                <p style={{ fontSize: 14, color: 'var(--text-secondary, #374151)', marginBottom: 28, lineHeight: 1.7 }}>
                  Thank you for choosing our OpenID Connect Server. This wizard will help you configure your identity provider in just a few steps.
                </p>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 36 }}>
                  {features.map((f, i) => (
                    <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <span style={{ color: '#0D9488', fontSize: 16 }}>{f.icon}</span>
                      <span style={{ fontSize: 14, color: 'var(--text-secondary, #374151)' }}>{f.text}</span>
                    </div>
                  ))}
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                  <button type="button" onClick={handleNext} style={btnPrimary}>
                    Get Started
                  </button>
                </div>
              </div>
            )}

            {step === 2 && (
              <div>
                <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary, #111)', marginBottom: 8 }}>
                  Server Configuration
                </div>
                <p style={{ fontSize: 14, color: 'var(--text-secondary, #374151)', marginBottom: 28 }}>
                  Configure the base settings for your OpenID server.
                </p>
                <div style={{ marginBottom: 20 }}>
                  <label style={labelStyle}>Issuer URL</label>
                  <input
                    type="url"
                    value={formData.issuer}
                    onChange={(e) => setFormData({ ...formData, issuer: e.target.value })}
                    placeholder="https://auth.example.com"
                    required
                    style={inputStyle}
                  />
                  <small style={{ fontSize: 12, color: 'var(--text-muted, #6b7280)', marginTop: 4, display: 'block' }}>
                    The base URL where your OpenID server will be accessible. It should match your domain configuration.
                  </small>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 32 }}>
                  <button type="button" onClick={handleBack} style={btnSecondary}>Back</button>
                  <button type="button" onClick={handleNext} style={btnPrimary}>Next</button>
                </div>
              </div>
            )}

            {step === 3 && (
              <div>
                <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary, #111)', marginBottom: 8 }}>
                  Create Admin Account
                </div>
                <p style={{ fontSize: 14, color: 'var(--text-secondary, #374151)', marginBottom: 28 }}>
                  Set up the initial administrator account for managing your server.
                </p>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 16, marginBottom: 8 }}>
                  <div>
                    <label style={labelStyle}>Username</label>
                    <input
                      type="text"
                      value={formData.adminUsername}
                      onChange={(e) => setFormData({ ...formData, adminUsername: e.target.value })}
                      placeholder="admin"
                      required
                      style={inputStyle}
                    />
                  </div>
                  <div>
                    <label style={labelStyle}>Email</label>
                    <input
                      type="email"
                      value={formData.adminEmail}
                      onChange={(e) => setFormData({ ...formData, adminEmail: e.target.value })}
                      placeholder="admin@example.com"
                      required
                      style={inputStyle}
                    />
                  </div>
                  <div>
                    <label style={labelStyle}>Full Name</label>
                    <input
                      type="text"
                      value={formData.adminName}
                      onChange={(e) => setFormData({ ...formData, adminName: e.target.value })}
                      placeholder="Admin User"
                      required
                      style={inputStyle}
                    />
                  </div>
                  <div>
                    <label style={labelStyle}>Password</label>
                    <input
                      type="password"
                      value={formData.adminPassword}
                      onChange={(e) => setFormData({ ...formData, adminPassword: e.target.value })}
                      placeholder="Enter a strong password"
                      minLength={8}
                      required
                      style={inputStyle}
                    />
                    <small style={{ fontSize: 12, color: 'var(--text-muted, #6b7280)', marginTop: 4, display: 'block' }}>
                      Password must be at least 8 characters long
                    </small>
                  </div>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 32 }}>
                  <button type="button" onClick={handleBack} style={btnSecondary}>Back</button>
                  <button
                    type="submit"
                    style={{ ...btnPrimary, opacity: setupMutation.isPending ? 0.7 : 1 }}
                    disabled={setupMutation.isPending}
                  >
                    {setupMutation.isPending ? 'Setting up...' : 'Complete Setup'}
                  </button>
                </div>
              </div>
            )}
          </form>
        </div>
      </div>
    </div>
  );
};

export default Setup;
