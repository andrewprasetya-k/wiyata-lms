# Wiyata LMS

Wiyata is a multi-school Learning Management System (LMS) designed to manage digital learning processes in a structured way. The platform supports academic management, class-based learning, assignment submission, grading, class communication, and notifications.

[Versi Indonesia](README.md)

# Wiyata LMS

Wiyata is a multi-school Learning Management System (LMS) designed to support digital learning, academic management, assignments, assessments, and classroom collaboration.

### Key Highlights

- Multi-tenant LMS architecture supporting multiple schools
- Backend built with Go (Gin), PostgreSQL, and GORM
- JWT authentication and role-based access control (RBAC)
- RESTful APIs for academic management workflows
- Layered architecture (Handler → Service → Repository → Domain)
- 22 database tables covering academic structure, learning content, communication, and access control
- Winner of 1st Place in a university-wide UI/UX competition

### Tech Stack

- Go (Gin)
- PostgreSQL
- GORM
- JWT Authentication
- Vue.js (In Progress)

## Table of Contents

- [Overview](#overview)
- [Repository Structure](#repository-structure)
- [Quick Start](#quick-start)
- [Technology](#technology)
- [Documentation](#documentation)
- [Architecture](#architecture)
- [Development](#development)
- [Code Conventions](#code-conventions)
- [Design Decisions](#important-design-decisions)
- [Known Issues](#known-issues)

## Overview

Wiyata LMS is an academic platform that integrates:

- School academic structure (academic years, terms, classes, subjects)
- Daily learning workspace per subject (subject class)
- Learning materials with progress tracking
- Assignment system with submission and grading
- Class communication and feedback
- Automatic notifications for learning activities
- User management and role-based access control (RBAC)

### Mental Model

```
School (root tenant)
  ├─ Academic Year → Term → Class
  │   └─ SubjectClass (class + subject + teacher)
  │       ├─ Learning materials
  │       └─ Assignments & grading
  ├─ Student/teacher enrollment to class
  ├─ Class feed (cross-subject communication)
  └─ Comments (can be attached to material, assignment, feed, or submission)
```

Key point: Materials and assignments live in **SubjectClass**, not in Class. Feed lives at **Class** level for communication across all subjects.

## Repository Structure

```
wiyata-lms/
├── README.md (Indonesian version)
├── README_EN.md (English version)
├── ANALYSIS_INDEX.md (documentation navigation guide)
├── CODEBASE_ANALYSIS.md (comprehensive technical analysis)
├── QUICK_REFERENCE.md (quick reference)
├── PRODUCT_SCOPE.md (product scope and decisions)
│
├── backend/
│   ├── cmd/api/main.go (application entry point)
│   ├── go.mod & go.sum (Go dependencies)
│   │
│   ├── internal/
│   │   ├── domain/ (entity models: User, School, Material, etc)
│   │   ├── dto/ (request/response contracts)
│   │   ├── handler/ (HTTP handlers, request parsing)
│   │   ├── service/ (business logic)
│   │   ├── repository/ (database access via GORM)
│   │   ├── middleware/ (auth JWT, RBAC, school context)
│   │   └── storage/ (file upload providers)
│   │
│   ├── schema.md (database schema in DBML format)
│   ├── AGENT.md (engineering context)
│   ├── PROJECT_CONTEXT.md (business context)
│   ├── TODO.md (pending work items)
│   └── .env.example (configuration example)
│
├── frontend/ (frontend application)
└── docs/ (additional documentation)
```

## Quick Start

### Prerequisites

- Go 1.21 or newer
- PostgreSQL 13 or newer

### Development Setup

1. Navigate to backend directory:

```bash
cd backend
```

2. Install dependencies:

```bash
go mod download
```

3. Setup .env file:

```bash
cp .env.example .env
# Edit .env and adjust values:
# DB_DSN=postgres://user:password@localhost:5432/wiyata_dev
# JWT_SECRET=your-secret-key-here
```

4. Run the application:

```bash
go run ./cmd/api
```

Server will run at `http://localhost:8080`

### Testing

```bash
# Run all tests
go test ./...

# Run with verbose
go test -v ./...

# Test specific package
go test ./internal/service/...
```

### Build

```bash
# Build for production
go build -o wiyata ./cmd/api

# Run the binary
./wiyata
```

### Code Formatting

```bash
# Check format
gofmt -l .

# Format all files
gofmt -w .
```

## Technology

### Backend Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin (HTTP routing and middleware)
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT (JSON Web Tokens)
- **Password Hashing**: bcrypt
- **Configuration**: dotenv (.env files)

### Architecture Pattern

Backend follows a strict layered architecture pattern:

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

Each layer has clear and separated responsibilities.

## Documentation

### Documentation Files

| File                       | Size   | Purpose                                        |
| -------------------------- | ------ | ---------------------------------------------- |
| README.md                  | 14 KB  | Main guide (Indonesian)                        |
| README_EN.md               | 14 KB  | Main guide (English)                           |
| ANALYSIS_INDEX.md          | 7.9 KB | Documentation navigation guide                 |
| CODEBASE_ANALYSIS.md       | 44 KB  | Comprehensive technical analysis (16 sections) |
| QUICK_REFERENCE.md         | 8.3 KB | Quick reference patterns and queries           |
| PRODUCT_SCOPE.md           | 17 KB  | Product scope and business decisions           |
| backend/schema.md          | -      | Database schema (DBML format)                  |
| backend/AGENT.md           | -      | Engineering context summary                    |
| backend/PROJECT_CONTEXT.md | -      | Business context                               |

### Reading Guide for New Developers

1. **README_EN.md** (this file) - overview
2. **QUICK_REFERENCE.md** - basic patterns and structure
3. **PRODUCT_SCOPE.md** - business context and requirements
4. **CODEBASE_ANALYSIS.md** - technical deep dive

For specific navigation: **ANALYSIS_INDEX.md**

## Architecture

### 4 Application Layers

#### 1. Handler Layer (23 handlers)

Handles HTTP requests and responses:

- Parse HTTP request body into DTO
- Extract user identity from JWT context
- Call service layer
- Map service results to response DTO
- Return HTTP response with appropriate status code

Examples: `internal/handler/material_handler.go`, `internal/handler/assignment_handler.go`

#### 2. Service Layer (21 services)

Contains business logic:

- Validate business rules
- Coordinate between repositories
- Trigger notifications (best-effort)
- Orchestrate complex workflows
- Handle file uploads to storage provider

Examples: `internal/service/material_service.go`, `internal/service/feed_service.go`

#### 3. Repository Layer (22 repositories)

Handles database access:

- Query database via GORM
- Handle soft deletes
- Preload related entities
- Validate RowsAffected for updates/deletes
- Abstract SQL from upper layers

Examples: `internal/repository/material_repo.go`, `internal/repository/assignment_repo.go`

#### 4. Domain Layer (19 domain models)

Defines entities and business rules:

- Define GORM models with table names and relationships
- Define constants and enums (NotificationType, SourceType, etc)
- No business logic, only data structures

Examples: `internal/domain/material.go`, `internal/domain/assignment.go`

### Middleware

Main middleware protecting routes:

1. **AuthRequired** - validate JWT token and extract user_id
2. **RequireSchoolMember** - verify user is school member
3. **RequireRole** - verify user has required role

Middleware chain on specific routes:

```
Route → AuthRequired → RequireSchoolMember → RequireRole("teacher") → Handler
```

### Database Schema

Database consists of 22 main tables:

**Academic Structure**: schools, academic_years, terms, subjects, classes, subject_classes

**Users & Access Control**: users, school_users, roles, user_roles, enrollments

**Learning Content**: materials, material_progress, assignments, submissions, assessments, assignment_categories

**Communication**: feeds, comments, attachments, medias

**System**: notifications, logs

See `backend/schema.md` for full details.

## Development

### Environment Variables

`.env` file in `backend/` directory:

```
# Database
DB_DSN=postgres://user:password@localhost:5432/wiyata_dev

# JWT
JWT_SECRET=your-super-secret-key-change-this-in-production
JWT_EXPIRY=24h

# Storage (optional)
STORAGE_PROVIDER=local
STORAGE_PATH=./uploads
```

### Running Locally

1. Ensure PostgreSQL is running
2. Create database: `createdb wiyata_dev`
3. Set environment variables in `.env`
4. Run: `go run ./cmd/api`

### Project Structure Best Practices

- Each package (domain, service, repository) is self-contained
- Dependency flows only from handler → service → repository → domain
- Circular dependencies must be avoided
- Interfaces used for abstraction in service and repository layer

### Adding New Feature

Follow this order:

1. **Database** - Add table to `schema.md` and run migration
2. **Domain** - Create model in `internal/domain/`
3. **DTO** - Create request/response DTO in `internal/dto/`
4. **Repository** - Implement data access in `internal/repository/`
5. **Service** - Implement business logic in `internal/service/`
6. **Handler** - Implement HTTP handlers in `internal/handler/`
7. **Routes** - Register routes in `cmd/api/main.go`

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

Use centralized error handler from `internal/handler/error_handler.go`:

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

Provide comments only on non-obvious logic:

```go
// Good: explains non-obvious logic
// Calculate late submission status by comparing submission time with deadline
isLate := submission.SubmittedAt.After(*assignment.Deadline)

// Unnecessary: obvious from code
user := getUser() // Get user
```

### Code Organization in File

```go
package service

import (
    "errors"
    "fmt"
    "gorm.io/gorm"

    "backend/internal/domain"
    "backend/internal/repository"
)

// Interface definition first
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

## Important Design Decisions

These design decisions are non-negotiable and should not be changed without approval:

1. **Class vs SubjectClass**: Materials and assignments live in SubjectClass, not Class. Feed lives at Class level.

2. **Soft Deletes**: All main entities (except linking tables) have soft delete via `deleted_at`. Data is recoverable.

3. **Upsert Semantics**: Submission and Assessment use upsert (1 per student per assignment). Resubmit/re-grade overwrites.

4. **Best-Effort Notifications**: Notifications don't cascade failures. Main action succeeds even if notification fails.

5. **School Multi-Tenancy**: School is root tenant. All data (except global user) is isolated per school.

6. **RBAC per School**: User can have different roles in different schools. Roles attach to school membership, not global user.

7. **Polymorphic Comments**: Comments can be attached to material, assignment, feed, submission, or another comment via SourceType + SourceID.

## Known Issues

1. **Route Ordering**: GET `/assignments/status/:id` can be swallowed by GET `/assignments/:assignmentId` due to dynamic route matching.

2. **No Unit Tests**: No test files yet. Structure is ready but not implemented.

3. **Code Formatting**: Many files don't follow gofmt standard.

4. **Missing Auth Checks**: Assessment doesn't verify teacher owns assignment. Submission doesn't verify student is enrolled in class.

5. **File Storage**: Storage provider is still a stub. Real S3/Supabase integration not implemented.

## Deferred Features (Out of Scope)

Following features are planned but not scoped for MVP:

- Realtime chat WebSocket
- Student personal notes per material
- Email notification delivery
- Signed/private file URLs
- Auto thumbnail generation from video
- Nested comment threading

## Contributing

When contributing to this project:

1. Follow existing layer structure (Handler → Service → Repository → Domain)
2. Look at similar files for pattern consistency
3. Run `gofmt` before committing
4. Write tests for new business logic
5. Update documentation if changing architecture

---
