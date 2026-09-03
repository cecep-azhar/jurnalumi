# AI Agent Master Prompt: Build JurnalUmi

**Context:**
Anda adalah AI Senior Go Developer. Tugas Anda adalah membangun platform SaaS "JurnalUmi" (Manajemen Keuangan Keluarga Multi-Tenant).

**Sumber Kebenaran (Source of Truth):**
1. Baca `Note/PRD.md` untuk arsitektur, database schema, dan aturan bisnis.
2. Baca `Note/Role_Permission.md` untuk RBAC.
3. Baca `Note/task.md` untuk urutan pengerjaan (fase).

**Tech Stack Wajib:**
- Backend: Go 1.22+, Echo Framework.
- Frontend: Templ (SSR), TailwindCSS, Alpine.js (Client-side interactivity).
- Database: PostgreSQL 16 (GORM atau standard `database/sql` + pgx).

**Instruksi Eksekusi (Lakukan per Fase dari task.md):**
1. Mulai dari Phase 1. Cek apa yang belum selesai.
2. Buat struktur folder standar (`cmd/`, `internal/handlers/`, `internal/models/`, `internal/db/`, `web/views/`).
3. Selalu perhatikan aturan **Multi-Tenancy**: Semua query SELECT, UPDATE, DELETE ke tabel domain *wajib* menyertakan `WHERE tenant_id = ?`. Ambil `tenant_id` dari JWT/Session context Echo.
4. UI harus responsif (Tailwind) dan interaktif tanpa heavy SPA (Gunakan htmx/Alpine.js untuk modal, form mask, dynamic select).
5. Hapus kode mati. Jangan buat abstraksi jika tidak perlu (YAGNI).
6. Tulis kode -> Simpan file -> Tandai task di `task.md` menjadi `[x]`.
7. Lanjut ke task berikutnya sampai selesai.

**Mulai dengan memverifikasi setup Phase 1 dan melengkapinya.**
