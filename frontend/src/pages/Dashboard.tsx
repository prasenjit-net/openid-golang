import { Row, Col, Card, Statistic, Typography, Spin, Alert } from 'antd';
import {
  UserOutlined,
  LockOutlined,
  KeyOutlined,
  LoginOutlined,
  SafetyOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { useStats } from '../hooks/useApi';

const { Title, Paragraph } = Typography;

const Dashboard = () => {
  const { data: stats, isLoading, error } = useStats();

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '50px' }}>
        <Spin size="large" tip="Loading statistics..." />
      </div>
    );
  }

  if (error) {
    return (
      <>
        <Title level={2} style={{ marginBottom: '24px' }}>
          Dashboard
        </Title>
        <Alert
          message="Error"
          description="Failed to load dashboard statistics. Please try again later."
          type="error"
          showIcon
        />
      </>
    );
  }

  return (
    <>
      <Title level={2} style={{ marginBottom: 24, marginTop: 0 }}>
        Dashboard
      </Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic
              title="Total Users"
              value={stats?.users || 0}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic
              title="OAuth Clients"
              value={stats?.clients || 0}
              prefix={<LockOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic
              title="Active Tokens"
              value={stats?.tokens || 0}
              prefix={<KeyOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic
              title="Recent Logins"
              value={stats?.logins || 0}
              suffix="(24h)"
              prefix={<LoginOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: '16px' }}>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic
              title="Total Signing Keys"
              value={stats?.total_keys || 0}
              prefix={<SafetyOutlined />}
              valueStyle={{ color: '#13c2c2' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic
              title="Active Keys"
              value={stats?.active_keys || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: '24px' }}>
        <Col span={24}>
          <Card title="System Overview" bordered={false}>
            <Paragraph type="secondary">
              Your OpenID Connect server is running with real-time statistics. 
              The dashboard displays live counts from your database including users, OAuth clients, 
              active tokens, and recent login activity from the last 24 hours.
            </Paragraph>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              Use the navigation menu to manage users, OAuth clients, signing keys, and server settings.
            </Paragraph>
          </Card>
        </Col>
      </Row>
    </>
  );
};

export default Dashboard;
