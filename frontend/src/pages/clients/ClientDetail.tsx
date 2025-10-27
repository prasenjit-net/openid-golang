import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Typography,
  Tag,
  Spin,
  Alert,
  message,
  Popconfirm,
  Modal,
  Input,
  List,
} from 'antd';
import {
  EditOutlined,
  ArrowLeftOutlined,
  DeleteOutlined,
  LockOutlined,
  ReloadOutlined,
  CopyOutlined,
  KeyOutlined,
} from '@ant-design/icons';
import { useClient, useDeleteClient, useRegenerateClientSecret } from '../../hooks/useApi';

const { Title, Paragraph } = Typography;

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
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/clients')}
          style={{ marginBottom: 16 }}
        >
          Back to Search
        </Button>
        <Alert
          message="Error"
          description={'Client not found'}
          type="error"
          showIcon
        />
      </>
    );
  }

  return (
    <>
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/clients')}
          >
            Back
          </Button>
          <Title level={2} style={{ margin: 0 }}>Client Details</Title>
        </Space>
        <Space>
          <Popconfirm
            title="Regenerate Client Secret"
            description="This will generate a new secret and invalidate the old one. The new secret will only be shown once."
            onConfirm={handleRegenerateSecret}
            okText="Regenerate"
            cancelText="Cancel"
            okButtonProps={{ danger: true }}
          >
            <Button icon={<ReloadOutlined />}>
              Regenerate Secret
            </Button>
          </Popconfirm>
          <Button
            type="primary"
            icon={<EditOutlined />}
            onClick={() => navigate(`/clients/${id}/edit`)}
          >
            Edit
          </Button>
          <Popconfirm
            title="Delete client"
            description="Are you sure you want to delete this client? This action cannot be undone."
            onConfirm={handleDelete}
            okText="Yes"
            cancelText="No"
            okButtonProps={{ danger: true }}
          >
            <Button danger icon={<DeleteOutlined />}>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      </div>

      {/* Basic Information */}
      <Card bordered={false} style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginTop: 0 }}>
          <LockOutlined /> Basic Information
        </Title>
        <Descriptions bordered column={2}>
          <Descriptions.Item label="Client ID" span={2}>
            <Space>
              <code>{client.client_id}</code>
              <Button
                type="text"
                size="small"
                icon={<CopyOutlined />}
                onClick={() => copyToClipboard(client.client_id)}
              />
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="Client Name" span={2}>
            {client.name || client.client_name || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Application Type">
            {client.application_type || 'web'}
          </Descriptions.Item>
          <Descriptions.Item label="Created">
            {new Date(client.created_at).toLocaleString()}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Redirect URIs */}
      <Card bordered={false} style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginTop: 0 }}>Redirect URIs</Title>
        <List
          dataSource={client.redirect_uris || []}
          renderItem={(uri: string) => (
            <List.Item>
              <Tag color="blue">{uri}</Tag>
            </List.Item>
          )}
        />
      </Card>

      {/* OAuth Configuration */}
      <Card bordered={false} style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginTop: 0 }}>OAuth Configuration</Title>
        <Descriptions bordered column={1}>
          <Descriptions.Item label="Grant Types">
            <Space wrap>
              {client.grant_types?.map((type: string, i: number) => (
                <Tag key={i} color="green">{type}</Tag>
              )) || '-'}
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="Response Types">
            <Space wrap>
              {client.response_types?.map((type: string, i: number) => (
                <Tag key={i} color="purple">{type}</Tag>
              )) || '-'}
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="Scope">
            {client.scope || 'openid profile email'}
          </Descriptions.Item>
          <Descriptions.Item label="Token Endpoint Auth Method">
            {client.token_endpoint_auth_method || 'client_secret_basic'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Additional Information */}
      {(client.contacts || client.client_uri || client.logo_uri || client.policy_uri || client.tos_uri || client.jwks_uri) && (
        <Card bordered={false} style={{ marginBottom: 24 }}>
          <Title level={4} style={{ marginTop: 0 }}>Additional Information</Title>
          <Descriptions bordered column={1}>
            {client.contacts && client.contacts.length > 0 && (
              <Descriptions.Item label="Contacts">
                {client.contacts.join(', ')}
              </Descriptions.Item>
            )}
            {client.client_uri && (
              <Descriptions.Item label="Client URI">
                <a href={client.client_uri} target="_blank" rel="noopener noreferrer">
                  {client.client_uri}
                </a>
              </Descriptions.Item>
            )}
            {client.logo_uri && (
              <Descriptions.Item label="Logo URI">
                <a href={client.logo_uri} target="_blank" rel="noopener noreferrer">
                  {client.logo_uri}
                </a>
              </Descriptions.Item>
            )}
            {client.policy_uri && (
              <Descriptions.Item label="Policy URI">
                <a href={client.policy_uri} target="_blank" rel="noopener noreferrer">
                  {client.policy_uri}
                </a>
              </Descriptions.Item>
            )}
            {client.tos_uri && (
              <Descriptions.Item label="Terms of Service URI">
                <a href={client.tos_uri} target="_blank" rel="noopener noreferrer">
                  {client.tos_uri}
                </a>
              </Descriptions.Item>
            )}
            {client.jwks_uri && (
              <Descriptions.Item label="JWKS URI">
                <a href={client.jwks_uri} target="_blank" rel="noopener noreferrer">
                  {client.jwks_uri}
                </a>
              </Descriptions.Item>
            )}
          </Descriptions>
        </Card>
      )}

      {/* New Secret Modal */}
      <Modal
        title={
          <Space>
            <KeyOutlined />
            <span>New Client Secret Generated</span>
          </Space>
        }
        open={secretModalVisible}
        onCancel={() => {
          setSecretModalVisible(false);
          setNewSecret('');
        }}
        footer={[
          <Button
            key="close"
            type="primary"
            onClick={() => {
              setSecretModalVisible(false);
              setNewSecret('');
            }}
          >
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
        <Paragraph>
          <strong>Client ID:</strong>
        </Paragraph>
        <Input
          value={client.client_id}
          readOnly
          addonAfter={
            <Button
              type="text"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => copyToClipboard(client.client_id)}
            >
              Copy
            </Button>
          }
        />
        <Paragraph style={{ marginTop: 16 }}>
          <strong>Client Secret:</strong>
        </Paragraph>
        <Input.TextArea
          value={newSecret}
          readOnly
          rows={3}
          style={{ fontFamily: 'monospace' }}
        />
        <Button
          icon={<CopyOutlined />}
          onClick={() => copyToClipboard(newSecret)}
          style={{ marginTop: 8 }}
        >
          Copy Secret
        </Button>
      </Modal>
    </>
  );
};

export default ClientDetail;
