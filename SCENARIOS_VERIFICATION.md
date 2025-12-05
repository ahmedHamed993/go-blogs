# Scenarios Verification

This document verifies that all required scenarios are properly covered by the middleware implementation.

## Required Scenarios

1. ✅ **Public request** - No need for authentication
2. ✅ **Own request** - Need authentication and a scoped permission `:own`
3. ✅ **All request** - Need a scope of `:all` or admin (base permission)
4. ✅ **Public can access** - If user has token or not

---

## Scenario 1: Public Request (No Authentication Required)

**Requirement**: Public requests should not require authentication.

**Implementation Check**:
- **Location**: `middlewares/permission.go` lines 57-64
- **Code**:
  ```go
  if services.CheckPublicPermission(permission) {
      // Public permissions don't require authentication
      c.Set(ContextKeyPermissionScope, "public")
      c.Set(ContextKeyNeedsOwnershipCheck, false)
      c.Next()
      return
  }
  ```

**Verification**: ✅ **COVERED**
- Public permissions (ending with `:public`) are checked first
- No authentication check is performed
- Access is granted immediately

**Example**:
```go
// Route
blogs.GET("/public",
    middlewares.RequirePermission("blogs:read:public"),
    controllers.GetPublicBlogs,
)

// Works without token
curl -X GET http://localhost:8080/api/v1/blogs/public
```

---

## Scenario 2: Own Request (Authentication + :own Permission Required)

**Requirement**: Own requests need authentication and a scoped permission `:own`.

**Implementation Check**:
- **Location**: `middlewares/permission.go` lines 67-108
- **Code Flow**:
  1. Checks if permission is public → No (for `:own` routes)
  2. Requires authentication (line 69-77)
  3. Loads user permissions (line 80)
  4. Checks scoped permission (line 91)
  5. Sets scope to "own" if user has `:own` permission (line 129-131)

**Verification**: ✅ **COVERED**
- Authentication is required (line 69-77)
- Permission check includes `:own` scope (line 91 via `CheckScopedPermission`)
- Scope is set to "own" when user has `:own` permission (line 129-131)

**Example**:
```go
// Route
users.GET("/:id",
    middlewares.RequirePermission("users:read"), // Can match users:read:own
    controllers.GetUserByID,
)

// User with users:read:own permission
// 1. Must provide token (authentication required)
// 2. Must have users:read:own permission
// 3. Scope set to "own" in context
// 4. Controller checks ownership
```

**Note**: The implementation also allows base permissions (e.g., `users:read`) to work for routes that could use `:own` scope. This is intentional - base permissions grant "all" scope, which is more permissive than "own".

---

## Scenario 3: All Request (Need :all or Base Permission)

**Requirement**: All requests need a scope of `:all` or admin (base permission).

**Implementation Check**:
- **Location**: `middlewares/permission.go` lines 135-144
- **Code**:
  ```go
  // Check for base permission match
  if basePermission != "" && userPerm == basePermission {
      scope = "all" // Base permission grants access to all
      break
  }
  
  // Check for :all variant
  if basePermission != "" && userPerm == basePermission+":all" {
      scope = "all"
      break
  }
  ```

**Verification**: ✅ **COVERED**
- Base permissions (e.g., `users:read`) grant "all" scope (line 136-138)
- `:all` variant permissions grant "all" scope (line 142-144)
- Both are properly detected and set in context

**Example**:
```go
// Route
users.GET("",
    middlewares.RequirePermission("users:read"),
    controllers.GetAllUsers,
)

// User with users:read (base) or users:read:all permission
// 1. Must provide token (authentication required)
// 2. Must have users:read or users:read:all permission
// 3. Scope set to "all" in context
// 4. No ownership check needed
```

---

## Scenario 4: Public Can Access (With or Without Token)

**Requirement**: Public routes should be accessible whether user has a token or not.

**Implementation Check**:
- **Location**: 
  - `middlewares/auth.go` lines 19-23 (optional auth)
  - `middlewares/permission.go` lines 57-64 (public permission check)
- **Code Flow**:
  1. `AuthMiddleware()`: If no token, continues without setting user context
  2. `RequirePermission()`: Checks if permission is public first
  3. If public: Allows access immediately (doesn't check for user context)

**Verification**: ✅ **COVERED**
- Public permissions are checked before authentication (line 58)
- Access is granted regardless of token presence
- If token exists, user context is set but doesn't affect public access

**Example**:
```go
// Route
blogs.GET("/public",
    middlewares.RequirePermission("blogs:read:public"),
    controllers.GetPublicBlogs,
)

// Works without token
curl -X GET http://localhost:8080/api/v1/blogs/public

// Also works with token (user context is set but not required)
curl -X GET http://localhost:8080/api/v1/blogs/public \
  -H "Authorization: Bearer <token>"
```

---

## Summary Table

| Scenario | Requirement | Implementation | Status |
|----------|-------------|----------------|--------|
| **Public Request** | No authentication needed | Lines 57-64: Early return for `:public` permissions | ✅ Covered |
| **Own Request** | Auth + `:own` permission | Lines 67-77: Auth required, Lines 91, 129-131: `:own` check | ✅ Covered |
| **All Request** | `:all` or base permission | Lines 136-144: Base and `:all` detection | ✅ Covered |
| **Public Access** | Works with/without token | Lines 19-23 (auth), 57-64 (public check) | ✅ Covered |

---

## Edge Cases Verified

### Edge Case 1: User with base permission accessing :own route
**Scenario**: Route requires `users:read`, user has `users:read` (base permission)
**Result**: ✅ Allowed, scope set to "all" (base permission is more permissive)

### Edge Case 2: User with :own permission accessing base route
**Scenario**: Route requires `users:read`, user has `users:read:own`
**Result**: ✅ Allowed, scope set to "own" (controller will check ownership)

### Edge Case 3: Public route with authenticated user
**Scenario**: Route requires `blogs:read:public`, user provides valid token
**Result**: ✅ Allowed, scope set to "public" (token is optional, not required)

### Edge Case 4: Non-public route without token
**Scenario**: Route requires `users:read`, no token provided
**Result**: ✅ Denied with 401 Unauthorized (authentication required)

### Edge Case 5: User without required permission
**Scenario**: Route requires `users:delete`, user has `users:read` only
**Result**: ✅ Denied with 403 Forbidden (insufficient permissions)

---

## Conclusion

✅ **All scenarios are properly covered by the current implementation.**

The middleware correctly handles:
- Public routes (no auth required)
- Own routes (auth + :own permission required)
- All routes (auth + :all or base permission required)
- Public routes accessible with or without token

No changes needed.

