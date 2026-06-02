# EDUVERSE LMS - CODEBASE ANALYSIS INDEX

## 📚 Documentation Reference

This folder contains comprehensive analysis of the Eduverse LMS backend codebase. Use this index to navigate and understand different aspects of the system.

### 1. **CODEBASE_ANALYSIS.md** (44 KB)
**→ START HERE for technical deep-dive**

Complete architectural analysis covering:
- **Section 1**: Routes Inventory (90+ endpoints mapped)
- **Section 2**: Handler Layer (23 handlers, request/response patterns)
- **Section 3**: Service Layer (21 services, business logic, notifications)
- **Section 4**: Repository Layer (22 repositories, GORM patterns)
- **Section 5**: Domain Models (19 entities, relationships)
- **Section 6**: DTOs (request/response validation)
- **Section 7**: Validation Rules (business constraints)
- **Section 8**: Database Tables (schema relationships)
- **Section 9**: Business Rules (core workflows, constraints)
- **Section 10**: Side Effects & Dependencies (cascading actions, risks)
- **Section 11**: Middleware Flow (auth, RBAC, context)
- **Section 12**: Error Handling (HTTP status mapping)
- **Section 13**: Request/Response Examples (complete flows)
- **Section 14**: Routing Warnings (known issues)
- **Section 15**: Layer Summary Table
- **Section 16**: Critical Findings & Notes

**Use when**: You need to understand complete system behavior, data flow, or need to trace a specific feature end-to-end.

---

### 2. **QUICK_REFERENCE.md** (8.3 KB)
**→ USE FOR FAST LOOKUPS**

Quick-access guide for:
- Architecture overview (visual)
- Entity relationships diagram
- Critical design decisions table
- Routes quick map (14 API groups)
- Authorization layers
- Key service patterns
- Common SQL queries (reference)
- Environment variables
- Error codes
- Side effects summary table
- Known issues & TODOs
- File structure

**Use when**: You need quick answers, common queries, or pattern references while implementing features.

---

### 3. **PRODUCT_SCOPE.md** (17 KB, existing)
**→ BUSINESS & PRODUCT CONTEXT**

High-level product definition including:
- Product vision and principles
- Target users (Super Admin, Admin, Teacher, Student)
- Domain model overview
- Main business workflows
- Frontend navigation scopes
- Backend contract priorities
- Non-negotiable product decisions

**Use when**: You need to understand business requirements, product design, or user workflows.

---

### 4. **backend/schema.md** (existing)
**→ DATABASE SCHEMA (DBML format)**

Complete PostgreSQL schema including:
- 22 main tables
- Enum types (material_type, source_type, owner_type, class_role, etc.)
- Foreign key relationships
- Unique constraints
- Table comments

**Use when**: You need to understand database structure, verify column names, or check constraints.

---

### 5. **backend/AGENT.md** (existing)
**→ ENGINEERING CONTEXT SUMMARY**

Engineering context report covering:
- What the app does (1-sentence summary)
- Tech stack and runtime (Go, Gin, GORM, PostgreSQL)
- Architecture and folder structure
- Key domains/modules
- Data flow and integrations
- Local setup commands
- Test/lint/build commands
- Coding conventions
- Potential pitfalls
- Recommended next steps

**Use when**: You're new to the project or need quick engineering overview.

---

## 🎯 Navigation by Task

### I need to understand the system
1. Read QUICK_REFERENCE.md (Architecture Overview + Entity Relationships)
2. Read PRODUCT_SCOPE.md (Business context)
3. Read CODEBASE_ANALYSIS.md (Sections 1-5 for layers)

### I need to add a new endpoint
1. QUICK_REFERENCE.md (Routes Quick Map)
2. CODEBASE_ANALYSIS.md (Section 1: similar route pattern)
3. CODEBASE_ANALYSIS.md (Section 2: handler structure)
4. CODEBASE_ANALYSIS.md (Section 3: service pattern)

### I need to add a new entity/feature
1. backend/schema.md (add table/relationships)
2. CODEBASE_ANALYSIS.md (Section 5: domain model pattern)
3. CODEBASE_ANALYSIS.md (Section 8: relationships)
4. Then: DTOs → Handler → Service → Repository

### I need to understand a specific flow
1. CODEBASE_ANALYSIS.md (Section 13: Request/Response Examples)
2. CODEBASE_ANALYSIS.md (Section 10: Side Effects & Dependencies)
3. CODEBASE_ANALYSIS.md (Section 9: Business Rules)

### I need to fix a bug
1. CODEBASE_ANALYSIS.md (Section 16: Critical Findings)
2. CODEBASE_ANALYSIS.md (Section 14: Routing Warnings)
3. CODEBASE_ANALYSIS.md (Section 10: Potential Side Effects)

### I need to add validation/authorization
1. CODEBASE_ANALYSIS.md (Section 7: Validation Rules)
2. CODEBASE_ANALYSIS.md (Section 11: Middleware Flow)
3. CODEBASE_ANALYSIS.md (Section 9: Business Rules)

### I need to understand notifications
1. CODEBASE_ANALYSIS.md (Section 3: NotificationService)
2. QUICK_REFERENCE.md (Notification Triggering pattern)
3. CODEBASE_ANALYSIS.md (Section 10: Notification Side Effects)

### I need to understand file uploads/storage
1. CODEBASE_ANALYSIS.md (Section 3: MaterialService)
2. QUICK_REFERENCE.md (Multipart Form Handling)
3. CODEBASE_ANALYSIS.md (Section 10: Storage Provider Integration)

---

## 🔍 Key Concepts by Document

### Authorization & Security
- CODEBASE_ANALYSIS.md Section 11: Middleware Flow
- CODEBASE_ANALYSIS.md Section 10: Authorization Dependencies
- QUICK_REFERENCE.md: Authorization Layers

### Data Relationships
- QUICK_REFERENCE.md: Key Entities Relationships
- CODEBASE_ANALYSIS.md Section 5: Domain Models
- CODEBASE_ANALYSIS.md Section 8: Database Tables & Relationships
- backend/schema.md

### Error Handling
- CODEBASE_ANALYSIS.md Section 12: Error Handling
- QUICK_REFERENCE.md: Error Code Summary

### Side Effects
- CODEBASE_ANALYSIS.md Section 10: Potential Side Effects & Dependencies
- QUICK_REFERENCE.md: Common Side Effects table

### Patterns
- CODEBASE_ANALYSIS.md Section 3: Service Patterns
- QUICK_REFERENCE.md: Key Service Patterns

---

## 📋 Checklist: Before Making Code Changes

- [ ] Read QUICK_REFERENCE.md (Architecture Overview)
- [ ] Read PRODUCT_SCOPE.md (Business context)
- [ ] Read CODEBASE_ANALYSIS.md (relevant section)
- [ ] Identify which layer(s) need changes (Handler/Service/Repository/Domain)
- [ ] Check CODEBASE_ANALYSIS.md Section 10 for side effects
- [ ] Check CODEBASE_ANALYSIS.md Section 16 for known issues
- [ ] Plan authorization strategy (middleware vs service level)
- [ ] Plan notification triggers (if any)
- [ ] Plan attachment management (if applicable)
- [ ] Review similar existing code as template

---

## ⚠️ Critical Information

### Design Decisions (Non-Negotiable)
1. **Class vs SubjectClass**: Materials/assignments live in SubjectClass, NOT Class
2. **Best-Effort Notifications**: Don't cascade failures
3. **Soft Deletes**: Data recoverable via deleted_at
4. **Upsert Semantics**: Submissions and assessments (last-write-wins)
5. **Polymorphic Comments**: SourceType + SourceID identify target

### Known Issues
1. ⚠️ Route ordering: GET /assignments/status/:id swallowed by GET /assignments/:assignmentId
2. ⚠️ No unit tests exist
3. ⚠️ gofmt non-compliance
4. ⚠️ Assessment doesn't verify teacher owns assignment (TODO)
5. ⚠️ Submission doesn't verify student enrolled (TODO)

### Out of Scope (Future)
- Real file storage (S3/Supabase)
- Realtime chat
- Student notes
- Email notifications
- Signed URLs
- Thumbnails
- Nested comments

---

## 📖 Documentation Statistics

| Document | Size | Sections | Purpose |
|----------|------|----------|---------|
| CODEBASE_ANALYSIS.md | 44 KB | 16 | Technical reference |
| QUICK_REFERENCE.md | 8.3 KB | 9 | Quick lookups |
| PRODUCT_SCOPE.md | 17 KB | 11 | Business context |
| schema.md | N/A | N/A | Database schema |
| AGENT.md | N/A | 10 | Engineering context |

---

## ✅ Analysis Complete

**No code was modified during this analysis.**

All information is read-only and suitable for:
- Feature planning
- Bug fixing
- Code review
- Onboarding
- Architecture discussions

**Next Step**: When ready to implement, refer to this index and the specific document sections.

---

Generated: 2026-06-02 02:45:39 UTC
