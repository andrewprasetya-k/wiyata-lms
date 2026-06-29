# Chat API

Base URL: `/api/chat`

REST chat mendukung satu room sekolah aktif, custom group room, dan direct
message untuk warga aktif di sekolah yang sama.

## Scope MVP

- School-wide chat selalu tersedia sebagai room utama sekolah.
- Custom group room dapat dibuat oleh warga aktif sekolah.
- Direct message dapat dibuka antar dua warga aktif di sekolah yang sama.
- Text-only messages.
- REST API only.
- Admin Sekolah, teacher, dan student boleh berpartisipasi jika masih menjadi
  member aktif sekolah tersebut.
- Tidak ada WebSocket/realtime.
- Tidak ada subject/class room, attachment, typing indicator,
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
