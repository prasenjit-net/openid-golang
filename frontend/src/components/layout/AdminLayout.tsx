import { useState, useEffect } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Drawer, Dropdown, Avatar, Tooltip } from 'antd';
import type { MenuProps } from 'antd';
import {
  DashboardOutlined,
  TeamOutlined,
  AppstoreOutlined,
  KeyOutlined,
  SettingOutlined,
  AuditOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SunOutlined,
  MoonOutlined,
  BellOutlined,
  UserOutlined,
  LogoutOutlined,
  LockOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import { useTheme } from '../../context/ThemeContext';
import { Logo } from '../Logo';

const NAV_ITEMS = [
  { key: '/dashboard', label: 'Dashboard', icon: <DashboardOutlined /> },
  { key: '/users', label: 'Users', icon: <TeamOutlined /> },
  { key: '/clients', label: 'Clients', icon: <AppstoreOutlined /> },
  { key: '/tokens', label: 'Tokens', icon: <SafetyCertificateOutlined /> },
  { key: '/keys', label: 'Keys', icon: <KeyOutlined /> },
  { key: '/settings', label: 'Settings', icon: <SettingOutlined /> },
  { key: '/audit', label: 'Audit Log', icon: <AuditOutlined /> },
];

function getUserInfo() {
  try {
    const raw = localStorage.getItem('user_info');
    if (raw) return JSON.parse(raw) as { name?: string; username?: string; email?: string };
  } catch {}
  return { name: 'Admin', username: 'admin' };
}

function getInitials(name?: string) {
  if (!name) return 'A';
  return name.split(' ').map(w => w[0]).slice(0, 2).join('').toUpperCase();
}

interface SidebarProps {
  collapsed: boolean;
  onCollapse: (v: boolean) => void;
}

function SidebarContent({ collapsed, onCollapse: _onCollapse }: SidebarProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const userInfo = getUserInfo();

  const isActive = (key: string) => {
    if (key === '/dashboard') return location.pathname === '/dashboard' || location.pathname === '/';
    return location.pathname.startsWith(key);
  };

  return (
    <div
      style={{
        width: collapsed ? 64 : 220,
        minWidth: collapsed ? 64 : 220,
        height: '100vh',
        background: 'var(--sidebar-bg)',
        display: 'flex',
        flexDirection: 'column',
        transition: 'width 250ms ease, min-width 250ms ease',
        overflow: 'hidden',
        position: 'fixed',
        left: 0,
        top: 0,
        bottom: 0,
        zIndex: 100,
        borderRight: '1px solid rgba(255,255,255,0.06)',
      }}
    >
      {/* Logo area */}
      <div
        style={{
          height: 'var(--header-height)',
          display: 'flex',
          alignItems: 'center',
          padding: '0 16px',
          borderBottom: '1px solid rgba(255,255,255,0.06)',
          flexShrink: 0,
          overflow: 'hidden',
        }}
      >
        <Logo size={28} variant={collapsed ? 'icon' : 'full'} />
      </div>

      {/* Nav items */}
      <nav style={{ flex: 1, padding: '12px 8px', overflowY: 'auto', overflowX: 'hidden' }}>
        {NAV_ITEMS.map(item => {
          const active = isActive(item.key);
          return (
            <Tooltip
              key={item.key}
              title={collapsed ? item.label : ''}
              placement="right"
            >
              <div
                onClick={() => navigate(item.key)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  height: 42,
                  padding: '0 12px',
                  marginBottom: 2,
                  borderRadius: 8,
                  cursor: 'pointer',
                  background: active ? 'var(--sidebar-active-bg)' : 'transparent',
                  borderLeft: active ? '3px solid var(--color-primary)' : '3px solid transparent',
                  color: active ? 'var(--sidebar-active-text)' : 'var(--sidebar-text)',
                  transition: 'background 150ms, color 150ms, border-color 150ms',
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                }}
                onMouseEnter={e => {
                  if (!active) (e.currentTarget as HTMLDivElement).style.background = 'var(--sidebar-hover-bg)';
                }}
                onMouseLeave={e => {
                  if (!active) (e.currentTarget as HTMLDivElement).style.background = 'transparent';
                }}
              >
                <span style={{ fontSize: 16, flexShrink: 0, lineHeight: 1, display: 'flex', alignItems: 'center' }}>
                  {item.icon}
                </span>
                {!collapsed && (
                  <span style={{ fontSize: 13, fontWeight: active ? 600 : 400, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {item.label}
                  </span>
                )}
              </div>
            </Tooltip>
          );
        })}
      </nav>

      {/* User info at bottom */}
      <div
        style={{
          padding: '12px 16px',
          borderTop: '1px solid rgba(255,255,255,0.06)',
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          overflow: 'hidden',
        }}
      >
        <Avatar
          size={28}
          style={{ background: 'var(--color-primary)', color: 'white', flexShrink: 0, fontSize: 11, fontWeight: 600 }}
        >
          {getInitials(userInfo.name || userInfo.username)}
        </Avatar>
        {!collapsed && (
          <div style={{ overflow: 'hidden' }}>
            <div style={{ fontSize: 12, fontWeight: 600, color: 'rgba(203,213,225,0.9)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              {userInfo.name || userInfo.username}
            </div>
            <div style={{ fontSize: 10, color: 'var(--sidebar-text-muted)', letterSpacing: '0.06em', textTransform: 'uppercase' }}>Admin</div>
          </div>
        )}
      </div>
    </div>
  );
}

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const { isDark, toggleTheme } = useTheme();
  const navigate = useNavigate();
  const userInfo = getUserInfo();

  useEffect(() => {
    const handler = () => setIsMobile(window.innerWidth < 768);
    window.addEventListener('resize', handler);
    return () => window.removeEventListener('resize', handler);
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('user_info');
    navigate('/logout');
  };

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: 'Profile',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'password',
      icon: <LockOutlined />,
      label: 'Change Password',
      onClick: () => navigate('/profile'),
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Sign Out',
      danger: true,
      onClick: handleLogout,
    },
  ];

  const sidebarWidth = isMobile ? 0 : (collapsed ? 64 : 220);

  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden', background: 'var(--bg)' }}>
      {/* Desktop Sidebar */}
      {!isMobile && (
        <SidebarContent collapsed={collapsed} onCollapse={setCollapsed} />
      )}

      {/* Mobile Drawer */}
      {isMobile && (
        <Drawer
          open={mobileOpen}
          onClose={() => setMobileOpen(false)}
          placement="left"
          width={220}
          styles={{
            body: { padding: 0, background: 'var(--sidebar-bg)' },
            header: { display: 'none' },
            wrapper: { boxShadow: 'var(--shadow-lg)' },
          }}
          closeIcon={null}
        >
          <SidebarContent collapsed={false} onCollapse={() => setMobileOpen(false)} />
        </Drawer>
      )}

      {/* Main content area */}
      <div
        style={{
          flex: 1,
          marginLeft: sidebarWidth,
          transition: 'margin-left 250ms ease',
          display: 'flex',
          flexDirection: 'column',
          minWidth: 0,
          height: '100vh',
          overflow: 'clip',
        }}
      >
        {/* Header */}
        <header
          style={{
            height: 'var(--header-height)',
            background: 'var(--surface)',
            borderBottom: '1px solid var(--border)',
            display: 'flex',
            alignItems: 'center',
            padding: '0 20px',
            gap: 8,
            flexShrink: 0,
            position: 'sticky',
            top: 0,
            zIndex: 50,
            boxShadow: 'var(--shadow-sm)',
          }}
        >
          {/* Collapse / hamburger button */}
          <button
            onClick={() => isMobile ? setMobileOpen(true) : setCollapsed(c => !c)}
            style={{
              width: 32,
              height: 32,
              border: 'none',
              background: 'transparent',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderRadius: 6,
              color: 'var(--text-secondary)',
              fontSize: 16,
              transition: 'background 150ms',
              flexShrink: 0,
            }}
            onMouseEnter={e => (e.currentTarget.style.background = 'var(--surface-hover)')}
            onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
          >
            {(!isMobile && collapsed) ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </button>

          {/* Spacer */}
          <div style={{ flex: 1 }} />

          {/* Theme toggle */}
          <Tooltip title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}>
            <button
              onClick={toggleTheme}
              style={{
                width: 32,
                height: 32,
                border: 'none',
                background: 'transparent',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                borderRadius: 6,
                color: 'var(--text-secondary)',
                fontSize: 15,
                transition: 'background 150ms',
              }}
              onMouseEnter={e => (e.currentTarget.style.background = 'var(--surface-hover)')}
              onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
            >
              {isDark ? <SunOutlined /> : <MoonOutlined />}
            </button>
          </Tooltip>

          {/* Notifications */}
          <Tooltip title="Notifications">
            <button
              style={{
                width: 32,
                height: 32,
                border: 'none',
                background: 'transparent',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                borderRadius: 6,
                color: 'var(--text-secondary)',
                fontSize: 15,
                transition: 'background 150ms',
              }}
              onMouseEnter={e => (e.currentTarget.style.background = 'var(--surface-hover)')}
              onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
            >
              <BellOutlined />
            </button>
          </Tooltip>

          {/* Divider */}
          <div style={{ width: 1, height: 24, background: 'var(--border)', margin: '0 4px' }} />

          {/* User dropdown */}
          <Dropdown menu={{ items: userMenuItems }} trigger={['click']} placement="bottomRight">
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                cursor: 'pointer',
                padding: '4px 8px',
                borderRadius: 8,
                transition: 'background 150ms',
              }}
              onMouseEnter={e => (e.currentTarget.style.background = 'var(--surface-hover)')}
              onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
            >
              <Avatar
                size={28}
                style={{ background: 'var(--color-primary)', color: 'white', fontSize: 11, fontWeight: 600 }}
              >
                {getInitials(userInfo.name || userInfo.username)}
              </Avatar>
              <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1 }}>
                <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-primary)' }}>
                  {userInfo.name || userInfo.username || 'Admin'}
                </span>
                <span
                  style={{
                    fontSize: 10,
                    fontWeight: 500,
                    color: 'var(--color-primary)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.04em',
                  }}
                >
                  Admin
                </span>
              </div>
            </div>
          </Dropdown>
        </header>

        {/* Page content */}
        <main
          style={{
            flex: 1,
            overflow: 'auto',
            padding: '24px',
            background: 'var(--bg)',
          }}
        >
          <Outlet />
        </main>
      </div>
    </div>
  );
}
