# Mobile RBAC Strategy: Enhanced Permission Control Model

## 📋 Tổng quan kiến trúc hệ thống

### 🚀 Tính năng mới: Chi tiết kiểm soát quyền từng role trên từng route

Hệ thống RBAC hiện đã được nâng cấp để hỗ trợ kiểm soát chi tiết như Iris framework:

#### 🎛️ Các loại permission control:
- **Allow(roles...)**: Chỉ cho phép các role cụ thể
- **AllowAll()**: Cho phép tất cả role đã đăng nhập
- **Forbid(roles...)**: Cấm các role cụ thể, các role khác được phép
- **ForbidAll()**: Cấm tất cả role (maintenance mode)

#### 📝 Cách sử dụng mới:

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/techmaster-vietnam/dd_goshare/rbac"
)

func setupRoutes(app *fiber.App) {
    api := app.Group("/api")

    // 1. Chỉ cho phép admin và moderator
    rbac.Get(api, "/admin/users", rbac.Allow(1, 2), true, adminHandler)

    // 2. Cấm guest user, các role khác được phép
    rbac.Post(api, "/content/create", rbac.Forbid(4), true, createHandler)

    // 3. Cho phép tất cả user đã đăng nhập
    rbac.Get(api, "/profile", rbac.AllowAll(), true, profileHandler)

    // 4. Cấm tất cả (maintenance)
    rbac.Delete(api, "/system/reset", rbac.ForbidAll(), true, resetHandler)

    // 5. Route public (không cần đăng nhập)
    rbac.Get(api, "/public/info", rbac.AllowAll(), false, publicHandler)
}
```

#### � Logic kiểm tra quyền:

1. **ForbidAll**: Cấm tất cả → Trả về 403
2. **Forbid**: Nếu user role trong danh sách cấm → Trả về 403
3. **AllowAll**: Cho phép tất cả → Tiếp tục
4. **Allow**: Nếu user role trong danh sách cho phép → Tiếp tục, ngược lại 403

### 🏗️ Microservices Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   dd_backend    │    │   dd_mobile     │
│  (Admin/Mgmt)   │    │ (Customer API)  │
│                 │    │                 │
│ • Employee mgmt │    │ • Customer API  │
│ • Data entry    │    │ • Mobile logic  │
│ • CRUD ops      │    │ • User features │
│ • Reporting     │    │ • Auth mobile   │
└─────────┬───────┘    └─────────┬───────┘
          │                      │
          └──────────────────────┘
                     │
            ┌─────────▼──────────┐
            │  Shared Database   │
            │   + dd_goshare     │
            │  (Auto-migration)  │
            └────────────────────┘
```

### 🎯 Phân chia trách nhiệm rõ ràng

1. **dd_backend (Employee-focused)**:
   - ✅ Quản lý employees, roles, permissions
   - ✅ CRUD operations cho tất cả entities
   - ✅ Data entry và content management
   - ✅ Analytics và reporting
   - ✅ System administration

2. **dd_mobile (Customer-focused)**:
   - ✅ Customer authentication (Firebase)
   - ✅ Customer-facing APIs
   - ✅ Mobile-specific features
   - ✅ User experience logic
   - ✅ Content consumption (không tạo/sửa)

## 🎯 RBAC Strategy cho Microservices

### dd_mobile: Customer-Only RBAC (Simplified) ✅

```sql
-- dd_mobile chỉ cần 1 role cho customer
customer (role) -> Tất cả mobile users đều có role này
```

**Lý do phù hợp với kiến trúc:**
- ✅ **Service boundary**: dd_mobile chỉ phục vụ customers
- ✅ **Simple auth flow**: Firebase → Customer role → Mobile APIs
- ✅ **No admin complexity**: Admin operations ở dd_backend
- ✅ **Shared database**: Customer data được manage bởi dd_backend

**Customer permissions trong dd_mobile:**
```go
// Customer chỉ có quyền:
// - Xem content (dialogs, topics)
// - Cập nhật profile của chính mình
// - Sử dụng features (translate, comments)
// - Download content (nếu có subscription)

// KHÔNG có quyền:
// - Tạo/sửa/xóa content (làm ở dd_backend)
// - Quản lý users khác
// - Admin operations
```

### Option 2: Multiple Roles cho Customer (Phức tạp hơn)

```sql
customer_basic    -> Customer miễn phí
customer_premium  -> Customer trả phí
customer_vip      -> Customer VIP
```

**Nhược điểm:**
- ❌ Tăng độ phức tạp không cần thiết
- ❌ UX/UI phải handle nhiều trường hợp
- ❌ Quản lý role assignments phức tạp

## 🏛️ Microservices RBAC Architecture

### dd_backend (Employee Management) - Complex RBAC
```
admin          -> Full system access + user management
manager        -> Business management + reporting
content_editor -> Content CRUD operations
support        -> Customer support + read access
viewer         -> Read-only access to all data
```

### dd_mobile (Customer API) - Simple RBAC
```
customer -> Consumer permissions only:
  ├── Read content (dialogs, topics)
  ├── Update own profile
  ├── Use features (translate, comments)
  ├── Download content (subscription-based)
  └── No admin/management operations
```

### Shared Database với Auto-Migration
```
dd_goshare provides:
├── Database models & migrations
├── Shared RBAC system
├── Common business logic
└── Utility functions
```

## 🔧 Implementation Strategy

### 1. Customer Authentication Flow
```go
// Customer đăng nhập -> Tự động assign role "customer"
user := authenticateCustomer(firebaseToken)
user.role = "customer"  // Always customer role

// Permission check dựa trên business logic
canDownload := checkSubscription(user.id) && checkPermission("mobile.dialog.download")
```

### 2. Feature-based Access Control
```go
// Trong handlers
func (h *DialogHandler) DownloadDialog(c *fiber.Ctx) error {
    user := getCurrentUser(c)
    
    // Check base permission (all customers have this)
    if !rbac.HasPermission(user.role, "customer.dialog.access") {
        return unauthorized()
    }
    
    // Check subscription for premium features
    if !subscriptionService.HasActiveSubscription(user.id) {
        return c.JSON(fiber.Map{
            "error": "Download requires active subscription",
            "upgrade_url": "/subscription/upgrade"
        })
    }
    
    return h.downloadDialog(c)
}
```

### 3. Database Schema
```sql
-- Customer table - no complex role management needed
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    firebase_uid VARCHAR(255) UNIQUE,
    email VARCHAR(255),
    username VARCHAR(255),
    subscription_type VARCHAR(50) DEFAULT 'free', -- free, premium, vip
    subscription_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User-role assignment (customers always get 'customer' role)
INSERT INTO user_roles (user_id, role_id) 
SELECT c.id, r.id FROM customers c, roles r 
WHERE r.name = 'customer';
```

## 🎉 Lợi ích của Single Customer Role

1. **Simplified UX**: Tất cả customer đều thấy cùng interface
2. **Easy Onboarding**: Không cần chọn role khi đăng ký
3. **Feature-based Monetization**: Premium features controlled by subscription, not roles
4. **Easier Support**: Support team không cần hiểu phức tạp về roles
5. **Scalable**: Dễ thêm features mới mà không ảnh hưởng role structure

## � Best Practices cho Enhanced RBAC

### 1. Ưu tiên sử dụng Allow thay vì Forbid
```go
// ✅ Tốt: Rõ ràng ai được phép
rbac.Get(api, "/admin/sensitive", rbac.Allow(1), true, handler)

// ❌ Tránh: Không rõ ai được phép khi có nhiều role
rbac.Get(api, "/admin/sensitive", rbac.Forbid(2, 3, 4, 5), true, handler)
```

### 2. Sử dụng AllowAll cho route cần authentication
```go
// ✅ Route cần đăng nhập nhưng tất cả role đều truy cập được
rbac.Get(api, "/profile", rbac.AllowAll(), true, profileHandler)

// ✅ Route public không cần đăng nhập
rbac.Get(api, "/public", rbac.AllowAll(), false, publicHandler)
```

### 3. ForbidAll cho maintenance mode
```go
// ✅ Tạm thời tắt tính năng
rbac.Post(api, "/dangerous-action", rbac.ForbidAll(), true, handler)
```

### 4. Forbid cho loại trừ role cụ thể
```go
// ✅ Cấm guest, các role khác được phép
rbac.Post(api, "/content/create", rbac.Forbid(4), true, handler)
```

### 5. Test Coverage
```go
// Đảm bảo test tất cả trường hợp
func TestRolePermissions(t *testing.T) {
    // Test Allow
    // Test Forbid  
    // Test AllowAll
    // Test ForbidAll
    // Test middleware logic
}
```

## 🚀 Kết luận

**Khuyến nghị:** Sử dụng **Enhanced RBAC** với **Allow/Forbid control**

- **Employee Backend**: Complex RBAC với Allow/Forbid/AllowAll/ForbidAll
- **Customer Mobile**: Simple RBAC với AllowAll + subscription/feature flags

### Migration từ hệ thống cũ:
1. Đăng ký route qua rbac.Get/Post/... thay vì app.Get/Post/...
2. Thêm permission control: Allow(), Forbid(), AllowAll(), ForbidAll()
3. Route sẽ tự động được kiểm tra quyền qua middleware

Cách này giúp hệ thống vừa có **kiểm soát chi tiết** (Allow/Forbid) vừa có **hiệu suất cao** (in-memory check).