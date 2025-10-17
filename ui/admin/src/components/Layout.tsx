import { Outlet, Link, useLocation } from 'react-router-dom';
import './Layout.css';

const Layout = () => {
  const location = useLocation();

  const isActive = (path: string) => {
    return location.pathname === path ? 'active' : '';
  };

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="sidebar-header">
          <h1>OpenID Admin</h1>
        </div>
        <nav className="sidebar-nav">
          <Link to="/dashboard" className={`nav-item ${isActive('/dashboard')}`}>
            <span className="icon">📊</span>
            <span>Dashboard</span>
          </Link>
          <Link to="/users" className={`nav-item ${isActive('/users')}`}>
            <span className="icon">👥</span>
            <span>Users</span>
          </Link>
          <Link to="/clients" className={`nav-item ${isActive('/clients')}`}>
            <span className="icon">🔑</span>
            <span>OAuth Clients</span>
          </Link>
          <Link to="/settings" className={`nav-item ${isActive('/settings')}`}>
            <span className="icon">⚙️</span>
            <span>Settings</span>
          </Link>
        </nav>
        <div className="sidebar-footer">
          <button className="logout-btn" onClick={() => {
            // Handle logout
            window.location.href = '/api/admin/logout';
          }}>
            <span className="icon">🚪</span>
            <span>Logout</span>
          </button>
        </div>
      </aside>
      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
};

export default Layout;
