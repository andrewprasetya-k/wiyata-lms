# Chat API

Base URL: `/api/chat`

REST Chat MVP mendukung satu chat room utama per sekolah aktif.

## Scope MVP

- School-wide chat only.
- One main room per active school.
- Text-only messages.
- REST API only.
- Admin Sekolah, teacher, dan student boleh berpartisipasi jika masih menjadi
  member aktif sekolah tersebut.
- Tidak ada WebSocket/realtime.
- Tidak ada subject/class room, DM, free group chat, attachment, typing
  indicator, online/offline, delete/unsend, moderation UI, atau notification
  integration.

Subject/class chat adalah ekspansi masa depan.

## Access Rules

Semua endpoint memerlukan:

- JWT authentication.
- Active `SchoolId` context.
- Active school membership (`school_users.deleted_at IS NULL`).

Room permission tidak memakai `chat_room_members` sebagai source of truth untuk
MVP. Akses hanya berbasis sekolah aktif:

- `chat_rooms.room_sch_id` harus sama dengan active school.
- `chat_rooms.room_type = "group"`.
- `chat_rooms.room_ref_type = "school"`.
- `chat_rooms.room_ref_id = activeSchoolID`.
- `chat_rooms.deleted_at IS NULL`.
- Super admin tidak ikut chat akademik sekolah kecuali juga memiliki membership
  aktif di sekolah tersebut.

## Endpoints

### List My Rooms

`GET /rooms`

Mengembalikan room sekolah yang sudah dibuka dan dapat diakses oleh user saat
ini. Untuk MVP, hasilnya maksimal satu room.

```json
{
  "rooms": [
    {
      "roomId": "uuid",
      "roomName": "Ruang sekolah",
      "roomType": "group",
      "roomRefType": "school",
      "roomRefId": "school-uuid",
      "schoolId": "school-uuid",
      "schoolName": "SMA EduVerse",
      "lastMessage": {
        "messageId": "uuid",
        "senderId": "uuid",
        "senderName": "Budi",
        "content": "Selamat pagi.",
        "createdAt": "2026-06-26T03:00:00Z"
      },
      "lastMessageAt": "2026-06-26T03:00:00Z",
      "unreadCount": 1,
      "canSend": true
    }
  ]
}
```

### Open School Room

`POST /school/open`

Membuka atau membuat room utama untuk active school.

```json
{
  "room": {
    "roomId": "uuid",
    "roomName": "Ruang sekolah",
    "roomType": "group",
    "roomRefType": "school",
    "roomRefId": "school-uuid",
    "schoolId": "school-uuid",
    "schoolName": "SMA EduVerse",
    "unreadCount": 0,
    "canSend": true
  }
}
```

### List Messages

`GET /rooms/:roomId/messages?limit=50&before=2026-06-26T03:00:00Z`

`limit` dibatasi maksimal 50. Response diurutkan oldest-to-newest untuk
keterbacaan percakapan. `nextBefore` bisa dipakai untuk mengambil pesan yang
lebih lama.

```json
{
  "messages": [
    {
      "messageId": "uuid",
      "roomId": "uuid",
      "senderId": "uuid",
      "senderName": "Budi",
      "senderRole": "student",
      "content": "Selamat pagi.",
      "messageType": "text",
      "createdAt": "2026-06-26T03:00:00Z",
      "isMine": true
    }
  ],
  "nextBefore": null,
  "hasMore": false
}
```

### Create Message

`POST /rooms/:roomId/messages`

```json
{
  "content": "Halo semua."
}
```

Rules:

- Content di-trim.
- Empty content ditolak.
- Maksimal 5.000 karakter.
- `messageType` selalu `text`.

Response adalah canonical `MessageDTO` dan dapat dipakai ulang nanti sebagai
payload WebSocket `new_message`.

### Mark Room Read

`PATCH /rooms/:roomId/read`

```json
{
  "lastReadMessageId": "uuid"
}
```

`lastReadMessageId` opsional. Endpoint ini idempotent dan hanya berlaku jika
current user memiliki akses ke room.
