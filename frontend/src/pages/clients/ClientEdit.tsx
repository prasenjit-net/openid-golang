import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  Button,
  Space,
  Typography,
  Spin,
  Alert,
  message,
  Select,
} from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';

const { Title } = Typography;
const { TextArea } = Input;

interface Client {
  id: string;
  client_id: string;
  name: string;
  redirect_uris: string[];
  grant_types?: string[];
  response_types?: string[];
  scope?: string;
  application_type?: string;
}

const ClientEdit = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [client, setClient] = useState<Client | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchClient();
  }, [id]);

  const fetchClient = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await fetch(`/api/admin/clients/${id}`);
      if (!response.ok) {
        if (response.status === 404) {
          throw new Error('Client not found');
        }
        throw new Error('Failed to fetch client');
      }
      const data = await response.json();
      setClient(data);
      form.setFieldsValue({
        ...data,
        redirect_uris_text: data.redirect_uris?.join('\n') || '',
      });
    } catch (err: any) {
      setError(err.message);
      message.error(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      setSubmitting(true);
      
      // Convert redirect URIs from text to array
      const redirect_uris = values.redirect_uris_text
        .split('\n')
        .map((uri: string) => uri.trim())
        .filter((uri: string) => uri.length > 0);

      const payload = {
        id,
        name: values.name,
        redirect_uris,
        grant_types: values.grant_types,
        response_types: values.response_types,
        scope: values.scope,
        application_type: values.application_type,
      };

      const response = await fetch(`/api/admin/clients/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      
      if (!response.ok) throw new Error('Failed to update client');
      message.success('Client updated successfully');
      navigate(`/clients/${id}`);
    } catch (error) {
      message.error('Failed to update client');
      console.error('Failed to update client:', error);
    } finally {
      setSubmitting(false);
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
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/clients')}
          style={{ marginBottom: 16 }}
        >
          Back to Search
        </Button>
        <Alert
          message="Error"
          description={error || 'Client not found'}
          type="error"
          showIcon
        />
      </>
    );
  }

  return (
    <>
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(`/clients/${id}`)}
          >
            Back to Details
          </Button>
          <Title level={2} style={{ margin: 0 }}>Edit Client</Title>
        </Space>
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
        <Card bordered={false} title="Basic Information" style={{ marginBottom: 24 }}>
          <Form.Item
            label="Client ID"
            help="Client ID cannot be changed"
          >
            <Input value={client.client_id} disabled />
          </Form.Item>

          <Form.Item
            label="Client Name"
            name="name"
            rules={[{ required: true, message: 'Please enter client name' }]}
          >
            <Input placeholder="My Application" />
          </Form.Item>

          <Form.Item
            label="Application Type"
            name="application_type"
          >
            <Select>
              <Select.Option value="web">Web</Select.Option>
              <Select.Option value="native">Native</Select.Option>
            </Select>
          </Form.Item>
        </Card>

        <Card bordered={false} title="Redirect URIs" style={{ marginBottom: 24 }}>
          <Form.Item
            label="Redirect URIs"
            name="redirect_uris_text"
            rules={[{ required: true, message: 'Please enter at least one redirect URI' }]}
            help="Enter one URI per line"
          >
            <TextArea
              rows={5}
              placeholder="http://localhost:3000/callback&#10;https://myapp.example.com/callback"
            />
          </Form.Item>
        </Card>

        <Card bordered={false} title="OAuth Configuration" style={{ marginBottom: 24 }}>
          <Form.Item
            label="Grant Types"
            name="grant_types"
          >
            <Select mode="multiple" placeholder="Select grant types">
              <Select.Option value="authorization_code">Authorization Code</Select.Option>
              <Select.Option value="implicit">Implicit</Select.Option>
              <Select.Option value="refresh_token">Refresh Token</Select.Option>
              <Select.Option value="client_credentials">Client Credentials</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Response Types"
            name="response_types"
          >
            <Select mode="multiple" placeholder="Select response types">
              <Select.Option value="code">code</Select.Option>
              <Select.Option value="token">token</Select.Option>
              <Select.Option value="id_token">id_token</Select.Option>
              <Select.Option value="id_token token">id_token token</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Scope"
            name="scope"
            help="Space-separated list of scopes"
          >
            <Input placeholder="openid profile email" />
          </Form.Item>
        </Card>

        <Card bordered={false}>
          <Space>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SaveOutlined />}
              loading={submitting}
            >
              Save Changes
            </Button>
            <Button onClick={() => navigate(`/clients/${id}`)}>
              Cancel
            </Button>
          </Space>
        </Card>
      </Form>
    </>
  );
};

export default ClientEdit;
