# Mobile RBAC Strategy: Microservices Architecture Permission Model

## 📋 Tổng quan kiến trúc hệ thống

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

## 🚀 Kết luận

**Khuyến nghị:** Sử dụng **Single Role** cho Customer với **Feature-based Access Control**

- Employee: Complex RBAC với multiple roles
- Customer: Simple RBAC với single role + subscription/feature flags

Cách này giúp hệ thống vừa có tính bảo mật cao (RBAC cho employee) vừa có UX đơn giản (single role cho customer).