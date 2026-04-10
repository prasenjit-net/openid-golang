import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from 'antd';
import { CheckOutlined } from '@ant-design/icons';
import { useAuth } from '../context/AuthContext';

const Logout = () => {
  const navigate = useNavigate();
  const { logout } = useAuth();

  useEffect(() => {
    logout();
  }, [logout]);

  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #0B1120 0%, #0F2027 50%, #162032 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div
        style={{
          width: 420,
          background: 'var(--surface, #fff)',
          borderRadius: 16,
          padding: 40,
          boxShadow: 'var(--shadow-lg, 0 20px 60px rgba(0,0,0,0.4))',
          textAlign: 'center',
        }}
      >
        <div
          style={{
            width: 64,
            height: 64,
            borderRadius: '50%',
            background: '#0D9488',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto',
          }}
        >
          <CheckOutlined style={{ fontSize: 32, color: '#fff' }} />
        </div>

        <div style={{ fontSize: 20, fontWeight: 600, color: 'var(--text-primary, #111)', marginTop: 20 }}>
          You've been signed out
        </div>
        <div style={{ fontSize: 14, color: 'var(--text-muted, #6b7280)', marginTop: 8, textAlign: 'center' }}>
          Your session has been securely terminated.
        </div>

        <div style={{ height: 1, background: 'var(--border, #e5e7eb)', margin: '24px 0' }} />

        <Button
          type="primary"
          block
          style={{ background: '#0D9488', borderColor: '#0D9488', height: 40, fontSize: 15 }}
          onClick={() => navigate('/signin')}
        >
          Sign In Again
        </Button>

        <div style={{ fontSize: 12, color: 'var(--text-muted, #6b7280)', marginTop: 20 }}>
          OpenID Connect Admin Portal
        </div>
      </div>
    </div>
  );
};

export default Logout;
