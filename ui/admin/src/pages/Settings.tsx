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
  server_host: string;
  server_port: number;
  storage_type: string;
  json_file_path: string;
  mongo_uri: string;
  jwt_expiry_minutes: number;
  jwt_private_key: string;
  jwt_public_key: string;
}

const Settings = () => {
  const [settings, setSettings] = useState<Settings>({
    issuer: '',
    server_host: '0.0.0.0',
    server_port: 8080,
    storage_type: 'json',
    json_file_path: 'data.json',
    mongo_uri: '',
    jwt_expiry_minutes: 60,
    jwt_private_key: 'config/keys/private.key',
    jwt_public_key: 'config/keys/public.key',
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

            <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>Server Configuration</Typography>
            
            <TextField
              label="Server Host"
              value={settings.server_host}
              onChange={(e) => setSettings({ ...settings, server_host: e.target.value })}
              placeholder="0.0.0.0"
              required
              fullWidth
              helperText="The host address the server binds to"
            />

            <TextField
              label="Server Port"
              type="number"
              value={settings.server_port}
              onChange={(e) => setSettings({ ...settings, server_port: parseInt(e.target.value) })}
              inputProps={{ min: 1, max: 65535 }}
              required
              fullWidth
              helperText="The port the server listens on (default: 8080)"
            />

            <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>Storage Configuration</Typography>
            
            <TextField
              label="Storage Type"
              value={settings.storage_type}
              onChange={(e) => setSettings({ ...settings, storage_type: e.target.value })}
              placeholder="json"
              required
              fullWidth
              helperText="Storage backend: 'json' or 'mongodb'"
            />

            <TextField
              label="JSON File Path"
              value={settings.json_file_path}
              onChange={(e) => setSettings({ ...settings, json_file_path: e.target.value })}
              placeholder="data.json"
              fullWidth
              helperText="Path to JSON storage file (only used if storage_type is 'json')"
            />

            <TextField
              label="MongoDB URI"
              value={settings.mongo_uri}
              onChange={(e) => setSettings({ ...settings, mongo_uri: e.target.value })}
              placeholder="mongodb://localhost:27017/openid"
              fullWidth
              helperText="MongoDB connection string (only used if storage_type is 'mongodb')"
            />

            <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>JWT Configuration</Typography>
            
            <TextField
              label="JWT Expiry (minutes)"
              type="number"
              value={settings.jwt_expiry_minutes}
              onChange={(e) => setSettings({ ...settings, jwt_expiry_minutes: parseInt(e.target.value) })}
              inputProps={{ min: 1, max: 1440 }}
              required
              fullWidth
              helperText="How long JWT tokens are valid (default: 60 minutes)"
            />

            <TextField
              label="JWT Private Key Path"
              value={settings.jwt_private_key}
              onChange={(e) => setSettings({ ...settings, jwt_private_key: e.target.value })}
              placeholder="config/keys/private.key"
              required
              fullWidth
              helperText="Path to RSA private key for signing JWTs"
            />

            <TextField
              label="JWT Public Key Path"
              value={settings.jwt_public_key}
              onChange={(e) => setSettings({ ...settings, jwt_public_key: e.target.value })}
              placeholder="config/keys/public.key"
              required
              fullWidth
              helperText="Path to RSA public key for verifying JWTs"
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
