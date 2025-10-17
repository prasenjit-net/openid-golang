import { useEffect, useState } from 'react';
import './Dashboard.css';

interface Stats {
  totalUsers: number;
  totalClients: number;
  activeTokens: number;
  recentLogins: number;
}

const Dashboard = () => {
  const [stats, setStats] = useState<Stats>({
    totalUsers: 0,
    totalClients: 0,
    activeTokens: 0,
    recentLogins: 0,
  });

  useEffect(() => {
    fetchStats();
  }, []);

  const fetchStats = async () => {
    try {
      const response = await fetch('/api/admin/stats');
      const data = await response.json();
      setStats(data);
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  return (
    <div className="dashboard">
      <h1>Dashboard</h1>
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon">👥</div>
          <div className="stat-content">
            <h3>Total Users</h3>
            <p className="stat-value">{stats.totalUsers}</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon">🔑</div>
          <div className="stat-content">
            <h3>OAuth Clients</h3>
            <p className="stat-value">{stats.totalClients}</p>
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
          <div className="stat-icon">📊</div>
          <div className="stat-content">
            <h3>Recent Logins</h3>
            <p className="stat-value">{stats.recentLogins}</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
