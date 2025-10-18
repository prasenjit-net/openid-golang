import { useStats } from '../hooks/useApi';
import './Dashboard.css';

const Dashboard = () => {
  const { data: stats, isLoading, error } = useStats();

  if (isLoading) {
    return (
      <div className="dashboard">
        <h1>Dashboard</h1>
        <div className="loading">Loading stats...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="dashboard">
        <h1>Dashboard</h1>
        <div className="error">Failed to load stats: {error.message}</div>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <h1>Dashboard</h1>
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon">👥</div>
          <div className="stat-content">
            <h3>Total Users</h3>
            <p className="stat-value">{stats?.users || 0}</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon">🔑</div>
          <div className="stat-content">
            <h3>OAuth Clients</h3>
            <p className="stat-value">{stats?.clients || 0}</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon">🎫</div>
          <div className="stat-content">
            <h3>Active Tokens</h3>
            <p className="stat-value">{stats.activeTokens}</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon">🎫</div>
          <div className="stat-content">
            <h3>Active Tokens</h3>
            <p className="stat-value">{stats?.tokens || 0}</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon">📊</div>
          <div className="stat-content">
            <h3>Recent Logins</h3>
            <p className="stat-value">{stats?.logins || 0}</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
