# Go Blogs - Authentication & Authorization Guide

A comprehensive guide to understanding the authentication and authorization system in this Go application. This guide is designed for beginners and explains every step of the process in detail.

## Table of Contents

1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [Authentication Flow](#authentication-flow)
4. [Authorization Flow](#authorization-flow)
5. [Permission System](#permission-system)
6. [Code Walkthrough](#code-walkthrough)
7. [Usage Examples](#usage-examples)
8. [Testing the System](#testing-the-system)

---

## Overview

This application implements a **Role-Based Access Control (RBAC)** system with **scoped permissions**. It uses:

- **JWT (JSON Web Tokens)** for authentication
- **RBAC (Role-Based Access Control)** for authorization
- **Scoped Permissions** (`:own`, `:all`, `:public`) for fine-grained access control

### Key Concepts

- **Authentication**: Verifying who you are (login)
- **Authorization**: Verifying what you can do (permissions)
- **Role**: A collection of permissions (e.g., admin, user)
- **Permission**: A specific action on a resource (e.g., `users:read`)
- **Scope**: Limits access to resources (`:own` = only your resources, `:all` = all resources, `:public` = no auth required)

---

## Project Structure

```
go-blogs/
├── controllers/          # Request handlers
│   ├── auth.go          # Login, Register
│   └── users.go         # User CRUD operations
├── middlewares/          # Request interceptors
│   ├── auth.go          # JWT token validation
│   ├── permission.go    # Permission checking
│   └── error-handler.go # Error handling
├── models/               # Database models
│   ├── user.go
│   ├── role.go
│   ├── permission.go
│   └── role_permission.go
├── routes/               # Route definitions
│   ├── auth.go
│   └── user.go
├── services/             # Business logic
│   ├── jwt.go           # Token generation
│   ├── password.go      # Password hashing
│   ├── permission.go    # Permission checking logic
│   └── response.go      # Response helpers
├── seeders/              # Database seeders
│   ├── permissions.go
│   ├── roles.go
│   └── seeder.go
└── database/            # Database connection
    └── database.go
```

---

## Authentication Flow

Authentication is the process of verifying a user's identity. Here's the complete flow:

### Step 1: User Registration

**Endpoint**: `POST /api/v1/auth/register`

**Flow**:
1. User sends registration request with `username`, `password`, and `role_id`
2. Server validates input
3. Server hashes the password (never store plain passwords!)
4. Server creates user in database
5. Server returns success response

**Code Location**: `controllers/auth.go` → `Register()`

**Example Request**:
```json
{
  "username": "john_doe",
  "password": "securepassword123",
  "role_id": 3
}
```

### Step 2: User Login

**Endpoint**: `POST /api/v1/auth/login`

**Flow**:
1. User sends login request with `username` and `password`
2. Server finds user by username in database
3. Server compares provided password with stored hash
4. If password matches:
   - Server generates JWT token containing `user_id` and `role_id`
   - Server returns token to client
5. Client stores token (usually in localStorage or cookies)

**Code Location**: `controllers/auth.go` → `Login()`

**Example Request**:
```json
{
  "username": "john_doe",
  "password": "securepassword123"
}
```

**Example Response**:
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Step 3: JWT Token Structure

The JWT token contains:
- **user_id**: The authenticated user's ID
- **role_id**: The user's role ID (determines permissions)
- **exp**: Token expiration time (72 hours)

**Code Location**: `services/jwt.go` → `GenerateToken()`

### Step 4: Token Validation (Auth Middleware)

**Middleware**: `middlewares/auth.go` → `AuthMiddleware()`

**Flow** (runs on every protected route):
1. Client sends request with token in `Authorization: Bearer <token>` header
2. Middleware extracts token from header
3. Middleware validates token signature
4. If valid:
   - Middleware extracts `user_id` and `role_id` from token
   - Middleware stores them in request context
   - Middleware loads user permissions from database
   - Middleware stores permissions in context
   - Request continues to next handler
5. If invalid:
   - Middleware returns 401 Unauthorized
   - Request is aborted

**Visual Flow**:
```
Client Request
    ↓
[Auth Middleware]
    ├─ Extract token from header
    ├─ Validate token signature
    ├─ Extract user_id & role_id
    ├─ Load permissions from DB
    └─ Store in context
    ↓
[Permission Middleware] (if needed)
    ↓
[Controller]
    ↓
Response
```

---

## Authorization Flow

Authorization determines what actions a user can perform. This happens after authentication.

### Step 1: Permission Loading

When a user authenticates, their permissions are loaded based on their role:

**Code Location**: `services/permission.go` → `LoadUserPermissions()`

**Process**:
1. Query `role_permissions` table joined with `permissions` table
2. Filter by user's `role_id`
3. Return list of permission names (e.g., `["users:read", "blogs:create"]`)

**Database Query**:
```sql
SELECT permissions.name
FROM role_permissions
JOIN permissions ON role_permissions.permission_id = permissions.id
WHERE role_permissions.role_id = ?
```

### Step 2: Permission Checking

When a protected route is accessed, the permission middleware checks if the user has the required permission.

**Code Location**: `middlewares/permission.go`

**Available Middleware Functions**:

#### a) `RequirePermission(permission string)`

Checks if user has a specific permission. Used for routes that don't need ownership checks.

**Example**: `GET /api/v1/users` requires `users:read`

**Flow**:
1. Check if permission is public (ends with `:public`)
2. Load user permissions from context (or database)
3. Check if user has the permission
4. If yes: allow request
5. If no: return 403 Forbidden

#### b) `RequirePermissionWithScope(permission string)`

**What It Does**: Checks permission and sets scope information in the request context. The controller can then use this scope information to perform ownership checks if needed. This middleware separates permission checking from ownership verification, giving controllers full control over how to extract and verify resource ownership.

**When to Use**: Use this middleware when:
- The route needs to support both `:own` and base permissions (e.g., `users:read:own` OR `users:read`)
- You want controllers to handle ownership checks with full access to request data (params, query, body)
- You need flexibility in how resource IDs are extracted and compared

**Complete Flow**:
1. **Check if permission is public** - If `:public`, set scope to `"public"` and allow access
2. **Load user permissions** - Get all permissions for the authenticated user's role
3. **Check scoped permission** - Verify user has the required permission (handles `:own`, `:all`, and base permissions)
4. **Determine actual scope** - Check user's permissions to determine their actual scope:
   - `"all"` - User has base permission or `:all` variant (can access all resources)
   - `"own"` - User has `:own` variant (can only access their own resources)
   - `"public"` - User has `:public` variant (public access)
5. **Set scope in context** - Store scope information for controller to use:
   - `ContextKeyPermissionScope` - The actual scope: `"all"`, `"own"`, or `"public"`
   - `ContextKeyNeedsOwnershipCheck` - Boolean indicating if ownership check is needed
6. **Allow access** - Proceed to controller, which will handle ownership verification if needed

**Example**: `GET /api/v1/users/:id` requires `users:read` or `users:read:own`

#### c) `AllowPublic(permission string)`

Allows access without authentication for public permissions.

**Example**: `GET /api/v1/blogs/public` with `blogs:read:public`

#### d) `RequirePermissionOrPublic(permission string)`

Allows access if user has permission OR if permission is public.

### Step 3: Permission Matching Logic

**Code Location**: `services/permission.go` → `CheckScopedPermission()`

The system uses intelligent permission matching:

1. **Exact Match**: User has `users:read` and route requires `users:read` → ✅ Allow
2. **Base Permission**: User has `users:read` and route requires `users:read:own` → ✅ Allow (base permission grants all scopes)
3. **Scope Variant**: User has `users:read:own` and route requires `users:read` on their own resource → ✅ Allow (with ownership check)
4. **All Variant**: User has `blogs:read:all` and route requires `blogs:read` → ✅ Allow
5. **No Match**: User has `users:read:own` and route requires `users:read` on someone else's resource → ❌ Deny

---

## Permission System

### Permission Naming Convention

Permissions follow this format: `resource:action[:scope]`

- **resource**: The resource type (e.g., `users`, `blogs`)
- **action**: The action (e.g., `read`, `create`, `update`, `delete`)
- **scope** (optional): Access scope (`:own`, `:all`, `:public`)

### Permission Types

#### 1. Base Permissions (No Scope)
- `users:read` - Can read all users
- `users:create` - Can create users
- `users:update` - Can update all users
- `users:delete` - Can delete all users

#### 2. Own Scope (`:own`)
- `users:read:own` - Can only read own profile
- `users:update:own` - Can only update own profile
- `blogs:read:own` - Can only read own blogs
- `blogs:update:own` - Can only update own blogs

#### 3. All Scope (`:all`)
- `blogs:read:all` - Can read all blogs
- `blogs:update:all` - Can update any blog
- `blogs:delete:all` - Can delete any blog

#### 4. Public Scope (`:public`)
- `blogs:read:public` - Can read public blogs without authentication

### Role-Permission Mapping

**Superadmin** (Full Access):
- `users:create`, `users:read`, `users:update`, `users:delete`
- `blogs:create`, `blogs:read:all`, `blogs:update:all`, `blogs:delete:all`

**Admin** (Limited Access):
- `users:read`, `users:update`
- `blogs:create`, `blogs:read:all`, `blogs:update:all`

**User** (Own Resources Only):
- `users:read:own`, `users:update:own`
- `blogs:create`, `blogs:read:own`, `blogs:update:own`, `blogs:delete:own`

**Code Location**: `seeders/roles.go`

---

## Code Walkthrough

### 1. Authentication Middleware (`middlewares/auth.go`)

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract token from Authorization header
        authHeader := c.GetHeader("Authorization")
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // 2. Parse and validate JWT token
        token, err := jwt.Parse(tokenString, ...)

        // 3. Extract user_id and role_id from token
        claims := token.Claims.(jwt.MapClaims)
        userID := uint(claims["user_id"].(float64))
        roleID := uint(claims["role_id"].(float64))

        // 4. Store in context for use in controllers
        c.Set("user_id", userID)
        c.Set("role_id", roleID)

        // 5. Preload permissions (performance optimization)
        permissions, err := services.LoadUserPermissions(roleID)
        c.Set("permissions", permissions)

        c.Next() // Continue to next handler
    }
}
```

**What happens**:
- Runs before every protected route
- Validates JWT token
- Extracts user information
- Loads permissions once (cached in context)

### 2. Permission Service (`services/permission.go`)

#### Loading Permissions

```go
func LoadUserPermissions(roleID uint) ([]string, error) {
    // Query database to get all permissions for this role
    var permissionNames []string
    
    database.DB.
        Table("role_permissions").
        Select("permissions.name").
        Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
        Where("role_permissions.role_id = ?", roleID).
        Pluck("permissions.name", &permissionNames)
    
    return permissionNames, nil
}
```

#### Checking Permissions

```go
func CheckScopedPermission(userPermissions []string, requiredPermission string) (bool, bool, bool) {
    // Returns: (hasPermission, isPublic, needsOwnershipCheck)
    
    // 1. Check if it's public
    if strings.HasSuffix(requiredPermission, ":public") {
        return false, true, false // Public, no ownership check needed
    }
    
    // 2. Check exact match
    if HasPermission(userPermissions, requiredPermission) {
        if strings.HasSuffix(requiredPermission, ":all") {
            return true, false, false // Has permission, no ownership check
        }
        if strings.HasSuffix(requiredPermission, ":own") {
            return true, false, true // Has permission, needs ownership check
        }
        return true, false, false // Base permission, no ownership check
    }
    
    // 3. Check for variant permissions
    // ... (see code for full logic)
    
    return false, false, false // No permission
}
```

### 3. Permission Middleware (`middlewares/permission.go`)

#### RequirePermission

```go
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Check if public
        if services.CheckPublicPermission(permission) {
            c.Next()
            return
        }
        
        // 2. Load permissions from context
        userPermissions, err := loadUserPermissions(c)
        
        // 3. Check permission
        hasPermission, isPublic, _ := services.CheckScopedPermission(userPermissions, permission)
        
        // 4. Allow or deny
        if !hasPermission {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

#### RequirePermissionWithScope

```go
func RequirePermissionWithScope(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Check if permission is public
        if services.CheckPublicPermission(permission) {
            c.Set(ContextKeyPermissionScope, "public")
            c.Set(ContextKeyNeedsOwnershipCheck, false)
            c.Next()
            return
        }

        // 2. Load user permissions from context
        userPermissions, err := loadUserPermissions(c)
        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to load permissions"})
            c.Abort()
            return
        }

        // 3. Check scoped permission
        // Returns: (hasPermission, isPublic, needsOwnershipCheck)
        hasPermission, isPublic, needsOwnershipCheck := services.CheckScopedPermission(
            userPermissions, 
            permission
        )

        if isPublic {
            c.Set(ContextKeyPermissionScope, "public")
            c.Set(ContextKeyNeedsOwnershipCheck, false)
            c.Next()
            return
        }

        if !hasPermission {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        // 4. Determine the actual scope from user's permissions
        // Check what permission variant the user actually has
        var scope string
        parts := strings.Split(permission, ":")
        basePermission := ""
        if len(parts) >= 2 {
            basePermission = parts[0] + ":" + parts[1]
        }

        // Check user's permissions to determine their actual scope
        for _, userPerm := range userPermissions {
            // Check for exact match
            if userPerm == permission {
                if strings.HasSuffix(userPerm, ":public") {
                    scope = "public"
                    break
                } else if strings.HasSuffix(userPerm, ":all") {
                    scope = "all"
                    break
                } else if strings.HasSuffix(userPerm, ":own") {
                    scope = "own"
                    break
                }
            }
            
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
            
            // Check for :own variant
            if basePermission != "" && userPerm == basePermission+":own" {
                scope = "own"
                break
            }
        }

        // Set scope information in context for controller to use
        c.Set(ContextKeyPermissionScope, scope)        // "all", "own", or "public"
        c.Set(ContextKeyNeedsOwnershipCheck, needsOwnershipCheck)
        
        c.Next() // Proceed to controller
    }
}
```

**Key Points**:
- The middleware **only checks permissions** and sets scope information in context
- **No ID extraction** happens in the middleware - that's the controller's responsibility
- The scope is determined from the user's **actual permissions**, not the requested permission
- Controllers can access scope information via:
  - `c.Get(middlewares.ContextKeyPermissionScope)` - Returns `"all"`, `"own"`, or `"public"`
  - `c.Get(middlewares.ContextKeyNeedsOwnershipCheck)` - Returns `true` if ownership check is needed

### 4. Route Protection (`routes/user.go`)

```go
func UserRoutes(rg *gin.RouterGroup) {
    users := rg.Group("/users")
    {
        // List all users - requires users:read
        users.GET("",
            middlewares.AuthMiddleware(),           // 1. Authenticate
            middlewares.RequirePermission("users:read"), // 2. Check permission
            controllers.GetAllUsers)                // 3. Handle request
        
        // Get specific user - requires users:read with ownership check
        users.GET("/:id",
            middlewares.AuthMiddleware(),
            middlewares.RequirePermissionWithScope("users:read"), // Sets scope in context
            controllers.GetUserByID) // Controller handles ownership check
    }
}
```

**Flow for `GET /users/:id`**:
1. `AuthMiddleware()` validates token, loads permissions
2. `RequirePermissionWithScope()` checks:
   - Does user have `users:read` or `users:read:own`?
   - Determines actual scope from user's permissions (`"all"` or `"own"`)
   - Sets scope information in context
3. `GetUserByID()` controller:
   - Extracts user ID from URL parameter
   - Checks ownership if scope is `"own"`
   - Returns user data if access is allowed
4. If permission check fails → 403 Forbidden returned by middleware
5. If ownership check fails → 403 Forbidden returned by controller

---

## Understanding RequirePermissionWithScope in Detail

### What is RequirePermissionWithScope?

`RequirePermissionWithScope` is a middleware that **checks permissions** and **sets scope information** in the request context. Unlike `RequirePermission`, it determines the user's actual scope (`"all"`, `"own"`, or `"public"`) and makes this information available to controllers, which then handle ownership verification.

### Why This Approach?

**Separation of Concerns**:
- **Middleware** handles authentication and authorization (permission checking)
- **Controllers** handle business logic (ownership verification, ID extraction)

**Benefits**:
1. **Simpler middleware** - No need to extract IDs or compare ownership
2. **More flexible** - Controllers can extract IDs from any source (params, query, body, database)
3. **Cleaner routes** - No need to pass extraction functions
4. **Better control** - Controllers have full access to request data for complex ownership logic

### How It Works

The middleware:
1. Checks if the user has the required permission
2. Determines the user's **actual scope** by examining their permissions:
   - `"all"` - User has base permission (e.g., `users:read`) or `:all` variant
   - `"own"` - User has `:own` variant (e.g., `users:read:own`)
   - `"public"` - User has `:public` variant (e.g., `blogs:read:public`)
3. Sets scope information in context for controllers to use

The controller:
1. Reads scope information from context
2. Extracts resource ID from request (params, query, body, etc.)
3. Performs ownership check if scope is `"own"`
4. Allows or denies access accordingly

### Context Keys Set by Middleware

The middleware sets two values in the request context:

1. **`ContextKeyPermissionScope`** (`"permission_scope"`):
   - Value: `"all"`, `"own"`, or `"public"`
   - Indicates the user's actual permission scope

2. **`ContextKeyNeedsOwnershipCheck`** (`"needs_ownership_check"`):
   - Value: `true` or `false`
   - Indicates if ownership verification is needed (true when scope is `"own"`)

### Using Scope in Controllers

Controllers that use `RequirePermissionWithScope` middleware can access the scope from context to determine if ownership checks are needed.

**Basic Pattern**:
```go
func GetUserByID(c *gin.Context) {
    // Get scope from context (set by RequirePermissionWithScope middleware)
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    // Extract resource ID (from params, query, body, etc.)
    id := c.Param("id")
    
    // Check ownership if scope is "own"
    if exists && scope == "own" {
        currentUserID, exists := c.Get(middlewares.ContextKeyUserID)
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "User not authenticated",
            })
            return
        }
        
        currentUserIDUint := currentUserID.(uint)
        // Parse resource ID and compare
        // resourceID, _ := strconv.ParseUint(id, 10, 32)
        // if currentUserIDUint != uint(resourceID) {
        //     c.JSON(http.StatusForbidden, gin.H{
        //         "success": false,
        //         "error":   "Access denied: You can only access your own resources",
        //     })
        //     return
        // }
    }
    
    // Implement your business logic here
    // (fetch data, process, return response, etc.)
}
```

**Current Implementation Example**:

The controllers currently access the scope like this:

```go
// GetUserByID - Uses RequirePermissionWithScope("users:read")
func GetUserByID(c *gin.Context) {
    // Get scope from context
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    // Extract user ID from URL parameter
    id := c.Param("id")
    
    // Check ownership if scope is "own"
    if exists && scope == "own" {
        currentUserID, exists := c.Get(middlewares.ContextKeyUserID)
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "User not authenticated",
            })
            return
        }
        // Compare IDs and deny access if they don't match
    }
    
    // Your implementation here
}
```

```go
// UpdateUser - Uses RequirePermissionWithScope("users:update")
func UpdateUser(c *gin.Context) {
    // Get scope from context
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    // Extract user ID from URL parameter
    id := c.Param("id")
    
    // Check ownership if scope is "own"
    if exists && scope == "own" {
        currentUserID, exists := c.Get(middlewares.ContextKeyUserID)
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "User not authenticated",
            })
            return
        }
        // Compare IDs and deny access if they don't match
    }
    
    // Your implementation here
}
```

**Scope Values**:
- `"all"` - User has base permission or `:all` variant, can access all resources (no ownership check needed)
- `"own"` - User has `:own` variant, can only access their own resources (ownership check required)
- `"public"` - User has `:public` variant, public access (no ownership check needed)

**Important Notes**:
- Only controllers using `RequirePermissionWithScope` middleware will have scope in context
- Controllers using `RequirePermission` middleware don't need scope checks (they already have full access)
- Always check if scope exists before using it
- When scope is `"own"`, you must verify `currentUserID == resourceOwnerID` before allowing access

### Complete Example: User Update Route

Let's trace through a complete example:

**Route Definition**:
```go
users.PUT("/:id",
    middlewares.AuthMiddleware(),
    middlewares.RequirePermissionWithScope("users:update"), // Sets scope in context
    controllers.UpdateUser, // Controller handles ownership check
)
```

**Controller Implementation**:
```go
func UpdateUser(c *gin.Context) {
    // Get scope from context
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    // Extract ID from URL parameter
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    
    // Check ownership if scope is "own"
    if exists && scope == "own" {
        currentUserID, _ := c.Get(middlewares.ContextKeyUserID)
        currentUserIDUint := currentUserID.(uint)
        if currentUserIDUint != uint(id) {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Access denied: You can only access your own resources"
            })
            return
        }
    }
    
    // ... implement your update logic here
}
```

**Scenario 1: Admin (has `users:update`) tries to update user 5**
1. `AuthMiddleware()` validates token, sets `user_id = 1` (admin's ID)
2. `RequirePermissionWithScope()`:
   - Loads permissions: `["users:read", "users:update", "users:delete"]`
   - Checks: Has `users:update`? ✅ Yes
   - Determines scope: `"all"` (has base permission)
   - Sets in context: `permission_scope = "all"`, `needs_ownership_check = false`
   - ✅ Allows access
3. `UpdateUser()` controller:
   - Gets scope from context: `"all"`
   - Since scope is `"all"`, no ownership check needed
   - ✅ Proceeds with update logic (to be implemented)

**Scenario 2: Regular User (has `users:update:own`) tries to update user 5 (their own ID)**
1. `AuthMiddleware()` validates token, sets `user_id = 5`
2. `RequirePermissionWithScope()`:
   - Loads permissions: `["users:read:own", "users:update:own"]`
   - Checks: Has `users:update` or `users:update:own`? ✅ Yes (`users:update:own`)
   - Determines scope: `"own"` (has `:own` variant)
   - Sets in context: `permission_scope = "own"`, `needs_ownership_check = true`
   - ✅ Allows access
3. `UpdateUser()` controller:
   - Gets scope from context: `"own"`
   - Checks ownership: `currentUserID (5) == resourceID (5)` → ✅ Match
   - ✅ Proceeds with update logic (to be implemented)

**Scenario 3: Regular User (has `users:update:own`) tries to update user 999 (someone else)**
1. `AuthMiddleware()` validates token, sets `user_id = 5`
2. `RequirePermissionWithScope()`:
   - Loads permissions: `["users:read:own", "users:update:own"]`
   - Checks: Has `users:update` or `users:update:own`? ✅ Yes
   - Determines scope: `"own"` (has `:own` variant)
   - Sets in context: `permission_scope = "own"`, `needs_ownership_check = true`
   - ✅ Allows access (permission check passed)
3. `UpdateUser()` controller:
   - Gets scope from context: `"own"`
   - Checks ownership: `currentUserID (5) == resourceID (999)` → ❌ No match
   - ❌ Returns 403 Forbidden: "Access denied: You can only access your own resources"

### Benefits of This Approach

1. **Separation of Concerns**: Middleware handles auth/authz, controllers handle business logic
2. **Simpler Middleware**: No ID extraction or comparison logic needed
3. **Flexible Controllers**: Can extract IDs from any source (params, query, body, database)
4. **Cleaner Routes**: No need to pass extraction functions
5. **Better Control**: Controllers have full access to request data for complex logic
6. **Easier Testing**: Controllers can be tested independently with mock contexts

### Common Patterns

#### Pattern 1: URL Parameter (Most Common)
```go
// Route
users.GET("/:id",
    middlewares.AuthMiddleware(),
    middlewares.RequirePermissionWithScope("users:read"),
    controllers.GetUserByID,
)

// Controller
func GetUserByID(c *gin.Context) {
    // Get scope from context
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    // Extract ID from URL parameter
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    
    // Check ownership if scope is "own"
    if exists && scope == "own" {
        currentUserID, _ := c.Get(middlewares.ContextKeyUserID)
        currentUserIDUint := currentUserID.(uint)
        if currentUserIDUint != uint(id) {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Access denied: You can only access your own resources"
            })
            return
        }
    }
    
    // Implement your logic here (fetch user, return data, etc.)
}
```

#### Pattern 2: Query Parameter
```go
// Route
posts.GET("",
    middlewares.AuthMiddleware(),
    middlewares.RequirePermissionWithScope("posts:read"),
    controllers.GetPostsByAuthor,
)

// Controller
func GetPostsByAuthor(c *gin.Context) {
    // Get scope from context
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    // Extract author ID from query parameter
    authorIDStr := c.Query("author_id")
    authorID, _ := strconv.ParseUint(authorIDStr, 10, 32)
    
    // Check ownership if scope is "own"
    if exists && scope == "own" {
        currentUserID, _ := c.Get(middlewares.ContextKeyUserID)
        currentUserIDUint := currentUserID.(uint)
        if currentUserIDUint != uint(authorID) {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Access denied: You can only access your own resources"
            })
            return
        }
    }
    
    // Implement your logic here (fetch posts, return data, etc.)
}
```

#### Pattern 3: Request Body
```go
// Route
blogs.PUT("",
    middlewares.AuthMiddleware(),
    middlewares.RequirePermissionWithScope("blogs:update"),
    controllers.UpdateBlog,
)

// Controller
func UpdateBlog(c *gin.Context) {
    // Get scope from context
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    var input struct {
        BlogID uint `json:"blog_id"`
    }
    c.ShouldBindJSON(&input)
    
    // Fetch blog to get author ID
    var blog models.Blog
    database.DB.First(&blog, input.BlogID)
    
    // Check ownership if scope is "own"
    if exists && scope == "own" {
        currentUserID, _ := c.Get(middlewares.ContextKeyUserID)
        currentUserIDUint := currentUserID.(uint)
        if currentUserIDUint != blog.AuthorID {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Access denied: You can only access your own resources"
            })
            return
        }
    }
    
    // Implement your update logic here
}
```

#### Pattern 4: Database Lookup
```go
// Route
posts.GET("/:post_id",
    middlewares.AuthMiddleware(),
    middlewares.RequirePermissionWithScope("posts:read"),
    controllers.GetPost,
)

// Controller
func GetPost(c *gin.Context) {
    // Get scope from context
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    postID, _ := strconv.ParseUint(c.Param("post_id"), 10, 32)
    
    var post models.Post
    database.DB.First(&post, uint(postID))
    
    // Check ownership using post's author_id from database
    if exists && scope == "own" {
        currentUserID, _ := c.Get(middlewares.ContextKeyUserID)
        currentUserIDUint := currentUserID.(uint)
        if currentUserIDUint != post.AuthorID {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Access denied: You can only access your own resources"
            })
            return
        }
    }
    
    // Implement your logic here (return post data, etc.)
}
```

---

## Usage Examples

### Example 1: Protected Route (List All Users)

**Route**: `GET /api/v1/users`

**Required Permission**: `users:read`

**Request**:
```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <token>"
```

**Flow**:
1. Token validated by `AuthMiddleware()`
2. Permissions loaded: `["users:read", "blogs:create", ...]`
3. `RequirePermission("users:read")` checks permissions
4. User has `users:read` → ✅ Allow
5. Controller returns list of users

**Who can access**:
- ✅ Superadmin (has `users:read`)
- ✅ Admin (has `users:read`)
- ❌ Regular user (only has `users:read:own`)

### Example 2: Scoped Route (Get Own Profile)

**Route**: `GET /api/v1/users/5`

**Required Permission**: `users:read` or `users:read:own`

**Request**:
```bash
curl -X GET http://localhost:8080/api/v1/users/5 \
  -H "Authorization: Bearer <token>"
```

**Scenario A**: User 5 requests their own profile
1. Token validated → user_id = 5
2. Permissions loaded: `["users:read:own", ...]`
3. `RequirePermissionWithScope()` checks:
   - Has `users:read:own`? ✅ Yes
   - Is resource owner (5) == current user (5)? ✅ Yes
4. ✅ Allow → Return user profile

**Scenario B**: User 3 requests user 5's profile
1. Token validated → user_id = 3
2. Permissions loaded: `["users:read:own", ...]`
3. `RequirePermissionWithScope()` checks:
   - Has `users:read:own`? ✅ Yes
   - Is resource owner (5) == current user (3)? ❌ No
4. ❌ Deny → Return 403 Forbidden

**Scenario C**: Superadmin requests user 5's profile
1. Token validated → user_id = 1 (superadmin)
2. Permissions loaded: `["users:read", ...]`
3. `RequirePermissionWithScope()` checks:
   - Has `users:read`? ✅ Yes (base permission)
   - Ownership check? ❌ Not needed (has base permission)
4. ✅ Allow → Return user profile

### Example 3: Public Route

**Route**: `GET /api/v1/blogs/public`

**Required Permission**: `blogs:read:public`

**Request**:
```bash
curl -X GET http://localhost:8080/api/v1/blogs/public
# No Authorization header needed!
```

**Flow**:
1. `AllowPublic("blogs:read:public")` checks permission name
2. Permission ends with `:public` → ✅ Allow
3. No authentication required
4. Controller returns public blogs

### Example 4: Create User (Admin Only)

**Route**: `POST /api/v1/users`

**Required Permission**: `users:create`

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "password": "password123",
    "role_id": 3
  }'
```

**Who can access**:
- ✅ Superadmin (has `users:create`)
- ❌ Admin (doesn't have `users:create`)
- ❌ Regular user (doesn't have `users:create`)

---

## Testing the System

### 1. Start the Application

```bash
go run main.go
```

The application will:
- Connect to PostgreSQL database
- Auto-migrate tables
- Seed roles and permissions
- Create default users

### 2. Default Users

After seeding, you'll have:

| Username | Password | Role | Permissions |
|----------|----------|------|-------------|
| superadmin | superadmin123 | superadmin | All permissions |
| admin | admin123 | admin | Limited admin permissions |
| user | user123 | user | Own resources only |

### 3. Test Authentication

**Login as superadmin**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "superadmin",
    "password": "superadmin123"
  }'
```

**Response**:
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 4. Test Authorization

**Save the token**:
```bash
TOKEN="<your-token-here>"
```

**Test protected route** (as superadmin):
```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN"
```

**Expected**: List of all users

**Test as regular user**:
```bash
# Login as user
USER_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "user123"}' | jq -r '.token')

# Try to list all users (should fail)
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $USER_TOKEN"
```

**Expected**: `{"success": false, "error": "Insufficient permissions"}`

**Test own profile access** (as regular user):
```bash
# Get your own profile (should work)
curl -X GET http://localhost:8080/api/v1/users/3 \
  -H "Authorization: Bearer $USER_TOKEN"
```

**Expected**: Your user profile

**Test someone else's profile** (as regular user):
```bash
# Try to get another user's profile (should fail)
curl -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer $USER_TOKEN"
```

**Expected**: `{"success": false, "error": "Access denied: You can only access your own resources"}`

---

## Summary

### Authentication Flow Summary

```
1. User registers → User created in database
2. User logs in → JWT token generated
3. User makes request → Token sent in Authorization header
4. AuthMiddleware validates token → Extracts user_id & role_id
5. Permissions loaded → Stored in context
6. Request continues to controller
```

### Authorization Flow Summary

```
1. Route requires permission → Permission middleware checks
2. Load user permissions → From context or database
3. Check permission match → Exact, base, or variant
4. If :own scope → Check resource ownership
5. Allow or deny → Continue or return 403
```

### Key Takeaways

1. **Authentication** = "Who are you?" (JWT token)
2. **Authorization** = "What can you do?" (Permissions)
3. **Roles** = Groups of permissions
4. **Scopes** = Limits on permissions (`:own`, `:all`, `:public`)
5. **Middleware** = Code that runs before controllers
6. **Context** = Storage for request data (user_id, permissions, etc.)

---

## Additional Resources

- [JWT.io](https://jwt.io/) - Learn about JWT tokens
- [RBAC Explained](https://en.wikipedia.org/wiki/Role-based_access_control) - Role-Based Access Control
- [Gin Framework Docs](https://gin-gonic.com/docs/) - Web framework documentation

---

**Happy Coding! 🚀**
