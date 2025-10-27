import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  Select,
  Button,
  Table,
  Space,
  Typography,
  Tag,
  message,
  Empty,
} from 'antd';
import {
  SearchOutlined,
  UserOutlined,
  EyeOutlined,
  PlusOutlined,
  ClearOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Title } = Typography;

interface User {
  id: string;
  username: string;
  email: string;
  name: string;
  role?: string;
  created_at: string;
}

interface SearchParams {
  username?: string;
  email?: string;
  name?: string;
  role?: string;
}

const UserSearch = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  const handleSearch = async (values: SearchParams) => {
    try {
      setLoading(true);
      setSearched(true);

      // Build query string
      const params = new URLSearchParams();
      if (values.username) params.append('username', values.username);
      if (values.email) params.append('email', values.email);
      if (values.name) params.append('name', values.name);
      if (values.role) params.append('role', values.role);

      const queryString = params.toString();
      const url = queryString ? `/api/admin/users?${queryString}` : '/api/admin/users';

      const response = await fetch(url);
      if (!response.ok) throw new Error('Failed to search users');
      const data = await response.json();
      setUsers(data || []);
    } catch (error) {
      message.error('Failed to search users');
      console.error('Failed to search users:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleClear = () => {
    form.resetFields();
    setUsers([]);
    setSearched(false);
  };

  const columns: ColumnsType<User> = [
    {
      title: 'Username',
      dataIndex: 'username',
      key: 'username',
      render: (text) => (
        <Space>
          <UserOutlined />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => (
        <Tag color={role === 'admin' ? 'red' : 'blue'}>{role?.toUpperCase() || 'USER'}</Tag>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record) => (
        <Button
          type="link"
          icon={<EyeOutlined />}
          onClick={() => navigate(`/users/${record.id}`)}
        >
          View Details
        </Button>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={2} style={{ margin: 0 }}>User Management</Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate('/users/new')}
        >
          Create User
        </Button>
      </div>

      <Card bordered={false} style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginTop: 0 }}>Search Users</Title>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSearch}
        >
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 16 }}>
            <Form.Item label="Username" name="username">
              <Input placeholder="Search by username" />
            </Form.Item>

            <Form.Item label="Email" name="email">
              <Input placeholder="Search by email" />
            </Form.Item>

            <Form.Item label="Name" name="name">
              <Input placeholder="Search by name" />
            </Form.Item>

            <Form.Item label="Role" name="role">
              <Select placeholder="Filter by role" allowClear>
                <Select.Option value="">All Roles</Select.Option>
                <Select.Option value="admin">Admin</Select.Option>
                <Select.Option value="user">User</Select.Option>
              </Select>
            </Form.Item>
          </div>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SearchOutlined />}
                loading={loading}
              >
                Search
              </Button>
              <Button
                icon={<ClearOutlined />}
                onClick={handleClear}
              >
                Clear
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {searched && (
        <Card bordered={false}>
          {users.length === 0 ? (
            <Empty description="No users found" />
          ) : (
            <Table
              columns={columns}
              dataSource={users}
              rowKey="id"
              loading={loading}
              pagination={{ pageSize: 10 }}
            />
          )}
        </Card>
      )}
    </>
  );
};

export default UserSearch;
