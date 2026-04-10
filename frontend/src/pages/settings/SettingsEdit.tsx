import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, Input, InputNumber, Button, Space, message, Select, Spin, Alert } from 'antd';
import { GlobalOutlined, DatabaseOutlined, SafetyOutlined, SaveOutlined } from '@ant-design/icons';
import { useSettings, useUpdateSettings } from '../../hooks/useApi';

const FormSection = ({ icon, title, children }: { icon: React.ReactNode; title: string; children: React.ReactNode }) => (
  <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', boxShadow: 'var(--shadow-card)', overflow: 'hidden', marginBottom: 24 }}>
    <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 8 }}>
      <span style={{ color: 'var(--color-primary)', fontSize: 16 }}>{icon}</span>
      <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>{title}</span>
    </div>
    <div style={{ padding: '24px' }}>{children}</div>
  </div>
);

const SettingsEdit = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const { data: settings, isLoading: loading } = useSettings();
  const updateSettingsMutation = useUpdateSettings();
  const [storageType, setStorageType] = useState<string>('json');

  useEffect(() => {
    if (settings) {
      form.setFieldsValue({
        issuer: settings.issuer,
        server_host: settings.server_host,
        server_port: settings.server_port,
        storage_type: settings.storage_type,
        json_file_path: settings.json_file_path,
        mongo_uri: settings.mongo_uri,
        jwt_expiry_minutes: settings.jwt_expiry_minutes,
      });
      setStorageType(settings.storage_type);
    }
  }, [settings, form]);

  const handleSubmit = async (values: {
    issuer: string;
    server_host: string;
    server_port: number;
    storage_type: string;
    json_file_path: string;
    mongo_uri: string;
    jwt_expiry_minutes: number;
  }) => {
    try {
      const result = await updateSettingsMutation.mutateAsync(values);
      message.success('Settings updated successfully');
      if (result.message && result.message.includes('not persisted')) {
        message.warning('Changes are in memory only. Restart may revert changes.', 5);
      }
      navigate('/settings');
    } catch (error) {
      message.error('Failed to update settings');
      console.error('Failed to update settings:', error);
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
      <div style={{ marginBottom: 4 }}>
        <button
          onClick={() => navigate('/settings')}
          style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: 14, color: 'var(--color-primary)', padding: 0, marginBottom: 12 }}
        >
          ← Settings
        </button>
        <div style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)', marginBottom: 16 }}>Edit Settings</div>
      </div>

      <Alert
        message="Important Notice"
        description="Settings changes are applied in memory only and may not persist after server restart. For permanent changes, update the configuration file."
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Form form={form} layout="vertical" onFinish={handleSubmit}>
        <FormSection icon={<GlobalOutlined />} title="Server Configuration">
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
        </FormSection>

        <FormSection icon={<DatabaseOutlined />} title="Storage Configuration">
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
        </FormSection>

        <FormSection icon={<SafetyOutlined />} title="JWT Configuration">
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
        </FormSection>

        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Space>
            <Button onClick={() => navigate('/settings')}>Cancel</Button>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SaveOutlined />}
              loading={updateSettingsMutation.isPending}
              style={{ background: 'var(--color-primary)', borderColor: 'var(--color-primary)' }}
            >
              Save Settings
            </Button>
          </Space>
        </div>
      </Form>
    </>
  );
};

export default SettingsEdit;
