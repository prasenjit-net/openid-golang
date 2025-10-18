import { useEffect, useState } from 'react';
import './Settings.css';

interface Settings {
  issuer: string;
  token_ttl: number;
  refresh_token_ttl: number;
  jwks_rotation_days: number;
}

const Settings = () => {
  const [settings, setSettings] = useState<Settings>({
    issuer: '',
    token_ttl: 3600,
    refresh_token_ttl: 2592000,
    jwks_rotation_days: 90,
  });
  const [keys, setKeys] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchSettings();
    fetchKeys();
  }, []);

  const fetchSettings = async () => {
    try {
      const response = await fetch('/api/admin/settings');
      const data = await response.json();
      setSettings(data);
    } catch (error) {
      console.error('Failed to fetch settings:', error);
    }
  };

  const fetchKeys = async () => {
    try {
      const response = await fetch('/api/admin/keys');
      const data = await response.json();
      setKeys(data);
    } catch (error) {
      console.error('Failed to fetch keys:', error);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await fetch('/api/admin/settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(settings),
      });
      alert('Settings saved successfully');
    } catch (error) {
      console.error('Failed to save settings:', error);
      alert('Failed to save settings');
    } finally {
      setLoading(false);
    }
  };

  const handleRotateKeys = async () => {
    if (!confirm('Are you sure you want to rotate signing keys? This will invalidate all existing tokens.')) {
      return;
    }
    setLoading(true);
    try {
      await fetch('/api/admin/keys/rotate', { method: 'POST' });
      alert('Keys rotated successfully');
      fetchKeys();
    } catch (error) {
      console.error('Failed to rotate keys:', error);
      alert('Failed to rotate keys');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="settings">
      <h1>Settings</h1>

      <div className="settings-section">
        <h2>Server Configuration</h2>
        <form onSubmit={handleSave}>
          <div className="form-group">
            <label>Issuer URL</label>
            <input
              type="url"
              value={settings.issuer}
              onChange={(e) => setSettings({ ...settings, issuer: e.target.value })}
              placeholder="https://auth.example.com"
              required
            />
            <small>The base URL of this OpenID Provider</small>
          </div>

          <div className="form-group">
            <label>Access Token TTL (seconds)</label>
            <input
              type="number"
              value={settings.token_ttl}
              onChange={(e) => setSettings({ ...settings, token_ttl: parseInt(e.target.value) })}
              min="60"
              max="86400"
              required
            />
            <small>How long access tokens are valid (default: 3600 = 1 hour)</small>
          </div>

          <div className="form-group">
            <label>Refresh Token TTL (seconds)</label>
            <input
              type="number"
              value={settings.refresh_token_ttl}
              onChange={(e) => setSettings({ ...settings, refresh_token_ttl: parseInt(e.target.value) })}
              min="3600"
              max="31536000"
              required
            />
            <small>How long refresh tokens are valid (default: 2592000 = 30 days)</small>
          </div>

          <div className="form-group">
            <label>JWKS Rotation Period (days)</label>
            <input
              type="number"
              value={settings.jwks_rotation_days}
              onChange={(e) => setSettings({ ...settings, jwks_rotation_days: parseInt(e.target.value) })}
              min="30"
              max="365"
              required
            />
            <small>How often to automatically rotate signing keys (default: 90 days)</small>
          </div>

          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? 'Saving...' : 'Save Settings'}
          </button>
        </form>
      </div>

      <div className="settings-section">
        <div className="section-header">
          <h2>Signing Keys</h2>
          <button className="btn-danger" onClick={handleRotateKeys} disabled={loading}>
            Rotate Keys
          </button>
        </div>
        <div className="keys-list">
          {keys.map((key) => (
            <div key={key.kid} className="key-card">
              <div className="key-info">
                <div className="key-detail">
                  <strong>Key ID:</strong> <code>{key.kid}</code>
                </div>
                <div className="key-detail">
                  <strong>Algorithm:</strong> {key.alg}
                </div>
                <div className="key-detail">
                  <strong>Use:</strong> {key.use}
                </div>
                <div className="key-detail">
                  <strong>Created:</strong> {new Date(key.created_at).toLocaleString()}
                </div>
              </div>
              {key.is_active && <span className="badge-active">Active</span>}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Settings;
