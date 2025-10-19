import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './Setup.css';

const Setup = () => {
  const navigate = useNavigate();
  const [step, setStep] = useState(1);
  const [formData, setFormData] = useState({
    issuer: window.location.origin,
    adminUsername: '',
    adminPassword: '',
    adminEmail: '',
    adminName: '',
  });

  const handleNext = () => {
    if (step < 3) {
      setStep(step + 1);
    }
  };

  const handleBack = () => {
    if (step > 1) {
      setStep(step - 1);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await fetch('/api/admin/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });
      navigate('/admin/login');
    } catch (error) {
      console.error('Setup failed:', error);
      alert('Setup failed. Please try again.');
    }
  };

  return (
    <div className="setup">
      <div className="setup-container">
        <div className="setup-header">
          <h1>OpenID Connect Server Setup</h1>
          <div className="progress-indicator">
            <div className={`step ${step >= 1 ? 'active' : ''}`}>1</div>
            <div className="line"></div>
            <div className={`step ${step >= 2 ? 'active' : ''}`}>2</div>
            <div className="line"></div>
            <div className={`step ${step >= 3 ? 'active' : ''}`}>3</div>
          </div>
        </div>

        <form onSubmit={handleSubmit}>
          {step === 1 && (
            <div className="setup-step">
              <h2>Welcome</h2>
              <p>
                Thank you for choosing our OpenID Connect Server. This wizard will help you
                configure your identity provider in just a few steps.
              </p>
              <ul className="features-list">
                <li>✅ OpenID Connect 1.0 compliant</li>
                <li>✅ OAuth 2.0 authorization server</li>
                <li>✅ JWT token signing</li>
                <li>✅ User and client management</li>
              </ul>
              <div className="form-actions">
                <button type="button" onClick={handleNext} className="btn-primary">
                  Get Started
                </button>
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="setup-step">
              <h2>Server Configuration</h2>
              <div className="form-group">
                <label>Issuer URL</label>
                <input
                  type="url"
                  value={formData.issuer}
                  onChange={(e) => setFormData({ ...formData, issuer: e.target.value })}
                  placeholder="https://auth.example.com"
                  required
                />
                <small>
                  This is the base URL where your OpenID server will be accessible.
                  It should match your domain configuration.
                </small>
              </div>
              <div className="form-actions">
                <button type="button" onClick={handleBack} className="btn-secondary">
                  Back
                </button>
                <button type="button" onClick={handleNext} className="btn-primary">
                  Next
                </button>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="setup-step">
              <h2>Create Admin Account</h2>
              <div className="form-group">
                <label>Username</label>
                <input
                  type="text"
                  value={formData.adminUsername}
                  onChange={(e) => setFormData({ ...formData, adminUsername: e.target.value })}
                  placeholder="admin"
                  required
                />
              </div>
              <div className="form-group">
                <label>Email</label>
                <input
                  type="email"
                  value={formData.adminEmail}
                  onChange={(e) => setFormData({ ...formData, adminEmail: e.target.value })}
                  placeholder="admin@example.com"
                  required
                />
              </div>
              <div className="form-group">
                <label>Full Name</label>
                <input
                  type="text"
                  value={formData.adminName}
                  onChange={(e) => setFormData({ ...formData, adminName: e.target.value })}
                  placeholder="Admin User"
                  required
                />
              </div>
              <div className="form-group">
                <label>Password</label>
                <input
                  type="password"
                  value={formData.adminPassword}
                  onChange={(e) => setFormData({ ...formData, adminPassword: e.target.value })}
                  placeholder="Enter a strong password"
                  minLength={8}
                  required
                />
                <small>Password must be at least 8 characters long</small>
              </div>
              <div className="form-actions">
                <button type="button" onClick={handleBack} className="btn-secondary">
                  Back
                </button>
                <button type="submit" className="btn-primary">
                  Complete Setup
                </button>
              </div>
            </div>
          )}
        </form>
      </div>
    </div>
  );
};

export default Setup;
