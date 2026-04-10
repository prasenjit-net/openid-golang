import { useState } from 'react';
import { Table, Button, Modal, Alert, Tag, Spin, message, InputNumber, Descriptions, Tooltip } from 'antd';
import {
  KeyOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  SafetyCertificateOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useKeys, useRotateKeys } from '../hooks/useApi';

interface CertInfo {
  subject: string;
  issuer: string;
  serial: string;
  not_before: string;
  not_after: string;
  fingerprint: string; // x5t#S256
}

interface SigningKey {
  id: string;
  kid: string;
  algorithm: string;
  is_active: boolean;
  created_at: string;
  expires_at?: string;
  status: 'active' | 'expired' | 'inactive';
  cert?: CertInfo;
}

const STATUS_CONFIG = {
  active:   { label: 'Active',   color: '#0D9488', bg: 'rgba(13,148,136,0.1)',  icon: <CheckCircleOutlined /> },
  expired:  { label: 'Expired',  color: '#EF4444', bg: 'rgba(239,68,68,0.1)',   icon: <CloseCircleOutlined /> },
  inactive: { label: 'Inactive', color: '#94A3B8', bg: 'rgba(148,163,184,0.1)', icon: <ClockCircleOutlined /> },
};

const fmt = (d?: string) =>
  d ? new Date(d).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' }) : '—';

const fmtFull = (d?: string) =>
  d ? new Date(d).toLocaleString() : '—';

const truncate = (s: string, n = 24) => s.length > n ? s.slice(0, n) + '…' : s;

const KeyManagement = () => {
  const { data: keys = [], isLoading: loading } = useKeys();
  const rotateKeysMutation = useRotateKeys();
  const [rotateModalVisible, setRotateModalVisible] = useState(false);
  const [validityDays, setValidityDays] = useState<number>(90);
  const [expandedKey, setExpandedKey] = useState<SigningKey | null>(null);

  const handleRotateKeys = async () => {
    try {
      const result = await rotateKeysMutation.mutateAsync(validityDays);
      message.success('RSA key rotated successfully');
      Modal.success({
        title: 'Key Rotation Complete',
        icon: <SafetyCertificateOutlined style={{ color: '#0D9488' }} />,
        content: (
          <div style={{ fontSize: 13, lineHeight: 1.7 }}>
            <p style={{ margin: '8px 0 4px', color: 'var(--text-secondary)' }}>{result.info}</p>
            <table style={{ width: '100%', borderCollapse: 'collapse', marginTop: 12 }}>
              {[
                ['Key ID (KID)', truncate(result.new_key_id, 32)],
                ['Valid from', fmtFull(result.not_before)],
                ['Expires',    fmtFull(result.not_after)],
                ['Validity',   `${result.validity_days} days`],
              ].map(([label, val]) => (
                <tr key={label}>
                  <td style={{ padding: '3px 8px 3px 0', color: 'var(--text-muted)', fontWeight: 500, whiteSpace: 'nowrap' }}>{label}</td>
                  <td style={{ padding: '3px 0', fontFamily: 'monospace', fontSize: 12 }}>{val}</td>
                </tr>
              ))}
            </table>
          </div>
        ),
      });
      setRotateModalVisible(false);
    } catch {
      message.error('Failed to rotate keys');
    }
  };

  const columns: ColumnsType<SigningKey> = [
    {
      title: 'Key ID (KID)',
      dataIndex: 'kid',
      key: 'kid',
      render: (kid: string) => (
        <Tooltip title={kid} placement="topLeft">
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--text-primary)', cursor: 'default' }}>
            {kid.length > 22 ? kid.slice(0, 22) + '…' : kid}
          </span>
        </Tooltip>
      ),
    },
    {
      title: 'Algorithm',
      dataIndex: 'algorithm',
      key: 'algorithm',
      width: 100,
      render: (alg: string) => <Tag style={{ fontFamily: 'var(--font-mono)', fontSize: 11 }}>{alg}</Tag>,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (status: 'active' | 'expired' | 'inactive') => {
        const cfg = STATUS_CONFIG[status] ?? STATUS_CONFIG.inactive;
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
      title: 'Certificate',
      key: 'cert',
      render: (_: unknown, record: SigningKey) => {
        if (!record.cert) return <span style={{ color: 'var(--text-muted)', fontSize: 12 }}>—</span>;
        return (
          <div style={{ fontSize: 12 }}>
            <div style={{ color: 'var(--text-secondary)', marginBottom: 2 }}>
              <SafetyCertificateOutlined style={{ marginRight: 4, color: '#0D9488' }} />
              {record.cert.subject}
            </div>
            <Tooltip title={record.cert.fingerprint} placement="bottomLeft">
              <div style={{ fontFamily: 'monospace', color: 'var(--text-muted)', fontSize: 11, cursor: 'default' }}>
                SHA-256: {record.cert.fingerprint.slice(0, 16)}…
              </div>
            </Tooltip>
          </div>
        );
      },
    },
    {
      title: 'Valid From',
      key: 'not_before',
      width: 130,
      render: (_: unknown, record: SigningKey) => (
        <span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>
          {record.cert ? fmt(record.cert.not_before) : fmt(record.created_at)}
        </span>
      ),
    },
    {
      title: 'Expires',
      key: 'expires',
      width: 130,
      render: (_: unknown, record: SigningKey) => {
        const expiryDate = record.cert?.not_after ?? record.expires_at;
        const expired = expiryDate ? new Date(expiryDate) < new Date() : false;
        return (
          <span style={{ fontSize: 12, color: expired ? '#EF4444' : 'var(--text-muted)' }}>
            {fmt(expiryDate)}
          </span>
        );
      },
    },
    {
      title: '',
      key: 'detail',
      width: 80,
      render: (_: unknown, record: SigningKey) =>
        record.cert ? (
          <Button
            size="small"
            type="text"
            icon={<InfoCircleOutlined />}
            onClick={() => setExpandedKey(record)}
            style={{ color: '#0D9488' }}
          >
            Details
          </Button>
        ) : null,
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
            RSA key pairs with X.509 certificates for JWT token signing
          </p>
        </div>
        <Button
          type="primary"
          icon={<ReloadOutlined />}
          onClick={() => setRotateModalVisible(true)}
          style={{ background: '#0D9488', borderColor: '#0D9488' }}
        >
          Rotate Key
        </Button>
      </div>

      {/* Info banner */}
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
        {'  '}The <strong>active key</strong> signs new tokens. Inactive keys remain in JWKS until their certificate expires, allowing verification of already-issued tokens during rotation.
        KIDs are derived from the certificate's SHA-256 fingerprint (RFC 7517 <code>x5t#S256</code>).
      </div>

      {/* Table */}
      <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', boxShadow: 'var(--shadow-card)' }}>
        <Table<SigningKey>
          columns={columns}
          dataSource={keys}
          rowKey="id"
          pagination={false}
          size="middle"
          locale={{
            emptyText: (
              <div style={{ padding: '60px 0', textAlign: 'center' }}>
                <KeyOutlined style={{ fontSize: 48, color: 'var(--text-muted)', marginBottom: 12, display: 'block' }} />
                <div style={{ color: 'var(--text-secondary)', fontSize: 'var(--text-sm)', marginBottom: 16 }}>
                  No signing keys found. Generate the first key to enable JWT signing.
                </div>
                <Button
                  type="primary"
                  icon={<SafetyCertificateOutlined />}
                  onClick={() => setRotateModalVisible(true)}
                  style={{ background: '#0D9488', borderColor: '#0D9488' }}
                >
                  Generate First Key
                </Button>
              </div>
            ),
          }}
        />
      </div>

      {/* Rotate / Generate Key Modal */}
      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <SafetyCertificateOutlined style={{ color: '#0D9488' }} />
            <span>{keys.length === 0 ? 'Generate Signing Key' : 'Rotate Signing Key'}</span>
          </div>
        }
        open={rotateModalVisible}
        onCancel={() => setRotateModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setRotateModalVisible(false)}>Cancel</Button>,
          <Button
            key="rotate"
            type="primary"
            icon={keys.length === 0 ? <KeyOutlined /> : <ReloadOutlined />}
            loading={rotateKeysMutation.isPending}
            onClick={handleRotateKeys}
            style={{ background: '#0D9488', borderColor: '#0D9488' }}
          >
            {keys.length === 0 ? 'Generate' : 'Rotate'}
          </Button>,
        ]}
      >
        {keys.length > 0 && (
          <Alert
            message="The current active key will be deactivated. It stays in JWKS until its certificate expires so existing tokens remain valid."
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-primary)', marginBottom: 8 }}>
            Certificate validity
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <InputNumber
              min={30}
              max={3650}
              value={validityDays}
              onChange={v => setValidityDays(v ?? 90)}
              addonAfter="days"
              style={{ width: 180 }}
            />
            <div style={{ display: 'flex', gap: 6 }}>
              {[{ label: '3 months', days: 90 }, { label: '6 months', days: 180 }, { label: '1 year', days: 365 }].map(p => (
                <Button
                  key={p.days}
                  size="small"
                  type={validityDays === p.days ? 'primary' : 'default'}
                  onClick={() => setValidityDays(p.days)}
                  style={validityDays === p.days ? { background: '#0D9488', borderColor: '#0D9488' } : {}}
                >
                  {p.label}
                </Button>
              ))}
            </div>
          </div>
          <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 6 }}>
            Expires: <strong>{new Date(Date.now() + validityDays * 86400_000).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}</strong>
          </div>
        </div>

        <Alert
          message="A 2048-bit RSA key and a self-signed X.509 certificate will be generated. The KID is derived from the certificate's SHA-256 fingerprint."
          type="info"
          showIcon
        />
      </Modal>

      {/* Cert Detail Modal */}
      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <SafetyCertificateOutlined style={{ color: '#0D9488' }} />
            <span>Certificate Details</span>
          </div>
        }
        open={!!expandedKey}
        onCancel={() => setExpandedKey(null)}
        footer={<Button onClick={() => setExpandedKey(null)}>Close</Button>}
        width={560}
      >
        {expandedKey && (
          <Descriptions column={1} size="small" bordered style={{ marginTop: 8 }}>
            <Descriptions.Item label="Key ID (KID)">
              <span style={{ fontFamily: 'monospace', fontSize: 11, wordBreak: 'break-all' }}>{expandedKey.kid}</span>
            </Descriptions.Item>
            <Descriptions.Item label="Algorithm">{expandedKey.algorithm}</Descriptions.Item>
            <Descriptions.Item label="Status">
              {(() => {
                const cfg = STATUS_CONFIG[expandedKey.status] ?? STATUS_CONFIG.inactive;
                return (
                  <span style={{ color: cfg.color, fontWeight: 600 }}>
                    {cfg.label}
                  </span>
                );
              })()}
            </Descriptions.Item>
            {expandedKey.cert && <>
              <Descriptions.Item label="Subject">{expandedKey.cert.subject}</Descriptions.Item>
              <Descriptions.Item label="Issuer">{expandedKey.cert.issuer}</Descriptions.Item>
              <Descriptions.Item label="Serial">
                <span style={{ fontFamily: 'monospace', fontSize: 11 }}>{expandedKey.cert.serial}</span>
              </Descriptions.Item>
              <Descriptions.Item label="Valid From">{fmtFull(expandedKey.cert.not_before)}</Descriptions.Item>
              <Descriptions.Item label="Expires">{fmtFull(expandedKey.cert.not_after)}</Descriptions.Item>
              <Descriptions.Item label="SHA-256 Fingerprint (x5t#S256)">
                <span style={{ fontFamily: 'monospace', fontSize: 11, wordBreak: 'break-all' }}>{expandedKey.cert.fingerprint}</span>
              </Descriptions.Item>
            </>}
            <Descriptions.Item label="Created">{fmtFull(expandedKey.created_at)}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </>
  );
};

export default KeyManagement;
