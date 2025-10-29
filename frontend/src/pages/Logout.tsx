import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Button, Result } from 'antd';
import { CheckCircleOutlined, LoginOutlined } from '@ant-design/icons';
import { useAuth } from '../context/AuthContext';
import './Logout.css';

const Logout = () => {
  const navigate = useNavigate();
  const { logout } = useAuth();

  useEffect(() => {
    // Perform logout when component mounts
    logout();
  }, [logout]);

  const handleBackToLogin = () => {
    navigate('/signin', { replace: true });
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Card
        style={{
          width: 450,
          boxShadow: '0 8px 24px rgba(0,0,0,0.15)',
          borderRadius: '8px',
        }}
        bodyStyle={{ padding: '40px' }}
      >
        <Result
          icon={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
          title="Successfully Logged Out"
          subTitle="You have been logged out of your account. All your session data has been cleared."
          extra={[
            <Button
              key="login"
              type="primary"
              size="large"
              icon={<LoginOutlined />}
              onClick={handleBackToLogin}
              style={{
                height: '40px',
                fontSize: '16px',
              }}
            >
              Log Back In
            </Button>,
          ]}
        />
        <div style={{ marginTop: '24px', textAlign: 'center', color: '#8c8c8c' }}>
          <p style={{ margin: 0, fontSize: '14px' }}>
            Thank you for using OpenID Connect Admin
          </p>
        </div>
      </Card>
    </div>
  );
};

export default Logout;
