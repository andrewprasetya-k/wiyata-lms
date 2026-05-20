# 🔔 Notification API Documentation

Base URL: `/api/notifications`

## Overview

Real-time notification system untuk inform users tentang activities seperti new assignments, grades, comments, dll.

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
      "createdAt": "12-03-2026 10:00:00"
    },
    {
      "notificationId": "uuid",
      "type": "assignment_graded", 
      "title": "Assignment Graded",
      "message": "Your assignment 'Quiz Chapter 1' has been graded: 85/100",
      "link": "/assignments/uuid",
      "isRead": true,
      "createdAt": "11-03-2026 15:30:00"
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
| `comment_added` | New comment on content | Owner of the commented content | Excluded |

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
- `comment_added`: if the commenter is the content owner, no notification is sent.
- `unread-count` increments automatically for each notification created.

**Supported comment source types for `comment_added`:**

| `sourceType` | Owner field |
|---|---|
| `feed` | feed creator |
| `material` | material creator |
| `assignment` | assignment creator |
| `submission` | student who submitted |

---

## Real-time Integration

### Frontend Polling (Current)
```javascript
// Poll every 30 seconds for new notifications
setInterval(async () => {
  const response = await fetch('/api/notifications/unread-count');
  const { unreadCount } = await response.json();
  updateBadge(unreadCount);
}, 30000);
```

### WebSocket (Future Enhancement)
```javascript
// Real-time notifications via WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/notifications');
ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  showNotification(notification);
  updateUnreadCount();
};
```

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

**Last Updated:** 2026-05-18
