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

### Bootstrap School Tenant with Initial Admin
Create a new school tenant and assign one initial school admin in one database
transaction. This is internal Super Admin MVP onboarding, not public landing-page
school request onboarding.

- **URL:** `/super-admin/school-bootstrap`
- **Method:** `POST`
- **Auth:** Required (`super_admin` on system school `sch_code = "000000"` only)

**Create new admin user:**
```json
{
  "school": {
    "schoolName": "SMA EduVerse",
    "schoolCode": "sma-eduverse",
    "schoolAddress": "Jl. Pendidikan No. 1",
    "schoolEmail": "admin@sma.sch.id",
    "schoolPhone": "08123456789",
    "schoolWebsite": "https://sma.sch.id"
  },
  "adminUser": {
    "mode": "new",
    "fullName": "Admin Sekolah",
    "email": "admin@sma.sch.id",
    "password": "InitialPassword123!"
  }
}
```

**Use existing global user:**
```json
{
  "school": {
    "schoolName": "SMA EduVerse",
    "schoolCode": "sma-eduverse",
    "schoolAddress": "Jl. Pendidikan No. 1",
    "schoolEmail": "admin@sma.sch.id",
    "schoolPhone": "08123456789"
  },
  "adminUser": {
    "mode": "existing",
    "userId": "uuid"
  }
}
```

**Response:**
```json
{
  "school": {
    "schoolId": "uuid",
    "schoolName": "SMA EduVerse",
    "schoolCode": "sma-eduverse"
  },
  "adminUser": {
    "userId": "uuid",
    "fullName": "Admin Sekolah",
    "email": "admin@sma.sch.id",
    "isActive": true
  },
  "schoolUserId": "uuid",
  "assignedRoles": ["admin"]
}
```

**Behavior:**
- Creates the school tenant.
- Creates a new global user or uses an existing active user.
- Creates `school_users` membership for the new school.
- Assigns only the `admin` role to that school membership.
- Rolls back the whole transaction if any step fails.
- Does not create public school request records.
- Does not assign `super_admin`.

---

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
| `/super-admin/school-bootstrap` | POST | ✅* | ❌ | ❌ | ❌ |
| `/rbac/roles` | POST | ✅ | ❌ | ❌ | ❌ |
| `/rbac/super-admin` | POST | ✅ | ❌ | ❌ | ❌ |

*Requires `super_admin` on the system school where `schools.sch_code = "000000"`.

### School Management
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/schools/:schoolCode` | GET | 📖 | 📖 | 📖 | 📖 |
| `/schools/:schoolCode` | PATCH | ❌* | ✅ | ❌ | ❌ |
| `/schools/:schoolCode` | DELETE | ❌* | ✅ | ❌ | ❌ |

*Super admin harus enroll sebagai admin di sekolah tersebut

### School Members
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/admin/school-members` | GET | ❌ | ✅ | ❌ | ❌ |
| `/admin/school-members` | POST | ❌ | ✅ | ❌ | ❌ |
| `/admin/school-members/:schoolUserId` | DELETE | ❌ | ✅ | ❌ | ❌ |
| `/admin/school-members/:schoolUserId/restore` | PATCH | ❌ | ✅ | ❌ | ❌ |
| `/admin/school-members/import/preview` | POST | ❌ | ✅ | ❌ | ❌ |
| `/admin/school-members/import/commit` | POST | ❌ | ✅ | ❌ | ❌ |

Warga sekolah hanya dikelola pada sekolah aktif milik Admin Sekolah. Role yang
diterima hanya `student`, `teacher`, dan `admin`; `super_admin` ditolak.
Membership aktif berarti `school_users.deleted_at IS NULL`. Menghapus warga
dari sekolah hanya mengisi `school_users.deleted_at`; akun global di `users`
tidak dihapus. Import/manual add dapat memulihkan membership yang pernah
soft-deleted pada sekolah aktif.

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
| `/materials` | POST | ❌ | ❌ | ✅* | ❌ |
| `/materials/:id` | PATCH/DELETE | ❌ | ✅** | ✅* | ❌ |
| `/assignments` | POST | ❌ | ❌ | ✅* | ❌ |
| `/assignments/:id` | GET | ❌ | ❌ | 📖* | ❌ |
| `/assignments/subject-class/:subjectClassId` | GET | ❌ | 📖** | 📖* | 📖*** |
| `/assignments/teacher-assignments` | GET | ❌ | ❌ | 📖* | ❌ |
| `/assignments/teacher-submissions` | GET | ❌ | ❌ | 📖* | ❌ |
| `/assignments/student-assignments` | GET | ❌ | ❌ | ❌ | 📖*** |
| `/assignments/:id` | PATCH/DELETE | ❌ | ✅** | ✅* | ❌ |
| `/assignments/submit/:id` | POST | ❌ | ❌ | ❌ | ✅ |
| `/assignments/submit/:id` | GET | ❌ | ❌ | 📖* | ❌ |
| `/assignments/submit/:id` | PATCH/DELETE | ❌ | ❌ | ❌ | ✅**** |
| `/assignments/assess/:id` | POST/PATCH/DELETE | ❌ | ❌ | ✅* | ❌ |
| `/medias/upload` | POST | ❌ | ✅** | ✅** | ✅** |
| `/medias/:id` | DELETE | ❌ | ✅** | ✅***** | ✅***** |
| `/notes` | GET | ❌ | ❌ | ❌ | ✅****** |
| `/notes/material/:materialId` | GET/PUT/DELETE | ❌ | ❌ | ❌ | ✅****** |
| `/notes/subject-class/:subjectClassId` | GET | ❌ | ❌ | ❌ | ✅****** |

*Teacher material/assignment creation, mutation, assignment detail, submission detail, and assessment access is limited to subject classes taught by the current teacher in the active `SchoolId` context.
**Admin and shared media access is scoped to active `SchoolId`.
***Student material/assignment read access is limited to subject classes in classes where the student is enrolled.
****Student submission mutation is limited to the current JWT user's own submission in the active school.
*****Non-admin media deletion is limited to media owned/uploaded by the current JWT user in the active school.
******Student material notes are private to the current JWT user. Access requires active `SchoolId` and active student enrollment (`left_at IS NULL`) in the material or requested subject class's class. Material note access excludes deleted materials, collection responses include only the current user's notes, notes are not exposed to teacher/admin roles, and deletion is a hard delete.

Media IDs attached to materials, assignments, and submissions are validated before linking: media must exist, belong to the active school, and be attachable by the actor. Non-admin users can attach only their own uploaded media; admins can attach active-school media where admin mutation is allowed. Assignment categories used by assignments must belong to the active school.

### Subject Class Assignment
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/subject-classes/assign` | POST | ❌ | ✅* | ❌ | ❌ |
| `/subject-classes/class/:classId` | GET | 📖 | 📖 | 📖 | 📖 |
| `/subject-classes/:id` | GET | 📖 | 📖 | 📖 | 📖 |
| `/subject-classes/:id` | PATCH | ❌ | ✅* | ❌ | ❌ |
| `/subject-classes/:id` | DELETE | ❌ | ✅* | ❌ | ❌ |
| `/subject-classes/my-teaching` | GET | ❌ | ❌ | 📖** | ❌ |

*Admin subject_class assignment requires active `SchoolId`; class, subject, and teacher school_user must belong to the active school. The teacher school_user must have school role `teacher` and must already be enrolled in the class with `class_role = teacher`.
Subject_class unassign is admin-only and active-school scoped. It is allowed only for empty setup mistakes; the API blocks removal when the subject_class already has materials or assignments.

**Teacher workspace access uses JWT user identity plus active `SchoolId`; teachers only see subject classes they teach while still actively enrolled in the class as `teacher` (`left_at IS NULL`).

### Grade Book
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/grades/my-grades/:classId` | GET | ❌ | ❌ | ❌ | 📖* |
| `/grades/weights/subject/:subjectId` | GET | 📖 | 📖 | 📖 | 📖 |
| `/grades/weights` | POST | ❌ | ✅ | ❌ | ❌ |
| `/grades/class/:classId/subject/:subjectId` | GET | ❌ | 📖 | 📖 | ❌ |

*Student gradebook access is current-user only. The student identity comes from JWT, the school context comes from `SchoolId`, and the class must be a class where the current student is enrolled.
Assessment weight management is admin-only for MVP. Weights are subject-level, school-scoped through subject/category ownership, and are used for provisional weighted grades.

### Feeds
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/feeds` | POST | ❌ | ✅* | ✅* | ❌ |
| `/feeds/class/:classId` | GET | ❌ | 📖* | 📖* | 📖* |
| `/feeds/:id` | GET | ❌ | 📖* | 📖* | 📖* |
| `/feeds/:id` | PATCH | ❌ | ✅* | ✅* | ❌ |
| `/feeds/:id` | DELETE | ❌ | ✅* | ✅* | ❌ |
| `/comments` | POST | ❌ | ✅** | ✅** | ✅** |
| `/comments?type=feed&id=` | GET | ❌ | 📖** | 📖** | 📖** |
| `/comments/:id` | GET | ❌ | 📖** | 📖** | 📖** |
| `/comments/:id` | PATCH | ❌ | ❌ | ✅*** | ✅*** |
| `/comments/:id` | DELETE | ❌ | ✅** | ✅*** | ✅*** |

*Feed access is scoped to active `SchoolId`. Feed is class-level. Admin can manage feeds in active-school classes. Teacher can create/read/update/delete only in classes they actively teach, and teacher update/delete is limited to their own feed posts. Student can read only feeds from classes where the current student has active enrollment (`left_at IS NULL`). Feed attachments, comments UI, reactions, and realtime are deferred from the MVP.

**Comments are feed-only for MVP. Admin can access active-school feed comments. Teacher can access comments only on feed posts for classes they actively teach. Student can access comments only on feed posts for classes where they are actively enrolled (`left_at IS NULL`). Non-feed comments are post-MVP and rejected.

***Teacher/student comment update/delete is limited to their own comments.

### User Management
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/users` | POST | ✅ | ❌ | ❌ | ❌ |
| `/users` | GET | 📖 | ✅ | ❌ | ❌ |
| `/users/:id` | GET | 📖 | ❌ | ❌ | ❌ |
| `/users/:id` | PATCH | ✅ | ❌ | ❌ | ❌ |
| `/users/change-password/:id` | PATCH | ✅ | ❌ | ❌ | ❌ |
| `/users/:id` | DELETE | ✅ | ❌ | ❌ | ❌ |

Global user creation is platform scope and only allowed for a `super_admin` membership
on the system school where `schools.sch_code = "000000"`.
School admins manage existing global users as school memberships and assign roles in
their active school context.
Public `/register` remains available for users creating their own plain global
account, but it does not grant school membership, roles, or enrollment.

`users.deleted_at` applies to the global login account. `school_users.deleted_at`
applies only to one school membership and must be filtered out for active school
contexts, role checks, and tenant-scoped member lists.

### Enrollment
| Endpoint | Method | super_admin | admin | teacher | student |
|----------|--------|-------------|-------|---------|---------|
| `/enrollments` | GET | 📖* | 📖* | 📖* | 📖* |
| `/enrollments` | POST | ❌ | ✅* | ❌ | ❌ |
| `/enrollments/:id` | PATCH | ❌ | ✅* | ❌ | ❌ |
| `/enrollments/:id` | DELETE | ❌ | ✅* | ❌ | ❌ |

*Enrollment access is scoped to the active `SchoolId`. Classes, school users,
and enrollment records must belong to the active school. Active enrollment means
`left_at IS NULL`; unenroll sets `left_at = now()` instead of deleting the row.
Re-enroll clears `left_at` on the same row and preserves the original
`joined_at`. Admin unenroll preserves academic history and blocks teacher removal
while the teacher is still assigned to any subject_class in the same class.

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
