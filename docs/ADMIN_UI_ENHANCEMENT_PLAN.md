# Admin UI Enhancement Plan

## Overview
Comprehensive plan to enhance the admin GUI with three main sections: User Management, Client Management, and Configuration Management, plus an improved Dashboard.

## Current State Analysis

### Existing Components
- Basic admin authentication
- Simple client listing
- Minimal UI structure in `frontend/src/pages/`

### Technology Stack
- **Frontend**: React + TypeScript + Vite
- **Styling**: CSS (consider upgrading to Tailwind CSS or Material-UI)
- **State Management**: React Context (AuthContext exists)
- **API Communication**: Custom useApi hook
- **Backend**: Go Echo framework with JSON/MongoDB storage

## Architecture Plan

### Phase 1: Foundation & Infrastructure (Week 1)

#### 1.1 Backend API Enhancements

**New API Endpoints Required:**

```
User Management APIs:
POST   /api/admin/users/search          - Search users with filters
POST   /api/admin/users                 - Create new user
GET    /api/admin/users/:id             - Get user details
PUT    /api/admin/users/:id             - Update user details
PUT    /api/admin/users/:id/password    - Change user password
PUT    /api/admin/users/:id/email       - Change user email
PUT    /api/admin/users/:id/status      - Enable/disable user
DELETE /api/admin/users/:id             - Delete user

Client Management APIs:
POST   /api/admin/clients/search        - Search clients with filters
POST   /api/admin/clients               - Create new client
GET    /api/admin/clients/:id           - Get client details
PUT    /api/admin/clients/:id           - Update client details
POST   /api/admin/clients/:id/secret    - Regenerate client secret
PUT    /api/admin/clients/:id/status    - Enable/disable client
DELETE /api/admin/clients/:id           - Delete client

Configuration Management APIs:
GET    /api/admin/config                - Get all configuration
PUT    /api/admin/config                - Update configuration
GET    /api/admin/config/certificates   - List all certificates
POST   /api/admin/config/certificates/rotate - Rotate certificates
DELETE /api/admin/config/certificates/purge  - Purge expired certificates
GET    /api/admin/config/jwks           - Get current JWKS

Dashboard APIs:
GET    /api/admin/dashboard/stats       - Get system statistics
GET    /api/admin/dashboard/activity    - Recent activity logs
GET    /api/admin/dashboard/health      - System health metrics
```

**Backend Files to Create/Modify:**
- `backend/pkg/handlers/admin_users.go` - User management handlers
- `backend/pkg/handlers/admin_clients.go` - Client management handlers (extend existing)
- `backend/pkg/handlers/admin_config.go` - Configuration management handlers
- `backend/pkg/handlers/admin_dashboard.go` - Dashboard handlers
- `backend/pkg/models/admin.go` - Admin-specific models
- `backend/pkg/storage/` - Add methods to storage interfaces

#### 1.2 Frontend Infrastructure

**UI Component Library Setup:**
```bash
# Install UI component library (recommend Ant Design or Material-UI)
npm install antd @ant-design/icons
# OR
npm install @mui/material @mui/icons-material @emotion/react @emotion/styled
```

**New Frontend Structure:**
```
frontend/src/
├── components/
│   ├── common/
│   │   ├── DataTable.tsx          - Reusable table component
│   │   ├── SearchBar.tsx          - Search with filters
│   │   ├── ConfirmDialog.tsx      - Confirmation modals
│   │   ├── FormField.tsx          - Reusable form fields
│   │   ├── StatusBadge.tsx        - Status indicators
│   │   └── LoadingSpinner.tsx     - Loading states
│   ├── users/
│   │   ├── UserList.tsx           - User listing table
│   │   ├── UserSearch.tsx         - User search filters
│   │   ├── UserForm.tsx           - Create/edit user form
│   │   ├── UserDetail.tsx         - User detail view
│   │   └── UserActions.tsx        - Action buttons/menu
│   ├── clients/
│   │   ├── ClientList.tsx         - Client listing table
│   │   ├── ClientSearch.tsx       - Client search filters
│   │   ├── ClientForm.tsx         - Create/edit client form
│   │   ├── ClientDetail.tsx       - Client detail view
│   │   └── ClientActions.tsx      - Action buttons/menu
│   ├── config/
│   │   ├── ConfigView.tsx         - Configuration viewer
│   │   ├── ConfigEditor.tsx       - Configuration editor
│   │   ├── CertificateList.tsx    - Certificate management
│   │   └── JWKSViewer.tsx         - JWKS viewer
│   └── dashboard/
│       ├── StatCard.tsx           - Stat display cards
│       ├── ActivityLog.tsx        - Recent activity
│       ├── HealthStatus.tsx       - System health
│       └── QuickActions.tsx       - Quick action buttons
├── pages/
│   ├── Dashboard.tsx              - Enhanced dashboard (exists)
│   ├── Users.tsx                  - User management page (exists)
│   ├── Clients.tsx                - Client management page (exists)
│   ├── Settings.tsx               - Config management page (exists)
│   └── UserDetail.tsx             - User detail page (new)
│   └── ClientDetail.tsx           - Client detail page (new)
├── hooks/
│   ├── useUsers.ts                - User data hooks
│   ├── useClients.ts              - Client data hooks
│   ├── useConfig.ts               - Config data hooks
│   └── useDashboard.ts            - Dashboard data hooks
├── services/
│   ├── userService.ts             - User API calls
│   ├── clientService.ts           - Client API calls
│   ├── configService.ts           - Config API calls
│   └── dashboardService.ts        - Dashboard API calls
└── types/
    ├── user.ts                    - User type definitions
    ├── client.ts                  - Client type definitions
    ├── config.ts                  - Config type definitions
    └── dashboard.ts               - Dashboard type definitions
```

### Phase 2: Dashboard Implementation (Week 2)

#### 2.1 Dashboard Statistics
- **Total Users** (active/inactive counts)
- **Total Clients** (active/inactive counts)
- **Total Tokens** (active tokens, tokens issued today/week)
- **Total Sessions** (active sessions)
- **Authorization Requests** (last 24h/7d/30d)
- **Token Issuance Rate** (chart)

#### 2.2 Recent Activity
- Recent user logins
- Recent client registrations
- Recent token issuances
- Recent authorization grants
- Failed authentication attempts

#### 2.3 System Health
- Storage backend status (JSON file/MongoDB)
- Certificate expiry warnings
- JWT signing key status
- Uptime and version info
- Memory/performance metrics (optional)

#### 2.4 Quick Actions
- Create new user (button)
- Create new client (button)
- View recent errors (button)
- Rotate certificates (button)

**Files to Implement:**
```
Backend:
- backend/pkg/handlers/admin_dashboard.go
- backend/pkg/models/dashboard.go

Frontend:
- frontend/src/pages/Dashboard.tsx (enhance existing)
- frontend/src/components/dashboard/StatCard.tsx
- frontend/src/components/dashboard/ActivityLog.tsx
- frontend/src/components/dashboard/HealthStatus.tsx
- frontend/src/components/dashboard/QuickActions.tsx
- frontend/src/hooks/useDashboard.ts
- frontend/src/services/dashboardService.ts
```

### Phase 3: User Management (Week 3)

#### 3.1 User List & Search
**Features:**
- Paginated table with sorting
- Search by: username, email, name
- Filter by: role (user/admin), status (active/disabled)
- Bulk actions (optional for v1)
- Export to CSV (optional)

**UI Components:**
- Search bar with filters
- Data table with columns: Username, Email, Name, Role, Status, Created, Actions
- Action menu: View, Edit, Delete, Enable/Disable

#### 3.2 Create User
**Fields:**
- Username (required, unique)
- Email (required, unique, validated)
- Password (required, strength validation)
- Confirm Password (required, must match)
- Name (optional)
- Given Name (optional)
- Family Name (optional)
- Role (user/admin)
- Status (active/disabled)
- Email Verified (checkbox)
- Address fields (optional)
- Picture URL (optional)

**Validation:**
- Username: alphanumeric + underscore/hyphen, 3-50 chars
- Email: valid email format
- Password: min 8 chars, complexity rules
- Duplicate checks

#### 3.3 View User Detail
**Display:**
- User information (all fields)
- User metadata (created_at, updated_at)
- Active sessions count
- Active tokens count
- Recent activity log
- Associated clients (optional)

#### 3.4 Edit User Detail
**Editable Fields:**
- Email
- Name, Given Name, Family Name
- Role
- Picture
- Address
- Email Verified flag

**Non-editable:**
- Username (immutable)
- User ID
- Created date

#### 3.5 Change Password
- Separate modal/page
- New password field
- Confirm password field
- Password strength indicator
- Option to force password change on next login

#### 3.6 Change Email
- Separate modal
- New email field
- Confirmation required
- Email validation
- Option to send verification email

#### 3.7 Enable/Disable User
- Toggle status
- Confirmation dialog
- When disabled:
  - User cannot login
  - All sessions invalidated
  - All tokens revoked

#### 3.8 Delete User
- Confirmation dialog with username verification
- Cascade options:
  - Revoke all tokens
  - Delete all sessions
  - Delete all consents
- Soft delete vs hard delete option

**Files to Implement:**
```
Backend:
- backend/pkg/handlers/admin_users.go
- backend/pkg/models/admin.go (request/response models)
- backend/pkg/storage/json.go (add user management methods)
- backend/pkg/storage/mongodb.go (add user management methods)

Frontend:
- frontend/src/pages/Users.tsx (enhance existing)
- frontend/src/pages/UserDetail.tsx
- frontend/src/components/users/UserList.tsx
- frontend/src/components/users/UserSearch.tsx
- frontend/src/components/users/UserForm.tsx
- frontend/src/components/users/UserDetail.tsx
- frontend/src/components/users/UserActions.tsx
- frontend/src/hooks/useUsers.ts
- frontend/src/services/userService.ts
- frontend/src/types/user.ts
```

### Phase 4: Client Management (Week 4)

#### 4.1 Client List & Filter
**Features:**
- Paginated table with sorting
- Search by: client_id, client_name
- Filter by: grant_types, response_types, status
- Display key info: ID, Name, Grant Types, Created, Status

#### 4.2 Create Client
**Basic Fields:**
- Client Name (required)
- Application Type (web/native)
- Redirect URIs (required, multi-value)
- Grant Types (checkboxes: authorization_code, implicit, client_credentials, refresh_token, password)
- Response Types (checkboxes: code, token, id_token)
- Scope (default: openid profile email)

**Advanced Fields:**
- Token Endpoint Auth Method
- JWKS URI or JWKS JSON
- Logo URI
- Client URI
- Policy URI
- Terms of Service URI
- Contacts (multi-value)

**Generated:**
- Client ID (UUID)
- Client Secret (for confidential clients)

#### 4.3 View Client Detail
**Display:**
- All client metadata
- Client secret (masked, with "Show" button)
- Registration date
- Last used date (optional)
- Active tokens count
- Recent authorizations

#### 4.4 Edit Client Detail
**Editable:**
- Client Name
- Redirect URIs
- Grant Types
- Response Types
- Scope
- All metadata fields

**Non-editable:**
- Client ID
- Registration date

#### 4.5 Regenerate Secret
- Button to regenerate
- Confirmation dialog with warning
- Display new secret (only once)
- Option to download/copy
- Old secret invalidated immediately

#### 4.6 Enable/Disable Client
- Toggle status
- Confirmation dialog
- When disabled:
  - Cannot request new tokens
  - Existing tokens remain valid (or option to revoke)
  - Cannot use any grant flows

#### 4.7 Delete Client
- Confirmation dialog with client_id verification
- Options:
  - Revoke all tokens
  - Delete all authorization codes
- Soft delete vs hard delete option

**Files to Implement:**
```
Backend:
- backend/pkg/handlers/admin_clients.go (enhance admin.go)
- backend/pkg/models/admin.go (client request/response models)

Frontend:
- frontend/src/pages/Clients.tsx (enhance existing)
- frontend/src/pages/ClientDetail.tsx
- frontend/src/components/clients/ClientList.tsx
- frontend/src/components/clients/ClientSearch.tsx
- frontend/src/components/clients/ClientForm.tsx
- frontend/src/components/clients/ClientDetail.tsx
- frontend/src/components/clients/ClientActions.tsx
- frontend/src/hooks/useClients.ts
- frontend/src/services/clientService.ts
- frontend/src/types/client.ts
```

### Phase 5: Configuration Management (Week 5)

#### 5.1 View All Configuration
**Sections:**
- **Server Configuration**
  - Issuer URL
  - Port
  - Base URL
  
- **Storage Configuration**
  - Type (JSON/MongoDB)
  - Connection details (masked)
  - Status indicator
  
- **JWT Configuration**
  - Expiry duration
  - Signing algorithm
  - Current key ID
  
- **Feature Flags**
  - Dynamic registration enabled
  - PKCE required
  - Implicit flow enabled
  - Password grant enabled
  
- **Session Configuration**
  - Session timeout
  - Cookie settings
  - CSRF protection settings

#### 5.2 View All Certificates
**Display:**
- Certificate list table:
  - Key ID
  - Algorithm (RS256, etc.)
  - Created date
  - Expiry date
  - Status (active/expired/rotating)
  - Actions (View, Set Active, Delete)
  
- Certificate details:
  - Public key (PEM format)
  - Fingerprint
  - Key size
  - Usage count (optional)

#### 5.3 Rotate Certificate
**Process:**
1. Generate new key pair
2. Assign new key ID
3. Add to JWKS with both old and new keys
4. Set new key as default for signing
5. Keep old key for verification (grace period)
6. Schedule old key removal

**UI:**
- Button to initiate rotation
- Confirmation dialog
- Progress indicator
- Grace period configuration (default 24h)
- Success notification with new key ID

#### 5.4 Purge Expired Certificates
**Features:**
- List expired certificates
- Confirm which to delete
- Warning if tokens still exist signed by these keys
- Batch delete operation

#### 5.5 Full Configuration Editor
**Features:**
- JSON editor with syntax highlighting
- Schema validation
- Diff view (before/after)
- Confirmation with validation
- Backup before save
- Rollback capability

**Safety Features:**
- Cannot break running system
- Validation before save
- Restart notification if required
- Export/import configuration

**Files to Implement:**
```
Backend:
- backend/pkg/handlers/admin_config.go
- backend/pkg/models/admin_config.go
- backend/pkg/crypto/rotation.go (key rotation logic)
- backend/pkg/configstore/editor.go (config editing logic)

Frontend:
- frontend/src/pages/Settings.tsx (enhance existing)
- frontend/src/components/config/ConfigView.tsx
- frontend/src/components/config/ConfigEditor.tsx
- frontend/src/components/config/CertificateList.tsx
- frontend/src/components/config/CertificateRotation.tsx
- frontend/src/components/config/JWKSViewer.tsx
- frontend/src/hooks/useConfig.ts
- frontend/src/services/configService.ts
- frontend/src/types/config.ts
```

### Phase 6: Common Components & UX (Week 6)

#### 6.1 Reusable Components

**DataTable Component:**
- Features: sorting, pagination, row selection, column visibility
- Props: data, columns, loading, onSort, onPageChange, actions
- Support for custom renderers

**SearchBar Component:**
- Search input with debouncing
- Filter dropdowns
- Clear filters button
- Export results button

**ConfirmDialog Component:**
- Customizable title, message, buttons
- Optional input field (for verification)
- Async action support
- Loading state

**FormField Component:**
- Label, input, validation message
- Support for text, email, password, select, checkbox, textarea
- Error state styling

**StatusBadge Component:**
- Color-coded status display
- Predefined statuses: active, inactive, disabled, expired, warning

#### 6.2 Navigation Enhancement

**Sidebar Navigation:**
```
📊 Dashboard
👥 User Management
   - List Users
   - Create User
🔐 Client Management
   - List Clients
   - Create Client
⚙️  Configuration
   - View Config
   - Certificates
   - JWKS
🚪 Logout
```

**Breadcrumbs:**
- Show current location
- Clickable navigation path

#### 6.3 Notifications & Feedback

**Toast Notifications:**
- Success messages (green)
- Error messages (red)
- Warning messages (orange)
- Info messages (blue)

**Loading States:**
- Skeleton loaders for tables
- Spinner for actions
- Progress bars for long operations

#### 6.4 Responsive Design

**Breakpoints:**
- Mobile: < 768px
- Tablet: 768px - 1024px
- Desktop: > 1024px

**Mobile Optimizations:**
- Collapsible sidebar
- Stacked forms
- Touch-friendly buttons
- Simplified tables (card view)

### Phase 7: Security & Validation (Week 7)

#### 7.1 Role-Based Access Control

**Permissions:**
```go
const (
    PermViewUsers      = "users:view"
    PermCreateUsers    = "users:create"
    PermEditUsers      = "users:edit"
    PermDeleteUsers    = "users:delete"
    PermViewClients    = "clients:view"
    PermCreateClients  = "clients:create"
    PermEditClients    = "clients:edit"
    PermDeleteClients  = "clients:delete"
    PermViewConfig     = "config:view"
    PermEditConfig     = "config:edit"
    PermRotateKeys     = "config:rotate"
)
```

**Implementation:**
- Admin role gets all permissions
- Regular users get read-only permissions
- UI components check permissions before rendering
- Backend enforces permissions on all admin endpoints

#### 7.2 Input Validation

**Frontend:**
- Real-time validation as user types
- Clear error messages
- Prevent submission with invalid data

**Backend:**
- Validate all inputs
- Sanitize strings
- Check business rules
- Return detailed validation errors

#### 7.3 Audit Logging

**Log Events:**
- User created/updated/deleted
- Client created/updated/deleted
- Configuration changes
- Certificate rotations
- Failed admin login attempts
- Permission denials

**Log Format:**
```go
type AuditLog struct {
    ID         string    `json:"id"`
    Timestamp  time.Time `json:"timestamp"`
    AdminID    string    `json:"admin_id"`
    Action     string    `json:"action"`
    Resource   string    `json:"resource"`
    ResourceID string    `json:"resource_id"`
    Details    string    `json:"details"`
    IPAddress  string    `json:"ip_address"`
    UserAgent  string    `json:"user_agent"`
}
```

### Phase 8: Testing & Documentation (Week 8)

#### 8.1 Backend Testing

**Unit Tests:**
- Admin handler tests
- User management CRUD tests
- Client management CRUD tests
- Configuration editing tests
- Certificate rotation tests

**Integration Tests:**
- Full admin workflow tests
- Permission enforcement tests
- Cascade delete tests

#### 8.2 Frontend Testing

**Component Tests:**
- User form validation
- Client form validation
- Data table functionality
- Search and filtering

**E2E Tests (Playwright/Cypress):**
- User creation flow
- Client creation flow
- Certificate rotation flow
- Configuration editing flow

#### 8.3 Documentation

**Admin Guide:**
- How to access admin UI
- User management walkthrough
- Client management walkthrough
- Configuration management walkthrough
- Certificate rotation best practices
- Troubleshooting guide

**API Documentation:**
- All admin endpoints documented
- Request/response examples
- Error codes and messages
- Authentication requirements

**Developer Guide:**
- Admin UI architecture
- Adding new admin features
- Permission system
- Audit logging

## Implementation Priority

### Must-Have (MVP)
1. ✅ Dashboard with basic stats
2. ✅ User list and search
3. ✅ Create/edit/delete users
4. ✅ Client list and search
5. ✅ Create/edit/delete clients
6. ✅ View configuration
7. ✅ Regenerate client secret

### Should-Have (v1.1)
1. Enhanced dashboard with charts
2. Certificate rotation UI
3. Advanced user filters
4. Advanced client filters
5. Configuration editor
6. Audit logging viewer
7. Email verification flow

### Nice-to-Have (v1.2)
1. Bulk operations
2. Export to CSV
3. Configuration import/export
4. Activity charts and graphs
5. Real-time metrics
6. Mobile app (PWA)
7. Dark mode

## Technical Decisions

### UI Framework
**Recommendation: Ant Design (antd)**
- Pros: Enterprise-ready, comprehensive components, TypeScript support, good documentation
- Cons: Larger bundle size
- Alternative: Material-UI (if prefer Material Design)

### State Management
**Recommendation: React Query + Context**
- React Query for server state (caching, refetching)
- Context for UI state (theme, notifications)
- Avoid Redux for this scope

### Form Handling
**Recommendation: React Hook Form**
- Excellent performance
- Easy validation
- TypeScript support
- Small bundle size

### Data Visualization
**Recommendation: Recharts**
- Simple API
- Responsive
- Good TypeScript support
- Alternative: Chart.js

## Estimated Timeline

| Phase | Duration | Deliverables |
|-------|----------|-------------|
| Phase 1: Foundation | 1 week | API structure, UI setup |
| Phase 2: Dashboard | 1 week | Enhanced dashboard |
| Phase 3: User Management | 1 week | Complete user CRUD |
| Phase 4: Client Management | 1 week | Complete client CRUD |
| Phase 5: Configuration | 1 week | Config viewer, cert rotation |
| Phase 6: Common Components | 1 week | Reusable UI components |
| Phase 7: Security | 1 week | RBAC, audit logging |
| Phase 8: Testing | 1 week | Tests and documentation |
| **Total** | **8 weeks** | **Production-ready admin UI** |

## Success Metrics

1. **Functionality**: All CRUD operations working
2. **Performance**: < 2s page load, < 500ms API responses
3. **Usability**: < 3 clicks to complete common tasks
4. **Security**: All admin actions require authentication + authorization
5. **Reliability**: 99.9% uptime, proper error handling
6. **Test Coverage**: > 80% backend, > 60% frontend

## Next Steps

1. **Review and approve this plan**
2. **Set up project tracking** (GitHub Projects/Jira)
3. **Create feature branches** for each phase
4. **Start with Phase 1** (Foundation & Infrastructure)
5. **Regular demos** at end of each phase
6. **Gather feedback** and adjust plan as needed

## Questions to Answer

1. **UI Framework**: Ant Design or Material-UI?
2. **Storage**: Will MongoDB support be required?
3. **Deployment**: Docker, systemd, or cloud-native?
4. **Multi-tenancy**: Single tenant or multi-tenant support?
5. **Email**: SMTP configuration for notifications?
6. **Backup**: Automated backup strategy?
7. **Monitoring**: Integration with monitoring tools (Prometheus, Grafana)?

---

**Document Version**: 1.0  
**Created**: October 27, 2025  
**Status**: Awaiting Approval
