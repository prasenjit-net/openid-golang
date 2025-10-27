import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  InputNumber,
  Button,
  Space,
  Typography,
  message,
  Select,
  Spin,
  Alert,
} from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';

const { Title } = Typography;

const SettingsEdit = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [storageType, setStorageType] = useState<string>('json');

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        setLoading(true);
        const response = await fetch('/api/admin/settings');
        if (!response.ok) throw new Error('Failed to fetch settings');
        const data = await response.json();
        
        form.setFieldsValue({
          issuer: data.issuer,
          server_host: data.server_host,
          server_port: data.server_port,
          storage_type: data.storage_type,
          json_file_path: data.json_file_path,
          mongo_uri: data.mongo_uri,
          jwt_expiry_minutes: data.jwt_expiry_minutes,
        });
        
        setStorageType(data.storage_type);
      } catch (error) {
        message.error('Failed to load settings');
        console.error('Failed to fetch settings:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchSettings();
  }, [form]);

  const handleSubmit = async (values: any) => {
    try {
      setSubmitting(true);
      
      const response = await fetch('/api/admin/settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      });
      
      if (!response.ok) throw new Error('Failed to update settings');
      
      const result = await response.json();
      message.success('Settings updated successfully');
      
      // Show warning if changes are not persisted
      if (result.message && result.message.includes('not persisted')) {
        message.warning('Changes are in memory only. Restart may revert changes.', 5);
      }
      
      navigate('/settings');
    } catch (error) {
      message.error('Failed to update settings');
      console.error('Failed to update settings:', error);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <>
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/settings')}
          >
            Back to Settings
          </Button>
          <Title level={2} style={{ margin: 0 }}>Edit Settings</Title>
        </Space>
      </div>

      <Alert
        message="Important Notice"
        description="Settings changes are applied in memory only and may not persist after server restart. For permanent changes, update the configuration file."
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
      >
        <Card bordered={false} title="Server Configuration" style={{ marginBottom: 24 }}>
          <Form.Item
            label="Issuer URL"
            name="issuer"
            rules={[
              { required: true, message: 'Please enter issuer URL' },
              { type: 'url', message: 'Please enter a valid URL' },
            ]}
            help="The base URL of this OpenID Provider"
          >
            <Input placeholder="https://auth.example.com" />
          </Form.Item>

          <Form.Item
            label="Server Host"
            name="server_host"
            rules={[{ required: true, message: 'Please enter server host' }]}
            help="The host address the server binds to"
          >
            <Input placeholder="0.0.0.0" />
          </Form.Item>

          <Form.Item
            label="Server Port"
            name="server_port"
            rules={[
              { required: true, message: 'Please enter server port' },
              { type: 'number', min: 1, max: 65535, message: 'Port must be between 1 and 65535' },
            ]}
            help="The port the server listens on"
          >
            <InputNumber min={1} max={65535} style={{ width: '100%' }} />
          </Form.Item>
        </Card>

        <Card bordered={false} title="Storage Configuration" style={{ marginBottom: 24 }}>
          <Form.Item
            label="Storage Type"
            name="storage_type"
            rules={[{ required: true, message: 'Please select storage type' }]}
            help="The type of storage backend to use"
          >
            <Select onChange={(value) => setStorageType(value)}>
              <Select.Option value="json">JSON File</Select.Option>
              <Select.Option value="mongodb">MongoDB</Select.Option>
            </Select>
          </Form.Item>

          {storageType === 'json' && (
            <Form.Item
              label="JSON File Path"
              name="json_file_path"
              rules={[{ required: true, message: 'Please enter JSON file path' }]}
              help="Path to the JSON file for data storage"
            >
              <Input placeholder="./data.json" />
            </Form.Item>
          )}

          {storageType === 'mongodb' && (
            <Form.Item
              label="MongoDB URI"
              name="mongo_uri"
              rules={[{ required: true, message: 'Please enter MongoDB URI' }]}
              help="MongoDB connection string"
            >
              <Input.Password placeholder="mongodb://localhost:27017/openid" />
            </Form.Item>
          )}
        </Card>

        <Card bordered={false} title="JWT Configuration" style={{ marginBottom: 24 }}>
          <Form.Item
            label="Token Expiry (Minutes)"
            name="jwt_expiry_minutes"
            rules={[
              { required: true, message: 'Please enter token expiry' },
              { type: 'number', min: 1, max: 43200, message: 'Expiry must be between 1 minute and 30 days' },
            ]}
            help="How long ID tokens remain valid"
          >
            <InputNumber min={1} max={43200} style={{ width: '100%' }} />
          </Form.Item>

          <Alert
            message="RSA Keys"
            description="RSA signing keys cannot be edited directly. Use the 'Rotate Keys' button on the detail page to generate new keys."
            type="info"
            showIcon
          />
        </Card>

        <Card bordered={false}>
          <Space>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SaveOutlined />}
              loading={submitting}
            >
              Save Settings
            </Button>
            <Button onClick={() => navigate('/settings')}>
              Cancel
            </Button>
          </Space>
        </Card>
      </Form>
    </>
  );
};

export default SettingsEdit;
