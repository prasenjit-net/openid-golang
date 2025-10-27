// Dashboard statistics and metrics
export interface DashboardStats {
  total_users: number;
  active_users: number;
  inactive_users: number;
  total_clients: number;
  active_clients: number;
  inactive_clients: number;
  total_tokens: number;
  active_tokens: number;
  tokens_today: number;
  active_sessions: number;
  auth_requests_24h: number;
  auth_requests_7d: number;
  auth_requests_30d: number;
}

export interface ActivityLogEntry {
  id: string;
  timestamp: string;
  admin_id: string;
  admin_username: string;
  action: string;
  resource: string;
  resource_id: string;
  details: string;
  ip_address: string;
}

export interface SystemHealth {
  storage_status: 'healthy' | 'degraded' | 'down';
  storage_type: 'json' | 'mongodb';
  certificate_status: 'valid' | 'expiring_soon' | 'expired';
  certificate_expiry?: string;
  uptime_seconds: number;
  version: string;
  memory_mb?: number;
}

export interface DashboardData {
  stats: DashboardStats;
  recent_activity: ActivityLogEntry[];
  system_health: SystemHealth;
}
