# Student Material Notes API

Base URL: `/api/notes`

Student Notes MVP is material-only and private to the current student. Notes use plain text, one row per student per material, manual save, and hard delete.

All endpoints require:

- JWT authentication.
- Active `SchoolId` context.
- Active school membership.
- `student` role.
- Active student enrollment (`enrollments.left_at IS NULL`) in the material's subject class.

Teacher and admin roles cannot read or mutate student notes.

## 1. Get Material Note

- **Method:** `GET`
- **URL:** `/material/:materialId`

The material must exist, belong to the active school, and be accessible to the current student.

**Response when a note exists:**

```json
{
  "note": {
    "noteId": "uuid",
    "materialId": "uuid",
    "content": "Ringkasan pribadi student.",
    "createdAt": "2026-06-24T10:00:00Z",
    "updatedAt": "2026-06-24T10:10:00Z"
  }
}
```

**Response when no note exists:**

```json
{
  "note": null
}
```

A missing note is not a `404`. A missing or inaccessible material remains an error.

## 2. Save Material Note

- **Method:** `PUT`
- **URL:** `/material/:materialId`

**Body:**

```json
{
  "content": "Catatan pribadi student."
}
```

Rules:

- Content is trimmed.
- Empty content is rejected.
- Maximum content length is 10,000 characters.
- `schoolId`, `userId`, and `materialId` are not accepted from the body.
- The row is created or updated using the existing unique key `(snt_usr_id, snt_mat_id)`.

The response uses the same `{ "note": ... }` shape as GET.

## 3. Delete Material Note

- **Method:** `DELETE`
- **URL:** `/material/:materialId`

The operation hard-deletes only the current JWT user's note for the material in the active school. Deleting when no note exists is idempotent and still succeeds.

```json
{
  "message": "Note deleted"
}
```

## 4. List Notes by Subject Class

- **Method:** `GET`
- **URL:** `/subject-class/:subjectClassId`

Returns the current student's private material notes for one accessible subject class. The student must have an active enrollment (`left_at IS NULL`) in the subject class's class.

The collection is loaded with one joined query over `student_notes` and active materials, scoped by the JWT user and active school. Notes for soft-deleted materials are excluded and results are sorted by `updatedAt` descending.

```json
{
  "notes": [
    {
      "noteId": "uuid",
      "materialId": "uuid",
      "materialTitle": "Pengenalan Ekosistem",
      "content": "Ringkasan pribadi student.",
      "createdAt": "2026-06-24T10:00:00Z",
      "updatedAt": "2026-06-24T10:10:00Z"
    }
  ]
}
```

When no notes exist, the endpoint returns `{ "notes": [] }`.

## Out of Scope

- Assignment notes.
- Subject notes.
- Global notes workspace.
- Teacher/admin visibility.
- Sharing or collaboration.
- Rich text, markdown rendering, attachments, search, tags, folders, and autosave.
