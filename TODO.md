RBAC Improvements (Future)
Audit dan implementasi UI multi-role assignment (checkbox/multi-select).
Ubah AdminUsers dari single-role menjadi multi-role editor.
Pastikan update role tidak lagi menghapus role lain secara tidak sengaja.
Evaluasi invitation & enroll untuk mendukung multi-role bila memang dibutuhkan.
Audit seluruh frontend terhadap asumsi single-role.

Request URL
http://localhost:8080/api/schools?page=1&limit=20&status=all&sortBy=createdAt&order=desc
Request Method
GET
Status Code
400 Bad Request
{"error":"School context required (SchoolId header or schoolCode param)"}

Request URL
http://localhost:8080/api/schools/summary
Request Method
GET
Status Code
400 Bad Request
{"error":"School context required (SchoolId header or schoolCode param)"}

admin invite anggota sekolah by email saja. ketika user mau accept, yang sudah punya akun lgsg klik "terima". yang belum, buat akun dulu

logging untuk admin sekolah mengenai sekolah (backend banyak bertambah, web socket) dan juga superadmin (lebih umum, ga sedetail admin sekolah)