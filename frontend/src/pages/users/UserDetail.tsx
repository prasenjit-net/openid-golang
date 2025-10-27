import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Typography,
  Tag,
  Spin,
  Alert,
  message,
  Popconfirm,
} from 'antd';
import {
  EditOutlined,
  ArrowLeftOutlined,
  DeleteOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useUser, useDeleteUser } from '../../hooks/useApi';

const { Title } = Typography;

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

  return (
    <>
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/users')}
          >
            Back
          </Button>
          <Title level={2} style={{ margin: 0 }}>User Details</Title>
        </Space>
        <Space>
          <Button
            type="primary"
            icon={<EditOutlined />}
            onClick={() => navigate(`/users/${id}/edit`)}
          >
            Edit
          </Button>
          <Popconfirm
            title="Delete user"
            description="Are you sure you want to delete this user? This action cannot be undone."
            onConfirm={handleDelete}
            okText="Yes"
            cancelText="No"
            okButtonProps={{ danger: true }}
          >
            <Button danger icon={<DeleteOutlined />}>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      </div>

      {/* Basic Information */}
      <Card bordered={false} style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginTop: 0 }}>
          <UserOutlined /> Basic Information
        </Title>
        <Descriptions bordered column={2}>
          <Descriptions.Item label="User ID" span={2}>
            <code>{user.id}</code>
          </Descriptions.Item>
          <Descriptions.Item label="Username">{user.username}</Descriptions.Item>
          <Descriptions.Item label="Role">
            <Tag color={user.role === 'admin' ? 'red' : 'blue'}>
              {user.role?.toUpperCase() || 'USER'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Name" span={2}>{user.name || '-'}</Descriptions.Item>
          <Descriptions.Item label="Email">{user.email || '-'}</Descriptions.Item>
          <Descriptions.Item label="Email Verified">
            {user.email_verified ? (
              <Tag color="success">Verified</Tag>
            ) : (
              <Tag color="warning">Not Verified</Tag>
            )}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Profile Claims */}
      <Card bordered={false} style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginTop: 0 }}>Profile Claims</Title>
        <Descriptions bordered column={2}>
          <Descriptions.Item label="Given Name">{user.given_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="Family Name">{user.family_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="Middle Name">{user.middle_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="Nickname">{user.nickname || '-'}</Descriptions.Item>
          <Descriptions.Item label="Preferred Username">{user.preferred_username || '-'}</Descriptions.Item>
          <Descriptions.Item label="Profile URL">{user.profile || '-'}</Descriptions.Item>
          <Descriptions.Item label="Picture URL">{user.picture || '-'}</Descriptions.Item>
          <Descriptions.Item label="Website">{user.website || '-'}</Descriptions.Item>
          <Descriptions.Item label="Gender">{user.gender || '-'}</Descriptions.Item>
          <Descriptions.Item label="Birthdate">{user.birthdate || '-'}</Descriptions.Item>
          <Descriptions.Item label="Timezone">{user.zoneinfo || '-'}</Descriptions.Item>
          <Descriptions.Item label="Locale">{user.locale || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Contact Information */}
      <Card bordered={false} style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginTop: 0 }}>Contact Information</Title>
        <Descriptions bordered column={2}>
          <Descriptions.Item label="Phone Number">{user.phone_number || '-'}</Descriptions.Item>
          <Descriptions.Item label="Phone Verified">
            {user.phone_number_verified ? (
              <Tag color="success">Verified</Tag>
            ) : (
              <Tag color="warning">Not Verified</Tag>
            )}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Address */}
      {user.address && (
        <Card bordered={false} style={{ marginBottom: 24 }}>
          <Title level={4} style={{ marginTop: 0 }}>Address</Title>
          <Descriptions bordered column={1}>
            <Descriptions.Item label="Formatted">{user.address.formatted || '-'}</Descriptions.Item>
            <Descriptions.Item label="Street Address">{user.address.street_address || '-'}</Descriptions.Item>
            <Descriptions.Item label="Locality">{user.address.locality || '-'}</Descriptions.Item>
            <Descriptions.Item label="Region">{user.address.region || '-'}</Descriptions.Item>
            <Descriptions.Item label="Postal Code">{user.address.postal_code || '-'}</Descriptions.Item>
            <Descriptions.Item label="Country">{user.address.country || '-'}</Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      {/* Metadata */}
      <Card bordered={false}>
        <Title level={4} style={{ marginTop: 0 }}>Metadata</Title>
        <Descriptions bordered column={2}>
          <Descriptions.Item label="Created At">
            {new Date(user.created_at).toLocaleString()}
          </Descriptions.Item>
          <Descriptions.Item label="Updated At">
            {user.updated_at ? new Date(user.updated_at).toLocaleString() : '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </>
  );
};

export default UserDetail;
