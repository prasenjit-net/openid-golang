// User type definitions
export interface Address {
  formatted?: string;
  street_address?: string;
  locality?: string;
  region?: string;
  postal_code?: string;
  country?: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  email_verified: boolean;
  password_hash?: string;
  role: 'user' | 'admin';
  name?: string;
  given_name?: string;
  family_name?: string;
  picture?: string;
  address?: Address;
  created_at: string;
  updated_at: string;
}

export interface UserSearchParams {
  query?: string;
  role?: 'user' | 'admin' | '';
  status?: 'active' | 'disabled' | '';
  page?: number;
  pageSize?: number;
}

export interface UserCreateRequest {
  username: string;
  email: string;
  password: string;
  name?: string;
  given_name?: string;
  family_name?: string;
  role: 'user' | 'admin';
  email_verified?: boolean;
  picture?: string;
  address?: Address;
}

export interface UserUpdateRequest {
  email?: string;
  name?: string;
  given_name?: string;
  family_name?: string;
  role?: 'user' | 'admin';
  email_verified?: boolean;
  picture?: string;
  address?: Address;
}

export interface PasswordChangeRequest {
  new_password: string;
}

export interface EmailChangeRequest {
  new_email: string;
}

export interface UserStatusRequest {
  enabled: boolean;
}

export interface UserListResponse {
  users: User[];
  total: number;
  page: number;
  pageSize: number;
}
