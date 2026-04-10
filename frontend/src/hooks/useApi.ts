import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

// API base URL
const API_BASE = '/api/admin'

// Types
interface CreateClientRequest {
  name: string;
  redirect_uris: string[];
}

export interface TokenEntry {
  id: string
  access_token_prefix: string
  refresh_token_present: boolean
  token_type: string
  client_id: string
  user_id: string
  username: string
  scope: string
  expires_at: string
  created_at: string
  is_active: boolean
}

export interface TokenFilter {
  active?: boolean
  client_id?: string
  user_id?: string
}

interface UpdateSettingsRequest {
  issuer?: string;
  server_host?: string;
  server_port?: number;
  storage_type?: string;
  json_file_path?: string;
  mongo_uri?: string;
  jwt_expiry_minutes?: number;
  jwt_private_key?: string;
  jwt_public_key?: string;
}

// Query keys
export const queryKeys = {
  stats: ['stats'] as const,
  users: ['users'] as const,
  user: (id: string) => ['user', id] as const,
  clients: ['clients'] as const,
  client: (id: string) => ['client', id] as const,
  settings: ['settings'] as const,
  keys: ['keys'] as const,
  setupStatus: ['setupStatus'] as const,
  tokens: (filter: TokenFilter) => ['tokens', filter] as const,
}

// Helper to get auth headers
function getAuthHeaders(): Record<string, string> {
  const token = localStorage.getItem('admin_token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

// Stats
export function useStats() {
  return useQuery({
    queryKey: queryKeys.stats,
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/stats`, {
        headers: {
          ...getAuthHeaders(),
        },
      })
      if (!res.ok) throw new Error('Failed to fetch stats')
      return res.json()
    },
  })
}

// Users
export function useUsers() {
  return useQuery({
    queryKey: queryKeys.users,
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/users`, {
        headers: {
          ...getAuthHeaders(),
        },
      })
      if (!res.ok) throw new Error('Failed to fetch users')
      return res.json()
    },
  })
}

export function useCreateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (user: { username: string; email: string; password: string; role: string }) => {
      const res = await fetch(`${API_BASE}/users`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify(user),
      })
      if (!res.ok) throw new Error('Failed to create user')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users })
    },
  })
}

export function useDeleteUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await fetch(`${API_BASE}/users/${id}`, {
        method: 'DELETE',
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to delete user')
    },
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users })
      queryClient.removeQueries({ queryKey: queryKeys.user(id) })
    },
  })
}

// Clients
export function useClients() {
  return useQuery({
    queryKey: queryKeys.clients,
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/clients`, {
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to fetch clients')
      return res.json()
    },
  })
}

export function useCreateClient() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (client: CreateClientRequest) => {
      const res = await fetch(`${API_BASE}/clients`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify(client),
      })
      if (!res.ok) throw new Error('Failed to create client')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clients })
    },
  })
}

export function useDeleteClient() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await fetch(`${API_BASE}/clients/${id}`, {
        method: 'DELETE',
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to delete client')
    },
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clients })
      queryClient.removeQueries({ queryKey: queryKeys.client(id) })
    },
  })
}

// Settings
export function useSettings() {
  return useQuery({
    queryKey: queryKeys.settings,
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/settings`, {
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to fetch settings')
      return res.json()
    },
  })
}

export function useUpdateSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (settings: UpdateSettingsRequest) => {
      const res = await fetch(`${API_BASE}/settings`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify(settings),
      })
      if (!res.ok) throw new Error('Failed to update settings')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings })
    },
  })
}

// Keys
export function useKeys() {
  return useQuery({
    queryKey: queryKeys.keys,
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/keys`, {
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to fetch keys')
      return res.json()
    },
  })
}

export function useRotateKeys() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (validityDays: number = 90) => {
      const res = await fetch(`${API_BASE}/settings/rotate-keys`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify({ validity_days: validityDays }),
      })
      if (!res.ok) throw new Error('Failed to rotate keys')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.keys })
    },
  })
}

export function useGenerateCSR() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (keyId: string) => {
      const res = await fetch(`${API_BASE}/keys/${keyId}/csr`, {
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to generate CSR')
      return res.json() as Promise<{ kid: string; csr_pem: string }>
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.keys })
    },
  })
}

export function useImportCert() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ keyId, certPem }: { keyId: string; certPem: string }) => {
      const res = await fetch(`${API_BASE}/keys/${keyId}/import-cert`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify({ cert_pem: certPem }),
      })
      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: 'Failed to import certificate' }))
        throw new Error(err.error ?? 'Failed to import certificate')
      }
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.keys })
    },
  })
}

// Setup Status
export function useSetupStatus() {
  return useQuery({
    queryKey: queryKeys.setupStatus,
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/setup/status`)
      if (!res.ok) throw new Error('Failed to fetch setup status')
      return res.json()
    },
  })
}

// Individual User
export function useUser(id: string) {
  return useQuery({
    queryKey: queryKeys.user(id),
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/users/${id}`, {
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to fetch user')
      return res.json()
    },
    enabled: !!id,
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...user }: { id: string; username?: string; email?: string; name?: string; password?: string; role?: string; [key: string]: unknown }) => {
      const res = await fetch(`${API_BASE}/users/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify(user),
      })
      if (!res.ok) throw new Error('Failed to update user')
      return res.json()
    },
    onSuccess: (data, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users })
      // Immediately update individual user cache so detail page shows fresh data
      queryClient.setQueryData(queryKeys.user(id), data)
    },
  })
}

// Individual Client
export function useClient(id: string) {
  return useQuery({
    queryKey: queryKeys.client(id),
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/clients/${id}`, {
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to fetch client')
      return res.json()
    },
    enabled: !!id,
  })
}

export function useUpdateClient() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...client }: { id: string; name?: string; redirect_uris?: string[]; grant_types?: string[]; response_types?: string[]; scope?: string; application_type?: string }) => {
      const res = await fetch(`${API_BASE}/clients/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify(client),
      })
      if (!res.ok) throw new Error('Failed to update client')
      return res.json()
    },
    onSuccess: (data, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clients })
      // Immediately update individual client cache so detail page shows fresh data
      queryClient.setQueryData(queryKeys.client(id), data)
    },
  })
}

export function useRegenerateClientSecret() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await fetch(`${API_BASE}/clients/${id}/regenerate-secret`, {
        method: 'POST',
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to regenerate client secret')
      return res.json()
    },
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clients })
      queryClient.invalidateQueries({ queryKey: queryKeys.client(id) })
    },
  })
}

// Setup
export function useSetup() {
  return useMutation({
    mutationFn: async (data: { username: string; email: string; password: string }) => {
      const res = await fetch(`${API_BASE}/setup`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      })
      if (!res.ok) throw new Error('Failed to complete setup')
      return res.json()
    },
  })
}

// Profile
export function useProfile() {
  return useQuery({
    queryKey: ['profile'],
    queryFn: async () => {
      const token = localStorage.getItem('admin_token')
      const res = await fetch(`${API_BASE}/profile`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })
      if (!res.ok) throw new Error('Failed to fetch profile')
      return res.json()
    },
  })
}

export function useUpdateProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: { email?: string; name?: string }) => {
      const token = localStorage.getItem('admin_token')
      const res = await fetch(`${API_BASE}/profile`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(data),
      })
      if (!res.ok) throw new Error('Failed to update profile')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profile'] })
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async (data: { currentPassword: string; newPassword: string }) => {
      const token = localStorage.getItem('admin_token')
      const res = await fetch(`${API_BASE}/profile/change-password`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(data),
      })
      if (!res.ok) {
        const error = await res.json()
        throw new Error(error.error || 'Failed to change password')
      }
      return res.json()
    },
  })
}


interface AuditFilter {
  limit?: number
  offset?: number
  action?: string
  actor?: string
}

interface AuditEntry {
  id: string
  timestamp: string
  action: string
  actor_type: string
  actor: string
  resource: string
  resource_id: string
  status: string
  ip_address: string
  user_agent: string
  details?: Record<string, unknown>
}

export function useAuditLogs(filter: AuditFilter = {}) {
  return useQuery({
    queryKey: ['audit', filter],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filter.limit) params.set('limit', String(filter.limit))
      if (filter.offset) params.set('offset', String(filter.offset))
      if (filter.action) params.set('action', filter.action)
      if (filter.actor) params.set('actor', filter.actor)
      const res = await fetch(`${API_BASE}/audit?${params.toString()}`, {
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to fetch audit logs')
      return res.json() as Promise<{ entries: AuditEntry[]; total: number; limit: number; offset: number }>
    },
  })
}

export function useTokens(filter: TokenFilter = { active: true }, enabled = false) {
  return useQuery({
    queryKey: queryKeys.tokens(filter),
    queryFn: async () => {
      const params = new URLSearchParams()
      params.set('active', filter.active === false ? 'false' : 'true')
      if (filter.client_id) params.set('client_id', filter.client_id)
      if (filter.user_id) params.set('user_id', filter.user_id)
      const res = await fetch(`${API_BASE}/tokens?${params.toString()}`, {
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to fetch tokens')
      return res.json() as Promise<{ tokens: TokenEntry[]; total: number }>
    },
    enabled,
  })
}

export function useRevokeToken() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (tokenId: string) => {
      const res = await fetch(`${API_BASE}/tokens/${tokenId}`, {
        method: 'DELETE',
        headers: { ...getAuthHeaders() },
      })
      if (!res.ok) throw new Error('Failed to revoke token')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
      queryClient.invalidateQueries({ queryKey: queryKeys.stats })
    },
  })
}
