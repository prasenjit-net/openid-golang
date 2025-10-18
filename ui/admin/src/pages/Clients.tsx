import { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
  Alert,
  CircularProgress,
  Chip,
  Tooltip,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  ContentCopy as CopyIcon,
  Edit as EditIcon,
} from '@mui/icons-material';

interface Client {
  id: string;
  client_id: string;
  client_secret: string;
  name: string;
  redirect_uris: string[];
  created_at: string;
}

const Clients = () => {
  const [clients, setClients] = useState<Client[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editingClient, setEditingClient] = useState<Client | null>(null);
  const [showSecret, setShowSecret] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    redirect_uris: '',
  });

  useEffect(() => {
    fetchClients();
  }, []);

  const fetchClients = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await fetch('/api/admin/clients');
      if (!response.ok) throw new Error('Failed to fetch clients');
      const data = await response.json();
      setClients(data);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to fetch clients');
      console.error('Failed to fetch clients:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const url = editingClient 
        ? `/api/admin/clients/${editingClient.id}`
        : '/api/admin/clients';
      const method = editingClient ? 'PUT' : 'POST';
      
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: formData.name,
          redirect_uris: formData.redirect_uris.split('\n').map(uri => uri.trim()).filter(Boolean),
        }),
      });
      
      if (!response.ok) throw new Error(editingClient ? 'Failed to update client' : 'Failed to create client');
      
      const resultClient = await response.json();
      handleCloseModal();
      
      if (!editingClient) {
        setShowSecret(resultClient.client_id);
      }
      
      fetchClients();
    } catch (error) {
      setError(error instanceof Error ? error.message : editingClient ? 'Failed to update client' : 'Failed to create client');
      console.error(editingClient ? 'Failed to update client:' : 'Failed to create client:', error);
    }
  };

  const handleEdit = (client: Client) => {
    setEditingClient(client);
    setFormData({
      name: client.name,
      redirect_uris: client.redirect_uris.join('\n'),
    });
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setEditingClient(null);
    setFormData({ name: '', redirect_uris: '' });
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this client?')) return;
    try {
      const response = await fetch(`/api/admin/clients/${id}`, { method: 'DELETE' });
      if (!response.ok) throw new Error('Failed to delete client');
      fetchClients();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to delete client');
      console.error('Failed to delete client:', error);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const getClientByClientId = (clientId: string) => {
    return clients.find(c => c.client_id === clientId);
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
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" fontWeight="bold">
          OAuth Clients
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setShowModal(true)}
        >
          Register Client
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      <TableContainer component={Paper} elevation={2}>
        <Table>
          <TableHead>
            <TableRow sx={{ bgcolor: 'grey.50' }}>
              <TableCell><strong>Name</strong></TableCell>
              <TableCell><strong>Client ID</strong></TableCell>
              <TableCell><strong>Redirect URIs</strong></TableCell>
              <TableCell><strong>Created</strong></TableCell>
              <TableCell align="center"><strong>Actions</strong></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {clients.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                  <Typography color="text.secondary">No OAuth clients registered</Typography>
                </TableCell>
              </TableRow>
            ) : (
              clients.map((client) => (
                <TableRow key={client.id} hover>
                  <TableCell>{client.name}</TableCell>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={1}>
                      <Typography
                        component="code"
                        sx={{
                          bgcolor: 'grey.100',
                          px: 1,
                          py: 0.5,
                          borderRadius: 1,
                          fontFamily: 'monospace',
                          fontSize: '0.875rem',
                        }}
                      >
                        {client.client_id}
                      </Typography>
                      <Tooltip title="Copy Client ID">
                        <IconButton
                          size="small"
                          onClick={() => copyToClipboard(client.client_id)}
                        >
                          <CopyIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Box display="flex" flexDirection="column" gap={0.5}>
                      {client.redirect_uris.map((uri, i) => (
                        <Chip
                          key={i}
                          label={uri}
                          size="small"
                          variant="outlined"
                          sx={{ maxWidth: 'fit-content' }}
                        />
                      ))}
                    </Box>
                  </TableCell>
                  <TableCell>{new Date(client.created_at).toLocaleDateString()}</TableCell>
                  <TableCell align="center">
                    <IconButton
                      color="primary"
                      size="small"
                      onClick={() => handleEdit(client)}
                      title="Edit client"
                    >
                      <EditIcon />
                    </IconButton>
                    <IconButton
                      color="error"
                      size="small"
                      onClick={() => handleDelete(client.id)}
                      title="Delete client"
                    >
                      <DeleteIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={showModal} onClose={handleCloseModal} maxWidth="sm" fullWidth>
        <form onSubmit={handleSubmit}>
          <DialogTitle>{editingClient ? 'Edit Client' : 'Register New Client'}</DialogTitle>
          <DialogContent>
            <Box display="flex" flexDirection="column" gap={2} mt={1}>
              <TextField
                label="Client Name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="My Application"
                required
                fullWidth
                autoFocus
              />
              <TextField
                label="Redirect URIs"
                value={formData.redirect_uris}
                onChange={(e) => setFormData({ ...formData, redirect_uris: e.target.value })}
                placeholder="https://example.com/callback"
                multiline
                rows={5}
                required
                fullWidth
                helperText="Enter one URI per line"
              />
            </Box>
          </DialogContent>
          <DialogActions sx={{ p: 2 }}>
            <Button onClick={handleCloseModal}>
              Cancel
            </Button>
            <Button type="submit" variant="contained">
              {editingClient ? 'Update Client' : 'Register Client'}
            </Button>
          </DialogActions>
        </form>
      </Dialog>

      <Dialog 
        open={!!showSecret} 
        onClose={() => setShowSecret(null)} 
        maxWidth="md" 
        fullWidth
      >
        <DialogTitle>Client Registered Successfully</DialogTitle>
        <DialogContent>
          <Alert severity="warning" sx={{ mb: 3 }}>
            Save these credentials securely. The secret will not be shown again!
          </Alert>
          <Box display="flex" flexDirection="column" gap={2}>
            <Box>
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                Client ID
              </Typography>
              <Paper variant="outlined" sx={{ p: 2, display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography
                  component="code"
                  sx={{ flexGrow: 1, fontFamily: 'monospace', wordBreak: 'break-all' }}
                >
                  {showSecret}
                </Typography>
                <IconButton onClick={() => showSecret && copyToClipboard(showSecret)}>
                  <CopyIcon />
                </IconButton>
              </Paper>
            </Box>
            <Box>
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                Client Secret
              </Typography>
              <Paper variant="outlined" sx={{ p: 2, display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography
                  component="code"
                  sx={{ flexGrow: 1, fontFamily: 'monospace', wordBreak: 'break-all' }}
                >
                  {getClientByClientId(showSecret || '')?.client_secret}
                </Typography>
                <IconButton onClick={() => {
                  const secret = getClientByClientId(showSecret || '')?.client_secret;
                  if (secret) copyToClipboard(secret);
                }}>
                  <CopyIcon />
                </IconButton>
              </Paper>
            </Box>
          </Box>
        </DialogContent>
        <DialogActions sx={{ p: 2 }}>
          <Button variant="contained" onClick={() => setShowSecret(null)}>
            I've Saved The Credentials
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Clients;
