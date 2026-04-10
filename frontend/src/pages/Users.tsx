import { useState } from 'react';
import {
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Popconfirm,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined, SearchOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from '../hooks/useApi';

interface User {
  id: string;
  username: string;
  email: string;
  name: string;
  role?: string;
  created_at: string;
}

const Users = () => {
  const [modalOpen, setModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [search, setSearch] = useState('');
  const [form] = Form.useForm();

  const { data: users = [], isLoading: loading } = useUsers();
  const createUserMutation = useCreateUser();
  const updateUserMutation = useUpdateUser();
  const deleteUserMutation = useDeleteUser();

  const filtered = search
    ? (users as User[]).filter((u) =>
        u.username.toLowerCase().includes(search.toLowerCase()) ||
        u.email.toLowerCase().includes(search.toLowerCase()) ||
        (u.name || '').toLowerCase().includes(search.toLowerCase())
      )
    : (users as User[]);

  const handleSubmit = async (values: { username: string; email: string; password: string; name: string; role: string }) => {
    try {
      if (editingUser) {
        await updateUserMutation.mutateAsync({ id: editingUser.id, ...values });
        message.success('User updated successfully');
      } else {
        await createUserMutation.mutateAsync(values);
        message.success('User created successfully');
      }
      setModalOpen(false);
      setEditingUser(null);
      form.resetFields();
    } catch (error) {
      message.error(editingUser ? 'Failed to update user' : 'Failed to create user');
      console.error('Failed to save user:', error);
    }
  };

  const handleEdit = (user: User) => {
    setEditingUser(user);
    form.setFieldsValue({ username: user.username, email: user.email, name: user.name, role: user.role || 'user' });
    setModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteUserMutation.mutateAsync(id);
      message.success('User deleted successfully');
    } catch (error) {
      message.error('Failed to delete user');
      console.error('Failed to delete user:', error);
    }
  };

  const handleCancel = () => {
    setModalOpen(false);
    setEditingUser(null);
    form.resetFields();
  };

  const columns: ColumnsType<User> = [
    {
      title: 'User',
      key: 'user',
      render: (_, record) => {
        const initial = (record.username || '?')[0].toUpperCase();
        return (
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <div style={{
              width: 32, height: 32, borderRadius: '50%',
              background: 'rgba(13,148,136,0.15)', color: '#0D9488',
              fontWeight: 600, fontSize: 13, display: 'flex',
              alignItems: 'center', justifyContent: 'center', flexShrink: 0,
            }}>
              {initial}
            </div>
            <div>
              <div style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: 'var(--text-sm)' }}>
                {record.username}
              </div>
              {record.name && (
                <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>{record.name}</div>
              )}
            </div>
          </div>
        );
      },
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      render: (email: string) => (
        <span style={{ fontSize: 'var(--text-sm)', color: 'var(--text-secondary)' }}>{email}</span>
      ),
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      width: 100,
      render: (role: string) => {
        const isAdmin = role === 'admin';
        return (
          <span style={{
            display: 'inline-block',
            background: isAdmin ? 'rgba(245,158,11,0.12)' : 'rgba(148,163,184,0.15)',
            color: isAdmin ? '#D97706' : '#64748B',
            borderRadius: 20, padding: '2px 10px',
            fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.04em',
          }}>
            {role || 'user'}
          </span>
        );
      },
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date: string) => (
        <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>
          {new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
        </span>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Space size={4}>
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
            style={{ color: 'var(--text-secondary)' }}
          />
          <Popconfirm
            title="Delete user"
            description="Are you sure you want to delete this user?"
            onConfirm={() => handleDelete(record.id)}
            okText="Delete"
            okButtonProps={{ danger: true }}
            cancelText="Cancel"
          >
            <Button
              type="text"
              size="small"
              danger
              icon={<DeleteOutlined />}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      {/* Page Header */}
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 16 }}>
        <div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-3xl)', fontWeight: 'var(--font-bold)', color: 'var(--text-primary)', lineHeight: 1.2 }}>
            User Management
          </h1>
          <p style={{ margin: '4px 0 0', fontSize: 'var(--text-sm)', color: 'var(--text-muted)' }}>
            Manage user accounts and roles
          </p>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => { setEditingUser(null); form.resetFields(); setModalOpen(true); }}
        >
          Add User
        </Button>
      </div>

      {/* Search Bar */}
      <div style={{ marginBottom: 16 }}>
        <Input
          prefix={<SearchOutlined style={{ color: 'var(--text-muted)' }} />}
          placeholder="Search by username, email, or name…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          allowClear
          style={{ maxWidth: 360 }}
        />
      </div>

      {/* Table */}
      <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', overflow: 'hidden', boxShadow: 'var(--shadow-card)' }}>
        <Table<User>
          columns={columns}
          dataSource={filtered}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10, style: { padding: '12px 16px' } }}
          style={{ borderRadius: 0 }}
          locale={{
            emptyText: (
              <div style={{ padding: '60px 0', textAlign: 'center' }}>
                <UserOutlined style={{ fontSize: 40, color: 'var(--text-muted)', marginBottom: 10, display: 'block' }} />
                <div style={{ color: 'var(--text-secondary)', fontSize: 'var(--text-sm)' }}>No users found</div>
              </div>
            ),
          }}
        />
      </div>

      {/* Create/Edit Modal */}
      <Modal
        title={editingUser ? 'Edit User' : 'Add User'}
        open={modalOpen}
        onCancel={handleCancel}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit} initialValues={{ role: 'user' }}>
          <Form.Item label="Username" name="username" rules={[{ required: true, message: 'Please enter username' }]}>
            <Input disabled={!!editingUser} />
          </Form.Item>
          <Form.Item label="Email" name="email" rules={[{ required: true }, { type: 'email' }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Name" name="name" rules={[{ required: true, message: 'Please enter name' }]}>
            <Input />
          </Form.Item>
          <Form.Item
            label="Password"
            name="password"
            rules={[{ required: !editingUser, message: 'Please enter password' }]}
            help={editingUser ? 'Leave blank to keep current password' : undefined}
          >
            <Input.Password placeholder={editingUser ? 'Leave blank to keep current' : 'Enter password'} />
          </Form.Item>
          <Form.Item label="Role" name="role">
            <Select>
              <Select.Option value="user">User</Select.Option>
              <Select.Option value="admin">Admin</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button onClick={handleCancel}>Cancel</Button>
              <Button type="primary" htmlType="submit">
                {editingUser ? 'Update' : 'Create'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default Users;
