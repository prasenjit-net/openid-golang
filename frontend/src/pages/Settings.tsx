import { Typography, Card, Descriptions, Alert } from 'antd';

const { Title } = Typography;

const Settings = () => {
  return (
    <>
      <Title level={2} style={{ marginBottom: 24, marginTop: 0 }}>Configuration</Title>
      
      <Card bordered={false}>
        <Alert
          message="Configuration Management"
          description="Configuration management features will be available in the next version. Currently, settings are managed through the backend configuration files."
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Descriptions title="Server Information" bordered column={1}>
          <Descriptions.Item label="Issuer">Configured via backend</Descriptions.Item>
          <Descriptions.Item label="Server Host">0.0.0.0</Descriptions.Item>
          <Descriptions.Item label="Server Port">8080</Descriptions.Item>
          <Descriptions.Item label="Storage Type">JSON / MongoDB</Descriptions.Item>
        </Descriptions>

        <Descriptions 
          title="JWT Configuration" 
          bordered 
          column={1}
          style={{ marginTop: 24 }}
        >
          <Descriptions.Item label="Token Expiry">60 minutes (default)</Descriptions.Item>
          <Descriptions.Item label="Signing Algorithm">RS256</Descriptions.Item>
          <Descriptions.Item label="Key Rotation">Manual via backend API</Descriptions.Item>
        </Descriptions>
      </Card>
    </>
  );
};

export default Settings;
