# Technical & Product Audit — JurnalUmi

- **Tanggal Audit:** 8 September 2026
- **Basis Kode:** commit `9913a35` (branch `main`), 2.830 baris (Go + Templ + HTML)
- **Metode:** pembacaan penuh seluruh file `cmd/`, `internal/`, `web/`, `Note/` + verifikasi grep per klaim
- **Verdict:** **±25-30% jalan** menuju produk yang bisa dijual — bukan 100% seperti tertulis di `task.md` lama.
  Fondasi (stack, struktur folder, model dasar, UI) sudah benar dan layak dilanjutkan.
  Yang hilang: **seluruh lapisan keamanan** dan **hampir semua logika bisnis**.

---

## 0. RINGKASAN EKSEKUTIF

| Dimensi | Status | Catatan |
|---|---|---|
| Struktur & stack | 🟢 Baik | Go + Echo + Templ + GORM + Postgres, folder layout standar, `go build` bersih |
| UI / tampilan | 🟡 Cukup | Rapi & konsisten, tapi CDN Tailwind, tanpa dark mode, dashboard tidak pakai `Layout` |
| Keamanan | 🔴 Kritis | Admin tanpa auth, session secret ter-commit, RBAC nol, tanpa CSRF |
| Logika bisnis | 🔴 Kritis | Snowball, budget, net worth, emergency fund, plan limit = teks/placeholder |
| Integritas data | 🔴 Kritis | Tombol bayar utang merusak data, tanpa edit/hapus transaksi, saldo tidak direkonsiliasi |
| Monetisasi | 🔴 Belum ada | Free tier tanpa enforcement, voucher 404, webhook Mayar tidak ada |
| Kualitas proses | 🔴 Buruk | 0 test, tanpa migration versioning, checklist `task.md` tidak mencerminkan realita |

---

## 1. BLOCKER KEAMANAN (P0 — jangan sampai kena internet publik)

| # | Temuan | Lokasi | Dampak |
|---|---|---|---|
| S1 | **Route `/admin/*` tanpa autentikasi apa pun** | `cmd/server/main.go:57-59` | Siapa pun buka `/admin/dashboard`: lihat seluruh tenant, upgrade diri jadi premium, generate voucher |
| S2 | **Session secret hardcoded & ter-commit ke git** (`jurnalumi-super-secret-key`) | `cmd/server/main.go:32` | Siapa pun yang melihat repo bisa memalsukan cookie jadi user/tenant mana pun → bypass total multi-tenancy |
| S3 | Cookie session tanpa `Secure` & `SameSite` | `internal/handlers/auth.go:37-41` | Session hijack via HTTP, rentan CSRF lintas situs |
| S4 | **RBAC nol persen terimplementasi** | seluruh handler | `Role_Permission.md` mendefinisikan 5 role × 15 baris hak akses; kode hanya punya `RequireAuth`. Role `member` (anak) bisa `POST /family` membuat user baru ber-role `owner` |
| S5 | Tidak ada CSRF token | semua form POST | Satu link jahat = transaksi/anggota keluarga baru di akun korban |
| S6 | Tidak ada rate limit di `/login` | `cmd/server/main.go:49` | Brute force bebas, tanpa lockout |
| S7 | `middleware.CORS()` wildcard + cookie session | `cmd/server/main.go:39` | Tidak dibutuhkan (SSR), memperluas attack surface |
| S8 | Error `db.Create()` diabaikan di 6 handler | `dashboard.go:132`, `dashboard.go:154`, `features.go:114`, `assets.go:69`, `debts.go:77`, `admin.go:66` | Gagal simpan → user tetap di-redirect seolah sukses. Fatal untuk aplikasi keuangan |
| S9 | Multi-tenancy hanya bergantung pada disiplin developer | seluruh query | PRD menjanjikan middleware auto-inject `tenant_id`; realitanya `WHERE tenant_id = ?` diketik manual. Satu handler lupa = data keluarga lain bocor |

**Rekomendasi kunci S9:** aktifkan **PostgreSQL Row-Level Security** + `SET LOCAL app.tenant_id` per request. Sekali pasang, murah, dan tidak bergantung pada developer ingat mengetik filter.

---

## 2. FITUR YANG DIKLAIM SELESAI TAPI TIDAK ADA DI KODE

Seluruh baris `task.md` lama bertanda `[x]`. Hasil verifikasi grep:

| Klaim task.md lama | Realita di kode |
|---|---|
| Dark/Light mode toggle | **0 occurrence** `dark:` di seluruh `.templ`/`.html`. Tidak ada |
| `docker-compose.yml` PostgreSQL | File tidak ada |
| Middleware RBAC | Tidak ada |
| CRUD Transactions (Income, Expense, **Transfer**) | Hanya Create. **Tidak ada edit/hapus.** Type `transfer` tidak diimplementasi |
| Budget capping check | `Category.BudgetLimit` tersimpan tapi **tidak pernah dibandingkan** dengan realisasi. Tidak ada indikator hijau/kuning/merah |
| Net Worth Dashboard (Assets − Debts) | Kartu dashboard literal: `"Sisa Utang (Coming Soon) — Rp 0.00"` (`dashboard.templ:64-67`). Aset & utang tidak pernah dijumlahkan |
| Debt Snowball/Avalanche calculator | Teks statis `"Fokus Pelunasan Terkecil First"` (`debts.templ:37-39`). **Nol baris perhitungan** |
| Sinking Funds & Emergency Fund tracker | Hanya kolom `target_amount` di form wallet. Tanpa progress bar, tanpa rasio 6x/9x/12x, tanpa health score |
| Cron worker Debt Reminder & Budget Alert | Tidak ada. `services.SendEmailNotification()` **tidak dipanggil dari mana pun** — dead code |
| Billing Mayar.id webhook | Tidak ada. Landing hanya `<a href>` statis ke link Mayar hardcoded (`landing.html:212`) |
| Voucher redemption | `landing.html:247` POST ke `/activate-voucher` → **route tidak terdaftar = 404**. Model `Voucher` ada, redeem-nya tidak |
| PDF Export engine | Hanya `window.print()` + CSV |
| Production Dockerfile + Caddy | Tidak ada. Hanya `Dockerfile.dev` |
| Free tier limit (1 dompet, 50 trx/bln, 1 user) | **Nol enforcement.** Semua user mendapat fitur premium gratis → model bisnis tidak eksis di kode |
| Plan expiry | `PlanExpiresAt` di-set saat upgrade, **tidak pernah dicek** → premium berlaku selamanya |

> Tindakan pertama bukan soal kode: **checklist yang tidak jujur membuat estimasi jarak ke production salah total.** `task.md` sudah ditulis ulang sesuai status riil.

---

## 3. BUG FUNGSIONAL

### 3.1 Merusak data user
- **`DebtPayPOST` membagi dua sisa utang** (`internal/handlers/debts.go:88-91`, komentarnya sendiri menulis *"mock implementation"*). Jika user menekan tombol Bayar, catatan utangnya rusak permanen. **Fitur yang aktif merusak lebih buruk daripada fitur yang belum ada — matikan tombolnya lebih dulu.**
- **Tidak ada edit/hapus transaksi.** Salah ketik nominal = salah selamanya. Ini penyebab nomor satu orang berhenti memakai aplikasi keuangan.
- **Saldo wallet denormalisasi tanpa rekonsiliasi.** `WalletPOST` (`dashboard.go:116-134`) bisa menetapkan balance sembarang tanpa jejak transaksi. Saldo dan histori akan divergen permanen.

### 3.2 Membuat klaim jualan tidak benar
- `https://api.logammulia.com/v1/price` (`internal/services/gold.go:17`) **bukan API yang eksis.** Praktis selalu jatuh ke fallback konstanta `1.450.000`. Artinya "Auto-Valuation Engine real-time" — Angle B, salah satu dari 3 hook iklan utama — sebenarnya angka mati. **Jangan diiklankan sebelum ada sumber harga nyata.**
- **`CalculateCommodityValue` memanggil `FetchLiveGoldPrice()` untuk setiap aset** (`internal/handlers/assets.go:31-35`). 20 aset + API mati = 20 × timeout 3 detik ≈ halaman `/assets` hang 60 detik. Harus: fetch sekali/hari, simpan di tabel `price_snapshots`.
- Perak hardcoded `16.500/gram`; dinar mengabaikan parameter `karatage`; UI menyebut "jumlah keping" tapi model meminta gram.

### 3.3 Kualitas, akurasi, skala
| # | Temuan | Lokasi |
|---|---|---|
| F1 | `FormatRupiah` menghasilkan `Rp 500000.00`, seharusnya `Rp 500.000` | `web/views/dashboard.templ:9-11` |
| F2 | Label **"Pemasukan/Pengeluaran Bulan Ini"** menjumlah *seluruh* transaksi sepanjang masa | `internal/handlers/dashboard.go:44-51` |
| F3 | `Find()` tanpa `LIMIT`/pagination di dashboard, reports, dan export CSV | `dashboard.go:36`, `features.go:33`, `features.go:52` |
| F4 | `Transaction.CategoryID` **tidak pernah diisi** — hanya `CategoryName` string. Akibatnya laporan per-kategori & budget vs aktual mustahil di-query benar. Ini akar kenapa budget capping tidak jalan | `internal/handlers/dashboard.go:87-96` |
| F5 | `float64` untuk uang. Rupiah tidak memakai sen → gunakan `int64` | `internal/models/models.go` (semua field nominal) |
| F6 | **`sw.js` cache-first untuk semua request termasuk `/dashboard`** — PRD sendiri meminta network-first untuk data finansial. Konsekuensi: setelah logout, data keuangan keluarga masih tersaji dari cache di HP yang dipinjam orang lain. **Bug privasi, bukan sekadar bug teknis** | `web/static/sw.js:20-26` |
| F7 | `sw.js` mem-precache `/dashboard` saat install padahal user belum login → yang tersimpan hanya redirect ke `/login` | `web/static/sw.js:2-9` |
| F8 | Offline queue IndexedDB (klaim PRD §6.2) tidak ada | — |
| F9 | Tailwind via `cdn.tailwindcss.com` — resmi dilarang untuk production; ini pula sebabnya dark mode tidak bisa dikonfigurasi | `layout.templ:16`, `dashboard.templ:21` |
| F10 | `manifest.json` icons menunjuk CDN Flaticon, satu file untuk 192 & 512 → gagal kriteria install PWA + isu lisensi | `web/static/manifest.json:9-19` |
| F11 | **`dashboard.templ` tidak memakai `Layout`** — nav di-copy-paste, sehingga halaman terpenting tidak punya manifest PWA, service worker, maupun `hx-boost` | `web/views/dashboard.templ:15-42` |
| F12 | `e.Static("/static", ...)` didaftarkan dua kali | `main.go:42` & `main.go:83` |
| F13 | **0 file test** di seluruh repo | — |
| F14 | `AutoMigrate` sebagai mekanisme migrasi production, tanpa versioning/rollback | `internal/db/db.go:29-41` |
| F15 | Harga tidak konsisten di 3 dokumen: PRD `39rb/bln – 390rb/thn`, STRATEGI_PROMOSI `190rb/thn`, link Mayar `jurnalumi-premium-39k` | lintas dokumen |
| F16 | `Dockerfile.dev` memakai `golang:1.23-alpine` sementara `go.mod` menyatakan `go 1.25.0` → build container gagal | `Dockerfile.dev:1` |
| F17 | Email `uniqueIndex` global (bukan per tenant) + error create diabaikan → penambahan anggota gagal senyap | `models.go:31`, `features.go:114` |

---

## 4. KRITIK LEVEL PRD & PRODUK

PRD lama rapi sebagai dokumen, tetapi berlabel **"Production-Ready"** tanpa satu pun *non-functional requirement*. Untuk aplikasi yang meminta orang menaruh **seluruh** keuangan rumah tangganya, ini gap terbesar:

1. **Tidak ada spesifikasi backup, restore, retensi, dan hapus akun.** Data keuangan keluarga adalah data pribadi spesifik di bawah UU PDP No. 27/2022. Tidak ada enkripsi at-rest, tidak ada rencana respons insiden. Justru inilah yang membuat orang mau membayar.
2. **Tidak ada audit log.** Produk dijual dengan janji "transparansi suami-istri", tapi tidak ada catatan siapa mengubah/menghapus apa. Fitur pembeda yang paling selaras dengan positioning justru absen.
3. **Tidak ada konsep periode / closing bulanan.** Ini akar masalah desain data: budget, laporan, rasio dana darurat semuanya butuh periode. `Category.BudgetLimit` satu nilai tanpa periode = tidak bisa histori, tidak bisa "budget vs aktual bulan lalu".
4. **Recurring transaction tidak punya model** padahal disebut di PRD §4.1. Gaji, SPP, cicilan, langganan — justru fitur yang paling menghemat waktu harian user.
5. **PRD sibuk di emas/dinar/snowball, mengabaikan retensi.** Aplikasi keuangan mati di friksi pencatatan harian. Penentu user bertahan: catat pengeluaran <5 detik, reminder harian, dan **import CSV/Excel** — target market pindah dari catatan Excel, dan jalur masuknya tidak ada sama sekali.
6. **Zakat hanya jadi nama kategori.** Padahal ini diferensiator syariah yang nyata dan tidak dimiliki Money Lover / Finansialku: kalkulator zakat maal otomatis (nisab 85gr emas, haul 1 tahun, 2,5%) yang menarik langsung dari aset emas + saldo yang sudah tercatat. Datanya sudah ada di sistem, tinggal dihitung. **Fitur berbayar paling meyakinkan yang bisa dimiliki JurnalUmi.**
7. **Multi-tenancy tanpa RLS.** Lihat S9.
8. **Free tier salah dimensi.** Membatasi 50 transaksi/bulan menghukum user paling rajin — persis user yang paling mungkin membayar. Balik logikanya: **transaksi unlimited, batasi di 1 user + histori 3 bulan**. Orang membayar untuk sync pasangan dan histori, bukan untuk jatah baris. Tambahkan trial premium 14 hari otomatis saat daftar agar mereka merasakan sync suami-istri lebih dulu.
9. **Kompetitor & switching cost tidak dibahas sama sekali** di PRD lama.
10. Klaim "data satu keluarga tidak akan pernah bocor ke keluarga lain (100%)" tanpa mekanisme yang membuktikannya (tanpa RLS, tanpa test, tanpa audit).

---

## 5. PRIORITAS PERBAIKAN

### P0 — sebelum boleh dipakai orang selain founder (±3 hari)
1. Proteksi `/admin/*` + role `superadmin`; session secret dari env; `Secure` + `SameSite`; CSRF; rate limit login
2. Middleware `RequireRole()` + terapkan matrix `Role_Permission.md`; aktifkan RLS PostgreSQL
3. Matikan `DebtPayPOST` mock → ganti pembayaran nyata (insert transaksi + potong wallet + kurangi sisa, satu DB transaction)
4. Semua `db.Create/Save` cek error & tampilkan ke user
5. Perbaiki `Dockerfile.dev` (Go 1.25) + tambah `docker-compose.yml`

### P1 — sebelum boleh dijual (±3-4 minggu)
6. Refactor data: `int64` untuk uang, `CategoryID` benar-benar dipakai, tabel `budgets(tenant, category, period)`, `recurring_rules`, `audit_logs`, `price_snapshots`; migration tool (goose/atlas) menggantikan AutoMigrate
7. Edit & hapus transaksi + pagination + `FormatRupiah` format Indonesia + filter periode di dashboard
8. Enforcement plan limit + expiry + `/activate-voucher` (kini 404) + webhook Mayar.id
9. Snowball/Avalanche benar-benar dihitung; net worth benar-benar dijumlah; emergency fund health score
10. Cron worker (`robfig/cron`) → debt reminder H-3/H-1 + budget alert; sambungkan `SendEmailNotification` yang selama ini menganggur
11. Tailwind CLI build + self-host Alpine/HTMX + dark mode + dashboard memakai `Layout` + `sw.js` network-first untuk data
12. `docker-compose`, Dockerfile production, deploy via Coolify, backup harian otomatis

### P2 — diferensiasi & retensi
13. Import CSV/Excel, kalkulator zakat maal, sumber harga emas nyata + grafik pertumbuhan aset, quick-add <5 detik, PDF asli (`maroto`/`gofpdf`), audit log UI ("aktivitas pasangan")

---

## 6. HAL YANG SUDAH BENAR (jangan dibongkar)

- Pemilihan stack: Go + Echo + Templ + Alpine + HTMX + Postgres — tepat untuk beban dan tim satu orang.
- Struktur folder `cmd/ internal/{db,handlers,models,middleware,services} web/{views,static}` sudah standar dan konsisten.
- Pola `Base` model dengan UUID + soft delete sudah benar.
- `RegisterPOST` sudah memakai DB transaction dan menyemai kategori + wallet default — pola yang tepat untuk onboarding.
- Password sudah bcrypt (bukan MD5/SHA).
- `Render()` wrapper Templ, penggunaan `@Layout` di halaman assets/debts/reports/family sudah rapi.
- Desain visual landing page & dashboard sudah di atas rata-rata produk lokal sejenis.
