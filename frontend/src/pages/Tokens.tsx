import { useState } from 'react'
import { Table, Tag, Button, Space, Input, Select, Tooltip, Popconfirm, Badge, message, Empty } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  KeyOutlined,
  SearchOutlined,
  StopOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  UserOutlined,
  AppstoreOutlined,
} from '@ant-design/icons'
import { useTokens, useRevokeToken } from '../hooks/useApi'
import type { TokenEntry, TokenFilter } from '../hooks/useApi'

const scopeColors: Record<string, string> = {
  openid: '#0D9488',
  profile: '#0891B2',
  email: '#7C3AED',
}

export default function Tokens() {
  // "draft" state — what the user is typing in the filter bar
  const [draft, setDraft] = useState<TokenFilter>({ active: true })
  const [draftSearch, setDraftSearch] = useState('')

  // "committed" state — applied only when Search is clicked
  const [filter, setFilter] = useState<TokenFilter>({ active: true })
  const [search, setSearch] = useState('')
  const [hasSearched, setHasSearched] = useState(false)

  const [messageApi, contextHolder] = message.useMessage()

  const { data, isLoading, refetch } = useTokens(filter, hasSearched)
  const revoke = useRevokeToken()

  const handleSearch = () => {
    setFilter(draft)
    setSearch(draftSearch)
    setHasSearched(true)
  }

  const handleClear = () => {
    setDraft({ active: true })
    setDraftSearch('')
    setFilter({ active: true })
    setSearch('')
    setHasSearched(false)
  }

  const handleRevoke = (tokenId: string, prefix: string) => {
    revoke.mutate(tokenId, {
      onSuccess: () => {
        messageApi.success(`Token ${prefix} revoked`)
        refetch()
      },
      onError: () => messageApi.error('Failed to revoke token'),
    })
  }

  // Client-side free-text search across token prefix, username, client_id, scope
  const tokens = (data?.tokens ?? []).filter(t => {
    if (!search) return true
    const q = search.toLowerCase()
    return (
      t.access_token_prefix.toLowerCase().includes(q) ||
      t.username.toLowerCase().includes(q) ||
      t.client_id.toLowerCase().includes(q) ||
      t.scope.toLowerCase().includes(q)
    )
  })

  const columns: ColumnsType<TokenEntry> = [
    {
      title: 'Token',
      dataIndex: 'access_token_prefix',
      key: 'token',
      render: (prefix: string) => (
        <code style={{
          fontFamily: 'monospace',
          fontSize: 12,
          background: 'var(--color-bg)',
          padding: '2px 8px',
          borderRadius: 4,
          color: 'var(--color-text-muted)',
          border: '1px solid var(--color-border)',
        }}>
          {prefix}
        </code>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'is_active',
      key: 'status',
      width: 100,
      render: (active: boolean) =>
        active ? (
          <Badge
            status="success"
            text={<span style={{ color: '#10B981', fontWeight: 600, fontSize: 12 }}>Active</span>}
          />
        ) : (
          <Badge
            status="error"
            text={<span style={{ color: '#EF4444', fontWeight: 600, fontSize: 12 }}>Expired</span>}
          />
        ),
    },
    {
      title: 'User',
      dataIndex: 'username',
      key: 'user',
      render: (username: string) =>
        username ? (
          <Space size={6}>
            <div style={{
              width: 26, height: 26, borderRadius: '50%',
              background: 'linear-gradient(135deg,#0D9488,#0F766E)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 11, fontWeight: 700, color: '#fff', flexShrink: 0,
            }}>
              {username[0].toUpperCase()}
            </div>
            <span style={{ fontSize: 13, fontWeight: 500 }}>{username}</span>
          </Space>
        ) : (
          <span style={{ color: 'var(--color-text-muted)', fontSize: 12 }}>
            <UserOutlined style={{ marginRight: 4 }} />
            Client credentials
          </span>
        ),
    },
    {
      title: 'Client',
      dataIndex: 'client_id',
      key: 'client',
      render: (clientId: string) => (
        <Space size={6}>
          <AppstoreOutlined style={{ color: 'var(--color-text-muted)', fontSize: 12 }} />
          <code style={{ fontSize: 12, color: 'var(--color-text-secondary)' }}>
            {clientId.length > 20 ? clientId.slice(0, 20) + '…' : clientId}
          </code>
        </Space>
      ),
    },
    {
      title: 'Scopes',
      dataIndex: 'scope',
      key: 'scope',
      render: (scope: string) => (
        <Space size={4} wrap>
          {scope.split(' ').filter(Boolean).map(s => (
            <Tag
              key={s}
              style={{
                fontSize: 11,
                padding: '1px 8px',
                borderRadius: 12,
                border: 'none',
                background: scopeColors[s] ? `${scopeColors[s]}20` : 'var(--color-bg)',
                color: scopeColors[s] ?? 'var(--color-text-muted)',
                fontWeight: 500,
              }}
            >
              {s}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'token_type',
      key: 'type',
      width: 90,
      render: (type: string) => (
        <Tag style={{ fontSize: 11, borderRadius: 8 }}>{type}</Tag>
      ),
    },
    {
      title: 'Refresh',
      dataIndex: 'refresh_token_present',
      key: 'refresh',
      width: 80,
      align: 'center' as const,
      render: (present: boolean) =>
        present ? (
          <Tooltip title="Has refresh token">
            <CheckCircleOutlined style={{ color: '#10B981', fontSize: 15 }} />
          </Tooltip>
        ) : (
          <Tooltip title="No refresh token">
            <CloseCircleOutlined style={{ color: '#475569', fontSize: 15 }} />
          </Tooltip>
        ),
    },
    {
      title: 'Expires',
      dataIndex: 'expires_at',
      key: 'expires',
      render: (ts: string) => {
        const date = new Date(ts)
        const expired = date < new Date()
        return (
          <span style={{ fontSize: 12, color: expired ? '#EF4444' : 'var(--color-text-muted)' }}>
            {date.toLocaleString()}
          </span>
        )
      },
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created',
      render: (ts: string) => (
        <span style={{ fontSize: 12, color: 'var(--color-text-muted)' }}>
          {new Date(ts).toLocaleString()}
        </span>
      ),
    },
    {
      title: '',
      key: 'actions',
      width: 80,
      render: (_: unknown, record: TokenEntry) => (
        <Popconfirm
          title="Revoke token?"
          description={`Revoke token ${record.access_token_prefix}? This cannot be undone.`}
          onConfirm={() => handleRevoke(record.id, record.access_token_prefix)}
          okText="Revoke"
          cancelText="Cancel"
          okButtonProps={{ danger: true }}
        >
          <Button
            size="small"
            danger
            icon={<StopOutlined />}
            title="Revoke token"
            style={{ borderRadius: 6 }}
          >
            Revoke
          </Button>
        </Popconfirm>
      ),
    },
  ]

  return (
    <>
      {contextHolder}

      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div style={{
            width: 40, height: 40, borderRadius: 10,
            background: 'linear-gradient(135deg,#0D9488,#0F766E)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            boxShadow: '0 4px 12px rgba(13,148,136,0.3)',
          }}>
            <KeyOutlined style={{ color: '#fff', fontSize: 18 }} />
          </div>
          <div>
            <h1 style={{ margin: 0, fontSize: 20, fontWeight: 700, color: 'var(--color-text)' }}>
              Token Management
            </h1>
            <p style={{ margin: 0, fontSize: 13, color: 'var(--color-text-muted)' }}>
              {hasSearched ? `${tokens.length} token${tokens.length !== 1 ? 's' : ''} found` : 'Use filters and click Search to load tokens'}
            </p>
          </div>
        </div>
      </div>

      {/* Filter Bar */}
      <div style={{
        display: 'flex',
        gap: 12,
        alignItems: 'center',
        flexWrap: 'wrap',
        marginBottom: 20,
        padding: '14px 16px',
        background: 'var(--color-surface)',
        border: '1px solid var(--color-border)',
        borderRadius: 12,
      }}>
        {/* Active toggle */}
        <Select
          value={draft.active === false ? 'all' : 'active'}
          onChange={val => setDraft(f => ({ ...f, active: val !== 'all' }))}
          style={{ width: 160 }}
          options={[
            { value: 'active', label: '🟢  Active only' },
            { value: 'all',    label: '🔘  All tokens' },
          ]}
        />

        {/* Free-text search box */}
        <Input
          prefix={<SearchOutlined style={{ color: 'var(--color-text-muted)' }} />}
          placeholder="Search user, client, scope…"
          value={draftSearch}
          onChange={e => setDraftSearch(e.target.value)}
          onPressEnter={handleSearch}
          allowClear
          style={{ width: 240, borderRadius: 8 }}
        />

        {/* Client ID filter */}
        <Input
          prefix={<AppstoreOutlined style={{ color: 'var(--color-text-muted)' }} />}
          placeholder="Filter by Client ID"
          value={draft.client_id ?? ''}
          onChange={e => setDraft(f => ({ ...f, client_id: e.target.value || undefined }))}
          onPressEnter={handleSearch}
          allowClear
          style={{ width: 200, borderRadius: 8 }}
        />

        {/* User ID filter */}
        <Input
          prefix={<UserOutlined style={{ color: 'var(--color-text-muted)' }} />}
          placeholder="Filter by User ID"
          value={draft.user_id ?? ''}
          onChange={e => setDraft(f => ({ ...f, user_id: e.target.value || undefined }))}
          onPressEnter={handleSearch}
          allowClear
          style={{ width: 200, borderRadius: 8 }}
        />

        {/* Search button */}
        <Button
          type="primary"
          icon={<SearchOutlined />}
          onClick={handleSearch}
          style={{
            borderRadius: 8,
            background: '#0D9488',
            borderColor: '#0D9488',
          }}
        >
          Search
        </Button>

        {/* Clear button — only shown after a search */}
        {hasSearched && (
          <Button
            onClick={handleClear}
            style={{ borderRadius: 8 }}
          >
            Clear
          </Button>
        )}
      </div>

      {/* Summary pills — only shown after search */}
      {hasSearched && (
        <div style={{ display: 'flex', gap: 10, marginBottom: 16, flexWrap: 'wrap' }}>
          {[
            { label: 'Total shown', value: tokens.length, color: '#0D9488' },
            { label: 'Active', value: tokens.filter(t => t.is_active).length, color: '#10B981' },
            { label: 'Expired', value: tokens.filter(t => !t.is_active).length, color: '#EF4444' },
            { label: 'With refresh', value: tokens.filter(t => t.refresh_token_present).length, color: '#F59E0B' },
          ].map(pill => (
            <div key={pill.label} style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              padding: '6px 14px',
              background: 'var(--color-surface)',
              border: `1px solid ${pill.color}30`,
              borderRadius: 20,
              fontSize: 12,
            }}>
              <span style={{
                width: 8, height: 8, borderRadius: '50%',
                background: pill.color, flexShrink: 0,
              }} />
              <span style={{ color: 'var(--color-text-muted)' }}>{pill.label}</span>
              <span style={{ fontWeight: 700, color: pill.color }}>{pill.value}</span>
            </div>
          ))}
        </div>
      )}

      {/* Table / empty state */}
      <div style={{
        background: 'var(--color-surface)',
        border: '1px solid var(--color-border)',
        borderRadius: 12,
      }}>
        {!hasSearched ? (
          <div style={{ padding: '60px 0' }}>
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                <span style={{ color: 'var(--color-text-muted)', fontSize: 14 }}>
                  Set your filters above and click <strong>Search</strong> to load tokens
                </span>
              }
            />
          </div>
        ) : (
          <Table
            className="token-table"
            dataSource={tokens}
            columns={columns}
            rowKey="id"
            loading={isLoading}
            scroll={{ x: 1100 }}
            pagination={{
              pageSize: 20,
              showSizeChanger: true,
              showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
              style: { padding: '12px 16px' },
            }}
            rowClassName={(record) => record.is_active ? '' : 'token-row-expired'}
            size="middle"
          />
        )}
      </div>

      <style>{`
        .token-row-expired td {
          opacity: 0.55;
        }
        .token-row-expired:hover td {
          opacity: 0.8;
        }
        /* Round the header corners without overflow:hidden on wrapper */
        .token-table .ant-table-thead > tr:first-child > th:first-child {
          border-top-left-radius: 11px !important;
        }
        .token-table .ant-table-thead > tr:first-child > th:last-child {
          border-top-right-radius: 11px !important;
        }
        /* Round bottom pagination area */
        .token-table .ant-table-wrapper {
          border-radius: 11px;
        }
      `}</style>
    </>
  )
}
