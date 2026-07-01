# 💬 Comment Module API Documentation

Base URL: `/api/comments`

Comments are feed-only for the school MVP. The comment table can technically store other source types, but `material`, `assignment`, and `submission` comments are post-MVP and are rejected by the API for now.

All comment endpoints require authenticated active school membership. The active school is read from the `SchoolId` context/header; `schoolId` in the request body is not trusted and must match the active school if provided.

## Authorization

For `sourceType = feed`, the source feed must exist, belong to the active school, and be accessible to the current actor:

- **Admin:** can list/create/delete comments on active-school feed posts.
- **Teacher:** can list/create comments only on feed posts for classes they actively teach. Active teacher enrollment requires `enrollments.left_at IS NULL`.
- **Student:** can list/create comments only on feed posts for classes where they are actively enrolled. Active student enrollment requires `enrollments.left_at IS NULL`.
- **Update:** only the comment owner can edit their own comment.
- **Delete:** comment owner can delete their own comment; admin can delete active-school comments.

Soft-deleted comments are excluded from list/detail responses by the existing soft-delete behavior.

## 1. Create Feed Comment

- **URL:** `(base URL)`
- **Method:** `POST`
- **Body:**

```json
{
  "sourceType": "feed",
  "sourceId": "feed-uuid",
  "content": "Comment text"
}
```

`schoolId` may be sent by older clients, but it must match active `SchoolId`.

## 2. Get Comments by Feed

- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Params:** `?type=feed&id=feed-uuid`
- **Response:** Array of comments ordered by `created_at ASC`.

```json
[
  {
    "commentId": "uuid",
    "sourceType": "feed",
    "sourceId": "feed-uuid",
    "content": "Comment text",
    "creatorName": "Student Name",
    "createdAt": "2026-06-23T10:00:00Z",
    "isMine": true
  }
]
```

## 3. Get Comment by ID

- **URL:** `/:id`
- **Method:** `GET`
- **Authorization:** Current actor must still have access to the source feed.

## 4. Update Own Comment

- **URL:** `/:id`
- **Method:** `PATCH`
- **Authorization:** Only the comment owner can update their own comment.
- **Body:**

```json
{
  "content": "Updated comment text"
}
```

## 5. Delete Comment

- **URL:** `/:id`
- **Method:** `DELETE`
- **Authorization:** Comment owner can delete their own comment; admin can delete active-school comments.
- **Note:** Soft delete.

## Post-MVP

- Comments on `material`, `assignment`, and `submission`.
- Nested replies.
- Attachments.
- Reactions.
- Realtime/WebSocket updates.
- Comment notification deep links.
