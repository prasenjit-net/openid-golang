import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Card,
  Form,
  Input,
  Button,
  Table,
  Space,
  Typography,
  Tag,
  message,
  Empty,
  Tooltip,
} from 'antd';
import {
  SearchOutlined,
  LockOutlined,
  EyeOutlined,
  PlusOutlined,
  ClearOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Title, Text } = Typography;

interface Client {
  id: string;
  client_id: string;
  name: string;
  redirect_uris: string[];
  grant_types?: string[];
  created_at: string;
}

interface SearchParams {
  client_id?: string;
  name?: string;
}

const ClientSearch = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useState<SearchParams | null>(null);

  const { data: clients = [], isLoading: loading } = useQuery({
    queryKey: ['clients', 'search', searchParams],
    queryFn: async () => {
      if (!searchParams) return [];
      
      const params = new URLSearchParams();
      if (searchParams.client_id) params.append('client_id', searchParams.client_id);
      if (searchParams.name) params.append('name', searchParams.name);

      const queryString = params.toString();
      const url = queryString ? `/api/admin/clients?${queryString}` : '/api/admin/clients';

      const response = await fetch(url);
      if (!response.ok) throw new Error('Failed to search clients');
      const data = await response.json();
      return data || [];
    },
    enabled: !!searchParams,
  });

  const handleSearch = (values: SearchParams) => {
    setSearchParams(values);
  };

  const handleClear = () => {
    form.resetFields();
    setSearchParams(null);
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
      title: 'Grant Types',
      dataIndex: 'grant_types',
      key: 'grant_types',
      render: (types: string[]) => (
        <Space wrap>
          {types?.slice(0, 2).map((type, i) => (
            <Tag key={i} color="blue">{type}</Tag>
          ))}
          {types?.length > 2 && <Text type="secondary">+{types.length - 2} more</Text>}
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
        <Button
          type="link"
          icon={<EyeOutlined />}
          onClick={() => navigate(`/clients/${record.client_id}`)}
        >
          View Details
        </Button>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={2} style={{ margin: 0 }}>Client Management</Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate('/clients/new')}
        >
          Create Client
        </Button>
      </div>

      <Card bordered={false} style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginTop: 0 }}>Search OAuth Clients</Title>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSearch}
        >
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: 16 }}>
            <Form.Item label="Client ID" name="client_id">
              <Input placeholder="Search by client ID" />
            </Form.Item>

            <Form.Item label="Client Name" name="name">
              <Input placeholder="Search by name" />
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

      {searchParams && (
        <Card bordered={false}>
          {clients.length === 0 ? (
            <Empty description="No clients found" />
          ) : (
            <Table
              columns={columns}
              dataSource={clients}
              rowKey="client_id"
              loading={loading}
              pagination={{ pageSize: 10 }}
            />
          )}
        </Card>
      )}
    </>
  );
};

export default ClientSearch;
