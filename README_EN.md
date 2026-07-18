# Wiyata LMS

Wiyata is a working multi-tenant Learning Management System (LMS) with a Go backend API and a Vue frontend for structured digital learning. It supports school onboarding, academic management, subject-class workspaces, materials, assignments, submissions, assessment, class feed, discussions, chat, and notifications.

[Versi Indonesia](README.md)

### Key Highlights

- Multi-tenant LMS architecture supporting multiple schools.
- Backend built with Go, Gin, GORM, and PostgreSQL/Supabase.
- Frontend built with Vue 3, TypeScript, Vite, and Tailwind CSS.
- JWT authentication and role-based access control through school memberships.
- Global user identity with per-school memberships and school-level roles.
- Runtime context uses one active school plus one active school role; platform `super_admin` is separate.
- REST APIs for academic, onboarding, notification, chat, and learning workflows.
- Layered backend architecture: Handler → Service → Repository → Domain.
- Winner of 1st Place in a university-wide UI/UX competition.

### Tech Stack

- Go, Gin, GORM, PostgreSQL/Supabase.
- Vue 3, TypeScript, Vite, Tailwind CSS.
- JWT authentication and bcrypt password hashing.
- SMTP email foundation with no-op fallback.
- Supabase-compatible media storage provider.

## Table of Contents

- [Overview](#overview)
- [Repository Structure](#repository-structure)
- [Quick Start](#quick-start)
- [Technology](#technology)
- [Documentation](#documentation)
- [Architecture](#architecture)
- [Development](#development)
- [Code Conventions](#code-conventions)
- [Important Design Decisions](#important-design-decisions)
- [Known Issues](#known-issues)

## Overview

Wiyata LMS integrates:

- Self-service school creation (verify email, then create a school and become its Admin instantly) and teacher/student invitations.
- School academic structure: academic years, terms, classes, and subjects.
- Daily learning workspace per subject class.
- Learning materials with progress tracking.
- Assignment submission and teacher assessment.
- Class feed, material/assignment/feed discussions, and chat.
- Notification Center and unread badges.
- Global user management, school memberships, and school-scoped RBAC.

### Mental Model

```
School (tenant root)
  ├─ Academic Year → Term → Class
  │   └─ SubjectClass (class + subject + teacher)
  │       ├─ Learning materials
  │       └─ Assignments & grading
  ├─ Student/teacher class placements
  ├─ Class feed (cross-subject communication)
  └─ Comments/discussions (feed, material, and assignment)
```

Materials and assignments live in **SubjectClass**, not directly in Class. Feed lives at **Class** level for communication across all subjects.

User identity is global. Academic access comes from `school_users` and `user_roles`; school-scoped requests use `SchoolId` and, in the current frontend, `Active-Role`.

## Repository Structure

```
wiyata-lms/
├── README.md (Indonesian)
├── README_EN.md (English)
├── ANALYSIS_INDEX.md (documentation navigation)
├── CODEBASE_ANALYSIS.md (technical analysis)
├── QUICK_REFERENCE.md (quick reference)
├── PRODUCT_SCOPE.md (product decisions)
│
├── backend/
│   ├── cmd/api/main.go (application entry point)
│   ├── go.mod & go.sum (Go dependencies)
│   ├── internal/
│   │   ├── domain/ (entity models)
│   │   ├── dto/ (request/response contracts)
│   │   ├── handler/ (HTTP handlers)
│   │   ├── service/ (business logic)
│   │   ├── repository/ (database access)
│   │   ├── middleware/ (JWT, RBAC, school context)
│   │   └── storage/ (file upload providers)
│   ├── schema.md (database schema in DBML format)
│   ├── AGENT.md (engineering context)
│   ├── PROJECT_CONTEXT.md (business context)
│   ├── TODO.md (current backend backlog)
│   └── .env.example (configuration example)
│
├── frontend/ (Vue application)
└── docs/ (additional documentation)
```

## Quick Start

### Prerequisites

- Go 1.21 or newer.
- PostgreSQL 13 or newer.
- Node.js and npm for the frontend.

### Development Setup

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

The backend runs at `http://localhost:8080`.

2. Setup frontend:

```bash
cd frontend
npm install
cp env.example .env
npm run dev
```

The frontend dev server follows the Vite configuration and `VITE_API_BASE_URL`.

### Testing

```bash
# Backend tests
cd backend
go test ./...

# Frontend production build
cd frontend
npm run build
```

### Build

```bash
cd backend
go build -o wiyata ./cmd/api
./wiyata
```

## Technology

### Backend Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL/Supabase with GORM
- **Authentication**: JWT
- **Password Hashing**: bcrypt
- **Configuration**: dotenv-compatible `.env` files

### Frontend Stack

- **Framework**: Vue 3
- **Language**: TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **State/Context**: Pinia store plus localStorage persistence for auth/session context

## Architecture

Backend follows a layered architecture:

```
HTTP Request
    ↓
Middleware (Auth, RBAC, School context)
    ↓
Handler (HTTP parsing, DTO binding)
    ↓
Service (Business logic, notifications, email best-effort)
    ↓
Repository (Database queries)
    ↓
PostgreSQL Database
```

Main middleware:

1. **AuthRequired** validates JWT and extracts the global user.
2. **RequireSchoolMember** validates active school membership from `SchoolId`.
3. **RequireRole** authorizes the selected `Active-Role` when present, with temporary legacy fallback where still supported.
4. **RequireSystemSuperAdmin** protects platform super admin routes.

## Documentation

Useful starting points:

- `docs/AI_HANDOFF.md` - read-first guide for future AI coding agents.
- `backend/schema.md` - database schema reference.
- `backend/docs/api/` - focused API documentation.
- `README.md` / `README_EN.md` - project overview and setup.

When documentation conflicts with the implementation, inspect current code and tests. Do not change working behavior only to match stale documentation.

## Development

### Environment Variables

Backend `.env` example:

```env
DB_DSN=postgres://user:password@localhost:5432/wiyata_dev
JWT_SECRET=your-super-secret-key-change-this-in-production
JWT_EXPIRY=24h
STORAGE_PROVIDER=disabled
APP_PUBLIC_URL=http://localhost:5173
SMTP_ENABLED=false
```

Do not commit real credentials, JWTs, invitation tokens, or SMTP passwords.

### Adding a Feature

Follow the existing code path:

1. Database/schema documentation.
2. Domain model.
3. DTO request/response contract.
4. Repository data access.
5. Service business logic.
6. Handler HTTP layer.
7. Route registration in `cmd/api/main.go`.
8. Frontend service/types/pages when needed.

## Code Conventions

- Database columns use snake_case and existing table prefixes.
- JSON fields use camelCase.
- Keep backend dependencies flowing Handler → Service → Repository → Domain.
- Keep notifications and email best-effort unless a workflow explicitly requires hard failure.
- Prefer existing UI patterns over broad redesign.

## Important Design Decisions

1. **Class vs SubjectClass**: Materials and assignments live in SubjectClass. Feed lives at Class level.
2. **Global User Identity**: A user account is global and can belong to multiple schools.
3. **School-Scoped RBAC**: Roles attach to school memberships, not to the global user record.
4. **Active Context**: Runtime frontend/backend behavior uses one active school plus one active school role. Platform `super_admin` is a separate context.
5. **Soft Deletes**: Main entities use soft deletes where the codebase already models them.
6. **Upsert Semantics**: Submission and assessment are one-per-student/per-assignment flows.
7. **Best-Effort Notifications and Email**: Main actions should not fail merely because notification/email delivery fails.
8. **Polymorphic Comments**: Feed, material, and assignment discussions use the shared comments system.
9. **Password Safety**: Passwords are never sent by email; existing user passwords are not overwritten during invitation/direct-create reuse.

## Known Issues

1. **Historical documentation**: Some older docs remain planning/analysis material. Treat current implementation and tests as source of truth.
2. **Enrollment role derivation**: Frontend class placement now infers student/teacher role from school member roles, but backend still receives the `role` payload and does not yet authoritatively derive it.
3. **Notification realtime**: Notification Center and unread state currently use REST/frontend refresh. WebSocket realtime exists for chat, not general notifications.
4. **File delivery**: Upload exists, but signed/private download URLs and thumbnail generation remain follow-up work.
5. **Large frontend bundles/warnings**: Build may report non-blocking chunk-size or CSS import-order warnings; inspect current build output before treating them as blockers.

## Deferred Work

- Assignment extension request/review flow.
- Signed/private file URLs and media thumbnails.
- Grade/transcript export.
- Notification preferences and optional realtime notification delivery.
- Nested comment threading if product decides to support it.
- Backend-authoritative class placement role derivation.

## Contributing

When contributing:

1. Follow the existing layer structure and local patterns.
2. Read comparable files before adding abstractions.
3. Run `gofmt` for backend changes and `npm run build` for frontend changes.
4. Run `git diff --check`.
5. Update documentation when contracts, context, or product behavior changes.
