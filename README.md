# Wiyata LMS

Wiyata adalah Learning Management System (LMS) multi-sekolah yang sudah memiliki backend API dan frontend Vue untuk mengelola proses pembelajaran digital secara terstruktur. Platform ini mendukung onboarding sekolah, manajemen akademik, pembelajaran berbasis kelas/mata pelajaran, pengumpulan tugas, penilaian, feed, diskusi, chat, dan notifikasi.

[English Version](README_EN.md)

### Key Highlights

- Multi-tenant LMS architecture supporting multiple schools
- Backend built with Go (Gin), PostgreSQL, and GORM
- Frontend built with Vue 3, TypeScript, Vite, and Tailwind CSS
- JWT authentication and role-based access control (RBAC) through school memberships
- Global user identity with per-school memberships and school-level roles
- Runtime context uses one active school plus one active school role; platform `super_admin` is separate
- RESTful APIs for academic management workflows
- Layered architecture (Handler → Service → Repository → Domain)
- Winner of 1st Place in a university-wide UI/UX competition

### Tech Stack

- Go (Gin), GORM, PostgreSQL/Supabase
- Vue 3, TypeScript, Vite, Tailwind CSS
- JWT Authentication
- SMTP email foundation with no-op fallback
- Supabase-compatible media storage provider

## Daftar Isi

- [Gambaran Umum](#gambaran-umum)
- [Struktur Repository](#struktur-repository)
- [Quick Start](#quick-start)
- [Teknologi](#teknologi)
- [Dokumentasi](#dokumentasi)
- [Arsitektur](#arsitektur)
- [Development](#development)
- [Konvensi Kode](#konvensi-kode)
- [Keputusan Desain](#keputusan-desain-penting)
- [Issue Diketahui](#issue-yang-diketahui)

## Gambaran Umum

Wiyata LMS adalah platform akademik yang mengintegrasikan:

- Onboarding sekolah, approval super admin, undangan admin sekolah, dan undangan guru/siswa
- Struktur akademik sekolah (tahun ajaran, semester, kelas, mata pelajaran)
- Ruang kerja pembelajaran harian per mata pelajaran (subject class)
- Materi pembelajaran dengan pelacakan progres
- Sistem tugas dengan pengumpulan dan penilaian
- Feed kelas, diskusi materi/tugas/feed, dan chat
- Notification Center dan unread badges
- Manajemen pengguna global, membership sekolah, dan kontrol akses berbasis peran (RBAC)

### Model Mental Utama

```
Sekolah (tenant root)
  ├─ Tahun Ajaran → Semester → Kelas
  │   └─ SubjectClass (kelas + mata pelajaran + guru)
  │       ├─ Materi pembelajaran
  │       └─ Tugas & penilaian
  ├─ Penempatan siswa/guru ke kelas
  ├─ Feed kelas (komunikasi lintas mata pelajaran)
  └─ Komentar/diskusi (feed, materi, dan tugas)
```

Poin penting: Materi dan tugas hidup di **SubjectClass**, bukan di kelas. Feed hidup di level **Kelas** untuk komunikasi yang meliputi semua mata pelajaran.

Identitas user bersifat global. Akses akademik berasal dari `school_users` dan `user_roles`; setiap request school-scoped menggunakan `SchoolId` dan, pada frontend saat ini, `Active-Role`.

## Struktur Repository

```
wiyata-lms/
├── README.md (Bahasa Indonesia)
├── README_EN.md (English)
├── ANALYSIS_INDEX.md (panduan navigasi dokumentasi)
├── CODEBASE_ANALYSIS.md (analisis teknis lengkap)
├── QUICK_REFERENCE.md (referensi cepat)
├── PRODUCT_SCOPE.md (scope produk dan keputusan produk)
│
├── backend/
│   ├── cmd/api/main.go (entry point aplikasi)
│   ├── go.mod & go.sum (dependensi Go)
│   │
│   ├── internal/
│   │   ├── domain/ (model entity: User, School, Material, dll)
│   │   ├── dto/ (request/response contracts)
│   │   ├── handler/ (HTTP handlers, parsing request)
│   │   ├── service/ (business logic)
│   │   ├── repository/ (akses database via GORM)
│   │   ├── middleware/ (auth JWT, RBAC, school context)
│   │   └── storage/ (file upload providers)
│   │
│   ├── schema.md (database schema dalam format DBML)
│   ├── AGENT.md (engineering context)
│   ├── PROJECT_CONTEXT.md (business context)
│   ├── TODO.md (daftar pekerjaan yang ditunda)
│   └── .env.example (contoh konfigurasi)
│
├── frontend/ (aplikasi frontend)
└── docs/ (dokumentasi tambahan)
```

## Quick Start

### Prerequisites

- Go 1.21 atau lebih baru
- PostgreSQL 13 atau lebih baru
- Node.js dan npm untuk frontend

### Setup Development

1. Setup backend:

```bash
cd backend
go mod download
cp .env.example .env
# Edit .env:
# DB_DSN=postgres://user:password@localhost:5432/wiyata_dev
# JWT_SECRET=your-secret-key-here
go run ./cmd/api
```

Server backend berjalan di `http://localhost:8080`.

2. Setup frontend:

```bash
cd frontend
npm install
cp env.example .env
npm run dev
```

Frontend dev server mengikuti konfigurasi Vite dan `VITE_API_BASE_URL`.

### Build

```bash
# Build untuk production
go build -o wiyata ./cmd/api

# Jalankan binary
./wiyata
```

### Code Formatting

```bash
# Check format
gofmt -l .

# Format semua file
gofmt -w .
```

## Teknologi

### Backend Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin (HTTP routing dan middleware)
- **Database**: PostgreSQL/Supabase dengan GORM ORM
- **Authentication**: JWT (JSON Web Tokens)
- **Password Hashing**: bcrypt
- **Configuration**: dotenv (.env files)

### Frontend Stack

- **Framework**: Vue 3
- **Language**: TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **State/Context**: Pinia store plus localStorage persistence for auth/session context

### Architecture Pattern

Backend mengikuti pola arsitektur berlapis yang ketat:

```
HTTP Request
    ↓
Middleware (Auth, RBAC, School context)
    ↓
Handler (HTTP parsing, DTO binding)
    ↓
Service (Business logic, notifications)
    ↓
Repository (Database queries)
    ↓
PostgreSQL Database
```

Setiap layer memiliki tanggung jawab yang jelas dan terpisah.

## Dokumentasi

### File-file Dokumentasi

| File | Tujuan |
| --- | --- |
| docs/AI_HANDOFF.md | Panduan read-first untuk AI coding agent dan developer baru |
| README.md | Panduan utama Bahasa Indonesia |
| README_EN.md | Panduan utama English |
| backend/schema.md | Referensi skema database dalam format DBML |
| backend/docs/api/ | Dokumentasi API yang lebih spesifik |
| docs/ANALYSIS_INDEX.md | Panduan navigasi dokumentasi historis |
| docs/CODEBASE_ANALYSIS.md | Analisis teknis historis; verifikasi ulang dengan kode saat ini |
| docs/QUICK_REFERENCE.md | Referensi cepat pola dan query |
| docs/PRODUCT_SCOPE.md | Scope produk dan keputusan bisnis |

### Urutan Membaca untuk Developer Baru

1. **docs/AI_HANDOFF.md** - konteks implementasi terkini dan aturan keamanan
2. **README.md** (file ini) - gambaran umum dan setup
3. **backend/docs/api/** - kontrak API spesifik
4. **backend/schema.md** - referensi skema
5. **Dokumentasi historis di docs/** - gunakan sebagai referensi, lalu verifikasi dengan kode

Jika butuh navigasi spesifik: **ANALYSIS_INDEX.md**

## Arsitektur

### 4 Layer Aplikasi

#### 1. Handler Layer

Menangani HTTP requests dan responses:

- Parse HTTP request body menjadi DTO
- Extract user identity dari JWT context
- Memanggil service layer
- Map hasil service ke response DTO
- Return HTTP response dengan status code sesuai

Contoh: `internal/handler/material_handler.go`, `internal/handler/assignment_handler.go`

#### 2. Service Layer

Mengandung business logic:

- Validasi business rules
- Koordinasi antar repository
- Trigger notifications (best-effort)
- Orchestrate complex workflows
- Handle file uploads ke storage provider

Contoh: `internal/service/material_service.go`, `internal/service/feed_service.go`

#### 3. Repository Layer

Menangani akses database:

- Query database via GORM
- Handle soft deletes
- Preload related entities
- Validasi RowsAffected untuk updates/deletes
- Abstraksi SQL dari layer atas

Contoh: `internal/repository/material_repo.go`, `internal/repository/assignment_repo.go`

#### 4. Domain Layer

Mendefinisikan entity dan business rules:

- Define GORM models dengan table names dan relationships
- Define constants dan enums (NotificationType, SourceType, dll)
- Tidak ada business logic, hanya struktur data

Contoh: `internal/domain/material.go`, `internal/domain/assignment.go`

### Middleware

Middleware utama melindungi routes:

1. **AuthRequired** - validasi JWT token dan extract user_id
2. **RequireSchoolMember** - verifikasi user adalah member sekolah
3. **RequireRole** - verifikasi user memiliki peran yang diperlukan

Middleware chain pada route tertentu:

```
Route → AuthRequired → RequireSchoolMember → RequireRole("teacher") → Handler
```

### Database Schema

Database memakai schema `edv` dengan tabel untuk:

**Academic Structure**: schools, academic_years, terms, subjects, classes, subject_classes

**Users & Access Control**: users, school_users, roles, user_roles, enrollments

**Learning Content**: materials, material_progress, assignments, submissions, assessments, assignment_categories

**Communication**: feeds, comments, attachments, medias, chat rooms/messages/receipts

**System**: notifications, logs

Lihat `backend/schema.md` untuk detail lengkap.

## Development

### Environment Variables

File `.env` di direktori `backend/`:

```
# Database
DB_DSN=postgres://user:password@localhost:5432/wiyata_dev

# JWT
JWT_SECRET=your-super-secret-key-change-this-in-production
JWT_EXPIRY=24h

# Storage (optional)
STORAGE_PROVIDER=disabled

# Public app URL and SMTP (optional)
APP_PUBLIC_URL=http://localhost:5173
SMTP_ENABLED=false
```

### Running Locally

1. Pastikan PostgreSQL sudah berjalan
2. Buat database: `createdb wiyata_dev`
3. Set environment variables di `.env`
4. Run: `go run ./cmd/api`

### Project Structure Best Practices

- Setiap paket (domain, service, repository) adalah self-contained
- Dependency hanya mengalir dari handler → service → repository → domain
- Circular dependencies harus dihindari
- Interface digunakan untuk abstraksi di service dan repository layer

### Menambah Feature Baru

Ikuti urutan ini:

1. **Database** - Tambah table ke `schema.md` dan jalankan migration
2. **Domain** - Buat model di `internal/domain/`
3. **DTO** - Buat request/response DTO di `internal/dto/`
4. **Repository** - Implement data access di `internal/repository/`
5. **Service** - Implement business logic di `internal/service/`
6. **Handler** - Implement HTTP handlers di `internal/handler/`
7. **Routes** - Register routes di `cmd/api/main.go`

### Testing

Struktur untuk testing sudah siap:

```bash
# Test akan ditempatkan di:
internal/service/material_service_test.go
internal/repository/material_repo_test.go
internal/handler/material_handler_test.go
```

Gunakan standard Go testing dengan package `testing`.

## Konvensi Kode

### Naming Conventions

#### Domain/Models

```go
type Material struct {
    ID             string
    SchoolID       string
    SubjectClassID string
    Title          string
    CreatedBy      string
    CreatedAt      time.Time
}

func (Material) TableName() string {
    return "edv.materials"
}
```

#### DTO Naming

```go
// Request DTO: Create + EntityName + DTO
type CreateMaterialDTO struct {
    SchoolID       string `json:"schoolId" binding:"required,uuid"`
    SubjectClassID string `json:"subjectClassId" binding:"required,uuid"`
    Title          string `json:"materialTitle" binding:"required"`
}

// Response DTO: EntityName + ResponseDTO
type MaterialResponseDTO struct {
    ID    string `json:"materialId"`
    Title string `json:"materialTitle"`
}
```

#### JSON Naming

- Database columns: snake_case (mat_id, mat_title)
- JSON fields: camelCase (materialId, materialTitle)

#### Service/Repository Methods

```go
// Service: logical action name
func (s *materialService) Create(...)
func (s *materialService) GetByID(...)
func (s *materialService) UpdateProgress(...)

// Repository: database operation + entity details
func (r *materialRepository) Create(mat *domain.Material) error
func (r *materialRepository) GetByID(id string) (*domain.Material, error)
func (r *materialRepository) UpsertProgress(prog *domain.MaterialProgress) error
```

### Error Handling

Gunakan centralized error handler dari `internal/handler/error_handler.go`:

```go
if err != nil {
    HandleError(c, err)
    return
}

if err := c.ShouldBindJSON(&input); err != nil {
    HandleBindingError(c, err)
    return
}
```

### Comments

Berikan comment hanya pada hal yang perlu klarifikasi:

```go
// Baik: menjelaskan non-obvious logic
// Calculate late submission status by comparing submission time with deadline
isLate := submission.SubmittedAt.After(*assignment.Deadline)

// Tidak perlu: obvious dari code
user := getUser() // Get user
```

### Code Organization dalam File

```go
package service

import (
    "errors"
    "fmt"
    "gorm.io/gorm"

    "backend/internal/domain"
    "backend/internal/repository"
)

// Interface definition dulu
type MaterialService interface {
    Create(...) error
    GetByID(...) (*domain.Material, error)
}

// Struct implementation
type materialService struct {
    repo              repository.MaterialRepository
    attachmentService AttachmentService
}

// Constructor
func NewMaterialService(...) MaterialService {
    return &materialService{...}
}

// Methods
func (s *materialService) Create(...) error {
    ...
}
```

## Keputusan Desain Penting

Keputusan desain ini adalah non-negotiable dan tidak boleh diubah tanpa persetujuan:

1. **Class vs SubjectClass**: Materi dan tugas hidup di SubjectClass, bukan Class. Feed hidup di Class level.

2. **Soft Deletes**: Semua entity utama (kecuali tabel linking) memiliki soft delete via `deleted_at`. Data recoverable.

3. **Upsert Semantics**: Submission dan Assessment menggunakan upsert (1 per student per assignment). Resubmit/re-grade overwrites.

4. **Best-Effort Notifications**: Notifikasi tidak cascade failures. Main action succeeds bahkan jika notifikasi gagal.

5. **School Multi-Tenancy**: School adalah root tenant. Semua data (kecuali user global) disolasi per school.

6. **RBAC per School**: User dapat memiliki role berbeda di sekolah berbeda. Role attach ke school membership, bukan user global.

7. **Polymorphic Comments**: Comment/diskusi saat ini aktif untuk feed, material, dan assignment via SourceType + SourceID. Nested replies/submission discussion masih perlu keputusan produk dan implementasi lanjutan.

## Issue yang Diketahui

1. **Dokumentasi historis**: Beberapa dokumen lama masih bersifat planning/analysis. Jika dokumen konflik dengan implementasi, cek kode dan test terlebih dahulu.

2. **Enrollment role derivation**: Frontend Penempatan Kelas sudah menginfer role dari school member role, tetapi backend masih menerima payload `role` dan belum menjadi sumber derivasi authoritative.

3. **Notification realtime**: Notification Center dan unread state menggunakan REST/frontend refresh. WebSocket realtime saat ini ada untuk chat, bukan notifikasi umum.

4. **File delivery**: Upload ke storage sudah ada, tetapi signed/private download URL dan thumbnail generation masih follow-up.

## Fitur yang Ditunda (Out of Scope)

Fitur-fitur berikut masih direncanakan atau perlu keputusan produk:

- Signed/private file URLs
- Auto thumbnail generation dari video
- Nested comment threading
- Assignment extension request/review flow
- Grade/transcript export
- Notification preferences dan realtime notification delivery

## Kontribusi

Ketika berkontribusi pada project ini:

1. Ikuti struktur layer yang ada (Handler → Service → Repository → Domain)
2. Lihat file yang sama untuk pattern consistency
3. Jalankan `gofmt` sebelum commit
4. Write tests untuk business logic baru
5. Update dokumentasi jika ada architectural changes

---
