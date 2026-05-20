# 📖 Material (Konten Belajar) Module API Documentation

Base URL: `/api/materials`

## 1. Create Material
Create a new learning material for a class with optional attachments.

- **URL:** `(base URL)`
- **Method:** `POST`

### Option A: JSON (with existing media IDs or inline media data)
- **Content-Type:** `application/json`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
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
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Form Fields:**
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `schoolId` | string | Yes | Must be a valid UUID |
| `subjectClassId` | string | Yes | UUID |
| `materialTitle` | string | Yes | |
| `materialDesc` | string | No | |
| `materialType` | string | Yes | `video`, `pdf`, `ppt`, `other` |
| `files` | file[] | No | Multiple files, max 10MB each |

**Object path in storage:** `schools/{schoolId}/{uuid}{ext}` (consistent with media upload)

**Response `501`** (storage not configured):
```json
{ "error": "File upload to storage is not configured" }
```

---

## 2. List Materials
- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Params:** `page`, `limit`, `search`, `subjectClassId`.
- **Note:** If `subjectClassId` is provided, response will be wrapped in `MaterialListWithSubjectDTO`.

---

## 3. Get Material Detail (with Attachments)
- **URL:** `/:id`
- **Method:** `GET`

---

## 4. Update Material
Update material details and its attachments.

- **URL:** `/:id`
- **Method:** `PATCH`
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
