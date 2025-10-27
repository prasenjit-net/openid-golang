import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

// API base URL
const API_BASE = '/api/admin'

// Types
interface CreateClientRequest {
  name: string;
  redirect_uris: string[];
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
  clients: ['clients'] as const,
  settings: ['settings'] as const,
  keys: ['keys'] as const,
  setupStatus: ['setupStatus'] as const,
}

// Stats
export function useStats() {
  return useQuery({
    queryKey: queryKeys.stats,
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/stats`)
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
      const res = await fetch(`${API_BASE}/users`)
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
        headers: { 'Content-Type': 'application/json' },
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
      })
      if (!res.ok) throw new Error('Failed to delete user')
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users })
    },
  })
}

// Clients
export function useClients() {
  return useQuery({
    queryKey: queryKeys.clients,
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/clients`)
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
        headers: { 'Content-Type': 'application/json' },
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
      })
      if (!res.ok) throw new Error('Failed to delete client')
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clients })
    },
  })
}

// Settings
export function useSettings() {
  return useQuery({
    queryKey: queryKeys.settings,
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/settings`)
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
        headers: { 'Content-Type': 'application/json' },
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
      const res = await fetch(`${API_BASE}/keys`)
      if (!res.ok) throw new Error('Failed to fetch keys')
      return res.json()
    },
  })
}

export function useRotateKeys() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_BASE}/keys/rotate`, {
        method: 'POST',
      })
      if (!res.ok) throw new Error('Failed to rotate keys')
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
    queryKey: ['user', id],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/users/${id}`)
      if (!res.ok) throw new Error('Failed to fetch user')
      return res.json()
    },
    enabled: !!id,
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...user }: { id: string; username?: string; email?: string; password?: string; role?: string }) => {
      const res = await fetch(`${API_BASE}/users/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(user),
      })
      if (!res.ok) throw new Error('Failed to update user')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users })
    },
  })
}

// Individual Client
export function useClient(id: string) {
  return useQuery({
    queryKey: ['client', id],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/clients/${id}`)
      if (!res.ok) throw new Error('Failed to fetch client')
      return res.json()
    },
    enabled: !!id,
  })
}

export function useUpdateClient() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...client }: { id: string; name?: string; redirect_uris?: string[] }) => {
      const res = await fetch(`${API_BASE}/clients/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(client),
      })
      if (!res.ok) throw new Error('Failed to update client')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clients })
    },
  })
}

export function useRegenerateClientSecret() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await fetch(`${API_BASE}/clients/${id}/regenerate-secret`, {
        method: 'POST',
      })
      if (!res.ok) throw new Error('Failed to regenerate client secret')
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clients })
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
