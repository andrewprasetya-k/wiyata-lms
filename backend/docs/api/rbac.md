# 🔐 RBAC (Role-Based Access Control) API Documentation

Base URL: `/api/rbac`

---

## Overview

Role-Based Access Control (RBAC) mengamankan API endpoints berdasarkan role user di setiap school. Sistem mendukung multi-school dengan role berbeda per school.

### Roles

| Role | Scope | Permissions |
|------|-------|-------------|
| `super_admin` | **System-wide** (bukan school-specific) | System management: create/manage schools, roles, super admins. Read access ke semua sekolah. **TIDAK bisa** melakukan operasi akademik (create assignments, materials, dll) tanpa role sekolah. |
| `admin` | School-specific | Manage sekolah tertentu (academic years, terms, users, subjects, classes) |
| `teacher` | School-specific | Manage kelas, materials, assignments yang diajar |
| `student` | School-specific | Akses read-only + submit assignment |

**Catatan Penting:**
- `super_admin` adalah **admin sistem aplikasi**, bukan admin sekolah
- `super_admin` bisa VIEW data semua sekolah (monitoring/troubleshooting)
- `super_admin` TIDAK bisa CREATE/UPDATE/DELETE konten akademik tanpa role sekolah
- Untuk operasi akademik, super_admin harus di-enroll sebagai `admin` atau `teacher` di sekolah tersebut

---

## 1. Role Management

### List All Roles
- **URL:** `/roles`
- **Method:** `GET`
- **Auth:** Required

**Response Example:**
```json
[
  {
    "roleId": "uuid",
    "roleName": "super_admin",
    "createdAt": "02-01-2006 15:04:05"
  },
  {
    "roleId": "uuid",
    "roleName": "teacher",
    "createdAt": "02-01-2006 15:04:05"
  }
]
```

### Create Role
- **URL:** `/roles`
- **Method:** `POST`
- **Auth:** Required (super_admin only)
- **Body:**
```json
{
  "roleName": "teacher"
}
```

### Get Role by ID
- **URL:** `/roles/:id`
- **Method:** `GET`
- **Auth:** Required

### Update Role Name
- **URL:** `/roles/:id`
- **Method:** `PATCH`
- **Auth:** Required (super_admin only)
- **Body:**
```json
{
  "roleName": "senior_teacher"
}
```

### Delete Role
- **URL:** `/roles/:id`
- **Method:** `DELETE`
- **Auth:** Required (super_admin only)

---

## 2. User Role Management (Assignments)

Assigning roles to users within a school context.

### Assign Role to User
- **URL:** `/user-roles`
- **Method:** `POST`
- **Auth:** Required (admin, super_admin)
- **Body:**
```json
{
  "schoolUserId": "uuid",
  "roleId": "uuid"
}
```

### Remove Role from User
- **URL:** `/user-roles?schoolUserId=...&roleId=...`
- **Method:** `DELETE`
- **Auth:** Required (admin, super_admin)

### List User's Roles
- **URL:** `/user-roles/:schoolUserId`
- **Method:** `GET`
- **Auth:** Required

**Response Example:**
```json
[
  {
    "urol_id": "uuid",
    "urol_scu_id": "uuid",
    "urol_rol_id": "uuid",
    "role": {
      "rol_id": "uuid",
      "rol_name": "teacher"
    }
  }
]
```

### Update User Roles (Sync)
Replace all roles for a user.
- **URL:** `/user-roles/:schoolUserId`
- **Method:** `PATCH`
- **Auth:** Required (admin, super_admin)
- **Body:**
```json
{
  "roleIds": ["role-uuid-1", "role-uuid-2"]
}
```

---

## 3. Super Admin Management

### Create Super Admin
Create a new super admin user (automatically enrolled to "admin" school with super_admin role).

- **URL:** `/super-admin`
- **Method:** `POST`
- **Auth:** Required (super_admin only)
- **Body:**
```json
{
  "fullName": "John Doe",
  "email": "john@admin.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "message": "Super admin created successfully"
}
```

**Note:** This endpoint can only be accessed by existing super_admin. For the first super_admin, use manual setup (see section 6).

---

## 4. RBAC Middleware

### School Context Header

**All protected endpoints require school context via:**

**Priority 1: Header (Recommended)**
```
SchoolId: uuid-school-id
```

**Priority 2: URL Parameter (Fallback)**
```
/api/schools/:schoolCode/...
```

### Request Example

```bash
POST /api/classes
Authorization: Bearer <token>
SchoolId: uuid-school-id
Content-Type: application/json

{
  "cls_code": "12-IPA-1",
  "cls_title": "Kelas 12 IPA 1"
}
```

---

## 4. Protected Endpoints

**Legend:**
- ✅ = Can perform action
- 📖 = Read-only access
- ❌ = No access

### System Management (Super Admin Only)
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/schools` | POST | ✅ | ❌ | ❌ | ❌ |
| `/rbac/roles` | POST | ✅ | ❌ | ❌ | ❌ |
| `/rbac/super-admin` | POST | ✅ | ❌ | ❌ | ❌ |

### School Management
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/schools/:schoolCode` | GET | 📖 | 📖 | 📖 | 📖 |
| `/schools/:schoolCode` | PATCH | ❌* | ✅ | ❌ | ❌ |
| `/schools/:schoolCode` | DELETE | ❌* | ✅ | ❌ | ❌ |

*Super admin harus enroll sebagai admin di sekolah tersebut

### Academic Structure
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/academic-years` | GET | 📖 | 📖 | 📖 | 📖 |
| `/academic-years` | POST | ❌ | ✅ | ❌ | ❌ |
| `/terms` | POST | ❌ | ✅ | ❌ | ❌ |
| `/subjects` | POST | ❌ | ✅ | ❌ | ❌ |

### Class Management
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/classes` | GET | 📖 | 📖 | 📖 | 📖 |
| `/classes` | POST | ❌ | ✅ | ✅ | ❌ |
| `/classes/:id` | PATCH | ❌ | ✅ | ✅ | ❌ |
| `/classes/:id` | DELETE | ❌ | ✅ | ❌ | ❌ |

### Learning Content
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/materials` | GET | 📖 | 📖 | 📖 | 📖 |
| `/materials` | POST | ❌ | ❌ | ✅ | ❌ |
| `/assignments` | POST | ❌ | ❌ | ✅ | ❌ |
| `/assignments/:id` | GET | ❌ | ❌ | 📖* | ❌ |
| `/assignments/submit/:id` | POST | ❌ | ❌ | ❌ | ✅ |
| `/assignments/submit/:id` | GET | ❌ | ❌ | 📖* | ❌ |
| `/assignments/assess/:id` | POST/PATCH/DELETE | ❌ | ❌ | ✅* | ❌ |

*Teacher assignment detail, submission detail, and assessment access is limited to assignments in subject classes taught by the current teacher in the active `SchoolId` context.

### User Management
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/users` | POST | 📖 | ✅ | ❌ | ❌ |
| `/users` | GET | 📖 | ✅ | ❌ | ❌ |
| `/users/:id` | DELETE | 📖 | ✅ | ❌ | ❌ |

### Enrollment
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/enrollments` | GET | 📖 | 📖 | 📖 | 📖 |
| `/enrollments` | POST | ❌ | ✅ | ✅ | ❌ |
| `/enrollments/:id` | DELETE | ❌ | ✅ | ✅ | ❌ |

---

## 5. Error Responses

### 400 Bad Request
School context tidak ditemukan.

```json
{
  "error": "School context required (SchoolId header or schoolCode param)"
}
```

### 401 Unauthorized
User tidak login atau token invalid.

```json
{
  "error": "Unauthorized"
}
```

### 403 Forbidden - Insufficient Permissions
User tidak punya role yang sesuai.

```json
{
  "error": "Forbidden: insufficient permissions"
}
```

### 403 Forbidden - Not School Member
User bukan member dari school yang diakses.

```json
{
  "error": "Forbidden: not a member of this school"
}
```

---

## 6. Setup & Testing

### Initial Setup (First Super Admin)

**Option 1: Via API (3 steps)**

1. **Create User**
```bash
POST /api/register
{
  "fullName": "Super Admin",
  "email": "admin@system.com",
  "password": "securepassword"
}
```

2. **Get School "admin" ID**
```bash
GET /api/schools
# Find school with name "admin", copy sch_id
```

3. **Enroll to admin school**
```bash
POST /api/school-users/enroll
{
  "scu_usr_id": "<user-id>",
  "scu_sch_id": "<admin-school-id>"
}
# Copy scu_id from response
```

4. **Get super_admin role ID**
```bash
GET /api/rbac/roles
# Find role with name "super_admin", copy rol_id
```

5. **Assign super_admin role**
```bash
POST /api/rbac/user-roles
{
  "schoolUserId": "<scu-id>",
  "roleId": "<super-admin-role-id>"
}
```

**Option 2: Via SQL (Quick)**
```sql
-- 1. Create user
INSERT INTO edv.users (usr_id, usr_nama_lengkap, usr_email, usr_password, is_active)
VALUES (gen_random_uuid(), 'Super Admin', 'admin@system.com', '<hashed-password>', true);

-- 2. Get IDs
SELECT usr_id FROM edv.users WHERE usr_email = 'admin@system.com';
SELECT sch_id FROM edv.schools WHERE sch_name = 'admin';
SELECT rol_id FROM edv.roles WHERE rol_name = 'super_admin';

-- 3. Enroll to admin school
INSERT INTO edv.school_users (scu_id, scu_usr_id, scu_sch_id)
VALUES (gen_random_uuid(), '<user-id>', '<admin-school-id>');

-- 4. Assign role
INSERT INTO edv.user_roles (urol_id, urol_scu_id, urol_rol_id)
VALUES (gen_random_uuid(), '<school-user-id>', '<role-id>');
```

### Create Additional Super Admins

Once you have one super_admin, use the endpoint:

```bash
POST /api/rbac/super-admin
Authorization: Bearer <super-admin-token>
SchoolId: <admin-school-id>

{
  "fullName": "Another Super Admin",
  "email": "admin2@system.com",
  "password": "securepassword"
}
```

This automatically:
- Creates user
- Enrolls to "admin" school
- Assigns super_admin role

### Initial Setup (Roles)
```bash
POST /api/rbac/roles
{
  "roleName": "super_admin"
}
```

2. **Enroll User to School**
```bash
POST /api/school-users/enroll
{
  "scu_usr_id": "user-uuid",
  "scu_sch_id": "school-uuid"
}
```

3. **Assign Role**
```bash
POST /api/rbac/user-roles
{
  "schoolUserId": "school-user-uuid",
  "roleId": "role-uuid"
}
```

### Frontend Integration

```javascript
// 1. User login
const { token } = await login(email, password);

// 2. Get user's schools
const schools = await fetch(`/api/school-users/user/${userId}`, {
  headers: { 'Authorization': `Bearer ${token}` }
});

// 3. User selects a school
const selectedSchoolId = schools[0].school.sch_id;

// 4. Make requests with SchoolId header
const response = await fetch('/api/classes', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'SchoolId': selectedSchoolId,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    cls_code: '12-IPA-1',
    cls_title: 'Kelas 12 IPA 1'
  })
});
```

### Test Examples

**Test as Admin (Success)**
```bash
POST /api/classes
Authorization: Bearer <admin-token>
SchoolId: <school-id>

{
  "cls_code": "12-IPA-1",
  "cls_title": "Kelas 12 IPA 1"
}

# Response: 200 OK
```

**Test as Student (Fail)**
```bash
POST /api/classes
Authorization: Bearer <student-token>
SchoolId: <school-id>

{
  "cls_code": "12-IPA-1",
  "cls_title": "Kelas 12 IPA 1"
}

# Response: 403 Forbidden
{
  "error": "Forbidden: insufficient permissions"
}
```

---

## 7. Implementation Notes

### Multi-School Support
- User dapat memiliki role berbeda di sekolah berbeda
- Frontend mengirim `SchoolId` header untuk specify context
- Middleware otomatis validate membership dan role

### Backward Compatible
- Endpoint dengan `schoolCode` di URL tetap work
- Tidak ada breaking changes pada API contract
- Handler code tidak perlu diubah

### Security Features
- Cross-tenant isolation (user tidak bisa akses school lain)
- Role-based permissions (action restricted by role)
- Fail-secure (default deny jika tidak ada role match)

### Future Enhancements
- [ ] Permission-based access (granular control)
- [ ] Resource ownership check (creator-only modifications)
- [ ] Class-level access (teacher/student specific to class)
- [ ] Audit logging untuk access attempts
