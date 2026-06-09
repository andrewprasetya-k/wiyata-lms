# 📁 Media & Metadata Module API Documentation

Base URL: `/api/medias`

## Storage Configuration

File upload requires a configured storage provider via environment variables.

| Env Variable | Required | Description |
| :--- | :--- | :--- |
| `STORAGE_PROVIDER` | Yes | `supabase` to enable uploads, `disabled` or empty to disable |
| `SUPABASE_URL` | If supabase | Your Supabase project URL |
| `SUPABASE_SERVICE_KEY` | If supabase | Supabase service role key (not anon key) |
| `SUPABASE_BUCKET` | If supabase | Target storage bucket name |

If `STORAGE_PROVIDER` is `disabled` or not set, upload endpoints return `501 Not Implemented`.

---

## 1. Upload File
Upload a file directly to storage (multipart form). The backend uploads to the configured storage provider and records metadata in the database atomically. If the DB record fails after a successful upload, the storage object is deleted (best-effort cleanup).

- **URL:** `/upload`
- **Method:** `POST`
- **Auth:** Required
- **School Context:** Requires `SchoolId` header
- **Role:** `admin`, `teacher`, or `student`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Content-Type:** `multipart/form-data`
- **Max file size:** 10MB

**Form Fields:**
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `file` | file | Yes | The file to upload |
| `schoolId` | string | Yes | Must be a valid UUID |
| `ownerType` | string | No | `user`, `school`, `material`, `assignment`, etc. |

`schoolId` must match the active `SchoolId` header. The backend stores `ownerId` from the JWT user.

**Object path in storage:** `schools/{schoolId}/{uuid}{ext}`

**Response `201`:**
```json
{
  "message": "File uploaded successfully",
  "mediaId": "uuid",
  "fileName": "example.pdf",
  "fileSize": 1024000,
  "mimeType": "application/pdf",
  "storagePath": "schools/uuid/uuid.pdf",
  "fileUrl": "https://your-supabase-url/storage/v1/object/public/bucket/schools/uuid/uuid.pdf",
  "ext": ".pdf"
}
```

**Response `501`** (storage not configured):
```json
{ "error": "File upload to storage is not configured" }
```

---

## 2. Record Media Metadata
Record metadata of a file already uploaded directly to external storage (e.g., via Supabase client SDK). No file transfer occurs — only a DB record is created.

- **URL:** `/metadata`
- **Method:** `POST`
- **Auth:** Required
- **School Context:** Requires `SchoolId` header
- **Role:** `admin`, `teacher`, or `student`
- **Content-Type:** `application/json`

**Body:**
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `schoolId` | uuid | Yes | |
| `mediaName` | string | Yes | |
| `fileSize` | int64 | Yes | In bytes |
| `mimeType` | string | Yes | e.g., `application/pdf` |
| `storagePath` | string | Yes | Path within the storage bucket |
| `fileUrl` | string | Yes | Public URL of the file |
| `thumbnailUrl` | string | No | |
| `isPublic` | boolean | No | Default: `true` |
| `ownerType` | string | Yes | `user`, `school`, `material`, `assignment`, etc. |
| `ownerId` | uuid | Yes | |

`schoolId` must match the active `SchoolId` header. Non-admin users can only record metadata where `ownerId` is their own JWT user ID.

---

## 3. Get Media Detail
- **URL:** `/:id`
- **Method:** `GET`
- **Auth:** Required

---

## 4. Delete Media
Deletes the storage object first, then soft-deletes the metadata record. If the storage object does not exist, deletion proceeds and the DB record is still removed.

- **URL:** `/:id`
- **Method:** `DELETE`
- **Auth:** Required
- **School Context:** Requires `SchoolId` header
- **Role:** `admin`, `teacher`, or `student`
- **Authorization:** Media must belong to the active school. Admin can delete active-school media. Non-admin users can delete only media where `ownerId` is their JWT user ID.
