import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Typography,
  message,
  Spin,
  Alert,
} from 'antd';
import {
  EditOutlined,
} from '@ant-design/icons';

const { Title } = Typography;

interface Settings {
  issuer: string;
  server_host: string;
  server_port: number;
  storage_type: string;
  json_file_path: string;
  mongo_uri: string;
  jwt_expiry_minutes: number;
  jwt_private_key: string;
  jwt_public_key: string;
}

const SettingsDetail = () => {
  const navigate = useNavigate();
  const [settings, setSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchSettings = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/admin/settings');
      if (!response.ok) throw new Error('Failed to fetch settings');
      const data = await response.json();
      setSettings(data);
    } catch (error) {
      message.error('Failed to load settings');
      console.error('Failed to fetch settings:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSettings();
  }, []);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!settings) {
    return (
      <Alert
        message="Settings Not Found"
        description="Unable to load server settings."
        type="error"
        showIcon
      />
    );
  }

  return (
    <>
      <div style={{ marginBottom: 24 }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Title level={2} style={{ margin: 0 }}>Server Settings</Title>
          <Button
            type="primary"
            icon={<EditOutlined />}
            onClick={() => navigate('/settings/edit')}
          >
            Edit Settings
          </Button>
        </Space>
      </div>

      <Card bordered={false} title="Server Configuration" style={{ marginBottom: 24 }}>
        <Descriptions column={1} bordered>
          <Descriptions.Item label="Issuer URL">
            {settings.issuer}
          </Descriptions.Item>
          <Descriptions.Item label="Server Host">
            {settings.server_host}
          </Descriptions.Item>
          <Descriptions.Item label="Server Port">
            {settings.server_port}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card bordered={false} title="Storage Configuration" style={{ marginBottom: 24 }}>
        <Descriptions column={1} bordered>
          <Descriptions.Item label="Storage Type">
            {settings.storage_type}
          </Descriptions.Item>
          {settings.storage_type === 'json' && (
            <Descriptions.Item label="JSON File Path">
              {settings.json_file_path || 'Not configured'}
            </Descriptions.Item>
          )}
          {settings.storage_type === 'mongodb' && (
            <Descriptions.Item label="MongoDB URI">
              {settings.mongo_uri ? '••••••••••••' : 'Not configured'}
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>

      <Card bordered={false} title="JWT Configuration" style={{ marginBottom: 24 }}>
        <Descriptions column={1} bordered>
          <Descriptions.Item label="Token Expiry">
            {settings.jwt_expiry_minutes} minutes
          </Descriptions.Item>
        </Descriptions>
        <Alert
          message="RSA Signing Keys"
          description="Signing keys are now managed separately. Go to the Keys page to view and rotate signing keys."
          type="info"
          showIcon
          style={{ marginTop: 16 }}
        />
      </Card>
    </>
  );
};

export default SettingsDetail;

