import { useState } from 'react';
import { Spin, Alert } from 'antd';
import {
  UserOutlined,
  AppstoreOutlined,
  KeyOutlined,
  LoginOutlined,
  SafetyOutlined,
  CheckCircleOutlined,
  PlusOutlined,
  AuditOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useStats, useSettings } from '../hooks/useApi';

interface StatCardProps {
  icon: React.ReactNode;
  iconBg: string;
  iconColor: string;
  value: number | string;
  label: string;
  badge?: string;
}

const StatCard = ({ icon, iconBg, iconColor, value, label, badge }: StatCardProps) => {
  const [hovered, setHovered] = useState(false);
  return (
    <div
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        background: 'var(--surface)',
        borderRadius: 12,
        padding: '20px 24px',
        border: '1px solid var(--border)',
        boxShadow: hovered
          ? '0 8px 24px rgba(0,0,0,0.12)'
          : 'var(--shadow-card)',
        display: 'flex',
        alignItems: 'center',
        gap: 16,
        transition: 'transform 200ms, box-shadow 200ms',
        transform: hovered ? 'translateY(-2px)' : 'none',
        cursor: 'default',
      }}
    >
      <div
        style={{
          width: 48,
          height: 48,
          borderRadius: 10,
          background: iconBg,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
        }}
      >
        <span style={{ fontSize: 22, color: iconColor }}>{icon}</span>
      </div>
      <div>
        <div style={{ fontSize: 28, fontWeight: 700, color: 'var(--text-primary)', lineHeight: 1 }}>
          {value}
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-secondary)', marginTop: 4 }}>{label}</div>
        {badge && (
          <div style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 3 }}>{badge}</div>
        )}
      </div>
    </div>
  );
};

const STATS_CONFIG = [
  {
    key: 'users',
    icon: <UserOutlined />,
    iconBg: 'rgba(13,148,136,0.1)',
    iconColor: '#0D9488',
    label: 'Total Users',
    badge: 'Total registered',
  },
  {
    key: 'clients',
    icon: <AppstoreOutlined />,
    iconBg: 'rgba(245,158,11,0.1)',
    iconColor: '#F59E0B',
    label: 'OAuth Clients',
    badge: 'Registered apps',
  },
  {
    key: 'tokens',
    icon: <KeyOutlined />,
    iconBg: 'rgba(14,165,233,0.1)',
    iconColor: '#0EA5E9',
    label: 'Active Tokens',
    badge: 'Currently valid',
  },
  {
    key: 'logins',
    icon: <LoginOutlined />,
    iconBg: 'rgba(16,185,129,0.1)',
    iconColor: '#10B981',
    label: 'Recent Logins',
    badge: 'Last 24 hours',
  },
  {
    key: 'total_keys',
    icon: <SafetyOutlined />,
    iconBg: 'rgba(139,92,246,0.1)',
    iconColor: '#8B5CF6',
    label: 'Total Keys',
    badge: 'All signing keys',
  },
  {
    key: 'active_keys',
    icon: <CheckCircleOutlined />,
    iconBg: 'rgba(34,197,94,0.1)',
    iconColor: '#22C55E',
    label: 'Active Keys',
    badge: 'Signing new tokens',
  },
];

const QUICK_ACTIONS = [
  { icon: <PlusOutlined />, label: 'Add User', path: '/users', color: '#0D9488', bg: 'rgba(13,148,136,0.08)' },
  { icon: <AppstoreOutlined />, label: 'Register Client', path: '/clients', color: '#F59E0B', bg: 'rgba(245,158,11,0.08)' },
  { icon: <ReloadOutlined />, label: 'Rotate Keys', path: '/keys', color: '#8B5CF6', bg: 'rgba(139,92,246,0.08)' },
  { icon: <AuditOutlined />, label: 'View Audit Log', path: '/audit', color: '#0EA5E9', bg: 'rgba(14,165,233,0.08)' },
];

const Dashboard = () => {
  const { data: stats, isLoading, error } = useStats();
  const { data: settings } = useSettings();
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '80px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <>
        <div style={{ marginBottom: 24 }}>
          <h1 style={{ margin: 0, fontSize: 'var(--text-3xl)', fontWeight: 'var(--font-bold)', color: 'var(--text-primary)' }}>
            Dashboard
          </h1>
        </div>
        <Alert
          message="Failed to load statistics"
          description="Could not fetch dashboard data. Please try refreshing."
          type="error"
          showIcon
        />
      </>
    );
  }

  return (
    <>
      {/* Page Header */}
      <div style={{ marginBottom: 28 }}>
        <h1 style={{ margin: 0, fontSize: 'var(--text-3xl)', fontWeight: 'var(--font-bold)', color: 'var(--text-primary)', lineHeight: 1.2 }}>
          Dashboard
        </h1>
        <p style={{ margin: '4px 0 0', fontSize: 'var(--text-sm)', color: 'var(--text-muted)' }}>
          System overview
        </p>
      </div>

      {/* Stats Grid */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))',
          gap: 16,
          marginBottom: 28,
        }}
      >
        {STATS_CONFIG.map((cfg) => (
          <StatCard
            key={cfg.key}
            icon={cfg.icon}
            iconBg={cfg.iconBg}
            iconColor={cfg.iconColor}
            value={stats?.[cfg.key] ?? 0}
            label={cfg.label}
            badge={cfg.badge}
          />
        ))}
      </div>

      {/* Quick Actions */}
      <div style={{ marginBottom: 28 }}>
        <h2 style={{ margin: '0 0 12px', fontSize: 'var(--text-lg)', fontWeight: 'var(--font-semibold)', color: 'var(--text-primary)' }}>
          Quick Actions
        </h2>
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
          {QUICK_ACTIONS.map((action) => (
            <QuickActionCard key={action.label} {...action} onClick={() => navigate(action.path)} />
          ))}
        </div>
      </div>

      {/* System Info */}
      <div
        style={{
          background: 'var(--surface)',
          borderRadius: 12,
          border: '1px solid var(--border)',
          boxShadow: 'var(--shadow-card)',
          overflow: 'hidden',
        }}
      >
        <div style={{ padding: '16px 24px', borderBottom: '1px solid var(--border)' }}>
          <span style={{ fontSize: 'var(--text-md)', fontWeight: 'var(--font-semibold)', color: 'var(--text-primary)' }}>
            System Info
          </span>
        </div>
        <div
          style={{
            padding: '20px 24px',
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            gap: '16px 32px',
          }}
        >
          <InfoRow label="Issuer URL" value={settings?.issuer || '—'} mono />
          <InfoRow label="Storage Type" value={settings?.storage_type || '—'} />
          <InfoRow label="JWT Expiry" value={settings?.jwt_expiry_minutes ? `${settings.jwt_expiry_minutes} minutes` : '—'} />
          <InfoRow
            label="Registration"
            value={settings?.registration_enabled !== false ? 'Open' : 'Closed'}
            valueColor={settings?.registration_enabled !== false ? '#10B981' : '#EF4444'}
          />
        </div>
      </div>
    </>
  );
};

interface QuickActionCardProps {
  icon: React.ReactNode;
  label: string;
  color: string;
  bg: string;
  onClick: () => void;
}

const QuickActionCard = ({ icon, label, color, bg, onClick }: QuickActionCardProps) => {
  const [hovered, setHovered] = useState(false);
  return (
    <div
      onClick={onClick}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        background: hovered ? bg : 'var(--surface)',
        border: `1px solid ${hovered ? color + '40' : 'var(--border)'}`,
        borderRadius: 10,
        padding: '12px 20px',
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        cursor: 'pointer',
        transition: 'all 150ms',
        boxShadow: 'var(--shadow-card)',
        minWidth: 148,
      }}
    >
      <span style={{ fontSize: 18, color: hovered ? color : 'var(--text-secondary)' }}>{icon}</span>
      <span style={{ fontSize: 'var(--text-sm)', fontWeight: 'var(--font-medium)', color: hovered ? color : 'var(--text-primary)' }}>
        {label}
      </span>
    </div>
  );
};

interface InfoRowProps {
  label: string;
  value: string;
  mono?: boolean;
  valueColor?: string;
}

const InfoRow = ({ label, value, mono, valueColor }: InfoRowProps) => (
  <div>
    <div style={{ fontSize: 11, fontWeight: 'var(--font-semibold)', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 3 }}>
      {label}
    </div>
    <div
      style={{
        fontSize: 'var(--text-sm)',
        color: valueColor || 'var(--text-primary)',
        fontFamily: mono ? 'var(--font-mono)' : undefined,
        wordBreak: 'break-all',
      }}
    >
      {value}
    </div>
  </div>
);

export default Dashboard;
