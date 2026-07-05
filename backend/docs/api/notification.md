# 🔔 Notification API Documentation

Base URL: `/api/notifications`

## Overview

REST notification system untuk inform users tentang activities seperti new assignments, grades, comments, feed posts, material updates, and discussion activity.

Notifications are currently scoped to the authenticated global user (`ntf_usr_id`), not to `school_user` or a specific school. Read/delete operations must use the JWT user context and cannot target another user's notification.

Current frontend behavior uses the Notification Center, dashboard preview, sidebar unread badge, manual refresh, visibility refresh, and optimistic read updates. General notification delivery does **not** currently use WebSocket or SSE. Chat has its own separate WebSocket flow.

---

## 1. Get User Notifications
Retrieve notifications untuk authenticated user dengan pagination.

- **URL:** `(base URL)`
- **Method:** `GET`
- **Auth:** Required
- **Query Parameters:**
  - `page` (default: `1`): Page number
  - `limit` (default: `20`): Items per page (max: 50)
  - `unreadOnly` (default: `false`): Show only unread notifications

**Response (200 OK):**
```json
{
  "data": [
    {
      "notificationId": "uuid",
      "type": "assignment_created",
      "title": "New Assignment Posted",
      "message": "New assignment 'Quiz Chapter 1' has been posted in Matematika",
      "link": "/assignments/uuid",
      "isRead": false,
      "createdAt": "2026-03-12T10:00:00Z"
    },
    {
      "notificationId": "uuid",
      "type": "assignment_graded", 
      "title": "Assignment Graded",
      "message": "Your assignment 'Quiz Chapter 1' has been graded: 85/100",
      "link": "/assignments/uuid",
      "isRead": true,
      "createdAt": "2026-03-11T15:30:00Z"
    }
  ],
  "unreadCount": 3,
  "totalItems": 25,
  "page": 1,
  "limit": 20,
  "totalPages": 2
}
```

---

## 2. Get Unread Count
Get jumlah unread notifications untuk badge display.

- **URL:** `/unread-count`
- **Method:** `GET`
- **Auth:** Required

**Response (200 OK):**
```json
{
  "unreadCount": 5
}
```

---

## 3. Mark Notification as Read
Mark specific notification sebagai sudah dibaca.

- **URL:** `/read/:id`
- **Method:** `PATCH`
- **Auth:** Required
- **Ownership:** Only the authenticated user's own notification can be marked as read.

**Response (200 OK):**
```json
{
  "message": "Notification marked as read"
}
```

---

## 4. Mark All Notifications as Read
Mark semua notifications user sebagai sudah dibaca.

- **URL:** `/read-all`
- **Method:** `PATCH`
- **Auth:** Required

**Response (200 OK):**
```json
{
  "message": "All notifications marked as read"
}
```

---

## 5. Delete Notification
Delete specific notification.

- **URL:** `/:id`
- **Method:** `DELETE`
- **Auth:** Required
- **Ownership:** Only the authenticated user's own notification can be deleted.

**Response (200 OK):**
```json
{
  "message": "Notification deleted"
}
```

---

## Notification Types

| Type | Description | Recipients | Self-notif |
|------|-------------|------------|------------|
| `assignment_created` | New assignment posted | Students enrolled in the class | N/A |
| `assignment_graded` | Submission has been graded | Student who submitted | N/A |
| `material_added` | New learning material posted | Students enrolled in the class | N/A |
| `feed_posted` | New announcement posted | All class members | Excluded |
| `comment_added` | New comment on feed/material/assignment discussion | Content creator or active class participants depending on source and commenter role | Excluded |

---

## Auto Triggers

Notifications are created **automatically** by the backend when the following events occur. No manual API call is needed.

| Event | Trigger point | Payload `relatedId` |
|---|---|---|
| Teacher creates assignment | `POST /assignments` | `assignmentId` |
| Teacher grades a submission | `POST /assignments/assess/:submissionId` | `submissionId` |
| Teacher creates material | `POST /materials` | `materialId` |
| Teacher/admin posts feed | `POST /feeds` | `feedId` |
| Anyone posts a comment | `POST /comments` | source content ID |

**Behavior:**
- All triggers are **best-effort** — if notification creation fails, the primary action (create assignment, grade, etc.) still succeeds.
- `feed_posted`: creator is excluded from recipients.
- `comment_added`: skips the commenter and deduplicates recipients. Feed behavior remains source-owner/class-aware according to the feed comment flow. Material/assignment discussion notifies active students when teacher/admin comments, and notifies the content creator/teacher when a student comments.
- `unread-count` increments automatically for each notification created.

**Supported comment source types for `comment_added`:**

| `sourceType` | Owner field |
|---|---|
| `feed` | feed creator |
| `material` | material creator |
| `assignment` | assignment creator |

---

## Frontend Refresh Behavior

The frontend currently uses shared notification services/composables:

- `frontend/src/services/notifications.ts`
- `frontend/src/composables/useNotificationUnreadCount.ts`
- Notification Center pages for student and teacher.
- Dashboard preview widgets and sidebar badge.

Unread state is refreshed by explicit calls and visibility/context refresh behavior. There is no `/ws/notifications` endpoint in the current backend.

Chat realtime is separate and uses the chat WebSocket implementation; do not treat that as a general notification WebSocket.

---

## Error Responses

### 404 Not Found
```json
{
  "error": "Notification not found"
}
```

### 403 Forbidden
```json
{
  "error": "Cannot access notification of another user"
}
```

---

## Usage Examples

### Get Recent Notifications
```bash
GET /api/notifications?page=1&limit=10
Authorization: Bearer <token>
```

### Get Only Unread
```bash
GET /api/notifications?unreadOnly=true
Authorization: Bearer <token>
```

### Mark as Read
```bash
PATCH /api/notifications/read/uuid-123
Authorization: Bearer <token>
```

---

## Frontend Integration

### Notification Badge
```javascript
// Show unread count in navigation
async function updateNotificationBadge() {
  const response = await fetch('/api/notifications/unread-count');
  const { unreadCount } = await response.json();
  
  const badge = document.getElementById('notification-badge');
  badge.textContent = unreadCount;
  badge.style.display = unreadCount > 0 ? 'block' : 'none';
}
```

### Notification List
```javascript
// Display notifications in dropdown/page
async function loadNotifications(page = 1) {
  const response = await fetch(`/api/notifications?page=${page}&limit=10`);
  const data = await response.json();
  
  renderNotifications(data.data);
  renderPagination(data.page, data.totalPages);
}
```

---

**Last Updated:** 2026-07-05
