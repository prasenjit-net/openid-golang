import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Button,
  Space,
  Tag,
  Spin,
  Alert,
  message,
  Popconfirm,
  Modal,
  Input,
} from 'antd';
import {
  EditOutlined,
  ArrowLeftOutlined,
  DeleteOutlined,
  ReloadOutlined,
  CopyOutlined,
  KeyOutlined,
  AppstoreOutlined,
  LinkOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { useClient, useDeleteClient, useRegenerateClientSecret } from '../../hooks/useApi';

const infoCard = (icon: React.ReactNode, title: string, children: React.ReactNode) => (
  <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', boxShadow: 'var(--shadow-card)', overflow: 'hidden', marginBottom: 24 }}>
    <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 8 }}>
      <span style={{ color: 'var(--color-primary)', fontSize: 16, display: 'flex' }}>{icon}</span>
      <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>{title}</span>
    </div>
    <div style={{ padding: '4px 0' }}>{children}</div>
  </div>
);

const infoRow = (label: string, value: React.ReactNode, last = false) => (
  <div style={{ display: 'flex', padding: '10px 20px', borderBottom: last ? 'none' : '1px solid var(--border-subtle)' }}>
    <span style={{ width: 200, flexShrink: 0, fontSize: 13, color: 'var(--text-secondary)', fontWeight: 500 }}>{label}</span>
    <span style={{ fontSize: 13, color: 'var(--text-primary)' }}>{value}</span>
  </div>
);

const ClientDetail = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [secretModalVisible, setSecretModalVisible] = useState(false);
  const [newSecret, setNewSecret] = useState<string>('');

  const { data: client, isLoading: loading } = useClient(id || '');
  const deleteClientMutation = useDeleteClient();
  const regenerateSecretMutation = useRegenerateClientSecret();

  const handleDelete = async () => {
    try {
      await deleteClientMutation.mutateAsync(id!);
      message.success('Client deleted successfully');
      navigate('/clients');
    } catch (error) {
      message.error('Failed to delete client');
      console.error('Failed to delete client:', error);
    }
  };

  const handleRegenerateSecret = async () => {
    try {
      const data = await regenerateSecretMutation.mutateAsync(id!);
      setNewSecret(data.client_secret);
      setSecretModalVisible(true);
      message.success('Client secret regenerated successfully');
    } catch (error) {
      message.error('Failed to regenerate secret');
      console.error('Failed to regenerate secret:', error);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('Copied to clipboard');
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '50px' }}>
        <Spin size="large" tip="Loading client details..." />
      </div>
    );
  }

  if (!client) {
    return (
      <>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/clients')} style={{ marginBottom: 16 }}>
          ← Clients
        </Button>
        <Alert message="Error" description="Client not found" type="error" showIcon />
      </>
    );
  }

  return (
    <>
      {/* Back button */}
      <div style={{ marginBottom: 16 }}>
        <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate('/clients')} style={{ color: 'var(--text-secondary)', paddingLeft: 0 }}>
          ← Clients
        </Button>
      </div>

      {/* Page header */}
      <div style={{ marginBottom: 28, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)' }}>
            {client.name || client.client_name || 'Unnamed Client'}
          </span>
          <code style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 6, padding: '2px 10px', fontFamily: 'monospace', fontSize: 12, color: 'var(--color-primary)' }}>
            {client.client_id}
          </code>
        </div>
        <Space>
          <Button type="primary" icon={<EditOutlined />} onClick={() => navigate(`/clients/${id}/edit`)}>
            Edit Client
          </Button>
          <Popconfirm
            title="Delete client"
            description="Are you sure you want to delete this client? This action cannot be undone."
            onConfirm={handleDelete}
            okText="Yes"
            cancelText="No"
            okButtonProps={{ danger: true }}
          >
            <Button danger icon={<DeleteOutlined />}>Delete Client</Button>
          </Popconfirm>
          <Popconfirm
            title="Regenerate Client Secret"
            description="This will invalidate the old secret. The new secret will only be shown once."
            onConfirm={handleRegenerateSecret}
            okText="Regenerate"
            cancelText="Cancel"
            okButtonProps={{ danger: true }}
          >
            <Button icon={<ReloadOutlined />} style={{ color: '#d97706', borderColor: '#d97706' }}>
              Regenerate Secret
            </Button>
          </Popconfirm>
        </Space>
      </div>

      {/* Basic Information */}
      {infoCard(<AppstoreOutlined />, 'Basic Information', <>
        {infoRow('Client Name', client.name || client.client_name || '-')}
        {infoRow('Application Type', client.application_type || 'web')}
        {infoRow('Created', new Date(client.created_at).toLocaleString(), true)}
      </>)}

      {/* Redirect URIs */}
      {infoCard(<LinkOutlined />, 'Redirect URIs',
        <div style={{ padding: '10px 20px', display: 'flex', flexWrap: 'wrap', gap: 8 }}>
          {(client.redirect_uris || []).length > 0
            ? client.redirect_uris.map((uri: string, i: number) => (
                <span key={i} style={{ background: '#eff6ff', border: '1px solid #bfdbfe', borderRadius: 6, padding: '3px 10px', fontSize: 13, color: '#1d4ed8', fontFamily: 'monospace' }}>{uri}</span>
              ))
            : <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>No redirect URIs configured</span>}
        </div>
      )}

      {/* OAuth Configuration */}
      {infoCard(<SettingOutlined />, 'OAuth Configuration', <>
        {infoRow('Grant Types',
          <Space wrap>
            {client.grant_types?.map((type: string, i: number) => <Tag key={i} color="green">{type}</Tag>) || '-'}
          </Space>
        )}
        {infoRow('Response Types',
          <Space wrap>
            {client.response_types?.map((type: string, i: number) => <Tag key={i} color="purple">{type}</Tag>) || '-'}
          </Space>
        )}
        {infoRow('Scope', client.scope || 'openid profile email')}
        {infoRow('Token Auth Method', client.token_endpoint_auth_method || 'client_secret_basic', true)}
      </>)}

      {/* Client Secret */}
      {infoCard(<KeyOutlined />, 'Client Secret',
        <div style={{ padding: '16px 20px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span style={{ fontSize: 13, color: 'var(--text-secondary)', maxWidth: 480 }}>
            The client secret is stored securely and cannot be retrieved. Use Regenerate to issue a new one.
          </span>
          <Popconfirm
            title="Regenerate Client Secret"
            description="This will invalidate the old secret. The new secret will only be shown once."
            onConfirm={handleRegenerateSecret}
            okText="Regenerate"
            cancelText="Cancel"
            okButtonProps={{ danger: true }}
          >
            <Button icon={<ReloadOutlined />} style={{ color: '#d97706', borderColor: '#d97706' }}>
              Regenerate Secret
            </Button>
          </Popconfirm>
        </div>
      )}

      {/* New Secret Modal */}
      <Modal
        title={<Space><KeyOutlined /><span>New Client Secret Generated</span></Space>}
        open={secretModalVisible}
        onCancel={() => { setSecretModalVisible(false); setNewSecret(''); }}
        footer={[
          <Button key="close" type="primary" onClick={() => { setSecretModalVisible(false); setNewSecret(''); }}>
            Close
          </Button>,
        ]}
      >
        <Alert
          message="Important"
          description="This is the only time the client secret will be displayed. Please save it securely."
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <p style={{ marginBottom: 4, fontWeight: 600, fontSize: 13 }}>Client ID:</p>
        <Input
          value={client.client_id}
          readOnly
          addonAfter={
            <Button type="text" size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(client.client_id)}>
              Copy
            </Button>
          }
        />
        <p style={{ marginTop: 16, marginBottom: 4, fontWeight: 600, fontSize: 13 }}>Client Secret:</p>
        <Input.TextArea value={newSecret} readOnly rows={3} style={{ fontFamily: 'monospace' }} />
        <Button icon={<CopyOutlined />} onClick={() => copyToClipboard(newSecret)} style={{ marginTop: 8 }}>
          Copy Secret
        </Button>
      </Modal>
    </>
  );
};

export default ClientDetail;
