# 📖 Material (Konten Belajar) Module API Documentation

Base URL: `/api/materials`

## 1. Create Material
Create a new learning material for a class with optional attachments.

- **URL:** `(base URL)`
- **Method:** `POST`
- **Auth:** Required
- **Role:** `teacher`
- **School Context:** Requires `SchoolId` header
- **Authorization:** The current teacher must teach the requested `subjectClassId`. Returns `403` if not.
- **Attachment Rule:** Existing `mediaIds` must exist, belong to the active school, and be owned/uploaded by the current teacher.

### Option A: JSON (with existing media IDs or inline media data)
- **Content-Type:** `application/json`
- **Auth Note:** Teacher identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Body:**
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `schoolId` | uuid | Yes | |
| `subjectClassId` | uuid | Yes | Link to subject, class, and teacher |
| `materialTitle`| string | Yes | |
| `materialDesc` | string | No | |
| `materialType`| string | Yes | `video`, `pdf`, `ppt`, `other` |
| `mediaIds` | uuid[] | No | List of already recorded Media IDs |
| `medias` | object[] | No | Inline media data (auto-create in medias table) |

`schoolId` must match the active `SchoolId` header.

**Inline Media Object:**
```json
{
  "name": "filename.pdf",
  "fileSize": 1024000,
  "mimeType": "application/pdf",
  "fileUrl": "https://supabase.co/storage/.../file.pdf",
  "thumbnailUrl": "https://..." // optional
}
```

### Option B: Multipart Form (with file uploads)
Files are uploaded to the configured storage provider (same as `POST /api/medias/upload`). Requires `STORAGE_PROVIDER=supabase` to be set. Returns `501` if storage is not configured.

Each file is uploaded to storage first. If upload succeeds but DB record fails, the storage object is deleted (best-effort cleanup). Max file size per file: **10MB**.

- **Content-Type:** `multipart/form-data`
- **Auth Note:** Teacher identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Form Fields:**
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `schoolId` | string | Yes | Must be a valid UUID |
| `subjectClassId` | string | Yes | UUID |
| `materialTitle` | string | Yes | |
| `materialDesc` | string | No | |
| `materialType` | string | Yes | `video`, `pdf`, `ppt`, `other` |
| `files` | file[] | No | Multiple files, max 10MB each |

`schoolId` must match the active `SchoolId` header.

**Object path in storage:** `schools/{schoolId}/{uuid}{ext}` (consistent with media upload)

**Response `501`** (storage not configured):
```json
{ "error": "File upload to storage is not configured" }
```

---

## 2. List Materials
- **URL:** `(base URL)`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `admin`, `teacher`, or `student`
- **School Context:** Requires `SchoolId` header
- **Query Params:** `page`, `limit`, `search`, `subjectClassId`.
- **Authorization:** `subjectClassId` is required. Admin can read active-school subject classes. Teacher can read only subject classes they teach. Student can read only subject classes in classes where they are enrolled.
- **Response:** Wrapped in `MaterialListWithSubjectDTO`.

---

## 3. Get Material Detail (with Attachments)
- **URL:** `/:id`
- **Method:** `GET`
- **Auth:** Required
- **Role:** `admin`, `teacher`, or `student`
- **School Context:** Requires `SchoolId` header
- **Authorization:** Same subject_class access rule as list materials.
- **Attachment Metadata:** Each valid attachment includes `mediaId`, `mediaName`, `fileSize`, `mimeType`, `fileUrl`, optional `thumbnailUrl`, `ownerType`, and `createdAt`.
- **Attachment Safety:** Media that has been soft-deleted or does not belong to the same active school is omitted. Non-HTTP(S) file and thumbnail URLs are returned as empty strings.

---

## 4. Update Material
Update material details and its attachments.

- **URL:** `/:id`
- **Method:** `PATCH`
- **Auth:** Required
- **Role:** `teacher` or `admin`
- **School Context:** Requires `SchoolId` header
- **Authorization:** Teacher must teach the material's subject class. Admin can update only active-school materials.
- **Attachment Rule:** If `mediaIds` is provided, every media must exist and belong to the active school. Teachers can attach only their own uploaded media; admins can attach active-school media.
- **Body:**
| Field | Type | Note |
| :--- | :--- | :--- |
| `materialTitle`| string | Optional |
| `materialDesc` | string | Optional |
| `materialType`| string | `video`, `pdf`, `ppt`, `other` (Optional) |
| `mediaIds` | uuid[] | New list of Media IDs (Will replace existing) |

---

## 5. Delete Material
- **URL:** `/:id`
- **Method:** `DELETE`
- **Auth:** Required
- **Role:** `teacher` or `admin`
- **School Context:** Requires `SchoolId` header
- **Authorization:** Teacher must teach the material's subject class. Admin can delete only active-school materials.

---

## 6. Update Progress
Mark a material as completed for a user.

- **URL:** `/progress`
- **Method:** `POST`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Body:**
```json
{
  "materialId": "uuid",
  "status": "completed"
}
```
