import { useState } from 'react';
import { Table, Button, Modal, Alert, Tag, Spin, message } from 'antd';
import { KeyOutlined, ReloadOutlined, CheckCircleOutlined, CloseCircleOutlined, ClockCircleOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useKeys, useRotateKeys } from '../hooks/useApi';

interface SigningKey {
  id: string;
  kid: string;
  algorithm: string;
  is_active: boolean;
  created_at: string;
  expires_at?: string;
  status: 'active' | 'expired' | 'inactive';
}

const STATUS_CONFIG = {
  active: { label: 'Active', color: '#0D9488', bg: 'rgba(13,148,136,0.1)', icon: <CheckCircleOutlined /> },
  expired: { label: 'Expired', color: '#EF4444', bg: 'rgba(239,68,68,0.1)', icon: <CloseCircleOutlined /> },
  inactive: { label: 'Inactive', color: '#94A3B8', bg: 'rgba(148,163,184,0.1)', icon: <ClockCircleOutlined /> },
};

const formatDate = (d?: string) =>
  d ? new Date(d).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' }) : '—';

const KeyManagement = () => {
  const { data: keys = [], isLoading: loading } = useKeys();
  const rotateKeysMutation = useRotateKeys();
  const [rotateModalVisible, setRotateModalVisible] = useState(false);

  const handleRotateKeys = async () => {
    try {
      const result = await rotateKeysMutation.mutateAsync();
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
    } catch (error) {
      message.error('Failed to rotate keys');
      console.error('Failed to rotate keys:', error);
    }
  };

  const columns: ColumnsType<SigningKey> = [
    {
      title: 'Key ID',
      dataIndex: 'kid',
      key: 'kid',
      render: (kid: string) => (
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--text-primary)' }}>{kid}</span>
      ),
    },
    {
      title: 'Algorithm',
      dataIndex: 'algorithm',
      key: 'algorithm',
      width: 120,
      render: (alg: string) => (
        <Tag style={{ fontFamily: 'var(--font-mono)', fontSize: 11 }}>{alg}</Tag>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: 'active' | 'expired' | 'inactive') => {
        const cfg = STATUS_CONFIG[status] || STATUS_CONFIG.inactive;
        return (
          <div style={{
            display: 'inline-flex', alignItems: 'center', gap: 6,
            background: cfg.bg, color: cfg.color,
            borderRadius: 20, padding: '2px 10px', fontSize: 12, fontWeight: 500,
          }}>
            {cfg.icon} {cfg.label}
          </div>
        );
      },
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 140,
      render: (d: string) => (
        <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>{formatDate(d)}</span>
      ),
    },
    {
      title: 'Expires',
      dataIndex: 'expires_at',
      key: 'expires_at',
      width: 140,
      render: (d: string) => (
        <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>{formatDate(d)}</span>
      ),
    },
  ];

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '80px' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <>
      {/* Page Header */}
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 16 }}>
        <div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-3xl)', fontWeight: 'var(--font-bold)', color: 'var(--text-primary)', lineHeight: 1.2 }}>
            Signing Keys
          </h1>
          <p style={{ margin: '4px 0 0', fontSize: 'var(--text-sm)', color: 'var(--text-muted)' }}>
            RSA key pairs for JWT token signing
          </p>
        </div>
        <Button
          type="primary"
          icon={<ReloadOutlined />}
          onClick={() => setRotateModalVisible(true)}
        >
          Rotate Keys
        </Button>
      </div>

      {/* Info Alert */}
      <div style={{
        background: 'rgba(14,165,233,0.07)',
        border: '1px solid rgba(14,165,233,0.25)',
        borderRadius: 8,
        padding: '10px 16px',
        marginBottom: 20,
        fontSize: 'var(--text-sm)',
        color: 'var(--text-secondary)',
      }}>
        <strong style={{ color: '#0EA5E9' }}>ℹ</strong>
        {'  '}Active keys sign new tokens. Expired keys validate existing tokens for 30 days.
      </div>

      {/* Table */}
      <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', overflow: 'hidden', boxShadow: 'var(--shadow-card)' }}>
        <Table<SigningKey>
          columns={columns}
          dataSource={keys}
          rowKey="id"
          pagination={false}
          style={{ borderRadius: 0 }}
          locale={{
            emptyText: (
              <div style={{ padding: '60px 0', textAlign: 'center' }}>
                <KeyOutlined style={{ fontSize: 48, color: 'var(--text-muted)', marginBottom: 12, display: 'block' }} />
                <div style={{ color: 'var(--text-secondary)', fontSize: 'var(--text-sm)', marginBottom: 16 }}>
                  No signing keys found
                </div>
                <Button type="primary" icon={<ReloadOutlined />} onClick={() => setRotateModalVisible(true)}>
                  Generate First Key
                </Button>
              </div>
            ),
          }}
        />
      </div>

      {/* Rotate Keys Modal */}
      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <KeyOutlined style={{ color: '#0D9488' }} />
            <span>Rotate Signing Keys</span>
          </div>
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
            icon={<ReloadOutlined />}
            loading={rotateKeysMutation.isPending}
            onClick={handleRotateKeys}
          >
            Rotate Keys
          </Button>,
        ]}
      >
        <Alert
          message="A new RSA key pair will be generated and the current active key will be marked inactive with a 30-day expiration. Existing tokens remain valid during this grace period."
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <p>Are you sure you want to rotate the signing keys?</p>
      </Modal>
    </>
  );
};

export default KeyManagement;
