# Scope-Based Claims Implementation Summary

## Overview
Implemented proper OpenID Connect scope-based claims filtering according to OIDC Core 1.0 Section 5.4. ID tokens and UserInfo responses now only include claims for explicitly requested scopes during authentication.

## Implementation Date
October 25, 2025

## Key Changes

### 1. User Model Extensions (`backend/pkg/models/models.go`)

#### New Address Struct
```go
type Address struct {
    Formatted     string `json:"formatted,omitempty"`      // Full mailing address
    StreetAddress string `json:"street_address,omitempty"` // Street address
    Locality      string `json:"locality,omitempty"`       // City/locality
    Region        string `json:"region,omitempty"`         // State/region
    PostalCode    string `json:"postal_code,omitempty"`    // Postal code
    Country       string `json:"country,omitempty"`        // Country
}
```

#### Updated User Struct
- Added `EmailVerified bool` field
- Added `Address *Address` field

### 2. JWT Claims (`backend/pkg/crypto/jwt.go`)

#### Enhanced IDTokenClaims
- Added `EmailVerified bool` field
- Added `Address *models.Address` field

#### Updated Token Generation Methods
Both methods now accept a `scope` parameter:

```go
func (jm *JWTManager) GenerateIDToken(user *models.User, clientID, nonce, scope string) (string, error)

func (jm *JWTManager) GenerateIDTokenWithClaims(user *models.User, clientID, nonce, scope string, 
    authTime time.Time, acr string, amr []string, accessToken, authCode string) (string, error)
```

#### New Scope Filtering Logic
Added `applyScopes()` helper function that filters claims based on requested scopes:

- **profile scope**: Includes `name`, `given_name`, `family_name`, `picture`
- **email scope**: Includes `email`, `email_verified`
- **address scope**: Includes `address` object (if user has address data)
- **No scope**: Only includes required claims (`sub`, `iss`, `aud`, `exp`, `iat`, `nonce`)

### 3. UserInfo Endpoint (`backend/pkg/handlers/userinfo.go`)

#### Updated UserInfoResponse
- Added `Address *models.Address` field

#### Enhanced buildUserInfoResponse()
- Checks for `address` scope before including address claims
- Uses actual `user.EmailVerified` value instead of hardcoding `true`
- Only includes claims for granted scopes

### 4. Token Generation Call Sites

Updated all token generation to pass scope parameter:

#### `backend/pkg/handlers/token.go`
- **Authorization Code Flow**: `generateIDTokenForAuthCode()` passes `authCode.Scope`
- **Refresh Token Flow**: Passes `oldToken.Scope`
- **Password Grant Flow**: Passes requested `scope`

#### `backend/pkg/handlers/authorize.go`
- **Implicit Flow**: Passes `authSession.Scope`

## Scope Behavior

### Profile Scope (`profile`)
When requested, includes:
- `name` - Full name
- `given_name` - First name
- `family_name` - Last name
- `picture` - Profile picture URL

### Email Scope (`email`)
When requested, includes:
- `email` - Email address
- `email_verified` - Email verification status

### Address Scope (`address`)
When requested, includes:
- `address` object with all address fields
- Only included if user has address data (null otherwise)

### Example Behaviors

#### Request: `openid profile`
```json
{
  "sub": "user123",
  "name": "John Doe",
  "given_name": "John",
  "family_name": "Doe",
  "picture": "https://example.com/pic.jpg"
  // No email or address
}
```

#### Request: `openid email`
```json
{
  "sub": "user123",
  "email": "john@example.com",
  "email_verified": true
  // No profile or address
}
```

#### Request: `openid profile email address`
```json
{
  "sub": "user123",
  "name": "John Doe",
  "given_name": "John",
  "family_name": "Doe",
  "email": "john@example.com",
  "email_verified": true,
  "address": {
    "formatted": "123 Main St, City, ST 12345, USA",
    "street_address": "123 Main St",
    "locality": "City",
    "region": "ST",
    "postal_code": "12345",
    "country": "USA"
  }
}
```

## Testing

### New Test File: `scope_claims_test.go`
Created 9 comprehensive test cases:

1. **TestIDToken_ProfileScope** - Verifies only profile claims included
2. **TestIDToken_EmailScope** - Verifies only email claims included
3. **TestIDToken_AddressScope** - Verifies only address claims included
4. **TestIDToken_MultipleScopesAtOnce** - Verifies combined scopes work
5. **TestIDToken_NoScopes** - Verifies only required claims with no optional scopes
6. **TestIDToken_AddressScopeWithoutAddressData** - Verifies nil address handling
7. **TestUserInfo_AddressScope** - Verifies UserInfo includes address with scope
8. **TestUserInfo_WithoutAddressScope** - Verifies UserInfo excludes address without scope
9. **TestUserInfo_WithoutEmailScope** - Verifies UserInfo excludes email without scope

### Test Results
✅ All 46 handler tests passing  
✅ All 7 crypto tests passing  
✅ All 5 model tests passing  
✅ All 3 session tests passing  

**Total: 61 tests, all passing**

## Compliance

### OpenID Connect Core 1.0
- ✅ Section 5.4 - Standard Claims (scope-based filtering)
- ✅ Section 5.1 - UserInfo Endpoint (proper claim filtering)
- ✅ Section 5.1.1 - Address Claim (proper structure)
- ✅ Section 3.1.3.3 - ID Token (proper claim inclusion)

### Privacy & Security
- ✅ Users control what data is shared via scope consent
- ✅ Minimal data disclosure (only requested claims)
- ✅ GDPR/CCPA friendly (data minimization)
- ✅ No PII leaked without explicit consent

## Migration Notes

### For Existing Deployments

#### Database Schema
No schema migration required. New fields are optional:
- `User.EmailVerified` defaults to `false`
- `User.Address` defaults to `nil`

#### Backward Compatibility
- Existing ID tokens continue to work
- Old clients without scope parameters still function
- No breaking changes to existing APIs

#### Recommended Actions
1. Update user records to set `EmailVerified` appropriately
2. Add address data for users who need it
3. Update client applications to request specific scopes
4. Update consent UI to show which claims each scope includes

## Future Enhancements

### Potential Additions
- [ ] Phone scope (`phone`, `phone_number_verified`)
- [ ] Custom claims based on additional scopes
- [ ] Locale support for address formatting
- [ ] Address verification system
- [ ] Granular consent per claim (not just per scope)

## References

- [OIDC Core 1.0 - Section 5.4 (Standard Claims)](https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims)
- [OIDC Core 1.0 - Section 5.1.1 (Address Claim)](https://openid.net/specs/openid-connect-core-1_0.html#AddressClaim)
- [RFC 6749 - OAuth 2.0 Authorization Framework](https://tools.ietf.org/html/rfc6749)

## Commit Information

- **Commit**: afa9e04
- **Branch**: main
- **Files Changed**: 8 files
- **Lines Added**: 616
- **Lines Removed**: 53
- **Status**: ✅ Pushed to origin/main
