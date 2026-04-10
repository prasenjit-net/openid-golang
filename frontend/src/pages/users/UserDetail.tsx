import { useNavigate, useParams } from 'react-router-dom';
import { Button, Space, Tag, Spin, Alert, message, Popconfirm } from 'antd';
import {
  EditOutlined,
  ArrowLeftOutlined,
  DeleteOutlined,
  UserOutlined,
  MailOutlined,
  ClockCircleOutlined,
  TagsOutlined,
  IdcardOutlined,
} from '@ant-design/icons';
import { useUser, useDeleteUser } from '../../hooks/useApi';

const InfoRow = ({ label, children }: { label: string; children: React.ReactNode }) => (
  <div style={{ display: 'flex', padding: '10px 20px', borderBottom: '1px solid var(--border-subtle)' }}>
    <span style={{ width: 160, flexShrink: 0, fontSize: 13, color: 'var(--text-secondary)', fontWeight: 500 }}>{label}</span>
    <span style={{ fontSize: 13, color: 'var(--text-primary)' }}>{children}</span>
  </div>
);

const InfoCard = ({ icon, title, children }: { icon: React.ReactNode; title: string; children: React.ReactNode }) => (
  <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', boxShadow: 'var(--shadow-card)', overflow: 'hidden' }}>
    <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 8 }}>
      <span style={{ color: 'var(--color-primary)', fontSize: 16, display: 'flex' }}>{icon}</span>
      <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>{title}</span>
    </div>
    <div style={{ padding: '4px 0' }}>{children}</div>
  </div>
);

const UserDetail = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: user, isLoading: loading, error: queryError } = useUser(id || '');
  const deleteUserMutation = useDeleteUser();

  const handleDelete = async () => {
    try {
      await deleteUserMutation.mutateAsync(id!);
      message.success('User deleted successfully');
      navigate('/users');
    } catch (error) {
      message.error('Failed to delete user');
      console.error('Failed to delete user:', error);
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '50px' }}>
        <Spin size="large" tip="Loading user details..." />
      </div>
    );
  }

  if (queryError || !user) {
    return (
      <>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/users')}
          style={{ marginBottom: 16 }}
        >
          Back to Search
        </Button>
        <Alert
          message="Error"
          description={queryError?.message || 'User not found'}
          type="error"
          showIcon
        />
      </>
    );
  }

  const initials = (user.username || '?')[0].toUpperCase();

  return (
    <>
      {/* Back button */}
      <Button
        type="link"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/users')}
        style={{ padding: 0, marginBottom: 20, color: 'var(--text-secondary)', fontWeight: 500 }}
      >
        Users
      </Button>

      {/* Page header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 28 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div style={{
            width: 56, height: 56, borderRadius: '50%',
            background: 'var(--color-primary)', display: 'flex', alignItems: 'center',
            justifyContent: 'center', fontSize: 22, fontWeight: 700, color: '#fff', flexShrink: 0,
          }}>
            {initials}
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary)' }}>{user.name || user.username}</span>
              <Tag color={user.role === 'admin' ? 'red' : 'blue'} style={{ margin: 0 }}>
                {user.role?.toUpperCase() || 'USER'}
              </Tag>
            </div>
            <div style={{ fontSize: 14, color: 'var(--text-muted)', marginTop: 2 }}>@{user.username}</div>
          </div>
        </div>
        <Space>
          <Button
            type="primary"
            icon={<EditOutlined />}
            onClick={() => navigate(`/users/${id}/edit`)}
            style={{ background: 'var(--color-primary)', borderColor: 'var(--color-primary)' }}
          >
            Edit User
          </Button>
          <Popconfirm
            title="Delete user"
            description="Are you sure you want to delete this user? This action cannot be undone."
            onConfirm={handleDelete}
            okText="Yes"
            cancelText="No"
            okButtonProps={{ danger: true }}
          >
            <Button danger icon={<DeleteOutlined />} loading={deleteUserMutation.isPending}>
              Delete User
            </Button>
          </Popconfirm>
        </Space>
      </div>

      {/* Two-column grid */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
        {/* Left column */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
          <InfoCard icon={<UserOutlined />} title="Basic Information">
            <InfoRow label="User ID"><code style={{ fontSize: 12 }}>{user.id}</code></InfoRow>
            <InfoRow label="Username">{user.username}</InfoRow>
            <InfoRow label="Full Name">{user.name || '-'}</InfoRow>
            <InfoRow label="Role">
              <Tag color={user.role === 'admin' ? 'red' : 'blue'} style={{ margin: 0 }}>
                {user.role?.toUpperCase() || 'USER'}
              </Tag>
            </InfoRow>
          </InfoCard>

          <InfoCard icon={<MailOutlined />} title="Contact">
            <InfoRow label="Email">{user.email || '-'}</InfoRow>
            <InfoRow label="Email Verified">
              {user.email_verified
                ? <Tag color="success" style={{ margin: 0 }}>Verified</Tag>
                : <Tag color="warning" style={{ margin: 0 }}>Not Verified</Tag>}
            </InfoRow>
            <InfoRow label="Phone">{user.phone_number || '-'}</InfoRow>
            <InfoRow label="Phone Verified">
              {user.phone_number_verified
                ? <Tag color="success" style={{ margin: 0 }}>Verified</Tag>
                : <Tag color="warning" style={{ margin: 0 }}>Not Verified</Tag>}
            </InfoRow>
          </InfoCard>
        </div>

        {/* Right column */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
          <InfoCard icon={<IdcardOutlined />} title="Profile Claims">
            <InfoRow label="Given Name">{user.given_name || '-'}</InfoRow>
            <InfoRow label="Family Name">{user.family_name || '-'}</InfoRow>
            <InfoRow label="Middle Name">{user.middle_name || '-'}</InfoRow>
            <InfoRow label="Nickname">{user.nickname || '-'}</InfoRow>
            <InfoRow label="Preferred Username">{user.preferred_username || '-'}</InfoRow>
            <InfoRow label="Profile URL">{user.profile || '-'}</InfoRow>
            <InfoRow label="Picture URL">{user.picture || '-'}</InfoRow>
            <InfoRow label="Website">{user.website || '-'}</InfoRow>
            <InfoRow label="Gender">{user.gender || '-'}</InfoRow>
            <InfoRow label="Birthdate">{user.birthdate || '-'}</InfoRow>
            <InfoRow label="Timezone">{user.zoneinfo || '-'}</InfoRow>
            <InfoRow label="Locale">{user.locale || '-'}</InfoRow>
            {user.address && (
              <>
                <InfoRow label="Street">{user.address.street_address || '-'}</InfoRow>
                <InfoRow label="City">{user.address.locality || '-'}</InfoRow>
                <InfoRow label="Region">{user.address.region || '-'}</InfoRow>
                <InfoRow label="Postal Code">{user.address.postal_code || '-'}</InfoRow>
                <InfoRow label="Country">{user.address.country || '-'}</InfoRow>
              </>
            )}
          </InfoCard>

          <InfoCard icon={<ClockCircleOutlined />} title="Metadata">
            <InfoRow label="Created At">{new Date(user.created_at).toLocaleString()}</InfoRow>
            <InfoRow label="Updated At">{user.updated_at ? new Date(user.updated_at).toLocaleString() : '-'}</InfoRow>
          </InfoCard>

          {user.address?.formatted && (
            <InfoCard icon={<TagsOutlined />} title="Address">
              <InfoRow label="Formatted">{user.address.formatted}</InfoRow>
            </InfoCard>
          )}
        </div>
      </div>
    </>
  );
};

export default UserDetail;
