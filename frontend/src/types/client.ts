// Client type definitions
export interface Client {
  client_id: string;
  client_secret?: string;
  client_secret_expires_at: number;
  redirect_uris: string[];
  grant_types?: string[];
  response_types?: string[];
  scope?: string;
  application_type?: string;
  contacts?: string[];
  client_name?: string;
  logo_uri?: string;
  client_uri?: string;
  policy_uri?: string;
  tos_uri?: string;
  jwks_uri?: string;
  jwks?: Record<string, any>;
  token_endpoint_auth_method?: string;
  client_id_issued_at?: number;
  created_at: string;
  updated_at: string;
}

export interface ClientSearchParams {
  query?: string;
  grant_type?: string;
  application_type?: string;
  page?: number;
  pageSize?: number;
}

export interface ClientCreateRequest {
  client_name: string;
  application_type?: 'web' | 'native';
  redirect_uris: string[];
  grant_types?: string[];
  response_types?: string[];
  scope?: string;
  contacts?: string[];
  logo_uri?: string;
  client_uri?: string;
  policy_uri?: string;
  tos_uri?: string;
  jwks_uri?: string;
  jwks?: Record<string, any>;
  token_endpoint_auth_method?: string;
}

export interface ClientUpdateRequest {
  client_name?: string;
  redirect_uris?: string[];
  grant_types?: string[];
  response_types?: string[];
  scope?: string;
  contacts?: string[];
  logo_uri?: string;
  client_uri?: string;
  policy_uri?: string;
  tos_uri?: string;
  jwks_uri?: string;
  jwks?: Record<string, any>;
}

export interface ClientSecretResponse {
  client_id: string;
  client_secret: string;
  client_secret_expires_at: number;
}

export interface ClientStatusRequest {
  enabled: boolean;
}

export interface ClientListResponse {
  clients: Client[];
  total: number;
  page: number;
  pageSize: number;
}
