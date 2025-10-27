import { useEffect } from 'react';
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
import { useClient, useUpdateClient } from '../../hooks/useApi';

const { Title } = Typography;
const { TextArea } = Input;

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

  const handleSubmit = async (values: any) => {
    if (!id) return;
    
    try {
      // Convert redirect URIs from text to array
      const redirect_uris = values.redirect_uris_text
        .split('\n')
        .map((uri: string) => uri.trim())
        .filter((uri: string) => uri.length > 0);

      const payload = {
        id,
        name: values.name,
        redirect_uris,
      };

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
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/clients')}
          style={{ marginBottom: 16 }}
        >
          Back to Search
        </Button>
        <Alert
          message="Error"
          description={error?.message || 'Client not found'}
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
              loading={updateClientMutation.isPending}
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
