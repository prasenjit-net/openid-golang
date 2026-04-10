import { useState } from 'react';
import {
  Table, Button, Modal, Alert, Tag, Spin, message,
  InputNumber, Descriptions, Tooltip, Input, Space,
} from 'antd';
import {
  KeyOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  SafetyCertificateOutlined,
  InfoCircleOutlined,
  DownloadOutlined,
  CopyOutlined,
  UploadOutlined,
  FileAddOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useKeys, useRotateKeys, useGenerateCSR, useImportCert } from '../hooks/useApi';

interface CertInfo {
  subject: string;
  issuer: string;
  serial: string;
  not_before: string;
  not_after: string;
  fingerprint: string;
  self_signed: boolean;
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
  has_csr: boolean;
}

const STATUS_CFG = {
  active:   { label: 'Active',   color: '#0D9488', bg: 'rgba(13,148,136,0.1)',  icon: <CheckCircleOutlined /> },
  expired:  { label: 'Expired',  color: '#EF4444', bg: 'rgba(239,68,68,0.1)',   icon: <CloseCircleOutlined /> },
  inactive: { label: 'Inactive', color: '#94A3B8', bg: 'rgba(148,163,184,0.1)', icon: <ClockCircleOutlined /> },
};

const fmt  = (d?: string) => d ? new Date(d).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' }) : '—';
const fmtL = (d?: string) => d ? new Date(d).toLocaleString() : '—';

function downloadText(filename: string, text: string) {
  const a = document.createElement('a');
  a.href = URL.createObjectURL(new Blob([text], { type: 'text/plain' }));
  a.download = filename;
  a.click();
}

async function copyText(text: string) {
  await navigator.clipboard.writeText(text);
  message.success('Copied to clipboard');
}

const KeyManagement = () => {
  const { data: keys = [], isLoading: loading } = useKeys();
  const rotateKeysMutation = useRotateKeys();
  const generateCSRMutation = useGenerateCSR();
  const importCertMutation  = useImportCert();

  // Rotate modal
  const [rotateOpen, setRotateOpen] = useState(false);
  const [validityDays, setValidityDays] = useState<number>(90);

  // Detail modal
  const [detailKey, setDetailKey] = useState<SigningKey | null>(null);

  // CSR modal
  const [csrOpen, setCsrOpen]   = useState(false);
  const [csrKeyId, setCsrKeyId] = useState('');
  const [csrPem, setCsrPem]     = useState('');

  // Import cert modal
  const [importOpen, setImportOpen]     = useState(false);
  const [importKeyId, setImportKeyId]   = useState('');
  const [importPem, setImportPem]       = useState('');

  // ── Rotate ──────────────────────────────────────────────────────────
  const handleRotate = async () => {
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
                ['Key ID', result.new_key_id?.slice(0, 32) + '…'],
                ['Valid from', fmtL(result.not_before)],
                ['Expires',    fmtL(result.not_after)],
                ['Validity',   `${result.validity_days} days`],
              ].map(([label, val]) => (
                <tr key={label as string}>
                  <td style={{ padding: '3px 8px 3px 0', color: 'var(--text-muted)', fontWeight: 500, whiteSpace: 'nowrap' }}>{label}</td>
                  <td style={{ padding: '3px 0', fontFamily: 'monospace', fontSize: 12 }}>{val}</td>
                </tr>
              ))}
            </table>
          </div>
        ),
      });
      setRotateOpen(false);
    } catch {
      message.error('Failed to rotate keys');
    }
  };

  // ── Generate CSR ─────────────────────────────────────────────────────
  const handleGenerateCSR = async (key: SigningKey) => {
    setCsrKeyId(key.id);
    setCsrPem('');
    setCsrOpen(true);
    try {
      const result = await generateCSRMutation.mutateAsync(key.id);
      setCsrPem(result.csr_pem);
    } catch {
      message.error('Failed to generate CSR');
      setCsrOpen(false);
    }
  };

  // ── Import cert ───────────────────────────────────────────────────────
  const openImport = (key: SigningKey) => {
    setImportKeyId(key.id);
    setImportPem('');
    setImportOpen(true);
  };

  const handleImport = async () => {
    const pem = importPem.trim();
    if (!pem) { message.error('Please paste the certificate PEM'); return; }
    try {
      const result = await importCertMutation.mutateAsync({ keyId: importKeyId, certPem: pem });
      message.success('Certificate imported successfully');
      Modal.success({
        title: 'Certificate Imported',
        icon: <SafetyCertificateOutlined style={{ color: '#0D9488' }} />,
        content: (
          <Descriptions column={1} size="small" style={{ marginTop: 8 }}>
            <Descriptions.Item label="Subject">{result.cert?.subject}</Descriptions.Item>
            <Descriptions.Item label="Issuer">{result.cert?.issuer}</Descriptions.Item>
            <Descriptions.Item label="Expires">{fmtL(result.cert?.not_after)}</Descriptions.Item>
            <Descriptions.Item label="New KID">
              <span style={{ fontFamily: 'monospace', fontSize: 11, wordBreak: 'break-all' }}>{result.kid}</span>
            </Descriptions.Item>
          </Descriptions>
        ),
      });
      setImportOpen(false);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Failed to import certificate';
      message.error(msg);
    }
  };

  // ── Table columns ─────────────────────────────────────────────────────
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
        const cfg = STATUS_CFG[status] ?? STATUS_CFG.inactive;
        return (
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 6, background: cfg.bg, color: cfg.color, borderRadius: 20, padding: '2px 10px', fontSize: 12, fontWeight: 500 }}>
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
              <SafetyCertificateOutlined style={{ marginRight: 4, color: record.cert.self_signed ? '#F59E0B' : '#0D9488' }} />
              {record.cert.subject}
              {record.cert.self_signed && <Tag color="orange" style={{ marginLeft: 6, fontSize: 10 }}>Self-signed</Tag>}
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
        const d = record.cert?.not_after ?? record.expires_at;
        const expired = d ? new Date(d) < new Date() : false;
        return <span style={{ fontSize: 12, color: expired ? '#EF4444' : 'var(--text-muted)' }}>{fmt(d)}</span>;
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 220,
      render: (_: unknown, record: SigningKey) => (
        <Space size={4}>
          <Tooltip title="Certificate details">
            <Button size="small" type="text" icon={<InfoCircleOutlined />} onClick={() => setDetailKey(record)} style={{ color: '#0D9488' }} />
          </Tooltip>
          <Tooltip title={record.has_csr ? 'View / regenerate CSR' : 'Generate CSR'}>
            <Button
              size="small"
              type="text"
              icon={<FileAddOutlined />}
              onClick={() => handleGenerateCSR(record)}
              loading={generateCSRMutation.isPending && csrKeyId === record.id}
              style={{ color: '#0EA5E9' }}
            >
              CSR
            </Button>
          </Tooltip>
          <Tooltip title="Import CA-signed certificate">
            <Button size="small" type="text" icon={<UploadOutlined />} onClick={() => openImport(record)} style={{ color: '#8B5CF6' }}>
              Import Cert
            </Button>
          </Tooltip>
        </Space>
      ),
    },
  ];

  if (loading) {
    return <div style={{ display: 'flex', justifyContent: 'center', padding: 80 }}><Spin size="large" /></div>;
  }

  return (
    <>
      {/* Header */}
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 16 }}>
        <div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-3xl)', fontWeight: 'var(--font-bold)', color: 'var(--text-primary)', lineHeight: 1.2 }}>
            Signing Keys
          </h1>
          <p style={{ margin: '4px 0 0', fontSize: 'var(--text-sm)', color: 'var(--text-muted)' }}>
            RSA key pairs with X.509 certificates for JWT token signing
          </p>
        </div>
        <Button type="primary" icon={<ReloadOutlined />} onClick={() => setRotateOpen(true)} style={{ background: '#0D9488', borderColor: '#0D9488' }}>
          Rotate Key
        </Button>
      </div>

      {/* Info banner */}
      <div style={{ background: 'rgba(14,165,233,0.07)', border: '1px solid rgba(14,165,233,0.25)', borderRadius: 8, padding: '10px 16px', marginBottom: 20, fontSize: 'var(--text-sm)', color: 'var(--text-secondary)' }}>
        <strong style={{ color: '#0EA5E9' }}>ℹ</strong>
        {'  '}The <strong>active key</strong> signs new tokens. Inactive keys remain in JWKS until their certificate expires.
        Generate a <strong>CSR</strong> to get a CA-signed certificate, then <strong>Import Cert</strong> to replace the self-signed one.
        KIDs are derived from the certificate SHA-256 fingerprint (<code>x5t#S256</code>).
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
                <Button type="primary" icon={<SafetyCertificateOutlined />} onClick={() => setRotateOpen(true)} style={{ background: '#0D9488', borderColor: '#0D9488' }}>
                  Generate First Key
                </Button>
              </div>
            ),
          }}
        />
      </div>

      {/* ── Rotate Modal ──────────────────────────────────────── */}
      <Modal
        title={<div style={{ display: 'flex', alignItems: 'center', gap: 8 }}><SafetyCertificateOutlined style={{ color: '#0D9488' }} />{keys.length === 0 ? 'Generate Signing Key' : 'Rotate Signing Key'}</div>}
        open={rotateOpen}
        onCancel={() => setRotateOpen(false)}
        footer={[
          <Button key="cancel" onClick={() => setRotateOpen(false)}>Cancel</Button>,
          <Button key="ok" type="primary" icon={keys.length === 0 ? <KeyOutlined /> : <ReloadOutlined />} loading={rotateKeysMutation.isPending} onClick={handleRotate} style={{ background: '#0D9488', borderColor: '#0D9488' }}>
            {keys.length === 0 ? 'Generate' : 'Rotate'}
          </Button>,
        ]}
      >
        {keys.length > 0 && (
          <Alert message="The current active key will be deactivated. It stays in JWKS until its certificate expires so existing tokens remain valid." type="warning" showIcon style={{ marginBottom: 16 }} />
        )}
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-primary)', marginBottom: 8 }}>Certificate validity</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
            <InputNumber min={30} max={3650} value={validityDays} onChange={v => setValidityDays(v ?? 90)} addonAfter="days" style={{ width: 180 }} />
            <div style={{ display: 'flex', gap: 6 }}>
              {[{ label: '3 months', days: 90 }, { label: '6 months', days: 180 }, { label: '1 year', days: 365 }].map(p => (
                <Button key={p.days} size="small" type={validityDays === p.days ? 'primary' : 'default'} onClick={() => setValidityDays(p.days)} style={validityDays === p.days ? { background: '#0D9488', borderColor: '#0D9488' } : {}}>
                  {p.label}
                </Button>
              ))}
            </div>
          </div>
          <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 6 }}>
            Expires: <strong>{new Date(Date.now() + validityDays * 86400_000).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}</strong>
          </div>
        </div>
        <Alert message="A 2048-bit RSA key and a self-signed X.509 certificate will be generated. You can later generate a CSR and import a CA-signed certificate." type="info" showIcon />
      </Modal>

      {/* ── CSR Modal ─────────────────────────────────────────── */}
      <Modal
        title={<div style={{ display: 'flex', alignItems: 'center', gap: 8 }}><FileAddOutlined style={{ color: '#0EA5E9' }} />Certificate Signing Request</div>}
        open={csrOpen}
        onCancel={() => setCsrOpen(false)}
        width={600}
        footer={
          csrPem ? [
            <Button key="copy" icon={<CopyOutlined />} onClick={() => copyText(csrPem)}>Copy</Button>,
            <Button key="download" icon={<DownloadOutlined />} onClick={() => downloadText('signing-key.csr', csrPem)}>Download .csr</Button>,
            <Button key="close" type="primary" onClick={() => setCsrOpen(false)} style={{ background: '#0D9488', borderColor: '#0D9488' }}>Done</Button>,
          ] : null
        }
      >
        {!csrPem ? (
          <div style={{ textAlign: 'center', padding: '40px 0' }}><Spin size="large" /><div style={{ marginTop: 12, color: 'var(--text-muted)' }}>Generating CSR…</div></div>
        ) : (
          <>
            <Alert
              message="Send this CSR to your Certificate Authority. Once you receive the signed certificate, use Import Cert to replace the self-signed one."
              type="info"
              showIcon
              style={{ marginBottom: 12 }}
            />
            <Input.TextArea
              value={csrPem}
              readOnly
              rows={12}
              style={{ fontFamily: 'monospace', fontSize: 11, background: 'var(--surface-alt)', color: 'var(--text-primary)' }}
            />
          </>
        )}
      </Modal>

      {/* ── Import Cert Modal ─────────────────────────────────── */}
      <Modal
        title={<div style={{ display: 'flex', alignItems: 'center', gap: 8 }}><UploadOutlined style={{ color: '#8B5CF6' }} />Import CA-signed Certificate</div>}
        open={importOpen}
        onCancel={() => setImportOpen(false)}
        footer={[
          <Button key="cancel" onClick={() => setImportOpen(false)}>Cancel</Button>,
          <Button key="import" type="primary" icon={<SafetyCertificateOutlined />} loading={importCertMutation.isPending} onClick={handleImport} style={{ background: '#8B5CF6', borderColor: '#8B5CF6' }}>
            Import Certificate
          </Button>,
        ]}
        width={580}
      >
        <Alert
          message="The certificate must correspond to this key's private key. The KID and expiry will be updated from the new certificate."
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-primary)', marginBottom: 8 }}>
          Certificate PEM
        </div>
        <Input.TextArea
          rows={10}
          placeholder={'-----BEGIN CERTIFICATE-----\n…\n-----END CERTIFICATE-----'}
          value={importPem}
          onChange={e => setImportPem(e.target.value)}
          style={{ fontFamily: 'monospace', fontSize: 11 }}
        />
        <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 6 }}>
          Paste the full PEM-encoded certificate including the <code>-----BEGIN CERTIFICATE-----</code> header and footer.
        </div>
      </Modal>

      {/* ── Detail Modal ─────────────────────────────────────────── */}
      <Modal
        title={<div style={{ display: 'flex', alignItems: 'center', gap: 8 }}><SafetyCertificateOutlined style={{ color: '#0D9488' }} />Key & Certificate Details</div>}
        open={!!detailKey}
        onCancel={() => setDetailKey(null)}
        footer={<Button onClick={() => setDetailKey(null)}>Close</Button>}
        width={560}
      >
        {detailKey && (
          <Descriptions column={1} size="small" bordered style={{ marginTop: 8 }}>
            <Descriptions.Item label="Key ID (KID)">
              <span style={{ fontFamily: 'monospace', fontSize: 11, wordBreak: 'break-all' }}>{detailKey.kid}</span>
            </Descriptions.Item>
            <Descriptions.Item label="Algorithm">{detailKey.algorithm}</Descriptions.Item>
            <Descriptions.Item label="Status">
              {(() => { const cfg = STATUS_CFG[detailKey.status] ?? STATUS_CFG.inactive; return <span style={{ color: cfg.color, fontWeight: 600 }}>{cfg.label}</span>; })()}
            </Descriptions.Item>
            {detailKey.cert && <>
              <Descriptions.Item label="Certificate Type">
                {detailKey.cert.self_signed
                  ? <Tag color="orange">Self-signed — consider getting a CA-signed cert via CSR</Tag>
                  : <Tag color="green">CA-signed</Tag>}
              </Descriptions.Item>
              <Descriptions.Item label="Subject">{detailKey.cert.subject}</Descriptions.Item>
              <Descriptions.Item label="Issuer">{detailKey.cert.issuer}</Descriptions.Item>
              <Descriptions.Item label="Serial"><span style={{ fontFamily: 'monospace', fontSize: 11 }}>{detailKey.cert.serial}</span></Descriptions.Item>
              <Descriptions.Item label="Valid From">{fmtL(detailKey.cert.not_before)}</Descriptions.Item>
              <Descriptions.Item label="Expires">{fmtL(detailKey.cert.not_after)}</Descriptions.Item>
              <Descriptions.Item label="SHA-256 Fingerprint (x5t#S256)">
                <span style={{ fontFamily: 'monospace', fontSize: 11, wordBreak: 'break-all' }}>{detailKey.cert.fingerprint}</span>
              </Descriptions.Item>
            </>}
            <Descriptions.Item label="Created">{fmtL(detailKey.created_at)}</Descriptions.Item>
            <Descriptions.Item label="CSR Generated">{detailKey.has_csr ? <Tag color="blue">Yes</Tag> : <Tag>No</Tag>}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </>
  );
};

export default KeyManagement;
