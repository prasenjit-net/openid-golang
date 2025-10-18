import { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Paper,
  TextField,
  Typography,
  Alert,
  CircularProgress,
  Divider,
  Card,
  CardContent,
  Chip,
} from '@mui/material';
import {
  Save as SaveIcon,
  Refresh as RefreshIcon,
  VpnKey as KeyIcon,
} from '@mui/icons-material';

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
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [rotating, setRotating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    fetchSettings();
    fetchKeys();
  }, []);

  const fetchSettings = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/admin/settings');
      if (!response.ok) throw new Error('Failed to fetch settings');
      const data = await response.json();
      setSettings(data);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to fetch settings');
      console.error('Failed to fetch settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchKeys = async () => {
    try {
      const response = await fetch('/api/admin/keys');
      if (!response.ok) throw new Error('Failed to fetch keys');
      const data = await response.json();
      setKeys(data);
    } catch (error) {
      console.error('Failed to fetch keys:', error);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setSuccess(null);
    try {
      const response = await fetch('/api/admin/settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(settings),
      });
      if (!response.ok) throw new Error('Failed to save settings');
      setSuccess('Settings saved successfully');
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to save settings');
      console.error('Failed to save settings:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleRotateKeys = async () => {
    if (!window.confirm('Are you sure you want to rotate signing keys? This will invalidate all existing tokens.')) {
      return;
    }
    setRotating(true);
    setError(null);
    setSuccess(null);
    try {
      const response = await fetch('/api/admin/keys/rotate', { method: 'POST' });
      if (!response.ok) throw new Error('Failed to rotate keys');
      setSuccess('Keys rotated successfully');
      fetchKeys();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to rotate keys');
      console.error('Failed to rotate keys:', error);
    } finally {
      setRotating(false);
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="60vh">
        <CircularProgress size={60} />
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom fontWeight="bold" sx={{ mb: 3 }}>
        Settings
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess(null)}>
          {success}
        </Alert>
      )}

      <Paper elevation={2} sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom fontWeight="bold">
          Server Configuration
        </Typography>
        <Divider sx={{ mb: 3 }} />
        <form onSubmit={handleSave}>
          <Box display="flex" flexDirection="column" gap={3}>
            <TextField
              label="Issuer URL"
              type="url"
              value={settings.issuer}
              onChange={(e) => setSettings({ ...settings, issuer: e.target.value })}
              placeholder="https://auth.example.com"
              required
              fullWidth
              helperText="The base URL of this OpenID Provider"
            />

            <TextField
              label="Access Token TTL (seconds)"
              type="number"
              value={settings.token_ttl}
              onChange={(e) => setSettings({ ...settings, token_ttl: parseInt(e.target.value) })}
              inputProps={{ min: 60, max: 86400 }}
              required
              fullWidth
              helperText="How long access tokens are valid (default: 3600 = 1 hour)"
            />

            <TextField
              label="Refresh Token TTL (seconds)"
              type="number"
              value={settings.refresh_token_ttl}
              onChange={(e) => setSettings({ ...settings, refresh_token_ttl: parseInt(e.target.value) })}
              inputProps={{ min: 3600, max: 31536000 }}
              required
              fullWidth
              helperText="How long refresh tokens are valid (default: 2592000 = 30 days)"
            />

            <TextField
              label="JWKS Rotation Period (days)"
              type="number"
              value={settings.jwks_rotation_days}
              onChange={(e) => setSettings({ ...settings, jwks_rotation_days: parseInt(e.target.value) })}
              inputProps={{ min: 30, max: 365 }}
              required
              fullWidth
              helperText="How often to automatically rotate signing keys (default: 90 days)"
            />

            <Box>
              <Button
                type="submit"
                variant="contained"
                startIcon={saving ? <CircularProgress size={20} color="inherit" /> : <SaveIcon />}
                disabled={saving}
              >
                {saving ? 'Saving...' : 'Save Settings'}
              </Button>
            </Box>
          </Box>
        </form>
      </Paper>

      <Paper elevation={2} sx={{ p: 3 }}>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
          <Typography variant="h6" fontWeight="bold">
            Signing Keys
          </Typography>
          <Button
            variant="outlined"
            color="error"
            startIcon={rotating ? <CircularProgress size={20} color="inherit" /> : <RefreshIcon />}
            onClick={handleRotateKeys}
            disabled={rotating}
          >
            {rotating ? 'Rotating...' : 'Rotate Keys'}
          </Button>
        </Box>
        <Divider sx={{ mb: 3 }} />
        <Box display="flex" flexDirection="column" gap={2}>
          {keys.length === 0 ? (
            <Typography color="text.secondary" textAlign="center">
              No signing keys found
            </Typography>
          ) : (
            keys.map((key) => (
              <Card key={key.kid} variant="outlined">
                <CardContent>
                  <Box display="flex" justifyContent="space-between" alignItems="flex-start">
                    <Box display="flex" gap={2}>
                      <Box
                        sx={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          width: 48,
                          height: 48,
                          borderRadius: 1,
                          bgcolor: 'primary.lighter',
                          color: 'primary.main',
                        }}
                      >
                        <KeyIcon />
                      </Box>
                      <Box>
                        <Box display="flex" gap={1} alignItems="center" mb={1}>
                          <Typography variant="subtitle2" color="text.secondary">
                            Key ID:
                          </Typography>
                          <Typography
                            component="code"
                            sx={{
                              bgcolor: 'grey.100',
                              px: 1,
                              py: 0.5,
                              borderRadius: 0.5,
                              fontFamily: 'monospace',
                              fontSize: '0.75rem',
                            }}
                          >
                            {key.kid}
                          </Typography>
                        </Box>
                        <Box display="flex" gap={2} flexWrap="wrap">
                          <Typography variant="body2" color="text.secondary">
                            <strong>Algorithm:</strong> {key.alg}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            <strong>Use:</strong> {key.use}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            <strong>Created:</strong> {new Date(key.created_at).toLocaleString()}
                          </Typography>
                        </Box>
                      </Box>
                    </Box>
                    {key.is_active && (
                      <Chip
                        label="Active"
                        color="success"
                        size="small"
                        sx={{ fontWeight: 'bold' }}
                      />
                    )}
                  </Box>
                </CardContent>
              </Card>
            ))
          )}
        </Box>
      </Paper>
    </Box>
  );
};

export default Settings;
