# 💬 Comment Module API Documentation

Base URL: `/api/comments`

## 1. Create Comment
- **URL:** `(base URL)`
- **Method:** `POST`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Body:**
```json
{
  "schoolId": "uuid",
  "sourceType": "feed|material|assignment|submission",
  "sourceId": "uuid",
  "content": "Comment text"
}
```

## 2. Get Comments by Source
- **URL:** `(base URL)`
- **Method:** `GET`
- **Query Params:** `?type=feed&id=uuid`
- **Response:** Array of comments ordered by creation time

## 3. Get Comment by ID
- **URL:** `/:id`
- **Method:** `GET`
- **Response:** Single comment with user info

## 4. Update Comment
- **URL:** `/:id`
- **Method:** `PATCH`
- **Body:**
```json
{
  "content": "Updated comment text"
}
```

## 5. Delete Comment
- **URL:** `/:id`
- **Method:** `DELETE`
- **Note:** Soft delete

---

## Features

- **Multi-Source:** Comments can be attached to feeds, materials, assignments, or submissions
- **User Context:** Creator name included in responses
- **Chronological Order:** Comments sorted by creation time (oldest first)
