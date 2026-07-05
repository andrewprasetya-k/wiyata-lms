# Wiyata Backend Backlog

This file tracks current backend follow-up work. It intentionally excludes features that are already implemented in the codebase.

## High Priority

- **Assignment Extensions**
  - Student request extension.
  - Teacher approve/reject.
  - Extended deadline handling.
  - Update submission domain/DTO/service/handler/repository contracts when implemented.

- **Protected File Delivery**
  - Generate signed/private URLs for media download.
  - Keep object path validation and storage cleanup behavior intact.
  - Add thumbnail generation as a separate follow-up.

- **Backend-Authoritative Enrollment Role Derivation**
  - Current frontend infers class placement role from school member roles.
  - Backend still receives `role` in enrollment payload.
  - Future backend work should derive/validate placement role authoritatively and keep `enr_role` compatible with downstream queries.

## Medium Priority

- **Grade Report / Transcript Export**
  - Export individual student transcript.
  - Export class grade report.
  - Generate report cards per term/year if product requires it.

- **Notification Preferences**
  - User-level settings for notification types.
  - Optional email vs in-app preferences.
  - Optional realtime notification delivery. Chat WebSocket already exists; general notification WebSocket/SSE is not implemented.

- **Rich Text Support**
  - Accept and sanitize HTML content for materials, assignments, and feed posts.
  - Add frontend editor only after backend sanitization rules are clear.

- **Nested Comments**
  - Add parent comment modeling if product chooses threaded discussion.
  - Current discussion/comment support is flat for feed, material, and assignment.

## Low Priority / Product Decisions

- Attendance/timetable.
- Quiz/exam module.
- Parent portal.
- Advanced search across materials and assignments.
- Engagement analytics beyond existing dashboard/grade features.

## Test Coverage Follow-Up

- Active school/role context middleware and `/api/me/context`.
- Invitation accept flows for admin, teacher, and student.
- Enrollment role mismatch and soft-unenroll behavior.
- Notification recipients for discussion comments.
- School registration approval email best-effort behavior.
