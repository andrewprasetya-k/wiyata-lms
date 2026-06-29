# 📰 Feed Module API Documentation

Base URL: `/api/feeds`

## 0. Feed Unread Badge

- **GET** `/unread-count`
- **Auth:** Required. Active-school scoped. Available for `student`, `teacher`, and `admin`.
- **Behavior:** Counts unread feed-related notifications for the current user and active school using existing notification types:
  - `feed_posted`
  - `comment_added`
- **Scope:** Uses the existing per-user notification rows as read state. The notification `relatedId` points to the feed row, and the endpoint joins that feed to the active school. It does not create a separate feed read receipt table.
- **Response:**

```json
{
  "unreadCount": 3
}
```

- **PATCH** `/read`
- **Auth:** Required. Active-school scoped. Available for `student`, `teacher`, and `admin`.
- **Behavior:** Marks unread feed-related notifications for the current user and active school as read. It does not mark assignment/material/general notifications.
- **Response:**

```json
{
  "message": "Feed notifications marked as read"
}
```

## 1. Create Feed
- **URL:** `(base URL)`
- **Method:** `POST`
- **Auth:** Required. Active school member with `teacher` or `admin` role.
- **Ownership:** Active `SchoolId` header is the source of truth. `schoolId` in the body must match the active school.
- **Teacher rule:** Teacher can create a feed only for a class where the teacher teaches at least one active subject_class.
- **Admin rule:** Admin can create a feed for active-school classes.
- **Body:**
```json
{
  "schoolId": "uuid",
  "classId": "uuid",
  "content": "Announcement text"
}
```

`mediaIds` / feed attachments are not supported in the current MVP.

## 2. Get Feeds by Class
- **URL:** `/class/:classId`
- **Method:** `GET`
- **Auth:** Required. Active school member with `admin`, `teacher`, or `student` role.
- **Access:**
  - Admin can read feeds for active-school classes.
  - Teacher can read feeds for classes they actively teach.
  - Student can read feeds only for classes where they have active enrollment (`left_at IS NULL`).
- **Query Params:** `?page=1&limit=10`
- **Response:** `ClassWithFeedsDTO` (with class header and paginated feeds)

## 3. Get Feed by ID
- **URL:** `/:id`
- **Method:** `GET`
- **Auth:** Required. Same active-school/class access rules as class feed list.
- **Response:** Includes attachments and comment count

## 4. Update Feed
- **URL:** `/:id`
- **Method:** `PATCH`
- **Auth:** Required. Active school member with `teacher` or `admin` role.
- **Access:**
  - Admin can update active-school feed posts.
  - Teacher can update only their own feed posts in classes they actively teach.
- **Body:** (all fields optional)
```json
{
  "content": "Updated announcement"
}
```

`mediaIds` / feed attachments are not supported in the current MVP.

## 5. Delete Feed
- **URL:** `/:id`
- **Method:** `DELETE`
- **Auth:** Required. Active school member with `teacher` or `admin` role.
- **Access:**
  - Admin can delete active-school feed posts.
  - Teacher can delete only their own feed posts in classes they actively teach.
- **Note:** Soft delete

---

## Features

- **Pagination:** Supports page and limit query params
- **Comment Count:** Automatically included in response
- **Class Context:** Feed list includes class header
- **Notifications:** `feed_posted` notification remains best-effort and does not block feed creation.
- **Feed comments:** Backend comment endpoints are feed-only and active-school scoped for MVP. UI exposure is deferred.
- **Deferred:** Reactions, realtime/WebSocket, nested replies, non-feed comments, and feed attachments are outside the current MVP.
