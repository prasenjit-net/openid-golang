import { useState } from 'react';
import {
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  message,
  Popconfirm,
  Tag,
  Tooltip,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, AppstoreOutlined, CopyOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useClients, useCreateClient, useUpdateClient, useDeleteClient } from '../hooks/useApi';

const { TextArea } = Input;

interface Client {
  id: string;
  client_id: string;
  client_secret: string;
  name: string;
  redirect_uris: string[];
  grant_types?: string[];
  response_types?: string[];
  scope?: string;
  created_at: string;
}

const GRANT_COLORS: Record<string, string> = {
  authorization_code: 'blue',
  client_credentials: 'purple',
  refresh_token: 'cyan',
  implicit: 'orange',
};

const Clients = () => {
  const [modalOpen, setModalOpen] = useState(false);
  const [editingClient, setEditingClient] = useState<Client | null>(null);
  const [form] = Form.useForm();

  const { data: clients = [], isLoading: loading } = useClients();
  const createClientMutation = useCreateClient();
  const updateClientMutation = useUpdateClient();
  const deleteClientMutation = useDeleteClient();

  const handleSubmit = async (values: { name: string; redirect_uris: string }) => {
    try {
      const payload = {
        ...values,
        redirect_uris: values.redirect_uris.split('\n').filter((uri: string) => uri.trim()),
      };
      if (editingClient) {
        await updateClientMutation.mutateAsync({ id: editingClient.client_id, ...payload });
        message.success('Client updated successfully');
      } else {
        await createClientMutation.mutateAsync(payload);
        message.success('Client created successfully');
      }
      setModalOpen(false);
      setEditingClient(null);
      form.resetFields();
    } catch (error) {
      message.error(editingClient ? 'Failed to update client' : 'Failed to create client');
      console.error('Failed to save client:', error);
    }
  };

  const handleEdit = (client: Client) => {
    setEditingClient(client);
    form.setFieldsValue({ name: client.name, redirect_uris: client.redirect_uris?.join('\n') || '' });
    setModalOpen(true);
  };

  const handleDelete = async (clientId: string) => {
    try {
      await deleteClientMutation.mutateAsync(clientId);
      message.success('Client deleted successfully');
    } catch (error) {
      message.error('Failed to delete client');
      console.error('Failed to delete client:', error);
    }
  };

  const handleCancel = () => {
    setModalOpen(false);
    setEditingClient(null);
    form.resetFields();
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('Copied to clipboard');
  };

  const columns: ColumnsType<Client> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <span style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: 'var(--text-sm)' }}>{name}</span>
      ),
    },
    {
      title: 'Client ID',
      dataIndex: 'client_id',
      key: 'client_id',
      render: (text: string) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--text-muted)' }}>
            {text}
          </span>
          <Tooltip title="Copy">
            <Button
              type="text"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => copyToClipboard(text)}
              style={{ color: 'var(--text-muted)', padding: 2 }}
            />
          </Tooltip>
        </div>
      ),
    },
    {
      title: 'Grant Types',
      dataIndex: 'grant_types',
      key: 'grant_types',
      render: (types: string[]) => (
        <Space size={4} wrap>
          {(types || []).map((t) => (
            <Tag key={t} color={GRANT_COLORS[t] || 'default'} style={{ fontSize: 11, margin: 0 }}>{t}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 130,
      render: (date: string) => (
        <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>
          {new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
        </span>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 100,
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
            title="Delete client"
            description="Are you sure you want to delete this OAuth client?"
            onConfirm={() => handleDelete(record.client_id)}
            okText="Delete"
            okButtonProps={{ danger: true }}
            cancelText="Cancel"
          >
            <Button type="text" size="small" danger icon={<DeleteOutlined />} />
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
            OAuth Clients
          </h1>
          <p style={{ margin: '4px 0 0', fontSize: 'var(--text-sm)', color: 'var(--text-muted)' }}>
            Registered OAuth 2.0 / OpenID Connect applications
          </p>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => { setEditingClient(null); form.resetFields(); setModalOpen(true); }}
        >
          Register Client
        </Button>
      </div>

      {/* Table */}
      <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', overflow: 'hidden', boxShadow: 'var(--shadow-card)' }}>
        <Table<Client>
          columns={columns}
          dataSource={clients as Client[]}
          rowKey="client_id"
          loading={loading}
          pagination={{ pageSize: 10, style: { padding: '12px 16px' } }}
          style={{ borderRadius: 0 }}
          locale={{
            emptyText: (
              <div style={{ padding: '60px 0', textAlign: 'center' }}>
                <AppstoreOutlined style={{ fontSize: 40, color: 'var(--text-muted)', marginBottom: 10, display: 'block' }} />
                <div style={{ color: 'var(--text-secondary)', fontSize: 'var(--text-sm)' }}>No OAuth clients registered yet</div>
              </div>
            ),
          }}
        />
      </div>

      {/* Create/Edit Modal */}
      <Modal
        title={editingClient ? 'Edit OAuth Client' : 'Register OAuth Client'}
        open={modalOpen}
        onCancel={handleCancel}
        width={600}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item label="Client Name" name="name" rules={[{ required: true, message: 'Please enter client name' }]}>
            <Input placeholder="My Application" />
          </Form.Item>
          <Form.Item
            label="Redirect URIs"
            name="redirect_uris"
            rules={[{ required: true, message: 'Please enter at least one redirect URI' }]}
            help="Enter one URI per line"
          >
            <TextArea rows={4} placeholder={`http://localhost:3000/callback\nhttps://myapp.com/callback`} />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button onClick={handleCancel}>Cancel</Button>
              <Button type="primary" htmlType="submit">
                {editingClient ? 'Update' : 'Register'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default Clients;
