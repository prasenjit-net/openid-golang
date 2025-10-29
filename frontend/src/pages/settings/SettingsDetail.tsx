import { useNavigate } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Typography,
  Spin,
  Alert,
} from 'antd';
import {
  EditOutlined,
} from '@ant-design/icons';
import { useSettings } from '../../hooks/useApi';

const { Title } = Typography;

const SettingsDetail = () => {
  const navigate = useNavigate();
  const { data: settings, isLoading: loading } = useSettings();

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

