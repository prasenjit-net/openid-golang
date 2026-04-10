import { useNavigate } from 'react-router-dom';
import { Button, Spin, Alert } from 'antd';
import { EditOutlined, GlobalOutlined, DatabaseOutlined, SafetyOutlined } from '@ant-design/icons';
import { useSettings } from '../../hooks/useApi';

const InfoRow = ({ label, value }: { label: string; value: React.ReactNode }) => (
  <div style={{ display: 'flex', padding: '10px 20px', borderBottom: '1px solid var(--border-subtle)' }}>
    <span style={{ width: 200, flexShrink: 0, fontSize: 13, color: 'var(--text-secondary)', fontWeight: 500 }}>{label}</span>
    <span style={{ fontSize: 13, color: 'var(--text-primary)' }}>{value}</span>
  </div>
);

const InfoCard = ({ icon, title, children }: { icon: React.ReactNode; title: string; children: React.ReactNode }) => (
  <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', boxShadow: 'var(--shadow-card)', overflow: 'hidden', marginBottom: 24 }}>
    <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 8 }}>
      <span style={{ color: 'var(--color-primary)', fontSize: 16 }}>{icon}</span>
      <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>{title}</span>
    </div>
    <div style={{ padding: '4px 0' }}>{children}</div>
  </div>
);

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
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 24 }}>
        <span style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)' }}>Server Settings</span>
        <Button
          type="primary"
          icon={<EditOutlined />}
          onClick={() => navigate('/settings/edit')}
          style={{ background: 'var(--color-primary)', borderColor: 'var(--color-primary)' }}
        >
          Edit Settings
        </Button>
      </div>

      <InfoCard icon={<GlobalOutlined />} title="Server Configuration">
        <InfoRow label="Issuer URL" value={settings.issuer} />
        <InfoRow label="Server Host" value={settings.server_host} />
        <InfoRow label="Server Port" value={settings.server_port} />
      </InfoCard>

      <InfoCard icon={<DatabaseOutlined />} title="Storage Configuration">
        <InfoRow label="Storage Type" value={settings.storage_type} />
        {settings.storage_type === 'json' && (
          <InfoRow label="JSON File Path" value={settings.json_file_path || 'Not configured'} />
        )}
        {settings.storage_type === 'mongodb' && (
          <InfoRow label="MongoDB URI" value={settings.mongo_uri ? '••••••••••••' : 'Not configured'} />
        )}
      </InfoCard>

      <InfoCard icon={<SafetyOutlined />} title="JWT Configuration">
        <InfoRow label="Token Expiry" value={`${settings.jwt_expiry_minutes} minutes`} />
        <div style={{ padding: '12px 20px' }}>
          <Alert
            message="RSA Signing Keys"
            description="Signing keys are now managed separately. Go to the Keys page to view and rotate signing keys."
            type="info"
            showIcon
          />
        </div>
      </InfoCard>
    </>
  );
};

export default SettingsDetail;

