import { useStats } from '../hooks/useApi';
import {
  Box,
  Card,
  CardContent,
  Typography,
  CircularProgress,
  Alert,
  Paper,
} from '@mui/material';
import {
  People as PeopleIcon,
  VpnKey as VpnKeyIcon,
  Token as TokenIcon,
  Login as LoginIcon,
} from '@mui/icons-material';

interface StatCardProps {
  title: string;
  value: number;
  icon: React.ReactNode;
  color: string;
}

const StatCard = ({ title, value, icon, color }: StatCardProps) => (
  <Card 
    elevation={2}
    sx={{ 
      height: '100%',
      transition: 'transform 0.2s, box-shadow 0.2s',
      '&:hover': {
        transform: 'translateY(-4px)',
        boxShadow: 6,
      }
    }}
  >
    <CardContent>
      <Box display="flex" alignItems="center" justifyContent="space-between">
        <Box>
          <Typography color="text.secondary" variant="overline" gutterBottom>
            {title}
          </Typography>
          <Typography variant="h3" component="div" fontWeight="bold">
            {value}
          </Typography>
        </Box>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: 64,
            height: 64,
            borderRadius: 2,
            bgcolor: `${color}.lighter`,
            color: `${color}.main`,
          }}
        >
          {icon}
        </Box>
      </Box>
    </CardContent>
  </Card>
);

const Dashboard = () => {
  const { data: stats, isLoading, error } = useStats();

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="60vh">
        <CircularProgress size={60} />
      </Box>
    );
  }

  if (error) {
    return (
      <Box>
        <Typography variant="h4" gutterBottom fontWeight="bold">
          Dashboard
        </Typography>
        <Alert severity="error" sx={{ mt: 2 }}>
          Failed to load stats: {error.message}
        </Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom fontWeight="bold" sx={{ mb: 3 }}>
        Dashboard
      </Typography>
      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', lg: 'repeat(4, 1fr)' },
          gap: 3,
        }}
      >
        <StatCard
          title="Total Users"
          value={stats?.users || 0}
          icon={<PeopleIcon sx={{ fontSize: 40 }} />}
          color="primary"
        />
        <StatCard
          title="OAuth Clients"
          value={stats?.clients || 0}
          icon={<VpnKeyIcon sx={{ fontSize: 40 }} />}
          color="success"
        />
        <StatCard
          title="Active Tokens"
          value={stats?.tokens || 0}
          icon={<TokenIcon sx={{ fontSize: 40 }} />}
          color="warning"
        />
        <StatCard
          title="Recent Logins"
          value={stats?.logins || 0}
          icon={<LoginIcon sx={{ fontSize: 40 }} />}
          color="info"
        />
      </Box>
      <Paper elevation={2} sx={{ p: 3, mt: 3 }}>
        <Typography variant="h6" gutterBottom fontWeight="bold">
          System Overview
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Your OpenID Connect server is running smoothly. Use the navigation menu to manage users,
          OAuth clients, and server settings.
        </Typography>
      </Paper>
    </Box>
  );
};

export default Dashboard;
