# Architecture Diagrams

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Application                       │
│                    (Web App, Mobile App, etc.)                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ HTTP/HTTPS
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                    OpenID Connect Server                         │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    HTTP Layer (Gorilla Mux)                 │ │
│  │  Endpoints: /authorize, /token, /userinfo, /.well-known    │ │
│  └─────────────────────────┬──────────────────────────────────┘ │
│                            │                                     │
│  ┌─────────────────────────▼──────────────────────────────────┐ │
│  │                      Middleware Layer                        │ │
│  │          Logging │ CORS │ Recovery │ Authentication         │ │
│  └─────────────────────────┬──────────────────────────────────┘ │
│                            │                                     │
│  ┌─────────────────────────▼──────────────────────────────────┐ │
│  │                    Handlers Layer                            │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────┐  │ │
│  │  │Authorize │  │  Token   │  │ UserInfo │  │ Discovery │  │ │
│  │  │ Handler  │  │ Handler  │  │ Handler  │  │  Handler  │  │ │
│  │  └──────────┘  └──────────┘  └──────────┘  └───────────┘  │ │
│  └─────────────────────────┬──────────────────────────────────┘ │
│                            │                                     │
│  ┌─────────────────────────▼──────────────────────────────────┐ │
│  │                    Business Logic Layer                      │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐ │ │
│  │  │    Crypto    │  │    Config    │  │     Models       │ │ │
│  │  │  JWT/PKCE    │  │  Management  │  │  User/Client/    │ │ │
│  │  │   bcrypt     │  │              │  │     Token        │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────┘ │ │
│  └─────────────────────────┬──────────────────────────────────┘ │
│                            │                                     │
│  ┌─────────────────────────▼──────────────────────────────────┐ │
│  │                    Storage Layer                             │ │
│  │                  (Interface-based)                           │ │
│  └─────────────────────────┬──────────────────────────────────┘ │
└────────────────────────────┼──────────────────────────────────┘
                             │
                ┌────────────┴─────────────┐
                │                          │
        ┌───────▼────────┐       ┌────────▼─────────┐
        │   JSON File    │       │    MongoDB       │
        │   (Default)    │       │   (Production)   │
        └────────────────┘       └──────────────────┘
```

## Authorization Code Flow

```
┌────────┐                                              ┌─────────────┐
│        │                                              │             │
│ Client │                                              │ OIDC Server │
│        │                                              │             │
└───┬────┘                                              └──────┬──────┘
    │                                                          │
    │  1. GET /authorize?client_id=...&redirect_uri=...       │
    │     &response_type=code&scope=openid profile email      │
    ├─────────────────────────────────────────────────────────>│
    │                                                          │
    │  2. Redirect to Login Page                              │
    │<─────────────────────────────────────────────────────────┤
    │                                                          │
    │  3. POST /login (username, password)                    │
    ├─────────────────────────────────────────────────────────>│
    │                                                          │
    │                                          [Authenticate]  │
    │                                          [Create Code]   │
    │                                                          │
    │  4. Redirect to callback with code                      │
    │     http://client/callback?code=abc123&state=xyz        │
    │<─────────────────────────────────────────────────────────┤
    │                                                          │
    │  5. POST /token (code, client_id, client_secret)        │
    ├─────────────────────────────────────────────────────────>│
    │                                                          │
    │                                          [Validate Code] │
    │                                          [Create Tokens] │
    │                                                          │
    │  6. Return tokens                                        │
    │     {access_token, id_token, refresh_token}             │
    │<─────────────────────────────────────────────────────────┤
    │                                                          │
    │  7. GET /userinfo (Authorization: Bearer <token>)       │
    ├─────────────────────────────────────────────────────────>│
    │                                                          │
    │                                        [Validate Token]  │
    │                                        [Get User Info]   │
    │                                                          │
    │  8. Return user info                                     │
    │     {sub, name, email, ...}                             │
    │<─────────────────────────────────────────────────────────┤
    │                                                          │
```

## Component Dependencies

```
┌─────────────────────────────────────────────────────────┐
│                     cmd/server/main.go                   │
│                  (Application Entry Point)               │
└────────┬─────────────────────────────────┬──────────────┘
         │                                 │
         │ imports                         │ imports
         │                                 │
    ┌────▼────────┐                  ┌─────▼──────────┐
    │   config    │                  │    storage     │
    │             │                  │                │
    │ - Load()    │                  │ - NewStorage() │
    │ - Validate()│                  │ - Interface    │
    └─────────────┘                  └────────┬───────┘
                                              │
         ┌────────────────────────────────────┼────────────────────┐
         │                                    │                    │
    ┌────▼────────┐                    ┌──────▼────────┐   ┌──────▼────────┐
    │   models    │                    │   mongodb     │   │   json        │
    │             │                    │               │   │               │
    │ - User      │◄───────────────────┤ - CreateUser()│   │ - CreateUser()│
    │ - Client    │                    │ - GetClient() │   │ - GetClient() │
    │ - Token     │                    │ - etc.        │   │ - etc.        │
    └─────────────┘                    └───────────────┘   └───────────────┘
         ▲
         │
         │ uses
         │
    ┌────┴────────┐
    │  handlers   │
    │             │
    │ - Authorize │
    │ - Token     │────────┐
    │ - UserInfo  │        │
    └─────┬───────┘        │ uses
          │                │
          │ uses      ┌────▼───────┐
          │           │   crypto   │
          │           │            │
          └──────────►│ - JWT      │
                      │ - PKCE     │
                      │ - bcrypt   │
                      └────────────┘
```

## Data Model Relationships

```
┌─────────────────┐
│      User       │
│─────────────────│
│ id (PK)         │◄─────────┐
│ username        │          │
│ email           │          │
│ password_hash   │          │
│ name            │          │
│ given_name      │          │
│ family_name     │          │
│ picture         │          │
└─────────────────┘          │
                             │
                             │ user_id (FK)
                             │
                  ┌──────────┴──────────┐
                  │                     │
       ┌──────────▼──────────┐   ┌──────▼──────────────┐
       │  Authorization Code │   │       Token         │
       │─────────────────────│   │─────────────────────│
       │ code (PK)           │   │ id (PK)             │
       │ client_id (FK)      │   │ access_token        │
       │ user_id (FK)        │   │ refresh_token       │
       │ redirect_uri        │   │ client_id (FK)      │
       │ scope               │   │ user_id (FK)        │
       │ nonce               │   │ scope               │
       │ code_challenge      │   │ expires_at          │
       │ expires_at          │   └─────────────────────┘
       └─────────┬───────────┘
                 │
                 │ client_id (FK)
                 │
       ┌─────────▼───────────┐
       │      Client         │
       │─────────────────────│
       │ id (PK)             │
       │ secret              │
       │ name                │
       │ redirect_uris       │
       │ grant_types         │
       │ response_types      │
       │ scope               │
       └─────────────────────┘

       ┌─────────────────────┐
       │      Session        │
       │─────────────────────│
       │ id (PK)             │
       │ user_id (FK) ───────┼─────┐
       │ expires_at          │     │
       └─────────────────────┘     │
                                   │
                        references │
                                   │
                       ┌───────────▼────────┐
                       │       User         │
                       └────────────────────┘
```

## Request Flow Details

### Token Generation Flow

```
┌─────────────┐
│ POST /token │
└──────┬──────┘
       │
       ▼
┌────────────────────────┐
│  Parse Request         │
│  - grant_type          │
│  - code/refresh_token  │
│  - client credentials  │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│  Authenticate Client   │
│  - Basic Auth or POST  │
│  - Validate credentials│
└──────┬─────────────────┘
       │
       ├─── authorization_code ──┐
       │                         │
       └─── refresh_token ────┐  │
                              │  │
              ┌───────────────▼──▼──────────────┐
              │  Grant Type Specific Validation │
              └───────────────┬──────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  Get User from Storage        │
              └───────────────┬───────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  Create New Token             │
              │  - Access Token               │
              │  - Refresh Token              │
              └───────────────┬───────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  Generate ID Token (JWT)      │
              │  - Sign with RS256            │
              │  - Include user claims        │
              └───────────────┬───────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  Return Token Response        │
              │  {access_token, id_token,     │
              │   refresh_token, expires_in}  │
              └───────────────────────────────┘
```

## Security Flow

```
┌──────────────────────────────────────────────────┐
│              Security Layers                      │
├──────────────────────────────────────────────────┤
│                                                  │
│  1. Transport Security (HTTPS)                   │
│     ├─ TLS/SSL encryption                        │
│     └─ Certificate validation                    │
│                                                  │
│  2. Client Authentication                        │
│     ├─ Client ID + Secret                        │
│     ├─ Basic Auth support                        │
│     └─ PKCE for public clients                   │
│                                                  │
│  3. User Authentication                          │
│     ├─ Username + Password                       │
│     ├─ bcrypt password hashing                   │
│     └─ Session management                        │
│                                                  │
│  4. Token Security                               │
│     ├─ JWT signing (RS256)                       │
│     ├─ Token expiration                          │
│     ├─ Nonce for replay protection               │
│     └─ State for CSRF protection                 │
│                                                  │
│  5. Authorization                                │
│     ├─ Scope validation                          │
│     ├─ Redirect URI validation                   │
│     └─ Client permission checks                  │
│                                                  │
│  6. Code Security                                │
│     ├─ One-time use codes                        │
│     ├─ Short expiration (10 min)                 │
│     └─ Bound to client & redirect URI            │
│                                                  │
└──────────────────────────────────────────────────┘
```
