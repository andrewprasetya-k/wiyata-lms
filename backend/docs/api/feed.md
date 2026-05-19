# 📰 Feed Module API Documentation

Base URL: `/api/feeds`

## 1. Create Feed
- **URL:** `/`
- **Method:** `POST`
- **Auth Note:** Actor identity is taken from the JWT token. Sending identity fields in the body is ignored or no longer required.
- **Body:**
```json
{
  "schoolId": "uuid",
  "classId": "uuid",
  "content": "Announcement text",
  "mediaIds": ["uuid"]
}
```

## 2. Get Feeds by Class
- **URL:** `/class/:classId`
- **Method:** `GET`
- **Query Params:** `?page=1&limit=10`
- **Response:** `ClassWithFeedsDTO` (with class header and paginated feeds)

## 3. Get Feed by ID
- **URL:** `/:id`
- **Method:** `GET`
- **Response:** Includes attachments and comment count

## 4. Update Feed
- **URL:** `/:id`
- **Method:** `PATCH`
- **Body:** (all fields optional)
```json
{
  "content": "Updated announcement",
  "mediaIds": ["uuid"]
}
```

## 5. Delete Feed
- **URL:** `/:id`
- **Method:** `DELETE`
- **Note:** Soft delete

---

## Features

- **Pagination:** Supports page and limit query params
- **Attachments:** Multiple media files per feed
- **Comment Count:** Automatically included in response
- **Class Context:** Feed list includes class header
