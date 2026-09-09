# Launch Gate Checklist — JurnalUmi (QA/Cron)

- **Dibuat:** 9 September 2026, oleh Claude atas permintaan Cecep (Prof), disepakati lewat sesi klarifikasi scope.
- **Basis:** seluruh P0 di `task.md` (16 item, tidak dikurangi) + kurasi P1 "kritis untuk publish" (31 item, dipilih dari total ±49 item P1 di `task.md`). Rasional kurasi ada di tiap section di bawah — **kalau Prof tidak setuju satu item dipindah kritis/non-kritis, edit langsung di file ini, Hermes ikut yang tertulis di sini, bukan asumsi sendiri.**
- **Yang TIDAK masuk sini:** P1 non-kritis & seluruh P2 (referensi di bagian paling bawah). `task.md` tetap sumber kebenaran untuk backlog lengkap; file ini adalah **subset gate "boleh publish"**, lebih sempit & lebih dieksekusi cron.
- **Prosedur kerja Hermes lengkap ada di `prompt-qa.md`** — file ini hanya data (checklist + log), bukan instruksi.

> ⚠️ **Catatan insiden (9 Sep 2026, sore):** versi pertama file ini tidak pernah di-commit ke `main` (cuma untracked file). Saat Hermes bikin branch `qa/p0-01-admin-auth` lalu `git add -A`, file ini ikut ter-commit DI DALAM branch itu — dan hilang dari `main` begitu Hermes pindah branch ke task berikutnya. Akibatnya task kedua (SESSION_SECRET) tidak menemukan file ini dan fallback nulis status ke `task.md` saja. **Fix:** file ini sekarang di-commit langsung ke `main` (lihat instruksi commit di akhir respons), dan `prompt-qa.md` diperbarui — Note/task-qa.md & Note/task.md tidak boleh lagi ikut ter-commit di dalam branch `qa/*`, update status selalu lewat commit terpisah langsung di `main`.

## Legend
`[ ]` belum · `[~]` sedang dikerjakan **atau PR sudah dibuka tapi belum di-merge Prof** · `[x]` selesai — kode **sudah merged ke `main`** (bukan cuma PR terbuka) · `[!]` blocked, butuh keputusan Prof (lihat catatan di item) · 🔴 = kalau di-skip: risiko keamanan/uang/hukum, bukan sekadar kurang fitur.

**Syarat centang `[x]`:** PR merged ke `main` + `go build ./...` lulus + dicoba manual end-to-end + error path ditangani. Kode ditulis & PR terbuka tapi belum di-merge = tetap `[~]`.

**Concurrency:** boleh lebih dari satu item `[~]` sekaligus (tiap item = branch independen) — yang membatasi adalah plafon PR terbuka (≤3, lihat `prompt-qa.md`), bukan "1 item pada satu waktu".

---

## P0 — BLOCKER (wajib 100% sebelum siapa pun selain Prof boleh pakai, termasuk soft launch)

*Sumber: `review.md` §1 dan §3.1. Urutan berikut = urutan pengerjaan, jangan diacak — item 4 (RequireRole) dan governance lain jadi prasyarat banyak item P1 di bawah.*

### Keamanan
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[x]` PR [#4](https://github.com/cecep-azhar/jurnalumi/pull/4) | QA-P0-01 🔴 | Proteksi seluruh `/admin/*`: middleware auth + role `superadmin` | `cmd/server/main.go:55-57` (saat ini tanpa middleware sama sekali) | — |
| `[x]` | QA-P0-02 🔴 | `SESSION_SECRET` dari env; refuse start jika kosong saat `APP_ENV=production` | `cmd/server/main.go:32` (hardcoded `jurnalumi-super-secret-key`, ter-commit) | — |
| `[x]` | QA-P0-03 🔴 | Cookie session: `Secure` + `SameSite=Lax` + rotasi session id saat login | `internal/handlers/auth.go:37-41` | — |
| `[~]` PR [#5](https://github.com/cecep-azhar/jurnalumi/pull/5) | QA-P0-04 🔴 | Middleware `RequireRole(...)` + terapkan matrix `Role_Permission.md` ke semua route | `internal/middleware/auth.go` (baru), semua route di `main.go` | QA-P0-01 |
| `[~]` PR [#6](https://github.com/cecep-azhar/jurnalumi/pull/6) | QA-P0-05 🔴 | `middleware.CSRF()` global + hidden token di semua form Templ | `cmd/server/main.go`, semua file `web/views/*.templ` yang punya `<form>` | — |
| `[~]` PR [#7](https://github.com/cecep-azhar/jurnalumi/pull/7) | QA-P0-06 🔴 | Rate limit `/login` & `/register` + lockout 5x gagal | `cmd/server/main.go:49-51` | — |
| `[x]` PR [#8](https://github.com/cecep-azhar/jurnalumi/pull/8) | QA-P0-07 | Hapus `middleware.CORS()` global (tidak perlu, ini SSR bukan API publik) | `cmd/server/main.go:39` | — |
| `[x]` PR [#9](https://github.com/cecep-azhar/jurnalumi/pull/9) | QA-P0-08 | Security header dasar: HSTS, X-Content-Type-Options nosniff, Referrer-Policy, CSP dasar | `cmd/server/main.go` | — |
| `[x]` PR [#10](https://github.com/cecep-azhar/jurnalumi/pull/10) | QA-P0-09 | Seed CLI untuk buat user `superadmin` (bukan lewat form publik) | `cmd/` (command baru, mis. `cmd/seed/`) | QA-P0-04 |

### Integritas data
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[~]` PR [#11](https://github.com/cecep-azhar/jurnalumi/pull/11) | QA-P0-10 🔴 | Matikan/ganti `DebtPayPOST` yang membagi dua sisa utang → bayar nyata: pilih wallet+nominal → insert transaksi + potong saldo + kurangi sisa, 1 DB transaction | `internal/handlers/debts.go:82-91` (komentar sendiri: "mock implementation") | — |
| `[ ]` | QA-P0-11 🔴 | Tangani error `db.Create/Save` di semua handler (kini diabaikan, redirect seolah sukses) | `dashboard.go:132,154`, `features.go:114`, `assets.go:69`, `debts.go:77`, `admin.go:66` | — |
| `[ ]` | QA-P0-12 | Validasi input server-side (nominal > 0, tanggal wajar, wallet/kategori milik tenant sendiri, enum tipe valid) | semua handler POST | — |

### Infrastruktur dasar
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P0-13 | Perbaiki `Dockerfile.dev` → `golang:1.25-alpine` (kini 1.23, mismatch `go.mod` → build gagal) | `Dockerfile.dev:1` | — |
| `[ ]` | QA-P0-14 | `docker-compose.yml` (Postgres 16 + Mailhog + app) untuk dev lokal | root (baru) | QA-P0-13 |
| `[ ]` | QA-P0-15 | `.env.example` + `README.md` cara menjalankan | root (baru) | QA-P0-02 |
| `[ ]` | QA-P0-16 | Hapus duplikasi `e.Static("/static", ...)` | `main.go:40` & `main.go:81` | — |

---

## P1-KRITIS — wajib sebelum publish (soft launch sekalipun)

*Kurasi: fitur yang (a) dipakai untuk KLAIM di landing page/pricing, (b) menyangkut kebenaran ANGKA UANG, (c) menyangkut PRIVASI data keluarga, atau (d) syarat hukum/teknis untuk online sama sekali. Fitur yang "sekadar kurang nyaman" (dark mode, transfer dompet, import CSV, quick-add, offline queue) sengaja DITUNDA — daftar lengkap di bagian bawah, boleh dikerjakan setelah publish.*

### A. Fondasi data (kerjakan duluan — banyak item lain menumpang di sini)
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P1-01 🔴 | Konversi seluruh nominal uang ke `int64` rupiah penuh (hapus `float64`) | `internal/models/models.go`, semua handler pemroses uang | — |
| `[ ]` | QA-P1-02 | Isi & pakai `category_id` di transaksi (kini cuma `category_name` string → laporan per kategori mustahil) | `internal/handlers/dashboard.go:87-96`, `models.go` | QA-P1-01 |
| `[ ]` | QA-P1-03 🔴 | Aktifkan PostgreSQL Row-Level Security + `SET LOCAL app.tenant_id` per request + helper `Scoped(c)` (lihat S9 di `review.md`) | `internal/db/`, semua handler (ganti `db.DB` langsung) | QA-P0-04 |
| `[ ]` | QA-P1-04 | Tabel baru: `budgets(tenant,category,period)`, `price_snapshots`, `payments` (recurring_rules & audit_logs DITUNDA, lihat non-kritis) | `internal/models/models.go`, migration | QA-P1-01, QA-P1-02 |
| `[ ]` | QA-P1-05 | Transaksi `opening_balance` saat wallet dibuat + job rekonsiliasi saldo harian (deteksi saldo vs histori divergen) | `internal/handlers/dashboard.go:116-134` | QA-P1-01 |
| `[ ]` | QA-P1-06 🔴 | Test minimal: perhitungan uang (rounding, konversi), isolasi tenant (RLS bocor?), RBAC (role rendah tidak bisa akses route tinggi) | `internal/**/*_test.go` (baru) | QA-P1-03, QA-P0-04 |

### B. Ledger inti (kejujuran angka & kebiasaan dasar user)
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P1-07 🔴 | Edit & hapus transaksi (soft delete) — penyebab #1 orang berhenti pakai app keuangan | `internal/handlers/dashboard.go`, route baru | QA-P1-01, QA-P1-05 |
| `[ ]` | QA-P1-08 | `FormatRupiah` format Indonesia (`Rp 500.000`, kini `Rp 500000.00`) | `web/views/dashboard.templ:9-11` | — |
| `[ ]` | QA-P1-09 🔴 | Filter periode di dashboard — label "Bulan Ini" kini menjumlah SELURUH transaksi sepanjang masa (data menyesatkan) | `internal/handlers/dashboard.go:44-51` | — |

### C. Fitur yang selama ini cuma teks statis di HTML (janji jual yang belum ditepati)
> Aturan per item: implementasi beneran ATAU non-aktifkan tombolnya + label jelas "segera hadir" (pilih salah satu, jangan biarkan terlihat berfungsi padahal tidak — Aturan Keras #8 di `prompt-qa.md`). Prioritas: implement kalau murah (< 1 hari kerja), disable+label kalau besar.
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P1-10 🔴 | Budget per kategori: hitung realisasi vs `BudgetLimit` di server + indikator hijau/kuning/merah (murah — implement) | `web/views/categories*.templ`, handler baru | QA-P1-02, QA-P1-04 |
| `[ ]` | QA-P1-11 🔴 | Net Worth = total aset − total utang, ganti kartu `"Sisa Utang (Coming Soon)"` (murah — implement, data sudah ada) | `web/views/dashboard.templ:64-67` | — |
| `[ ]` | QA-P1-12 | Kalkulator Snowball & Avalanche beneran menghitung urutan pelunasan (ganti teks statis, murni kalkulasi atas data Debt yang sudah ada — implement) | `web/views/debts.templ:37-39`, service baru | — |
| `[ ]` | QA-P1-13 | Sinking Fund / Emergency Fund: versi dasar (target vs setoran terkumpul → progress %). Health score 6x/9x/12x boleh menyusul post-publish | `internal/handlers/*`, model wallet target | QA-P1-01 |

### D. Aset & harga logam mulia (bug fungsional aktif, bukan cuma fitur kurang)
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P1-14 🔴 | Cron harian ambil harga → `price_snapshots`; halaman `/assets` HANYA baca snapshot (fix bug: 20 aset = 20×timeout 3 detik = halaman hang ~60 detik) | `internal/handlers/assets.go:31-35`, scheduler baru | QA-P1-04, QA-P1-19 |
| `[ ]` | QA-P1-15 🔴 | Ganti/verifikasi sumber harga (`api.logammulia.com` tidak eksis, selalu fallback konstanta) ATAU label jelas "harga manual, update berkala" — jangan sebut "real-time" kalau bohong | `internal/services/gold.go:17` | QA-P1-14 |
| `[ ]` | QA-P1-16 | Perbaiki perhitungan dinar (keping→gram) & perak (kini hardcoded 16.500/gram) | `internal/services/gold.go`, `assets.go` | — |

### E. Akun dasar (bukan email marketing — ini keamanan akun)
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P1-17 | Verifikasi email saat register | `internal/handlers/auth.go`, `internal/services/email.go` (sudah ada SMTP client, belum dipakai) | QA-P0-15 |
| `[ ]` | QA-P1-18 | Reset password (lupa password) | `internal/handlers/auth.go` (baru) | QA-P1-17 |

### F. Scheduler dasar (infra minimal, HANYA untuk 2 job kritis di bawah — bukan email reminder)
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P1-19 | Setup `robfig/cron` + DB lock (agar tidak dobel-jalan kalau ada >1 instance) — dipakai oleh QA-P1-14 & QA-P1-21 saja di tahap ini | `internal/scheduler/` (baru) | — |

### G. Monetisasi — ini "keterikatan flow bisnis" yang Prof maksud, saat ini putus total
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P1-20 🔴 | Middleware `RequirePlan(feature)` + enforcement limit Free tier (tanpa ini semua user dapat premium gratis selamanya — model bisnis tidak eksis) | semua handler yang membatasi fitur premium | QA-P0-04, QA-P1-04 |
| `[ ]` | QA-P1-21 | Cron turunkan plan saat `plan_expires_at` lewat (pakai QA-P1-19) | `internal/handlers/admin.go` | QA-P1-19, QA-P1-20 |
| `[ ]` | QA-P1-22 🔴 | Route `POST /activate-voucher` — form di `landing.html:247` KINI MENUJU 404 di landing page publik, ini bug yang langsung kelihatan user | `cmd/server/main.go`, handler baru | — |
| `[ ]` | QA-P1-23 🔴 | Webhook Mayar.id `payment.success` (verifikasi signature, idempotent, tabel `payments`) — TANPA INI: orang bayar di Mayar, JurnalUmi tidak pernah tahu, user tidak ter-upgrade. Ini flow bisnis paling kritis yang bolong | `internal/handlers/` (baru) | QA-P1-04 |
| `[ ]` | QA-P1-24 | Satu sumber harga (env/konstanta) dipakai konsisten di kode + landing + materi promosi (kini 3 angka beda: 39rb/bln, 190rb/thn, link `jurnalumi-premium-39k`) | lintas file, lihat `review.md` F15 | — |

### H. Frontend — hanya yang menyangkut privasi/konsistensi, murah
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P1-25 | `dashboard.templ` pakai `@Layout` (kini nav di-copy-paste manual, hilang manifest PWA & hx-boost di halaman terpenting) | `web/views/dashboard.templ:15-42` | — |
| `[ ]` | QA-P1-26 🔴 | `sw.js`: network-first untuk route data + hapus cache saat logout — BUG PRIVASI: data keuangan keluarga masih tampil dari cache di HP yang dipinjam orang lain setelah logout | `web/static/sw.js:20-26` | — |

### I. Kepatuhan & operasional — syarat hukum & syarat "beneran bisa diakses publik"
| Status | ID | Task | Lokasi | Depends on |
|---|---|---|---|---|
| `[ ]` | QA-P1-27 🔴 | Halaman Kebijakan Privasi & Syarat Ketentuan (UU PDP No. 27/2022 — app ini nyimpan seluruh data keuangan keluarga orang) | `web/views/` (baru) | — |
| `[ ]` | QA-P1-28 🔴 | Hapus akun mandiri + ekspor data mandiri (hak subjek data, prasyarat UU PDP) | `internal/handlers/` (baru) | QA-P1-01 |
| `[ ]` | QA-P1-29 🔴 | Backup `pg_dump` harian terenkripsi + **uji restore minimal 1×, catat tanggal ujinya di sini** | infra deploy (Coolify scheduled task / cron VPS) | QA-P0-14 |
| `[ ]` | QA-P1-30 🔴 | Dockerfile production + deploy via **Coolify** + domain + TLS — ini yang bikin JurnalUmi BENERAN online | `Dockerfile` (baru, production, bukan `.dev`) | QA-P0-13 |
| `[ ]` | QA-P1-31 | FAQ keamanan data di landing page (bundling murah dengan QA-P1-27) | `web/views/landing.html` | QA-P1-27 |

---

## SETELAH PUBLISH — jangan dikerjakan Hermes sebelum semua di atas `[x]`

Referensi saja (detail lengkap tetap di `task.md`), supaya Prof tahu apa yang SENGAJA ditunda, bukan kelupaan:

**P1 non-kritis:** migrasi ke goose (AutoMigrate cukup untuk soft launch), transfer antar dompet, pagination, quick entry <5 detik, import CSV/Excel, `recurring_rules` + auto-post, `audit_logs` + halaman "Aktivitas Keluarga", debt reminder & budget alert email, ringkasan bulanan email, trial premium 14 hari otomatis, Tailwind CLI build (ganti CDN), self-host Alpine/HTMX, dark mode, offline queue IndexedDB, PWA icon self-host, CI (`go build`/`vet`/`test`/`templ generate --check`).

**P2:** kalkulator zakat maal, rekap tahunan zakat/infaq, grafik pertumbuhan aset, PDF export asli (`maroto`/`gofpdf`), alokasi 50/30/20 saran, reminder piutang WhatsApp, onboarding wizard, push notification harian.

---

## Temuan Baru Selama QA

> Hermes: kalau nemu bug/mock/gap baru yang TIDAK ada di daftar atas selagi kerja, JANGAN diam-diam diperbaiki di luar scope task yang sedang dikerjakan. Catat di sini dengan format di bawah, biar Prof yang putuskan masuk P0/P1-kritis/ditunda.

<!-- Format: ### [tanggal] — ditemukan saat kerjakan QA-XXX
Deskripsi singkat + lokasi file. Usulan prioritas: P0/P1-kritis/ditunda. -->

(belum ada)

---

## Log Eksekusi

> Hermes: append di sini tiap run selesai (baik berhasil, blocked, maupun tidak ada item yang bisa dikerjakan). Jangan overwrite entri lama.

<!-- Format:
### <tanggal jam> WIB — QA-<id>
Status: selesai & terverifikasi | PR dibuka, menunggu merge | blocked | skip (dependency belum merge)
PR: <link>
Ringkasan: apa yang berubah, bagaimana diverifikasi (build/test/manual flow).
Catatan: kendala, keputusan yang diambil, hal yang perlu Prof tahu.
-->

### 2026-09-09 08:15 WIB — QA-P0-01
Status: ditolak Prof (PR ditutup)
PR: https://github.com/cecep-azhar/jurnalumi/pull/1
Ringkasan: Menambahkan auth & role middleware (superadmin) untuk route /admin/*, verifikasi direct hit redirect ke /login.
Catatan: dikerjakan di branch `qa/p0-01-admin-auth`. PR ditutup tanpa merge. Dikembalikan ke [ ] untuk dikerjakan ulang.

### 2026-09-09 09:50 WIB — QA-P0-01 (Retry)
Status: PR dibuka, menunggu merge
PR: https://github.com/cecep-azhar/jurnalumi/pull/4
Ringkasan: Memisahkan route /admin menjadi group, menerapkan middleware pada group. Verifikasi: curl hit ke endpoint dashboard, tenant/upgrade, dan vouchers/generate return 302 redirect.
Catatan: Branch sebelumnya di-reset, patch PR ini meng-cover seluruh route /admin.

### 2026-09-09 08:34 WIB — QA-P0-02
Status: PR dibuka, menunggu merge Prof
PR: https://github.com/cecep-azhar/jurnalumi/pull/2
Ringkasan: SESSION_SECRET dipindah ke env var, `log.Fatal` kalau kosong saat `APP_ENV=production`. Verifikasi: dicoba dengan `APP_ENV=production` tanpa `SESSION_SECRET` set, server refuse start.
Catatan: dikerjakan di branch `qa/p0-02-session-secret`. **Bug ditemukan setelah run ini**: task-qa.md sempat hilang dari `main` (lihat catatan insiden di atas file), jadi run ini sebenarnya update statusnya nyasar ke `task.md`, bukan ke sini — sudah direkonsiliasi manual oleh Claude 9 Sep 2026 sore.

### 2026-09-09 08:45 WIB — QA-P0-03
Status: PR dibuka, menunggu merge Prof
PR: https://github.com/cecep-azhar/jurnalumi/pull/3
Ringkasan: Cookie diset `Secure: true` saat `APP_ENV=production` dan `SameSite=Lax`. Session ID dirotasi saat login/register dengan reset session sebelum set isi data. Verifikasi: `go build` sukses.
Catatan: Menambah import `os` untuk cek env var.

### 2026-09-09 10:20 WIB — QA-P0-05
Status: PR dibuka, menunggu merge
PR: https://github.com/cecep-azhar/jurnalumi/pull/6
Ringkasan: Menerapkan middleware CSRF global dan menyisipkan hidden input csrf_token di seluruh form. `go build ./...` lulus.
Catatan: Menambah dependency go get github.com/labstack/echo/v4/middleware dan update templ.

### 2026-09-09 13:40 WIB — QA-P0-06
Status: PR dibuka, menunggu merge
PR: https://github.com/cecep-azhar/jurnalumi/pull/7
Ringkasan: Rate limiter 5x/detik ditambahkan di endpoint `POST /login` dan `POST /register`.
Catatan: Plafon belum tercapai (PR masih ≤ 3 yang terbuka, setelah run ini mungkin plafon penuh jadi 4 PR terbuka — cek run berikut).
