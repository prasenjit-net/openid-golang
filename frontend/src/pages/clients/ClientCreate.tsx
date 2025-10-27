import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  Button,
  Space,
  Typography,
  message,
  Select,
  Modal,
  Alert,
} from 'antd';
import { ArrowLeftOutlined, SaveOutlined, CopyOutlined, KeyOutlined } from '@ant-design/icons';
import { useCreateClient } from '../../hooks/useApi';

const { Title, Paragraph } = Typography;
const { TextArea } = Input;

interface ClientResponse {
  client_id: string;
  client_secret: string;
  name: string;
}

const ClientCreate = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const createClientMutation = useCreateClient();
  const [createdClient, setCreatedClient] = useState<ClientResponse | null>(null);
  const [secretModalVisible, setSecretModalVisible] = useState(false);

  const handleSubmit = async (values: any) => {
    try {
      // Convert redirect URIs from text to array
      const redirect_uris = values.redirect_uris_text
        .split('\n')
        .map((uri: string) => uri.trim())
        .filter((uri: string) => uri.length > 0);

      const payload = {
        name: values.name,
        redirect_uris,
      };

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
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/clients')}
          >
            Back to Search
          </Button>
          <Title level={2} style={{ margin: 0 }}>Create New Client</Title>
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

        <Card bordered={false} title="OAuth Configuration (Optional)" style={{ marginBottom: 24 }}>
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
              loading={createClientMutation.isPending}
            >
              Create Client
            </Button>
            <Button onClick={() => navigate('/clients')}>
              Cancel
            </Button>
          </Space>
        </Card>
      </Form>

      {/* Client Secret Modal */}
      <Modal
        title={
          <Space>
            <KeyOutlined />
            <span>Client Created Successfully</span>
          </Space>
        }
        open={secretModalVisible}
        onCancel={handleModalClose}
        footer={[
          <Button
            key="close"
            type="primary"
            onClick={handleModalClose}
          >
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
            <Paragraph>
              <strong>Client ID:</strong>
            </Paragraph>
            <Input
              value={createdClient.client_id}
              readOnly
              addonAfter={
                <Button
                  type="text"
                  size="small"
                  icon={<CopyOutlined />}
                  onClick={() => copyToClipboard(createdClient.client_id)}
                >
                  Copy
                </Button>
              }
            />
            <Paragraph style={{ marginTop: 16 }}>
              <strong>Client Secret:</strong>
            </Paragraph>
            <Input.TextArea
              value={createdClient.client_secret}
              readOnly
              rows={3}
              style={{ fontFamily: 'monospace' }}
            />
            <Button
              icon={<CopyOutlined />}
              onClick={() => copyToClipboard(createdClient.client_secret)}
              style={{ marginTop: 8 }}
            >
              Copy Secret
            </Button>
          </>
        )}
      </Modal>
    </>
  );
};

export default ClientCreate;
