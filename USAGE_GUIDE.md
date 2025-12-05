# Middleware Usage Guide

A practical guide for using authentication and authorization middlewares in your Go application.

## Quick Start

### 1. Authentication Middleware (`AuthMiddleware`)

**Purpose**: Optionally authenticates users. Sets user context if token is valid, but doesn't abort if token is missing.

**When to Use**: Apply to route groups that may need user context (even for public routes).

**Example**:
```go
users := rg.Group("/users")
users.Use(middlewares.AuthMiddleware()) // Optional - sets context if token exists
{
    // Your routes here
}
```

**What It Does**:
- ✅ If token is valid: Sets `user_id`, `role_id`, and `permissions` in context
- ✅ If token is missing/invalid: Continues without setting user context (allows public access)
- ❌ Does NOT abort requests (unlike old behavior)

---

### 2. Authorization Middleware (`RequirePermission`)

**Purpose**: Handles authorization for both public and protected routes. Checks permissions and sets scope in context.

**When to Use**: Apply to individual routes that need permission checks.

**Example**:
```go
users.GET("",
    middlewares.RequirePermission("users:read"),
    controllers.GetAllUsers,
)
```

**What It Does**:
- ✅ If permission is public (`:public`): Allows access without authentication
- ✅ If permission is not public: Requires authentication, then checks permissions
- ✅ Sets scope in context if determinable (`own`, `all`, or `public`)

---

## Common Patterns

### Pattern 1: Protected Route (Requires Authentication)

**Route**: Only authenticated users with specific permission can access.

```go
users := rg.Group("/users")
users.Use(middlewares.AuthMiddleware()) // Optional auth
{
    // List all users - requires users:read permission
    users.GET("",
        middlewares.RequirePermission("users:read"),
        controllers.GetAllUsers,
    )
    
    // Create user - requires users:create permission
    users.POST("",
        middlewares.RequirePermission("users:create"),
        controllers.CreateUser,
    )
}
```

**Flow**:
1. `AuthMiddleware()`: Sets user context if token exists
2. `RequirePermission()`: Checks if user is authenticated → Checks permissions → Allows/denies

---

### Pattern 2: Scoped Route (Ownership Check)

**Route**: Users can access their own resources or all resources (based on permissions).

```go
users := rg.Group("/users")
users.Use(middlewares.AuthMiddleware())
{
    // Get user by ID - supports both :own and base permissions
    users.GET("/:id",
        middlewares.RequirePermission("users:read"),
        controllers.GetUserByID,
    )
    
    // Update user - supports both :own and base permissions
    users.PUT("/:id",
        middlewares.RequirePermission("users:update"),
        controllers.UpdateUser,
    )
}
```

**Controller Implementation**:
```go
func GetUserByID(c *gin.Context) {
    // Get scope from context (set by RequirePermission if determinable)
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    // Extract resource ID
    id := c.Param("id")
    userID, _ := strconv.ParseUint(id, 10, 32)
    
    // Check ownership if scope is "own"
    if exists && scope == "own" {
        currentUserID, _ := c.Get(middlewares.ContextKeyUserID)
        currentUserIDUint := currentUserID.(uint)
        
        if currentUserIDUint != uint(userID) {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error":   "Access denied: You can only access your own resources",
            })
            return
        }
    }
    
    // Your business logic here
    // ...
}
```

**Scope Values**:
- `"all"`: User has base permission or `:all` variant → No ownership check needed
- `"own"`: User has `:own` variant → Ownership check required
- `"public"`: Public permission → No ownership check needed

---

### Pattern 3: Public Route (No Authentication Required)

**Route**: Anyone can access, no authentication needed.

```go
blogs := rg.Group("/blogs")
blogs.Use(middlewares.AuthMiddleware()) // Still apply for optional user context
{
    // Public blogs - no authentication required
    blogs.GET("/public",
        middlewares.RequirePermission("blogs:read:public"),
        controllers.GetPublicBlogs,
    )
}
```

**Flow**:
1. `AuthMiddleware()`: No token provided → Continues without user context
2. `RequirePermission("blogs:read:public")`: Detects `:public` suffix → Allows access
3. Controller: Can access public data

---

### Pattern 4: Completely Public Route (No Middleware)

**Route**: No authentication or authorization needed at all.

```go
// Health check - no middleware needed
v1.GET("/health", controllers.HealthCheck)

// Public API info - no middleware needed
v1.GET("/info", controllers.GetAPIInfo)
```

---

## Route Examples

### Example 1: User Routes

```go
func UserRoutes(rg *gin.RouterGroup) {
    users := rg.Group("/users")
    users.Use(middlewares.AuthMiddleware()) // Optional auth for all user routes
    
    {
        // GET /users - List all users (requires users:read)
        users.GET("",
            middlewares.RequirePermission("users:read"),
            controllers.GetAllUsers,
        )
        
        // GET /users/:id - Get user by ID (supports :own scope)
        users.GET("/:id",
            middlewares.RequirePermission("users:read"),
            controllers.GetUserByID,
        )
        
        // POST /users - Create user (requires users:create)
        users.POST("",
            middlewares.RequirePermission("users:create"),
            controllers.CreateUser,
        )
        
        // PUT /users/:id - Update user (supports :own scope)
        users.PUT("/:id",
            middlewares.RequirePermission("users:update"),
            controllers.UpdateUser,
        )
        
        // DELETE /users/:id - Delete user (requires users:delete)
        users.DELETE("/:id",
            middlewares.RequirePermission("users:delete"),
            controllers.DeleteUser,
        )
    }
}
```

---

### Example 2: Blog Routes (Mixed Public/Protected)

```go
func BlogRoutes(rg *gin.RouterGroup) {
    blogs := rg.Group("/blogs")
    blogs.Use(middlewares.AuthMiddleware()) // Optional auth
    
    {
        // Public route - no authentication required
        blogs.GET("/public",
            middlewares.RequirePermission("blogs:read:public"),
            controllers.GetPublicBlogs,
        )
        
        // Protected route - requires authentication
        blogs.GET("",
            middlewares.RequirePermission("blogs:read"),
            controllers.GetAllBlogs,
        )
        
        // Protected route with ownership - supports :own scope
        blogs.GET("/:id",
            middlewares.RequirePermission("blogs:read"),
            controllers.GetBlogByID,
        )
        
        // Create blog - requires authentication
        blogs.POST("",
            middlewares.RequirePermission("blogs:create"),
            controllers.CreateBlog,
        )
        
        // Update blog - supports :own scope
        blogs.PUT("/:id",
            middlewares.RequirePermission("blogs:update"),
            controllers.UpdateBlog,
        )
    }
}
```

---

## Controller Patterns

### Pattern 1: Basic Protected Controller

```go
func GetAllUsers(c *gin.Context) {
    // No scope check needed - user already has permission
    // Just implement your business logic
    
    var users []models.User
    database.DB.Find(&users)
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    users,
    })
}
```

---

### Pattern 2: Controller with Ownership Check

```go
func GetUserByID(c *gin.Context) {
    // Get scope from context
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    // Extract user ID from URL
    id := c.Param("id")
    userID, _ := strconv.ParseUint(id, 10, 32)
    
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
        if currentUserIDUint != uint(userID) {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error":   "Access denied: You can only access your own resources",
            })
            return
        }
    }
    
    // Fetch user from database
    var user models.User
    if err := database.DB.First(&user, userID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "error":   "User not found",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    user,
    })
}
```

---

### Pattern 3: Controller with Query Parameter Ownership

```go
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
                "success": false,
                "error":   "Access denied: You can only access your own posts",
            })
            return
        }
    }
    
    // Fetch posts
    var posts []models.Post
    database.DB.Where("author_id = ?", authorID).Find(&posts)
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    posts,
    })
}
```

---

### Pattern 4: Controller with Database Lookup Ownership

```go
func UpdateBlog(c *gin.Context) {
    // Get scope from context
    scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
    
    // Extract blog ID from URL
    blogID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    
    // Fetch blog to get author ID
    var blog models.Blog
    if err := database.DB.First(&blog, blogID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "error":   "Blog not found",
        })
        return
    }
    
    // Check ownership if scope is "own"
    if exists && scope == "own" {
        currentUserID, _ := c.Get(middlewares.ContextKeyUserID)
        currentUserIDUint := currentUserID.(uint)
        
        if currentUserIDUint != blog.AuthorID {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error":   "Access denied: You can only update your own blogs",
            })
            return
        }
    }
    
    // Update blog logic here
    // ...
}
```

---

## Context Keys

Access these values in your controllers:

```go
// User ID (set by AuthMiddleware if authenticated)
userID, exists := c.Get(middlewares.ContextKeyUserID)
// Returns: uint or nil

// Role ID (set by AuthMiddleware if authenticated)
roleID, exists := c.Get(middlewares.ContextKeyRoleID)
// Returns: uint or nil

// Permissions (set by AuthMiddleware if authenticated)
permissions, exists := c.Get(middlewares.ContextKeyPermissions)
// Returns: []string or nil

// Permission Scope (set by RequirePermission if determinable)
scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
// Returns: "all", "own", "public", or nil

// Needs Ownership Check (set by RequirePermission if determinable)
needsCheck, exists := c.Get(middlewares.ContextKeyNeedsOwnershipCheck)
// Returns: bool or nil
```

---

## Error Responses

### 401 Unauthorized
Returned when:
- Non-public route accessed without authentication

```json
{
  "success": false,
  "error": "Authentication required"
}
```

### 403 Forbidden
Returned when:
- User doesn't have required permission
- Ownership check fails (in controller)

```json
{
  "success": false,
  "error": "Insufficient permissions"
}
```

---

## Best Practices

1. **Apply `AuthMiddleware` at group level**: Reduces repetition and ensures consistent behavior
   ```go
   users.Use(middlewares.AuthMiddleware())
   ```

2. **Always check if scope exists**: Scope may not be set for all routes
   ```go
   scope, exists := c.Get(middlewares.ContextKeyPermissionScope)
   if exists && scope == "own" {
       // Check ownership
   }
   ```

3. **Use descriptive permission names**: Follow `resource:action[:scope]` format
   - ✅ `users:read`
   - ✅ `users:read:own`
   - ✅ `blogs:read:public`
   - ❌ `read_users`

4. **Handle public routes explicitly**: Use `:public` suffix for public permissions
   ```go
   middlewares.RequirePermission("blogs:read:public")
   ```

5. **Keep ownership checks in controllers**: Middleware handles permissions, controllers handle business logic

---

## Testing Examples

### Test Public Route
```bash
# No token needed
curl -X GET http://localhost:8080/api/v1/blogs/public
```

### Test Protected Route
```bash
# Token required
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <your-token>"
```

### Test Scoped Route (Own Resource)
```bash
# User 5 accessing their own profile
curl -X GET http://localhost:8080/api/v1/users/5 \
  -H "Authorization: Bearer <user5-token>"
```

### Test Scoped Route (Someone Else's Resource)
```bash
# User 3 trying to access user 5's profile (should fail if user 3 has :own permission)
curl -X GET http://localhost:8080/api/v1/users/5 \
  -H "Authorization: Bearer <user3-token>"
```

---

## Summary

- **`AuthMiddleware()`**: Optional authentication - sets user context if token exists
- **`RequirePermission(permission)`**: Unified authorization - handles public and protected routes
- **Public routes**: Use `:public` suffix in permission name
- **Scoped routes**: Check `ContextKeyPermissionScope` in controller for ownership checks
- **Always check if values exist**: Context values may not be set for all routes

For more details, see [readme.md](readme.md).

