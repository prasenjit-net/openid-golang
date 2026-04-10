import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Form,
  Input,
  Button,
  Space,
  message,
  Select,
  Modal,
  Alert,
} from 'antd';
import { ArrowLeftOutlined, CopyOutlined, KeyOutlined, AppstoreOutlined, LinkOutlined, SettingOutlined } from '@ant-design/icons';
import { useCreateClient } from '../../hooks/useApi';

const { TextArea } = Input;

interface ClientResponse {
  client_id: string;
  client_secret: string;
  name: string;
}

const sectionCard = (icon: React.ReactNode, title: string, children: React.ReactNode) => (
  <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', boxShadow: 'var(--shadow-card)', overflow: 'hidden', marginBottom: 24 }}>
    <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 8 }}>
      <span style={{ color: 'var(--color-primary)', fontSize: 16, display: 'flex' }}>{icon}</span>
      <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>{title}</span>
    </div>
    <div style={{ padding: '16px 20px' }}>{children}</div>
  </div>
);

const ClientCreate = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const createClientMutation = useCreateClient();
  const [createdClient, setCreatedClient] = useState<ClientResponse | null>(null);
  const [secretModalVisible, setSecretModalVisible] = useState(false);

  const handleSubmit = async (values: { name: string; redirect_uris_text: string }) => {
    try {
      const redirect_uris = values.redirect_uris_text
        .split('\n')
        .map((uri: string) => uri.trim())
        .filter((uri: string) => uri.length > 0);

      const payload = { name: values.name, redirect_uris };
      const data = await createClientMutation.mutateAsync(payload);
      setCreatedClient(data);
      setSecretModalVisible(true);
      message.success('Client created successfully');
    } catch (error) {
      message.error('Failed to create client');
      console.error('Failed to create client:', error);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('Copied to clipboard');
  };

  const handleModalClose = () => {
    setSecretModalVisible(false);
    if (createdClient) {
      navigate(`/clients/${createdClient.client_id}`);
    }
  };

  return (
    <>
      {/* Back button */}
      <div style={{ marginBottom: 16 }}>
        <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate('/clients')} style={{ color: 'var(--text-secondary)', paddingLeft: 0 }}>
          ← Clients
        </Button>
      </div>

      {/* Page header */}
      <div style={{ marginBottom: 28 }}>
        <span style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)' }}>Create New Client</span>
      </div>

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{
          grant_types: ['authorization_code', 'implicit'],
          response_types: ['code', 'token', 'id_token', 'id_token token'],
          scope: 'openid profile email',
          application_type: 'web',
        }}
      >
        {sectionCard(<AppstoreOutlined />, 'Basic Information', <>
          <Form.Item
            label="Client Name"
            name="name"
            rules={[{ required: true, message: 'Please enter client name' }]}
          >
            <Input placeholder="My Application" />
          </Form.Item>
          <Form.Item label="Application Type" name="application_type" style={{ marginBottom: 0 }}>
            <Select>
              <Select.Option value="web">Web</Select.Option>
              <Select.Option value="native">Native</Select.Option>
            </Select>
          </Form.Item>
        </>)}

        {sectionCard(<LinkOutlined />, 'Redirect URIs',
          <Form.Item
            label="Redirect URIs"
            name="redirect_uris_text"
            rules={[{ required: true, message: 'Please enter at least one redirect URI' }]}
            help="Enter one URI per line"
            style={{ marginBottom: 0 }}
          >
            <TextArea rows={5} placeholder={"http://localhost:3000/callback\nhttps://myapp.example.com/callback"} />
          </Form.Item>
        )}

        {sectionCard(<SettingOutlined />, 'OAuth Configuration', <>
          <Form.Item label="Grant Types" name="grant_types">
            <Select mode="multiple" placeholder="Select grant types">
              <Select.Option value="authorization_code">Authorization Code</Select.Option>
              <Select.Option value="implicit">Implicit</Select.Option>
              <Select.Option value="refresh_token">Refresh Token</Select.Option>
              <Select.Option value="client_credentials">Client Credentials</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="Response Types" name="response_types">
            <Select mode="multiple" placeholder="Select response types">
              <Select.Option value="code">code</Select.Option>
              <Select.Option value="token">token</Select.Option>
              <Select.Option value="id_token">id_token</Select.Option>
              <Select.Option value="id_token token">id_token token</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="Scope" name="scope" help="Space-separated list of scopes" style={{ marginBottom: 0 }}>
            <Input placeholder="openid profile email" />
          </Form.Item>
        </>)}

        {/* Action row */}
        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Space>
            <Button onClick={() => navigate('/clients')}>Cancel</Button>
            <Button type="primary" htmlType="submit" loading={createClientMutation.isPending}>
              Create Client
            </Button>
          </Space>
        </div>
      </Form>

      {/* Client Secret Modal */}
      <Modal
        title={<Space><KeyOutlined /><span>Client Created Successfully</span></Space>}
        open={secretModalVisible}
        onCancel={handleModalClose}
        footer={[
          <Button key="close" type="primary" onClick={handleModalClose}>
            Continue to Client Details
          </Button>,
        ]}
        closable={false}
      >
        <Alert
          message="Important"
          description="This is the only time the client secret will be displayed. Please save it securely."
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        {createdClient && (
          <>
            <p style={{ marginBottom: 4, fontWeight: 600, fontSize: 13 }}>Client ID:</p>
            <Input
              value={createdClient.client_id}
              readOnly
              addonAfter={
                <Button type="text" size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(createdClient.client_id)}>
                  Copy
                </Button>
              }
            />
            <p style={{ marginTop: 16, marginBottom: 4, fontWeight: 600, fontSize: 13 }}>Client Secret:</p>
            <Input.TextArea value={createdClient.client_secret} readOnly rows={3} style={{ fontFamily: 'monospace' }} />
            <Button icon={<CopyOutlined />} onClick={() => copyToClipboard(createdClient.client_secret)} style={{ marginTop: 8 }}>
              Copy Secret
            </Button>
          </>
        )}
      </Modal>
    </>
  );
};

export default ClientCreate;
