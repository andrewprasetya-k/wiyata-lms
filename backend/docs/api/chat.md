# Chat API

Base URL: `/api/chat`

WebSocket URL: `/api/ws/chat`

REST chat mendukung satu room sekolah aktif, custom group room, dan direct
message untuk warga aktif di sekolah yang sama.

## Scope MVP

- School-wide chat selalu tersedia sebagai room utama sekolah.
- Custom group room dapat dibuat oleh warga aktif sekolah.
- Direct message dapat dibuka antar dua warga aktif di sekolah yang sama.
- Text messages dan attachment file/gambar melalui upload-first media flow.
- REST API remains the source of truth.
- WebSocket tersedia hanya sebagai realtime event transport.
- Chat room summary dan badge dipicu oleh event websocket serta refresh saat visibility/context change; tidak ada polling berkala untuk endpoint room summary di frontend saat ini.
- Admin Sekolah, teacher, dan student boleh berpartisipasi jika masih menjadi
  member aktif sekolah tersebut.
- Tidak ada subject/class room, typing indicator,
  online/offline, delete/unsend, moderation UI, atau notification integration.

Subject/class chat adalah ekspansi masa depan.

## Access Rules

Semua endpoint memerlukan:

- JWT authentication.
- Active `SchoolId` context.
- Active school membership (`school_users.deleted_at IS NULL`).

School room permission berbasis sekolah aktif:

- `chat_rooms.room_sch_id` harus sama dengan active school.
- `chat_rooms.room_type = "group"`.
- `chat_rooms.room_ref_type = "school"`.
- `chat_rooms.room_ref_id = activeSchoolID`.
- `chat_rooms.deleted_at IS NULL`.
- Super admin tidak ikut chat akademik sekolah kecuali juga memiliki membership
  aktif di sekolah tersebut.

Custom group room permission:

- `chat_rooms.room_sch_id` harus sama dengan active school.
- `chat_rooms.room_type = "group"`.
- `chat_rooms.room_ref_type IS NULL`.
- `chat_rooms.room_ref_id IS NULL`.
- User harus active school member.
- User juga harus active `chat_room_members` dengan `left_at IS NULL`.

Direct message permission:

- `chat_rooms.room_sch_id` harus sama dengan active school.
- `chat_rooms.room_type = "dm"`.
- `chat_rooms.room_ref_type IS NULL`.
- `chat_rooms.room_ref_id IS NULL`.
- User harus active school member.
- User juga harus active `chat_room_members` dengan `left_at IS NULL`.

## Endpoints

### List My Rooms

`GET /rooms?search=`

Mengembalikan room sekolah dan custom group room yang dapat diakses oleh user
saat ini, termasuk direct message yang user ikuti. Query `search` opsional,
case-insensitive, dan mencari berdasarkan nama room, nama sekolah, atau nama
/ email target DM. Ordering tetap berdasarkan aktivitas terakhir.

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
      "schoolName": "SMA Wiyata",
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

### List Chat Members

`GET /members?search=nama&excludeRoomId=uuid`

Mengembalikan warga aktif di active school untuk member picker group chat.
Tidak mengekspos membership dari sekolah lain. `excludeRoomId` opsional dipakai
saat menambah anggota ke grup, sehingga user yang masih aktif di grup tersebut
tidak ikut muncul.

```json
{
  "members": [
    {
      "userId": "uuid",
      "fullName": "Budi Santoso",
      "email": "budi@siswa.sch.id",
      "roles": ["student"]
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
    "schoolName": "SMA Wiyata",
    "unreadCount": 0,
    "canSend": true
  }
}
```

### Open Direct Message

`POST /dm/open`

```json
{
  "targetUserId": "uuid"
}
```

Rules:

- Current user harus active school member.
- Target user harus active school member di sekolah aktif yang sama.
- Tidak boleh DM diri sendiri.
- Jika room DM dengan dua anggota aktif yang sama sudah ada, endpoint akan
  mengembalikan room yang sama.
- Jika belum ada, endpoint akan membuat room baru bertipe `dm`, lalu menambahkan
  kedua user ke `chat_room_members`.

Response:

```json
{
  "room": {
    "roomId": "uuid",
    "roomName": "Budi Santoso",
    "roomType": "dm",
    "roomRefType": null,
    "roomRefId": null,
    "schoolId": "school-uuid",
    "schoolName": "Sekolah Wiyata",
    "dmTargetUserId": "uuid",
    "dmTargetName": "Budi Santoso",
    "dmTargetEmail": "budi@example.com",
    "unreadCount": 0,
    "canSend": true
  }
}
```

### Create Group Room

`POST /groups`

```json
{
  "roomName": "Grup Belajar Fisika",
  "memberUserIds": ["user-uuid"]
}
```

Rules:

- Current user harus active school member.
- Semua `memberUserIds` harus active school member pada active school.
- Duplicate member ditolak.
- Creator selalu dimasukkan sebagai room member dengan `crm_role = "admin"`.
- Member terpilih dimasukkan dengan `crm_role = "member"`.
- `room_ref_type` dan `room_ref_id` disimpan `NULL`.

Response:

```json
{
  "room": {
    "roomId": "uuid",
    "roomName": "Grup Belajar Fisika",
    "roomType": "group",
    "roomRefType": null,
    "roomRefId": null,
    "schoolId": "school-uuid",
    "schoolName": "SMA Wiyata",
    "unreadCount": 0,
    "canSend": true
  }
}
```

### Get Group Info

`GET /groups/:roomId`

Mengembalikan info room, pembuat, admin, anggota aktif, waktu dibuat, dan jumlah
anggota. Hanya active member grup yang boleh membaca info grup.

```json
{
  "group": {
    "roomId": "uuid",
    "roomName": "Grup Belajar Fisika",
    "roomType": "group",
    "schoolId": "school-uuid",
    "schoolName": "SMA Wiyata",
    "creator": {
      "userId": "uuid",
      "fullName": "Budi",
      "email": "budi@siswa.sch.id",
      "roles": ["student"]
    },
    "admins": [],
    "members": [],
    "createdAt": "2026-06-26T03:00:00Z",
    "memberCount": 3
  }
}
```

### Rename Group Room

`PATCH /groups/:roomId`

```json
{
  "roomName": "Grup Persiapan UTS"
}
```

Rules:

- School room tidak bisa diubah namanya.
- Hanya admin/creator grup aktif yang boleh rename.
- Nama di-trim.
- Minimal 3 karakter.
- Maksimal 150 karakter.

### Leave Group Room

`POST /groups/:roomId/leave`

Menandai membership current user dengan `left_at`. Setelah keluar, user tidak
melihat room di daftar dan tidak bisa membaca/mengirim pesan ke grup tersebut.

Ownership transfer:

- Jika creator keluar dan masih ada admin aktif lain, creator dipindahkan ke
  admin aktif paling lama.
- Jika tidak ada admin lain, anggota aktif paling lama dipromosikan menjadi
  admin dan menjadi creator baru.
- Jika tidak ada anggota tersisa, room di-soft-delete.

### Add Group Members

`POST /groups/:roomId/members`

```json
{
  "memberUserIds": ["user-uuid"]
}
```

Rules:

- Hanya admin grup.
- Semua user harus active school member di sekolah aktif.
- Active duplicate ditolak.
- Membership yang sebelumnya `left_at` akan direstore sebagai `member`.

### Remove Group Member

`DELETE /groups/:roomId/members/:userId`

Rules:

- Hanya admin grup.
- Admin tidak bisa mengeluarkan diri sendiri; gunakan endpoint leave.
- Jika target adalah admin terakhir, anggota aktif paling lama dipromosikan
  sebelum target dikeluarkan.
- Removed member langsung kehilangan akses karena `left_at` diisi.

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
      "attachments": [],
      "createdAt": "2026-06-26T03:00:00Z",
      "isMine": true
    }
  ],
  "nextBefore": null,
  "hasMore": false
}
```

### Get Read Summary

`GET /rooms/:roomId/read-summary`

Mengembalikan ringkasan read receipt untuk room yang dapat diakses current
user. Endpoint ini dipakai UI untuk menampilkan indikator `Terkirim`,
`Dibaca`, dan `Dibaca X orang`.

Rules:

- Memerlukan akses ke room melalui permission chat yang sama dengan message
  list/send.
- School room hanya mengembalikan active school members.
- Group room dan direct message hanya mengembalikan active `chat_room_members`.
- Removed school member (`school_users.deleted_at IS NOT NULL`) tidak ikut
  dihitung.
- Tidak mengekspos member dari sekolah lain.

```json
{
  "roomId": "uuid",
  "lastReadMessageId": "uuid",
  "lastReadAt": "2026-06-26T03:05:00Z",
  "members": [
    {
      "userId": "uuid",
      "fullName": "Budi Santoso",
      "email": "budi@siswa.sch.id",
      "lastReadMessageId": "uuid",
      "lastReadAt": "2026-06-26T03:05:00Z"
    }
  ]
}
```

### Create Message

`POST /rooms/:roomId/messages`

```json
{
  "content": "Halo semua.",
  "mediaIds": ["media-uuid"]
}
```

Rules:

- Content di-trim.
- Empty content ditolak jika `mediaIds` juga kosong.
- Maksimal 5.000 karakter.
- Maksimal 5 attachment per pesan.
- Duplicate `mediaIds` ditolak.
- Jika `mediaIds` dikirim, message disimpan sebagai `messageType = "file"`.
- Setiap media harus sudah di-upload melalui `POST /api/medias/upload`, masih
  aktif, berada di active school yang sama, dan dimiliki oleh current user untuk
  non-admin/non-owner flow MVP.
- Frontend memakai `ownerType = "user"` untuk upload lampiran chat karena enum
  `owner_type` belum memiliki nilai khusus `chat`.

Response adalah canonical `MessageDTO` dan dapat dipakai ulang nanti sebagai
payload WebSocket `new_message`.

Attachment DTO:

```json
{
  "attachmentId": "uuid",
  "mediaId": "uuid",
  "fileName": "catatan.png",
  "mimeType": "image/png",
  "sizeBytes": 123456,
  "url": "https://storage.example/schools/.../catatan.png"
}
```

Known limitation: media provider saat ini dapat mengembalikan public URL.
Permission room tetap diterapkan di message list/send, tetapi direct file URL
masih bisa dibuka jika URL bocor. Private bucket, signed URL, atau protected
download proxy disiapkan untuk hardening berikutnya.

### Mark Room Read

`PATCH /rooms/:roomId/read`

```json
{
  "lastReadMessageId": "uuid"
}
```

`lastReadMessageId` opsional. Endpoint ini idempotent dan hanya berlaku jika
current user memiliki akses ke room.

Jika `lastReadMessageId` dikirim, message harus berada di room tersebut.
Endpoint menyimpan `last_read_msg_id` dan memperbarui `last_read_at`. Jika body
kosong, endpoint hanya memperbarui `last_read_at` tanpa menghapus
`last_read_msg_id` yang sudah tersimpan.

Response:

```json
{
  "message": "Chat room marked as read",
  "roomId": "uuid",
  "userId": "uuid",
  "lastReadMessageId": "uuid",
  "lastReadAt": "2026-06-26T03:05:00Z"
}
```

Unread count pada `GET /rooms` dihitung dari read receipt current user,
menggunakan `last_read_msg_id` jika tersedia atau `last_read_at` sebagai
fallback, dan tidak menghitung pesan yang dikirim oleh current user.

## WebSocket Realtime Transport

### Connect Chat WebSocket

`GET /api/ws/chat?token=<jwt>&schoolId=<school-uuid>`

Endpoint ini membuat koneksi WebSocket untuk event realtime chat. REST tetap
menjadi source of truth; message tetap dibuat melalui
`POST /api/chat/rooms/:roomId/messages`.

Handshake:

- JWT wajib valid.
- Browser client dapat mengirim token melalui query param `token` karena
  WebSocket tidak mudah mengirim `Authorization` header.
- `schoolId` wajib berisi active school context.
- Token tidak boleh dilog.
- User harus active school member (`school_users.deleted_at IS NULL`).
- Super admin hanya dapat terkoneksi jika juga memiliki membership aktif di
  school tersebut.

Event `new_message`:

```json
{
  "type": "new_message",
  "roomId": "uuid",
  "schoolId": "school-uuid",
  "payload": {
    "messageId": "uuid",
    "roomId": "uuid",
    "senderId": "uuid",
    "senderName": "Budi",
    "senderRole": "student",
    "content": "Halo.",
    "messageType": "text",
    "attachments": [],
    "createdAt": "2026-06-26T03:00:00Z",
    "isMine": false
  }
}
```

Event `message_read`:

```json
{
  "type": "message_read",
  "roomId": "uuid",
  "schoolId": "school-uuid",
  "payload": {
    "roomId": "uuid",
    "userId": "uuid",
    "lastReadMessageId": "uuid",
    "lastReadAt": "2026-06-26T03:05:00Z"
  }
}
```

Event `room_updated`:

```json
{
  "type": "room_updated",
  "roomId": "uuid",
  "schoolId": "school-uuid",
  "payload": {
    "reason": "new_message"
  }
}
```

`room_updated.payload.reason` saat ini berisi `new_message` atau
`message_read`.

Broadcast eligibility:

- School room dikirim hanya ke active school members di school yang sama.
- Group room dan DM dikirim hanya ke active `chat_room_members` yang masih
  memiliki active school membership.
- Removed school member tidak menerima event.
- Event tidak pernah dibroadcast lintas sekolah.

Sprint 18B mengirim event `new_message`, `message_read`, dan `room_updated`.
Message creation tetap melalui REST. Typing indicator, presence, notifications,
browser notification, dan message creation via WebSocket belum diimplementasikan.
