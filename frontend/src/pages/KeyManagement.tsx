import { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Typography,
  message,
  Modal,
  Alert,
  Tag,
  Spin,
} from 'antd';
import {
  KeyOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Title, Text } = Typography;

interface SigningKey {
  id: string;
  kid: string;
  algorithm: string;
  is_active: boolean;
  created_at: string;
  expires_at?: string;
  status: 'active' | 'expired' | 'inactive';
}

const KeyManagement = () => {
  const [keys, setKeys] = useState<SigningKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [rotating, setRotating] = useState(false);
  const [rotateModalVisible, setRotateModalVisible] = useState(false);

  const fetchKeys = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/admin/keys');
      if (!response.ok) throw new Error('Failed to fetch keys');
      const data = await response.json();
      setKeys(data);
    } catch (error) {
      message.error('Failed to load signing keys');
      console.error('Failed to fetch keys:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchKeys();
  }, []);

  const handleRotateKeys = async () => {
    try {
      setRotating(true);
      const response = await fetch('/api/admin/settings/rotate-keys', {
        method: 'POST',
      });
      if (!response.ok) throw new Error('Failed to rotate keys');
      
      const result = await response.json();
      message.success('RSA keys rotated successfully');
      Modal.info({
        title: 'Key Rotation Complete',
        content: (
          <div>
            <p>{result.message}</p>
            <p>{result.info}</p>
            <p><strong>New Key ID:</strong> {result.new_key_id}</p>
          </div>
        ),
      });
      setRotateModalVisible(false);
      
      // Refresh keys list
      await fetchKeys();
    } catch (error) {
      message.error('Failed to rotate keys');
      console.error('Failed to rotate keys:', error);
    } finally {
      setRotating(false);
    }
  };

  const getStatusTag = (status: string) => {
    switch (status) {
      case 'active':
        return <Tag icon={<CheckCircleOutlined />} color="success">Active</Tag>;
      case 'expired':
        return <Tag icon={<CloseCircleOutlined />} color="error">Expired</Tag>;
      case 'inactive':
        return <Tag icon={<ClockCircleOutlined />} color="warning">Inactive</Tag>;
      default:
        return <Tag>{status}</Tag>;
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleString();
  };

  const columns: ColumnsType<SigningKey> = [
    {
      title: 'Key ID',
      dataIndex: 'kid',
      key: 'kid',
      render: (kid: string) => <Text code>{kid}</Text>,
    },
    {
      title: 'Algorithm',
      dataIndex: 'algorithm',
      key: 'algorithm',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: 'Created At',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => formatDate(date),
    },
    {
      title: 'Expires At',
      dataIndex: 'expires_at',
      key: 'expires_at',
      render: (date: string) => formatDate(date),
    },
  ];

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <>
      <div style={{ marginBottom: 24 }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Title level={2} style={{ margin: 0 }}>
            <KeyOutlined /> Signing Keys
          </Title>
          <Button
            type="primary"
            danger
            icon={<ReloadOutlined />}
            onClick={() => setRotateModalVisible(true)}
          >
            Rotate Keys
          </Button>
        </Space>
      </div>

      <Alert
        message="Key Rotation Strategy"
        description="When you rotate keys, the old active key is marked as inactive and will expire in 30 days. This allows existing tokens to be validated during the grace period. Only the newest active key is used for signing new tokens."
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Card bordered={false}>
        <Table
          columns={columns}
          dataSource={keys}
          rowKey="id"
          pagination={false}
          locale={{
            emptyText: (
              <div style={{ padding: '40px' }}>
                <KeyOutlined style={{ fontSize: 48, color: '#ccc', marginBottom: 16 }} />
                <p>No signing keys found</p>
                <Button type="primary" icon={<ReloadOutlined />} onClick={() => setRotateModalVisible(true)}>
                  Generate First Key
                </Button>
              </div>
            ),
          }}
        />
      </Card>

      {/* Rotate Keys Modal */}
      <Modal
        title={
          <Space>
            <KeyOutlined />
            <span>Rotate Signing Keys</span>
          </Space>
        }
        open={rotateModalVisible}
        onCancel={() => setRotateModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setRotateModalVisible(false)}>
            Cancel
          </Button>,
          <Button
            key="rotate"
            type="primary"
            danger
            icon={<ReloadOutlined />}
            loading={rotating}
            onClick={handleRotateKeys}
          >
            Rotate Keys
          </Button>,
        ]}
      >
        <Alert
          message="Important"
          description="Rotating signing keys will generate a new RSA key pair and mark the current active key as inactive."
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <p><strong>What happens when you rotate keys:</strong></p>
        <ul>
          <li>A new RSA key pair is generated and marked as active</li>
          <li>The old active key is marked as inactive with a 30-day expiration</li>
          <li>New tokens will be signed with the new key</li>
          <li>Existing tokens remain valid and can be verified with the old key</li>
          <li>After 30 days, the old key will expire and tokens signed with it will be rejected</li>
        </ul>
        <p>Are you sure you want to rotate the signing keys?</p>
      </Modal>
    </>
  );
};

export default KeyManagement;
