import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Form,
  Input,
  Button,
  Space,
  Spin,
  Alert,
  message,
  Select,
} from 'antd';
import { ArrowLeftOutlined, AppstoreOutlined, LinkOutlined, SettingOutlined } from '@ant-design/icons';
import { useClient, useUpdateClient } from '../../hooks/useApi';

const { TextArea } = Input;

const sectionCard = (icon: React.ReactNode, title: string, children: React.ReactNode) => (
  <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', boxShadow: 'var(--shadow-card)', overflow: 'hidden', marginBottom: 24 }}>
    <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 8 }}>
      <span style={{ color: 'var(--color-primary)', fontSize: 16, display: 'flex' }}>{icon}</span>
      <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>{title}</span>
    </div>
    <div style={{ padding: '16px 20px' }}>{children}</div>
  </div>
);

const ClientEdit = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const { data: client, isLoading: loading, error } = useClient(id || '');
  const updateClientMutation = useUpdateClient();

  useEffect(() => {
    if (client) {
      form.setFieldsValue({
        ...client,
        redirect_uris_text: client.redirect_uris?.join('\n') || '',
      });
    }
  }, [client, form]);

  const handleSubmit = async (values: { name: string; redirect_uris_text: string }) => {
    if (!id) return;
    try {
      const redirect_uris = values.redirect_uris_text
        .split('\n')
        .map((uri: string) => uri.trim())
        .filter((uri: string) => uri.length > 0);

      const payload = { id, name: values.name, redirect_uris };
      await updateClientMutation.mutateAsync(payload);
      message.success('Client updated successfully');
      navigate(`/clients/${id}`);
    } catch (error) {
      message.error('Failed to update client');
      console.error('Failed to update client:', error);
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '50px' }}>
        <Spin size="large" tip="Loading client..." />
      </div>
    );
  }

  if (error || !client) {
    return (
      <>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/clients/${id}`)} style={{ marginBottom: 16 }}>
          ← Back
        </Button>
        <Alert message="Error" description={error?.message || 'Client not found'} type="error" showIcon />
      </>
    );
  }

  return (
    <>
      {/* Back button */}
      <div style={{ marginBottom: 16 }}>
        <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate(`/clients/${id}`)} style={{ color: 'var(--text-secondary)', paddingLeft: 0 }}>
          ← Back
        </Button>
      </div>

      {/* Page header */}
      <div style={{ marginBottom: 28 }}>
        <span style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)' }}>Edit Client</span>
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
          <Form.Item label="Client ID" help="Client ID cannot be changed">
            <Input value={client.client_id} disabled />
          </Form.Item>
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
            <Button onClick={() => navigate(`/clients/${id}`)}>Cancel</Button>
            <Button type="primary" htmlType="submit" loading={updateClientMutation.isPending}>
              Save Changes
            </Button>
          </Space>
        </div>
      </Form>
    </>
  );
};

export default ClientEdit;
