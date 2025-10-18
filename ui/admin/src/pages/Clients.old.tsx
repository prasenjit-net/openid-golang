import { useEffect, useState } from 'react';
import './Clients.css';

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
  const [showSecret, setShowSecret] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    redirect_uris: '',
  });

  useEffect(() => {
    fetchClients();
  }, []);

  const fetchClients = async () => {
    try {
      const response = await fetch('/api/admin/clients');
      const data = await response.json();
      setClients(data);
    } catch (error) {
      console.error('Failed to fetch clients:', error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/admin/clients', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: formData.name,
          redirect_uris: formData.redirect_uris.split('\n').map(uri => uri.trim()).filter(Boolean),
        }),
      });
      const newClient = await response.json();
      setShowModal(false);
      setFormData({ name: '', redirect_uris: '' });
      setShowSecret(newClient.client_id);
      fetchClients();
    } catch (error) {
      console.error('Failed to create client:', error);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this client?')) return;
    try {
      await fetch(`/api/admin/clients/${id}`, { method: 'DELETE' });
      fetchClients();
    } catch (error) {
      console.error('Failed to delete client:', error);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const getClientByClientId = (clientId: string) => {
    return clients.find(c => c.client_id === clientId);
  };

  return (
    <div className="clients">
      <div className="page-header">
        <h1>OAuth Clients</h1>
        <button className="btn-primary" onClick={() => setShowModal(true)}>
          + Register Client
        </button>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Client ID</th>
              <th>Redirect URIs</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {clients.map((client) => (
              <tr key={client.id}>
                <td>{client.name}</td>
                <td>
                  <code className="client-id">{client.client_id}</code>
                  <button
                    className="copy-btn"
                    onClick={() => copyToClipboard(client.client_id)}
                    title="Copy to clipboard"
                  >
                    📋
                  </button>
                </td>
                <td>
                  <div className="redirect-uris">
                    {client.redirect_uris.map((uri, i) => (
                      <div key={i} className="uri">{uri}</div>
                    ))}
                  </div>
                </td>
                <td>{new Date(client.created_at).toLocaleDateString()}</td>
                <td>
                  <button
                    className="btn-danger-sm"
                    onClick={() => handleDelete(client.id)}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>Register New Client</h2>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Client Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="My Application"
                  required
                />
              </div>
              <div className="form-group">
                <label>Redirect URIs (one per line)</label>
                <textarea
                  value={formData.redirect_uris}
                  onChange={(e) => setFormData({ ...formData, redirect_uris: e.target.value })}
                  placeholder="https://example.com/callback"
                  rows={5}
                  required
                />
              </div>
              <div className="form-actions">
                <button type="button" onClick={() => setShowModal(false)} className="btn-secondary">
                  Cancel
                </button>
                <button type="submit" className="btn-primary">
                  Register Client
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showSecret && (
        <div className="modal-overlay" onClick={() => setShowSecret(null)}>
          <div className="modal secret-modal" onClick={(e) => e.stopPropagation()}>
            <h2>Client Registered Successfully</h2>
            <div className="secret-info">
              <p className="warning">⚠️ Save these credentials securely. The secret will not be shown again!</p>
              <div className="credential">
                <label>Client ID:</label>
                <div className="credential-value">
                  <code>{showSecret}</code>
                  <button className="copy-btn" onClick={() => copyToClipboard(showSecret)}>
                    📋 Copy
                  </button>
                </div>
              </div>
              <div className="credential">
                <label>Client Secret:</label>
                <div className="credential-value">
                  <code>{getClientByClientId(showSecret)?.client_secret}</code>
                  <button
                    className="copy-btn"
                    onClick={() => copyToClipboard(getClientByClientId(showSecret)?.client_secret || '')}
                  >
                    📋 Copy
                  </button>
                </div>
              </div>
            </div>
            <div className="form-actions">
              <button onClick={() => setShowSecret(null)} className="btn-primary">
                I've Saved The Credentials
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Clients;
