# 💬 Comment Module API Documentation

Base URL: `/api/comments`

Comments support class feed posts, materials, and assignments for the school MVP. The comment table can technically store other source types, but `submission` and nested `comment` discussions are still post-MVP and are rejected by the API for now.

All comment endpoints require authenticated active school membership. The active school is read from the `SchoolId` context/header; `schoolId` in the request body is not trusted and must match the active school if provided.

## Authorization

The source must exist, belong to the active school, and be accessible to the current actor.

- **Feed:** class-level comments. Admin can access active-school feed comments. Teacher access requires teaching the feed class. Student access requires active class enrollment.
- **Material:** subject-class discussion. Admin can access active-school material comments. Teacher access requires teaching the material subject class. Student access requires active enrollment in that subject class.
- **Assignment:** class-wide assignment discussion. Admin can access active-school assignment comments. Teacher access requires teaching the assignment subject class. Student access requires active enrollment in that subject class.
- **Update:** only the comment owner can edit their own comment.
- **Delete:** comment owner can delete their own comment; admin can delete active-school comments.

Soft-deleted comments are excluded from list/detail responses by the existing soft-delete behavior.

## 1. Create Comment

- **URL:** `(base URL)`
- **Method:** `POST`
- **Body:**

```json
{
  "sourceType": "feed",
  "sourceId": "source-uuid",
  "content": "Comment text"
}
```

Supported `sourceType` values are `feed`, `material`, and `assignment`.
`schoolId` may be sent by older clients, but it must match active `SchoolId`.

## 2. Get Comments by Source

- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Params:** `?type=feed|material|assignment&id=source-uuid`
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
- **Authorization:** Current actor must still have access to the comment source.

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

- Comments on `submission`.
- Nested replies.
- Attachments.
- Reactions.
- Realtime/WebSocket updates.
