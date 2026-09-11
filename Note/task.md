# Task Tracker JurnalUmi

- **Direset:** 8 September 2026, berdasarkan audit kode di `review.md`.
- **Kenapa direset:** versi sebelumnya menandai **seluruh** task `[x]` padahal sebagian besar fitur inti (RBAC, snowball, budget, net worth, cron, billing, dark mode, docker-compose) tidak ada di kode. Checklist yang tidak jujur membuat estimasi jarak ke production salah total.

## Aturan menandai task
> `[x]` hanya boleh dicentang bila: **(1)** kode-nya ada & terpanggil dari route/UI, **(2)** sudah dicoba manual dan berhasil, **(3)** error path ditangani. Kode yang ditulis tapi belum diverifikasi = `[~]` (in progress). Teks statis di HTML **bukan** fitur.

**Legend:** `[ ]` belum · `[~]` sedang dikerjakan / belum terverifikasi · `[x]` selesai & terverifikasi · 🔴 blocker keamanan

---

## SUDAH ADA & BERFUNGSI (hasil verifikasi audit)

- [x] Boilerplate Go 1.25 + Echo + Templ + Air, struktur folder `cmd/ internal/ web/`
- [x] Koneksi PostgreSQL + GORM + AutoMigrate (sementara; akan diganti goose)
- [x] Register (buat tenant + owner + wallet default + seed 8 kategori, dalam DB transaction)
- [x] Login/Logout berbasis session cookie + bcrypt
- [x] Middleware `RequireAuth` (autentikasi saja, **belum** otorisasi)
- [x] Create transaksi income/expense + update saldo wallet dalam satu DB transaction
- [x] Create wallet, create kategori, create aset komoditas, create utang/piutang
- [x] Halaman: dashboard, assets, debts, reports, family, admin, landing (UI rapi)
- [x] Export CSV transaksi
- [x] Layout Templ + nav + HTMX `hx-boost`
- [x] `manifest.json` + `sw.js` terdaftar (**strateginya salah**, lihat P1)
- [x] Landing page + section demo video YouTube

---

## P0 — BLOCKER: sebelum boleh dipakai orang selain founder

### Keamanan
- [ ] 🔴 Proteksi seluruh route `/admin/*`: middleware auth + role `superadmin` (`main.go:57-59` saat ini terbuka publik)
- [x] 🔴 Pindahkan `SESSION_SECRET` ke env; refuse start bila kosong saat `APP_ENV=production` (`main.go:32` hardcoded & ter-commit)
- [ ] 🔴 Cookie session: `Secure` + `SameSite=Lax` + rotasi session id saat login
- [ ] 🔴 Middleware `RequireRole(...)` + terapkan matrix `Role_Permission.md` ke semua route (kini role sama sekali tidak dicek)
- [ ] 🔴 `middleware.CSRF()` global + hidden token di semua form Templ
- [ ] 🔴 Rate limit `/login` & `/register` + lockout setelah 5 kali gagal
- [ ] Hapus `middleware.CORS()` global (tidak dibutuhkan untuk SSR)
- [ ] Tambah security header (HSTS, nosniff, Referrer-Policy, CSP dasar)
- [ ] Seed CLI untuk membuat user `superadmin` (jangan lewat form publik)

### Integritas data
- [ ] 🔴 Matikan/ganti `DebtPayPOST` yang membagi dua sisa utang (`debts.go:88-91`) → form bayar nyata: pilih wallet + nominal → insert transaksi + potong saldo + kurangi sisa + catat `debt_payments`, satu DB transaction
- [ ] 🔴 Tangani error `db.Create/Save` di semua handler (kini diabaikan di 6 tempat) + tampilkan flash message ke user
- [ ] Validasi input server-side (nominal > 0, tanggal wajar, wallet/kategori milik tenant sendiri, enum tipe)

### Infrastruktur dasar
- [ ] Perbaiki `Dockerfile.dev` (pakai `golang:1.25-alpine`, kini 1.23 → build gagal vs `go.mod`)
- [ ] Buat `docker-compose.yml` (Postgres 16 + Mailhog + app)
- [ ] Buat `.env.example` + `README` cara menjalankan
- [ ] Hapus duplikasi `e.Static("/static")` (`main.go:42` & `:83`)

---

## P1 — sebelum boleh dijual ke publik

### Fondasi data (kerjakan lebih dulu, semua fitur lain menumpang di sini)
- [ ] Migrasi ke **goose**, matikan `AutoMigrate` di production
- [ ] Konversi seluruh nominal ke `int64` rupiah penuh (hapus `float64` untuk uang)
- [ ] Isi & gunakan `category_id` di transaksi (kini hanya `category_name` string → laporan per kategori mustahil)
- [ ] Tabel baru: `budgets`, `recurring_rules`, `debt_payments`, `assets`, `price_snapshots`, `audit_logs`, `payments`
- [ ] Aktifkan **Row-Level Security** + `SET LOCAL app.tenant_id` per request + helper `Scoped(c)`
- [ ] Transaksi `opening_balance` untuk saldo awal wallet + job rekonsiliasi saldo harian
- [ ] Test: perhitungan uang, isolasi tenant, RBAC, plan limit

### Ledger & UX inti
- [ ] Edit & hapus transaksi (soft delete + audit log) — penyebab utama user berhenti memakai app keuangan
- [ ] Transfer antar dompet (`type='transfer'` + `to_wallet_id`)
- [ ] Pagination di dashboard, reports, dan export
- [ ] `FormatRupiah` format Indonesia (`Rp 500.000`, kini `Rp 500000.00`)
- [ ] Filter periode di dashboard (label "Bulan Ini" kini menjumlah seluruh transaksi sepanjang masa)
- [ ] Quick entry < 5 detik: FAB mobile, default wallet & kategori terakhir, shortcut nominal
- [ ] Import CSV/Excel dari catatan lama (mapping + preview + rollback)

### Fitur yang selama ini hanya teks di HTML
- [ ] Budget per periode + indikator hijau/kuning/merah **dihitung di server**
- [ ] Net Worth = total aset − total utang (ganti kartu "Coming Soon" di `dashboard.templ:64`)
- [ ] Kalkulator **Snowball & Avalanche** yang benar-benar menghitung + unit test
- [ ] Emergency Fund Health Score (rata-rata pengeluaran 3 bulan × faktor 6/9/12)
- [ ] Sinking funds: target + tanggal target → setoran bulanan yang dibutuhkan + progress bar
- [ ] Audit log + halaman "Aktivitas Keluarga" (bukti nyata janji transparansi pasangan)

### Aset & harga logam mulia
- [ ] Cron harian ambil harga → `price_snapshots`; halaman aset **hanya baca snapshot** (kini 1 HTTP call per aset, timeout 3 detik masing-masing)
- [ ] Ganti/verifikasi sumber harga (`api.logammulia.com` tidak eksis → selalu fallback konstanta 1.450.000)
- [ ] Fallback harga manual per aset selama sumber belum tersedia
- [ ] Perbaiki perhitungan dinar (keping → gram) & perak (kini hardcoded 16.500/gram)

### Notifikasi & scheduler
- [ ] Scheduler `robfig/cron` + DB lock
- [ ] Debt reminder H-3 & H-1 (sambungkan `SendEmailNotification` yang kini dead code)
- [ ] Budget alert 80% & 100% (maks 1 email per kategori per ambang per bulan)
- [ ] Ringkasan bulanan email tanggal 1
- [ ] Verifikasi email + reset password

### Monetisasi
- [ ] Middleware `RequirePlan(feature)` + enforcement limit Free (2 dompet, 1 user, histori 3 bulan)
- [ ] Trial Premium 14 hari otomatis saat register
- [ ] Cron penurunan plan saat `plan_expires_at` lewat
- [ ] Route `POST /activate-voucher` (form di `landing.html:247` kini menuju 404) + redeem idempotent
- [ ] Webhook Mayar.id `payment.success` (verifikasi signature, idempotent, tabel `payments`)
- [ ] Satu sumber harga (env/konstanta) untuk kode + landing + materi promosi

### Frontend & PWA
- [ ] Tailwind via CLI build (hapus `cdn.tailwindcss.com` — dilarang untuk production)
- [ ] Self-host Alpine.js & HTMX
- [ ] Dark mode (`darkMode: 'class'` + toggle + localStorage) — prasyaratnya Tailwind CLI
- [ ] `dashboard.templ` memakai `@Layout` (kini nav di-copy-paste, tanpa PWA & hx-boost)
- [ ] `sw.js`: network-first untuk route data, cache-first hanya aset statis, **jangan precache halaman ber-auth**
- [ ] Bersihkan cache saat logout (`caches.delete`) — kini data keuangan bisa tersaji di HP yang dipinjam
- [ ] Offline queue IndexedDB + indikator "menunggu sinkronisasi"
- [ ] Icon PWA di-host sendiri, 192 & 512 terpisah + maskable (kini menunjuk CDN Flaticon)

### Kepatuhan & operasional (gate publik launch)
- [ ] Halaman Kebijakan Privasi & Syarat Ketentuan (UU PDP 27/2022)
- [ ] Hapus akun mandiri + ekspor data mandiri
- [ ] `pg_dump` harian terenkripsi + **uji restore minimal 1× (catat tanggalnya di sini)**
- [ ] FAQ keamanan data di landing page
- [ ] Dockerfile production + deploy via Coolify + domain + TLS
- [ ] CI: `go build`, `go vet`, `go test`, `templ generate --check`

---

## P2 — diferensiasi & retensi (setelah ada pengguna berbayar)

- [ ] **Kalkulator Zakat Maal** (nisab 85 gr emas, haul, 2,5%) dari data aset & saldo yang sudah tercatat
- [ ] Rekap tahunan zakat/infaq siap cetak
- [ ] Grafik pertumbuhan aset logam mulia dari histori `price_snapshots`
- [ ] Export **PDF asli** (`maroto`/`gofpdf`) menggantikan `window.print()`
- [ ] Recurring transaction auto-post + notifikasi
- [ ] Alokasi 50/30/20 sebagai saran satu klik saat pemasukan besar masuk
- [ ] Reminder piutang via WhatsApp deep link (`wa.me`)
- [ ] Onboarding wizard (status keluarga, dompet awal, target dana darurat)
- [ ] Push notification harian "sudah catat pengeluaran hari ini?"

---

## Catatan operasional

- Update file ini **setelah** fitur diverifikasi berjalan, bukan setelah kode ditulis.
- Setiap kali sebuah item P0 selesai, jalankan ulang audit singkat pada area itu.
- Sumber kebenaran: `PRD.md` (apa yang dibangun) · `Role_Permission.md` (siapa boleh apa) · `review.md` (kenapa) · `timeline.md` (kapan).
