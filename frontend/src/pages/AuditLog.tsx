import { useState } from 'react';
import { Table, Tag, Select, Input, Button, Tooltip } from 'antd';
import { SearchOutlined, AuditOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useAuditLogs } from '../hooks/useApi';

interface AuditEntry {
  id: string;
  timestamp: string;
  action: string;
  actor_type: string;
  actor: string;
  resource: string;
  resource_id: string;
  status: string;
  ip_address: string;
  user_agent: string;
  details?: Record<string, unknown>;
}

const ACTION_COLORS: Record<string, string> = {
  'user.login': 'blue',
  'user.login_failed': 'red',
  'user.consent_granted': 'green',
  'user.consent_denied': 'orange',
  'token.issued': 'cyan',
  'token.revoked': 'volcano',
  'client.registered': 'purple',
  'admin.login': 'geekblue',
  'admin.user.created': 'green',
  'admin.user.updated': 'gold',
  'admin.user.deleted': 'red',
  'admin.client.created': 'green',
  'admin.client.updated': 'gold',
  'admin.client.deleted': 'red',
  'admin.settings.updated': 'magenta',
  'admin.keys.rotated': 'lime',
  'admin.password.changed': 'orange',
};

const ALL_ACTIONS = Object.keys(ACTION_COLORS);
const PAGE_SIZE = 50;

function formatTimestamp(ts: string) {
  const d = new Date(ts);
  return d.toLocaleString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
    hour: '2-digit', minute: '2-digit', hour12: false,
  }).replace(',', ' ·');
}

export default function AuditLog() {
  const [page, setPage] = useState(1);
  const [actionFilter, setActionFilter] = useState<string | undefined>(undefined);
  const [actorInput, setActorInput] = useState('');
  const [actorFilter, setActorFilter] = useState('');

  const offset = (page - 1) * PAGE_SIZE;
  const { data, isLoading } = useAuditLogs({
    limit: PAGE_SIZE,
    offset,
    action: actionFilter,
    actor: actorFilter || undefined,
  });

  const hasFilters = !!actionFilter || !!actorFilter;

  const clearFilters = () => {
    setActionFilter(undefined);
    setActorInput('');
    setActorFilter('');
    setPage(1);
  };

  const columns: ColumnsType<AuditEntry> = [
    {
      title: 'Timestamp',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 200,
      render: (ts: string) => (
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--text-muted)' }}>
          {formatTimestamp(ts)}
        </span>
      ),
      sorter: (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
      defaultSortOrder: 'descend',
    },
    {
      title: 'Action',
      dataIndex: 'action',
      key: 'action',
      width: 210,
      render: (action: string) => (
        <Tag color={ACTION_COLORS[action] || 'default'} style={{ fontSize: 11 }}>{action}</Tag>
      ),
    },
    {
      title: 'Actor',
      dataIndex: 'actor',
      key: 'actor',
      width: 160,
      render: (actor: string, record) => {
        const initial = (actor || '?')[0].toUpperCase();
        return (
          <Tooltip title={`Type: ${record.actor_type}`}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <div style={{
                width: 28, height: 28, borderRadius: '50%',
                background: 'rgba(13,148,136,0.15)',
                color: '#0D9488', fontWeight: 600, fontSize: 12,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                flexShrink: 0,
              }}>
                {initial}
              </div>
              <span style={{ fontSize: 13, color: 'var(--text-primary)' }}>{actor || '—'}</span>
            </div>
          </Tooltip>
        );
      },
    },
    {
      title: 'Resource',
      key: 'resource',
      width: 220,
      render: (_: unknown, record) =>
        record.resource ? (
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <Tag style={{ margin: 0 }}>{record.resource}</Tag>
            <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--text-muted)' }}>
              {record.resource_id?.slice(0, 18)}{record.resource_id?.length > 18 ? '…' : ''}
            </span>
          </div>
        ) : '—',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const isSuccess = status === 'success';
        return (
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <span style={{
              width: 8, height: 8, borderRadius: '50%',
              background: isSuccess ? '#10B981' : '#EF4444',
              display: 'inline-block',
            }} />
            <span style={{ fontSize: 13, color: isSuccess ? '#10B981' : '#EF4444', fontWeight: 500 }}>
              {status}
            </span>
          </div>
        );
      },
    },
    {
      title: 'IP Address',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 130,
      responsive: ['md'] as ('md')[],
      render: (ip: string) => (
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--text-muted)' }}>
          {ip || '—'}
        </span>
      ),
    },
  ];

  const emptyState = (
    <div style={{ padding: '60px 0', textAlign: 'center' }}>
      <AuditOutlined style={{ fontSize: 48, color: 'var(--text-muted)', marginBottom: 12, display: 'block' }} />
      <div style={{ color: 'var(--text-secondary)', fontSize: 'var(--text-sm)' }}>No audit events found</div>
      {hasFilters && (
        <Button type="link" onClick={clearFilters} style={{ marginTop: 8 }}>
          Clear filters
        </Button>
      )}
    </div>
  );

  return (
    <>
      {/* Page Header */}
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 16 }}>
        <div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-3xl)', fontWeight: 'var(--font-bold)', color: 'var(--text-primary)', lineHeight: 1.2 }}>
            Audit Log
          </h1>
          <p style={{ margin: '4px 0 0', fontSize: 'var(--text-sm)', color: 'var(--text-muted)' }}>
            Security events and admin actions
          </p>
        </div>
      </div>

      {/* Filters Bar */}
      <div style={{
        display: 'flex', flexWrap: 'wrap', gap: 10, alignItems: 'center',
        marginBottom: 16,
        background: 'var(--surface)',
        border: '1px solid var(--border)',
        borderRadius: 10,
        padding: '12px 16px',
        boxShadow: 'var(--shadow-card)',
      }}>
        <Input
          prefix={<SearchOutlined style={{ color: 'var(--text-muted)' }} />}
          placeholder="Search by actor…"
          value={actorInput}
          onChange={(e) => setActorInput(e.target.value)}
          onPressEnter={() => { setActorFilter(actorInput); setPage(1); }}
          allowClear
          onClear={() => { setActorFilter(''); setPage(1); }}
          style={{ width: 220 }}
        />
        <Select
          placeholder="Filter by action"
          allowClear
          value={actionFilter}
          style={{ width: 240 }}
          onChange={(v) => { setActionFilter(v); setPage(1); }}
          options={ALL_ACTIONS.map((a) => ({ label: a, value: a }))}
        />
        {hasFilters && (
          <Button type="link" onClick={clearFilters} style={{ padding: 0, height: 'auto', color: 'var(--text-muted)' }}>
            Clear filters
          </Button>
        )}
      </div>

      {/* Table */}
      <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', overflow: 'hidden', boxShadow: 'var(--shadow-card)' }}>
        <Table<AuditEntry>
          columns={columns}
          dataSource={data?.entries ?? []}
          rowKey="id"
          loading={isLoading}
          locale={{ emptyText: emptyState }}
          pagination={{
            current: page,
            pageSize: PAGE_SIZE,
            total: data?.total ?? 0,
            onChange: (p) => setPage(p),
            showTotal: (total) => `${total} entries`,
            style: { padding: '12px 16px' },
          }}
          scroll={{ x: 900 }}
          style={{ borderRadius: 0 }}
        />
      </div>
    </>
  );
}
