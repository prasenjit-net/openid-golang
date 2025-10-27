import { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  message,
  Popconfirm,
  Typography,
  Card,
  Tag,
  Tooltip,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, LockOutlined, CopyOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Title, Text } = Typography;
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

const Clients = () => {
  const [clients, setClients] = useState<Client[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingClient, setEditingClient] = useState<Client | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    fetchClients();
  }, []);

  const fetchClients = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/admin/clients');
      if (!response.ok) throw new Error('Failed to fetch clients');
      const data = await response.json();
      setClients(data || []);
    } catch (error) {
      message.error('Failed to fetch OAuth clients');
      console.error('Failed to fetch clients:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      const payload = {
        ...values,
        redirect_uris: values.redirect_uris.split('\n').filter((uri: string) => uri.trim()),
      };

      if (editingClient) {
        const response = await fetch(`/api/admin/clients/${editingClient.client_id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
        if (!response.ok) throw new Error('Failed to update client');
        message.success('Client updated successfully');
      } else {
        const response = await fetch('/api/admin/clients', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
        if (!response.ok) throw new Error('Failed to create client');
        message.success('Client created successfully');
      }
      setModalOpen(false);
      setEditingClient(null);
      form.resetFields();
      fetchClients();
    } catch (error) {
      message.error(editingClient ? 'Failed to update client' : 'Failed to create client');
      console.error('Failed to save client:', error);
    }
  };

  const handleEdit = (client: Client) => {
    setEditingClient(client);
    form.setFieldsValue({
      name: client.name,
      redirect_uris: client.redirect_uris?.join('\n') || '',
    });
    setModalOpen(true);
  };

  const handleDelete = async (clientId: string) => {
    try {
      const response = await fetch(`/api/admin/clients/${clientId}`, { method: 'DELETE' });
      if (!response.ok) throw new Error('Failed to delete client');
      message.success('Client deleted successfully');
      fetchClients();
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
      title: 'Client ID',
      dataIndex: 'client_id',
      key: 'client_id',
      render: (text) => (
        <Space>
          <LockOutlined />
          <Text code>{text}</Text>
          <Tooltip title="Copy">
            <Button
              type="text"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => copyToClipboard(text)}
            />
          </Tooltip>
        </Space>
      ),
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Redirect URIs',
      dataIndex: 'redirect_uris',
      key: 'redirect_uris',
      render: (uris: string[]) => (
        <Space direction="vertical" size="small">
          {uris?.slice(0, 2).map((uri, i) => (
            <Tag key={i}>{uri}</Tag>
          ))}
          {uris?.length > 2 && <Text type="secondary">+{uris.length - 2} more</Text>}
        </Space>
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
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            Edit
          </Button>
          <Popconfirm
            title="Delete client"
            description="Are you sure you want to delete this OAuth client?"
            onConfirm={() => handleDelete(record.client_id)}
            okText="Yes"
            cancelText="No"
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={2}>OAuth Client Management</Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => {
            setEditingClient(null);
            form.resetFields();
            setModalOpen(true);
          }}
        >
          Add Client
        </Button>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={clients}
          rowKey="client_id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title={editingClient ? 'Edit OAuth Client' : 'Create OAuth Client'}
        open={modalOpen}
        onCancel={handleCancel}
        width={600}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            label="Client Name"
            name="name"
            rules={[{ required: true, message: 'Please enter client name' }]}
          >
            <Input placeholder="My Application" />
          </Form.Item>

          <Form.Item
            label="Redirect URIs"
            name="redirect_uris"
            rules={[{ required: true, message: 'Please enter at least one redirect URI' }]}
            help="Enter one URI per line"
          >
            <TextArea
              rows={4}
              placeholder="http://localhost:3000/callback&#10;https://myapp.com/callback"
            />
          </Form.Item>

          <Form.Item>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button onClick={handleCancel}>Cancel</Button>
              <Button type="primary" htmlType="submit">
                {editingClient ? 'Update' : 'Create'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Clients;
