# Task Finishing — JurnalUmi Siap Launching

Status dasar: seluruh **P0 (16/16)** dan **P1-KRITIS (31/31)** di `Note/task-qa.md` sudah `[x]` (lihat `LAUNCH_REVIEW.md`). Semua branch `qa/*` sudah digabung ke `main` per 2026-09-16.

**Legend:** `[ ]` belum · `[~]` dikerjakan / belum terverifikasi · `[x]` selesai & terverifikasi
**Prioritas:** 🔴 LAUNCH-BLOCKER (wajib sebelum publish) · 🟠 LAUNCH-PENTING (minggu pertama setelah publish) · ⚪ POST-LAUNCH

---

## 🔴 LAUNCH-BLOCKER

### F-01 — Hapus Tailwind CDN, ganti Tailwind CLI build
- [ ] Pasang `tailwindcss` CLI (standalone binary, jangan tambah Node toolchain ke image produksi kalau bisa dihindari)
- [ ] Buat `tailwind.config.js` dengan `content: ["./web/views/**/*.templ", "./web/views/**/*.html", "./web/static/**/*.js"]` dan `darkMode: 'class'`
- [ ] Build ke `web/static/css/app.css` (minified), commit hasilnya ATAU jalankan di stage build `Dockerfile`
- [ ] Hapus **semua** `<script src="https://cdn.tailwindcss.com...">` di `web/views/layout.templ`, `admin.templ`, `auth.templ`, `landing.html`, `privacy.html`, `terms.html`, dan referensinya di `web/static/sw.js`
- [ ] Perketat CSP di `internal/middleware`: hapus izin `cdn.tailwindcss.com` dan `'unsafe-eval'` yang tadinya dibutuhkan Tailwind CDN
- **Kenapa blocker:** Tailwind CDN dilarang untuk production (ukuran, latensi, dan memaksa CSP longgar).
- **Verifikasi:** `grep -rn "cdn.tailwindcss.com" web/` → 0 hasil · buka dashboard/landing/auth/admin, styling tidak rusak · DevTools Console 0 CSP violation.

### F-02 — Self-host Alpine.js & HTMX
- [ ] Unduh Alpine.js & HTMX versi terkunci ke `web/static/js/` (catat versi di komentar file atau `README`)
- [ ] Ganti semua `<script src="https://unpkg.com/...">` / `cdn.jsdelivr.net` ke path lokal
- [ ] Update `web/static/sw.js` agar meng-cache aset lokal, bukan URL CDN
- [ ] Update CSP `script-src` agar cukup `'self'`
- **Verifikasi:** `grep -rn "unpkg.com\|jsdelivr" web/` → 0 hasil · matikan koneksi eksternal (DevTools block third-party) lalu cek HTMX submit transaksi & Alpine dropdown tetap jalan.

### F-03 — Ikon PWA di-host sendiri
- [ ] Buat `icon-192.png`, `icon-512.png`, dan versi `maskable` di `web/static/icons/`
- [ ] Update `manifest.json`: `purpose: "any"` dan `purpose: "maskable"` terpisah
- [ ] Hapus semua referensi CDN Flaticon
- **Verifikasi:** Lighthouse PWA → "Installable" hijau · install ke home screen Android/Chrome, ikon tampil benar.

### F-04 — Migrasi ke goose, matikan AutoMigrate di production
- [ ] Tambah `github.com/pressly/goose/v3`, buat direktori `migrations/`
- [ ] Generate migration awal dari skema saat ini (dump schema DB dev → `00001_init.sql` dengan `-- +goose Up/Down`)
- [ ] `internal/db/db.go`: jalankan `AutoMigrate` hanya kalau `APP_ENV != "production"`; di production jalankan `goose.Up`
- [ ] Tambah `cmd/migrate` (atau flag `-migrate`) supaya bisa dipanggil terpisah
- [ ] Dokumentasikan prosedur migrasi + rollback di `README.md`
- **Kenapa blocker:** `AutoMigrate` di produksi bisa diam-diam mengubah/drop kolom, tidak ada jejak & tidak bisa rollback.
- **Verifikasi:** DB kosong baru → `goose up` menghasilkan skema identik dengan dev (bandingkan `\d+` tiap tabel) · `goose down-to 0` bersih tanpa error.

### F-05 — CI GitHub Actions
- [ ] `.github/workflows/ci.yml` pada `push` + `pull_request` ke `main`
- [ ] Step: `go build ./...`, `go vet ./...`, `go test ./... -race`, `templ generate && git diff --exit-code`
- [ ] Cache modul Go, pin `go-version: '1.26'`
- [ ] Jadikan CI **required status check** di branch protection `main`
- **Verifikasi:** buat PR dummy dengan `.templ` yang belum di-generate → CI harus MERAH; setelah di-generate → HIJAU.

### F-06 — Verifikasi produksi end-to-end di Coolify
- [ ] Isi env produksi di Coolify: `SMTP_USERNAME`, `SMTP_PASSWORD`, `APP_ENV=production`, `SESSION_SECRET`, kredensial S3/MinIO backup (**jangan** di repo)
- [ ] Pastikan auto-deploy dari push `main` benar-benar jalan (cek log deploy Coolify, bukan asumsi)
- [ ] Domain custom terpasang + TLS valid (bukan self-signed), HTTP → HTTPS redirect aktif
- [ ] Smoke test di domain produksi: register → verifikasi email (email asli sampai ke inbox) → login → buat dompet → catat transaksi → lihat dashboard → reset password
- [ ] Health check endpoint terpasang & dipakai Coolify
- **Verifikasi:** tempel hasil smoke test (waktu + hasil tiap langkah) di Log.

### F-07 — Uji restore backup produksi sungguhan
- [ ] Jalankan `scripts/backup.sh` melawan DB produksi → artefak terenkripsi masuk S3/MinIO
- [ ] `scripts/restore.sh` ke DB staging kosong → data cocok (bandingkan `COUNT(*)` tabel utama + spot-check 1 tenant)
- [ ] Jadwalkan backup harian (cron Coolify / scheduled container) + verifikasi jadwal jalan minimal 1 siklus
- [ ] Catat RPO/RTO aktual (berapa menit backup, berapa menit restore) di `README.md`
- **Verifikasi:** tempel jumlah baris sebelum/sesudah restore.

### F-08 — Finalisasi konten hukum & proses pembayaran manual
- [ ] Prof me-review & finalisasi teks `/privacy` dan `/terms` (draft UU PDP saat ini belum final)
- [ ] Landing page: rekening BSI + instruksi transfer manual jelas, termasuk SLA aktivasi (mis. "maks 1×24 jam kerja")
- [ ] Tulis runbook singkat di `README.md`: siapa cek mutasi, cara upgrade akun user manual lewat admin panel, cara handle salah nominal/refund
- **Kenapa blocker:** Mayar DITUNDA (`QA-DECISIONS.md`), jadi jalur pembayaran aktif 100% manual — tanpa runbook, user bayar tapi tidak ter-upgrade.
- **Verifikasi:** simulasi 1 siklus: user baru transfer (dummy) → admin upgrade → user melihat fitur Premium terbuka.

---

## 🟠 LAUNCH-PENTING

### F-09 — Transfer antar dompet
- [ ] `type='transfer'` + `to_wallet_id`, satu transaksi memindahkan saldo atomik (dalam DB transaction)
- [ ] Validasi: dompet asal ≠ tujuan, saldo cukup, nominal > 0, kedua dompet milik tenant yang sama
- [ ] Transfer **tidak boleh** terhitung sebagai pemasukan/pengeluaran di dashboard, budget, maupun laporan
- **Verifikasi:** unit test: saldo A+B konstan sebelum/sesudah transfer; `totalIncome`/`totalExpense` tidak berubah.

### F-10 — Pagination (dashboard, reports, export)
- [ ] Limit default 25 baris + navigasi halaman berbasis HTMX (tanpa reload penuh)
- [ ] Query pakai `LIMIT/OFFSET` (atau keyset) — jangan load semua lalu potong di Go
- [ ] Export tetap memuat seluruh data periode, bukan cuma halaman aktif
- **Verifikasi:** seed 2.000 transaksi → dashboard render < 500 ms; cek query count tidak N+1.

### F-11 — Notifikasi email (aktifkan `SendEmailNotification` yang kini dead code)
- [ ] Reminder utang H-3 & H-1 (cron harian)
- [ ] Alert budget 80% & 100% — **maks 1 email per kategori per ambang per bulan** (butuh tabel penanda supaya tidak spam)
- [ ] Ringkasan bulanan tiap tanggal 1
- [ ] Semua job idempotent: kalau cron jalan dua kali, email tidak terkirim dobel
- **Verifikasi:** unit test penanda anti-dobel; uji kirim ke Mailhog lalu 1× ke inbox asli.

### F-12 — Trial Premium 14 hari otomatis saat register
- [ ] Set `plan='premium'` + `plan_expires_at = now + 14 hari` saat tenant dibuat
- [ ] Cron downgrade jam-an yang sudah ada (`downgradeExpiredPlans`) harus menangani ini tanpa perubahan khusus — verifikasi
- [ ] Banner sisa hari trial di dashboard + email H-3 sebelum trial habis
- **Verifikasi:** buat tenant dengan `plan_expires_at` kemarin → cron menurunkannya ke free, fitur Premium terkunci.

### F-13 — Quick entry < 5 detik
- [ ] FAB di mobile, form fokus otomatis ke nominal
- [ ] Default ke dompet & kategori yang terakhir dipakai user
- [ ] Shortcut nominal (10rb / 20rb / 50rb / 100rb)
- **Verifikasi:** stopwatch manual di layar mobile — buka app → transaksi tersimpan < 5 detik.

### F-14 — Audit log + halaman "Aktivitas Keluarga"
- [ ] Tabel `audit_logs` (tenant, user, aksi, entitas, id entitas, ringkasan perubahan, waktu)
- [ ] Catat minimal: create/update/delete transaksi, ubah dompet, undang/hapus anggota, upgrade plan
- [ ] Halaman read-only untuk role `owner`/`spouse`, ter-paginate
- **Verifikasi:** lakukan 5 aksi berbeda → semuanya muncul dengan pelaku & waktu yang benar.

### F-15 — Dark mode
- [ ] Prasyarat F-01 selesai · `darkMode: 'class'` + toggle + simpan preferensi di `localStorage`
- [ ] Hormati `prefers-color-scheme` saat belum ada preferensi tersimpan
- **Verifikasi:** semua halaman utama terbaca di kedua mode (kontras ≥ 4.5:1 untuk teks isi).

---

## ⚪ POST-LAUNCH (jangan dikerjakan sebelum semua 🔴 selesai)

- [ ] Offline queue IndexedDB + indikator "menunggu sinkronisasi"
- [ ] Import CSV/Excel dari catatan lama (mapping + preview + rollback)
- [ ] Emergency Fund Health Score (rata-rata pengeluaran 3 bulan × faktor 6/9/12)
- [ ] Kalkulator Zakat Maal (nisab 85 gr emas, haul, 2,5%)
- [ ] Rekap tahunan zakat/infaq siap cetak
- [ ] Grafik pertumbuhan aset logam mulia dari histori `price_snapshots`
- [ ] Export PDF asli (`maroto`/`gofpdf`) menggantikan `window.print()`
- [ ] Recurring transaction auto-post + notifikasi
- [ ] Alokasi 50/30/20 sebagai saran satu klik saat pemasukan besar masuk
- [ ] Reminder piutang via WhatsApp deep link (`wa.me`)
- [ ] Onboarding wizard (status keluarga, dompet awal, target dana darurat)
- [ ] Push notification harian "sudah catat pengeluaran hari ini?"

---

## Gate Publish

Publish hanya setelah **F-01 s/d F-08 semuanya `[x]`**. F-09 s/d F-15 boleh menyusul di minggu pertama.

## Log

2026-09-16 — konsolidasi — gabungkan 9 branch `qa/*` ke `main`; `qa/launch-review-summary` membawa `LAUNCH_REVIEW.md`, 8 branch fitur lainnya sudah tergantikan implementasi yang lebih baru di `main` sehingga dimerge dengan strategi `ours` (tidak ada regresi kode). verified: `go build ./...` OK, `go test ./...` OK (semua paket lulus).
