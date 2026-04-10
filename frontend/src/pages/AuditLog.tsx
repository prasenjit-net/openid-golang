import { useState } from 'react';
import {
  Table,
  Typography,
  Card,
  Tag,
  Select,
  Space,
  Input,
  Button,
  Tooltip,
} from 'antd';
import { AuditOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useAuditLogs } from '../hooks/useApi';

const { Title } = Typography;

interface AuditEntry {
  id: string;
  timestamp: string;
  action: string;
  actor_type: string;
  actor: string;
  resource_type: string;
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

export default function AuditLog() {
  const [page, setPage] = useState(1);
  const [actionFilter, setActionFilter] = useState<string | undefined>(undefined);
  const [actorFilter, setActorFilter] = useState('');

  const offset = (page - 1) * PAGE_SIZE;
  const { data, isLoading, refetch } = useAuditLogs({
    limit: PAGE_SIZE,
    offset,
    action: actionFilter,
    actor: actorFilter || undefined,
  });

  const columns: ColumnsType<AuditEntry> = [
    {
      title: 'Timestamp',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (ts: string) =>
        new Date(ts).toLocaleString(undefined, { hour12: false }),
      sorter: (a, b) =>
        new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
      defaultSortOrder: 'descend',
    },
    {
      title: 'Action',
      dataIndex: 'action',
      key: 'action',
      width: 200,
      render: (action: string) => (
        <Tag color={ACTION_COLORS[action] || 'default'}>{action}</Tag>
      ),
    },
    {
      title: 'Actor',
      dataIndex: 'actor',
      key: 'actor',
      width: 150,
      render: (actor: string, record) => (
        <Tooltip title={`Type: ${record.actor_type}`}>{actor || '—'}</Tooltip>
      ),
    },
    {
      title: 'Resource',
      key: 'resource',
      width: 200,
      render: (_: unknown, record) =>
        record.resource_type ? (
          <span>
            <Tag>{record.resource_type}</Tag>
            <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
              {record.resource_id?.slice(0, 20)}
              {record.resource_id?.length > 20 ? '…' : ''}
            </span>
          </span>
        ) : (
          '—'
        ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: string) => (
        <Tag color={status === 'success' ? 'success' : 'error'}>{status}</Tag>
      ),
    },
    {
      title: 'IP Address',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 130,
      render: (ip: string) => ip || '—',
    },
    {
      title: 'Details',
      dataIndex: 'details',
      key: 'details',
      render: (details: Record<string, unknown>) =>
        details ? (
          <Tooltip
            title={
              <pre style={{ fontSize: 11, margin: 0 }}>
                {JSON.stringify(details, null, 2)}
              </pre>
            }
          >
            <span style={{ cursor: 'pointer', color: '#1890ff' }}>view</span>
          </Tooltip>
        ) : (
          '—'
        ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <Space style={{ marginBottom: 16, width: '100%', justifyContent: 'space-between' }}>
          <Title level={4} style={{ margin: 0 }}>
            <AuditOutlined style={{ marginRight: 8 }} />
            Audit Log
          </Title>
          <Space>
            <Input.Search
              placeholder="Filter by actor"
              allowClear
              style={{ width: 200 }}
              onSearch={(v) => { setActorFilter(v); setPage(1); }}
            />
            <Select
              placeholder="Filter by action"
              allowClear
              style={{ width: 220 }}
              onChange={(v) => { setActionFilter(v); setPage(1); }}
              options={ALL_ACTIONS.map((a) => ({ label: a, value: a }))}
            />
            <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
              Refresh
            </Button>
          </Space>
        </Space>

        <Table<AuditEntry>
          columns={columns}
          dataSource={data?.entries ?? []}
          rowKey="id"
          loading={isLoading}
          size="small"
          pagination={{
            current: page,
            pageSize: PAGE_SIZE,
            total: data?.total ?? 0,
            onChange: (p) => setPage(p),
            showTotal: (total) => `${total} entries`,
          }}
          scroll={{ x: 1100 }}
        />
      </Card>
    </div>
  );
}
