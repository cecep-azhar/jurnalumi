# Product Requirement Document (PRD) — JurnalUmi
## Platform Manajemen Keuangan Rumah Tangga & Multi-Tenant SaaS Keluarga

- **Versi:** 3.0.0 — *revisi pasca audit teknis 8 Sep 2026 (lihat `review.md`)*
- **Status:** Draft Aktif untuk Implementasi (menggantikan v2.0.0 yang berlabel "Production-Ready" secara prematur)
- **Target Infra:** Dev lokal Podman/Docker (Fedora) → Production VPS Hostinger via **Coolify**
- **Tech Stack:** Go 1.25 | Echo v4 | Templ (SSR type-safe) | HTMX + Alpine.js | Tailwind CSS (CLI build, **bukan CDN**) | PostgreSQL 16 | Air (dev)

> **Aturan dokumen ini:** sebuah fitur baru boleh ditulis `selesai` bila (a) ada kode yang mengeksekusinya, (b) ada route/UI yang memanggilnya, dan (c) ada test atau bukti verifikasi manual. Teks statis di HTML **bukan** fitur.

---

## 1. RINGKASAN EKSEKUTIF & VALUE PROPOSITION

**JurnalUmi** adalah SaaS pencatatan keuangan rumah tangga untuk keluarga muslim Indonesia yang ingin **transparansi finansial antara suami-istri** tanpa harus paham akuntansi.

**Tiga hal yang membuat JurnalUmi berbeda (dan menjadi alasan orang membayar):**
1. **Sync Pasangan + Audit Log** — satu dompet keluarga, dua HP, dan setiap perubahan tercatat siapa pelakunya. Ini janji "transparansi" yang dibuktikan mekanismenya, bukan sekadar klaim.
2. **Modul Syariah yang benar-benar menghitung** — kalkulator **Zakat Maal** otomatis (nisab 85 gr emas, haul 1 tahun, 2,5%) yang menarik data dari aset & saldo yang sudah tercatat. Tidak dimiliki kompetitor (Money Lover, Finansialku, Catatan Keuangan).
3. **Tracker Logam Mulia** — emas/dinar/perak dinilai dengan harga pasar harian ber-snapshot, lengkap dengan grafik pertumbuhan aset.

**Yang menentukan retensi (dan sering dilupakan):** kecepatan mencatat pengeluaran harian (**target < 5 detik**), reminder harian, dan **import CSV/Excel** dari catatan lama.

**Non-goals v1:** integrasi rekening bank otomatis, OCR struk, aplikasi native, multi-currency, akuntansi bisnis/UMKM.

---

## 2. NON-FUNCTIONAL REQUIREMENTS (WAJIB — tidak ada di v2.0.0)

### 2.1 Keamanan (gate rilis, tanpa pengecualian)
| Kode | Requirement | Kriteria Diterima |
|---|---|---|
| SEC-1 | Semua secret dari environment variable | `SESSION_SECRET`, `DATABASE_URL`, `SMTP_*`, `MAYAR_*`. Build gagal/refuse start jika `SESSION_SECRET` kosong di production |
| SEC-2 | Cookie session `HttpOnly` + `Secure` + `SameSite=Lax`, MaxAge 30 hari, di-rotate saat login | Diverifikasi via response header |
| SEC-3 | CSRF token pada seluruh form POST | `middleware.CSRF()` aktif global; form Templ menyertakan hidden field |
| SEC-4 | Rate limit `/login` & `/register` (10 req/menit/IP) + lockout 15 menit setelah 5 gagal | Test integrasi |
| SEC-5 | **RBAC ditegakkan di server**, bukan hanya menyembunyikan menu | Middleware `RequireRole(...)` di setiap route; test per role sesuai `Role_Permission.md` |
| SEC-6 | **PostgreSQL Row-Level Security** aktif di seluruh tabel domain + `SET LOCAL app.tenant_id` per request | Test: query tanpa konteks tenant mengembalikan 0 baris |
| SEC-7 | Route `/admin/*` hanya untuk role `superadmin` (user terpisah, di-seed via CLI) | Test: akses anonim & role `owner` → 403 |
| SEC-8 | Semua error DB ditangani & disampaikan ke user; tidak ada `_ =` pada operasi tulis | Code review + `errcheck` di CI |
| SEC-9 | Security header: HSTS, X-Content-Type-Options, Referrer-Policy, CSP dasar | Dicek via `curl -I` |

### 2.2 Privasi & Kepatuhan (UU PDP No. 27/2022)
- Data keuangan keluarga = **data pribadi spesifik**. Wajib: kebijakan privasi, dasar pemrosesan, dan alamat kontak DPO/PIC.
- **Hapus akun mandiri**: menghapus tenant + seluruh data turunan dalam ≤ 7 hari, dengan konfirmasi ganda.
- **Ekspor data mandiri** (portabilitas): CSV seluruh transaksi + aset + utang kapan saja.
- Enkripsi in-transit (TLS via Caddy/Coolify) & at-rest (disk terenkripsi + backup terenkripsi).
- Retensi log aplikasi 30 hari; log tidak boleh memuat nominal atau email lengkap.

### 2.3 Backup, Restore & Ketersediaan
- `pg_dump` harian otomatis, terenkripsi, retensi 14 harian + 6 bulanan, disimpan off-server.
- **Uji restore wajib minimal 1× sebelum publik launch** dan dicatat tanggalnya di `task.md`.
- Target ketersediaan realistis: 99% bulanan (produk satu operator). Halaman status sederhana + healthcheck `/health`.

### 2.4 Performa
- P95 render halaman < 400 ms pada 5.000 transaksi/tenant.
- Semua listing wajib pagination (default 50 baris).
- **Tidak boleh ada HTTP call ke pihak ketiga di dalam loop request.** Harga logam mulia diambil oleh cron 1×/hari → tabel `price_snapshots`.

### 2.5 Kualitas & Observability
- Test wajib untuk: perhitungan uang, RBAC, isolasi tenant, dan enforcement plan limit. Target: **semua service domain punya unit test** sebelum publik launch.
- Migration berversi (**goose**), `AutoMigrate` dilarang di production.
- Structured logging (request id, tenant id, user id) + error tracking.
- CI: `go build`, `go vet`, `go test`, `templ generate --check`.

---

## 3. ARSITEKTUR

### 3.1 High-Level
```
[ Browser / PWA (network-first untuk data, cache-first untuk aset statis) ]
                         │ HTTPS
                         ▼
              [ Coolify + Caddy (TLS) ]
                         │
        ┌────────────────────────────────────┐
        │      Go (Echo) SSR Application     │
        │  Templ + HTMX + Alpine (no CDN)    │
        │  Middleware: Session → Tenant(RLS) │
        │              → RBAC → CSRF         │
        │  Domain Services (uang = int64)    │
        │  Scheduler (robfig/cron)           │
        └───────┬───────────────┬────────────┘
                │               │
        [PostgreSQL 16]   [SMTP Transaksional]
         (RLS per tenant)  (reminder & alert)
                │
        [Cron harian: harga logam mulia → price_snapshots]
```

### 3.2 Multi-Tenancy (dua lapis, bukan satu)
1. **Lapis aplikasi:** middleware `TenantScope` menyuntikkan `tenant_id` dari session ke konteks; helper query `Scoped(c)` wajib dipakai — akses langsung ke `db.DB` di handler dilarang oleh code review.
2. **Lapis database:** **Row-Level Security** pada semua tabel domain. Bila lapis aplikasi lalai, database tetap menolak. Inilah bukti dari klaim "data satu keluarga tidak bocor ke keluarga lain".

### 3.3 Prinsip Uang & Ledger
- **Semua nominal `int64` dalam rupiah penuh** (tanpa sen). `float64` dilarang untuk uang.
- **Saldo wallet = turunan, bukan kebenaran.** Setiap perubahan saldo wajib punya baris transaksi (termasuk saldo awal → transaksi `opening_balance`). Kolom `balance` boleh di-cache tetapi wajib di-update dalam DB transaction yang sama dan direkonsiliasi oleh job harian.
- **Transfer antar dompet** dicatat sebagai satu transaksi bertipe `transfer` dengan `wallet_id` (sumber) dan `to_wallet_id` (tujuan).
- Transaksi **dapat diedit & dihapus** (soft delete), setiap perubahan menulis `audit_logs`.

---

## 4. USER ROLES & RBAC

Matrix hak akses lengkap: lihat `Role_Permission.md` (sumber kebenaran untuk implementasi middleware).

| Role | Ringkasan |
|---|---|
| `superadmin` | Platform owner. Kelola tenant, billing, voucher, variabel global. **Tidak bisa melihat isi transaksi tenant** (hanya metadata & agregat) |
| `owner` | Kepala keluarga. Full access + kelola anggota + setel budget/target |
| `spouse` | Pasangan. Full access transaksi, wallet, aset, dana darurat. Utang/piutang: **input & lihat**, hapus hanya oleh `owner` |
| `member` | Anak/tanggungan. Hanya catat pengeluaran pada wallet yang ditugaskan. Tidak melihat utang & aset |
| `auditor` | Read-only untuk perencana keuangan. Tidak bisa menulis apa pun |

Aturan: **role dicek di server pada setiap handler.** Menyembunyikan menu di UI bukan kontrol akses.

---

## 5. MODUL DOMAIN

### 5.1 Pencatatan Cepat (Quick Entry) — *prioritas retensi tertinggi*
- Tombol tambah selalu terjangkau (FAB mobile). Target dari buka app → tersimpan: **< 5 detik, ≤ 4 tap**.
- Default cerdas: tanggal = hari ini, wallet & kategori = yang terakhir dipakai.
- Input nominal dengan pemisah ribuan otomatis (Alpine) + shortcut `5rb / 25rb / 100rb`.
- **Import CSV/Excel** dari catatan lama (mapping kolom + preview + rollback batch). Ini penghilang friksi terbesar untuk pengguna yang pindah dari Excel.

### 5.2 Pemasukan
- Sumber: gaji, bonus/THR, freelance, hasil usaha, bagi hasil, dividen, sewa.
- **Recurring rules** (tabel tersendiri): frekuensi bulanan/mingguan, tanggal eksekusi, auto-post oleh cron + notifikasi "sudah dicatat".
- **Alokasi 50/30/20** sebagai *saran* saat pemasukan besar masuk (bukan pemotongan otomatis): tampilkan usulan pengisian pos Needs/Wants/Savings, user menyetujui satu klik.

### 5.3 Pengeluaran & Budget
- Kategori 2 level (parent → sub), dengan seed default Indonesia (Dapur, Listrik & Air, SPP, BPJS, Transport, Zakat/Infaq, Sosial, Hiburan, Langganan Digital).
- **Budget per periode**: tabel `budgets(tenant_id, category_id, period_month, amount_idr)` — bukan satu angka statis di kategori. Ini prasyarat laporan "budget vs aktual bulan lalu".
- Indikator realisasi: hijau < 70%, kuning 70-90%, merah > 90% — **dihitung di server** dari `SUM(transactions)` per periode, bukan di UI saja.
- Alert email pada 80% dan 100% (maksimal 1 email per kategori per periode per ambang, agar tidak spam).

### 5.4 Utang & Piutang
- Parameter: sisa pokok, margin/bunga (%), tenor, jatuh tempo bulanan, counterparty.
- **Pembayaran nyata**: form bayar → pilih wallet → insert transaksi `expense` + kurangi saldo + kurangi sisa pokok, **satu DB transaction**, tercatat di `debt_payments`. (Perilaku "bagi dua otomatis" pada v2 adalah bug dan dihapus.)
- **Kalkulator Snowball & Avalanche yang benar-benar menghitung**: input dana ekstra bulanan → output urutan pelunasan, estimasi bulan bebas utang, total bunga, dan perbandingan kedua metode. Wajib ada unit test dengan skenario tetap.
- Piutang: histori cicilan + tombol *Reminder* (WhatsApp deep link `wa.me` + email).

### 5.5 Aset & Net Worth
- Aset likuid (kas, bank, e-wallet), investasi (reksadana, saham, deposito, SBN), properti/kendaraan.
- **Logam mulia:** emas batangan (gram, kadar), dinar (keping → gram, 4,25 gr 22K per dinar), perak/dirham.
- **Valuasi:** cron harian mengambil harga → `price_snapshots(commodity, price_idr, source, fetched_at)`. Halaman aset **hanya membaca snapshot terakhir** (tidak pernah memanggil API saat request).
  - Sumber harga wajib nyata & tercatat sumbernya di UI ("Harga Antam per 8 Sep 2026"). Selama sumber belum tersedia, tampilkan **harga manual yang diinput user** dan **jangan mengiklankan valuasi real-time**.
- **Net Worth = Total Aset − Total Utang**, tampil di dashboard dengan tren bulanan. (Kartu "Coming Soon" pada v2 dihapus.)
- Grafik pertumbuhan aset logam mulia dari histori snapshot.

### 5.6 Proteksi: Dana Darurat & Sinking Funds
- **Emergency Fund Health Score**: target = rata-rata pengeluaran 3 bulan terakhir × faktor (single 6×, menikah 9×, menikah+anak 12×). Status: Danger < 3 bln, Warning 3-6 bln, Safe > 6 bln.
- **Sinking Funds**: pos dana khusus (Kurban, Pajak Kendaraan, Mudik, Masuk Sekolah, Liburan) dengan target nominal + tanggal target → sistem menghitung **setoran bulanan yang dibutuhkan** dan progress bar.

### 5.7 Modul Syariah (diferensiator)
- **Kalkulator Zakat Maal**: agregasi kas + tabungan + emas/perak + investasi − utang jatuh tempo; bandingkan dengan nisab (85 gr emas, memakai harga snapshot); cek haul; hitung 2,5%; hasilkan pengingat tahunan.
- Pencatatan Zakat Fitrah, Infaq/Sedekah dengan rekap tahunan siap cetak.
- Penanda kategori halal/syariah pada laporan tahunan.

### 5.8 Laporan & Ekspor
- Filter periode + kategori + wallet + pencarian, dengan pagination server-side.
- **PDF asli** (`maroto`/`gofpdf`) untuk ringkasan bulanan — bukan `window.print()`.
- CSV lengkap (portabilitas data, sekaligus memenuhi NFR privasi).
- **Ringkasan bulanan otomatis via email setiap tanggal 1.**

### 5.9 Audit Log & Transparansi Pasangan
- Setiap create/update/delete pada transaksi, wallet, utang, aset, dan anggota → baris `audit_logs`.
- Halaman "Aktivitas Keluarga": siapa, kapan, mengubah apa, dari nilai berapa ke berapa. **Ini implementasi konkret dari janji pemasaran "transparansi suami-istri".**

---

## 6. PWA & NOTIFIKASI

### 6.1 PWA (koreksi dari v2)
- `sw.js` **network-first untuk seluruh route data** (`/dashboard`, `/reports`, `/assets`, `/debts`), cache-first hanya untuk aset statis ber-hash.
- **Dilarang mem-precache halaman ber-autentikasi.** Cache milik user wajib dibersihkan saat logout (`caches.delete`) — mencegah data keuangan tersaji di HP yang dipinjam.
- Offline: form quick entry disimpan di **IndexedDB** dan disinkronkan saat online, dengan indikator "menunggu sinkronisasi".
- Icon PWA di-host sendiri (`/static/icons/`), ukuran 192 & 512 terpisah + maskable.

### 6.2 Email Transaksional
- Provider: SMTP relay / Resend; Mailhog untuk lokal. Pengiriman **asinkron via worker queue**, dengan retry.
- Trigger: (1) reminder utang H-3 & H-1, (2) budget alert 80% & 100%, (3) ringkasan bulanan tanggal 1, (4) dana darurat terpakai, (5) verifikasi email & reset password.
- Semua email punya tautan berhenti berlangganan untuk notifikasi non-transaksional.
- Scheduler: `robfig/cron` di dalam proses, dengan **lock berbasis DB** agar aman bila ada 2 instance.

---

## 7. SAAS: PAKET, LIMIT & BILLING

### 7.1 Paket (revisi — batasan diubah dimensinya)
| | **Free** | **Premium Household** |
|---|---|---|
| Harga | Rp 0 | **Rp 190.000/tahun (promo launching)** — normal Rp 390.000/tahun |
| User | 1 (kepala keluarga) | **Hingga 5 anggota keluarga (sync suami-istri)** |
| Transaksi | **Unlimited** | Unlimited |
| Dompet | 2 | Unlimited |
| Histori | **3 bulan terakhir** | Penuh + laporan tahunan |
| Aset & logam mulia | – | ✔ + grafik pertumbuhan |
| Utang & Snowball/Avalanche | Catat saja | ✔ + kalkulator strategi |
| Dana darurat & sinking funds | – | ✔ |
| Zakat maal | – | ✔ |
| Email reminder & ringkasan bulanan | – | ✔ |
| Ekspor CSV | ✔ (hak portabilitas data) | ✔ + PDF |
| Audit log keluarga | – | ✔ |

**Alasan perubahan:** membatasi jumlah transaksi menghukum user paling rajin — persis orang yang paling mungkin membayar. Batasi pada **jumlah user + kedalaman histori**; itulah nilai yang benar-benar dibeli.

**Trial:** setiap pendaftar mendapat **Premium 14 hari otomatis**, agar sempat merasakan sync pasangan sebelum turun ke Free.

**Enforcement (wajib, saat ini belum ada sama sekali):** middleware `RequirePlan(feature)` + pengecekan `plan_expires_at` di setiap request; cron harian menurunkan tenant kedaluwarsa ke `free` (data tidak dihapus, hanya dibatasi aksesnya).

### 7.2 Pembayaran
- **Mayar.id**: payment link + **webhook `payment.success`** (verifikasi signature, idempotent, dicatat di tabel `payments`) → set `plan='premium'`, `plan_expires_at = now + 1 tahun`.
- **Voucher**: `superadmin` generate kode; route **`POST /activate-voucher`** (pada v2 form-nya ada tapi route-nya 404 — wajib dibuat), redeem idempotent + dicatat pemakainya.
- Semua harga di kode, landing page, dan materi promosi **wajib bersumber dari satu konstanta/env** agar tidak lagi berbeda-beda antar dokumen.

### 7.3 Landing Page
Hero (hook emosional) → bukti sosial → showcase 3 diferensiator (sync+audit, zakat, logam mulia) → demo video YouTube (`aspect-video`, `loading="lazy"`) → matriks harga (transparan, sesuai §7.1) → FAQ keamanan data (**wajib**: di mana data disimpan, siapa yang bisa lihat, cara hapus akun) → CTA daftar/voucher.

---

## 8. SKEMA DATABASE (PostgreSQL 16) — revisi

> Perubahan utama vs v2: uang `BIGINT` (rupiah penuh), `category_id` benar-benar dipakai, `transfer` didukung, plus tabel `budgets`, `recurring_rules`, `debt_payments`, `price_snapshots`, `audit_logs`, `payments`, `vouchers`. Semua tabel domain diaktifkan RLS.

```sql
-- Tenant (Keluarga)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',      -- free, premium
    plan_expires_at TIMESTAMPTZ,
    household_status VARCHAR(30) DEFAULT 'married', -- single, married, married_kids (faktor dana darurat)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,  -- NULL untuk superadmin
    name VARCHAR(255) NOT NULL,
    email CITEXT UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',    -- superadmin, owner, spouse, member, auditor
    email_verified_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,                     -- cash, bank, ewallet, investment, sinking, emergency
    balance_idr BIGINT NOT NULL DEFAULT 0,         -- cache; kebenaran = SUM(transactions)
    target_idr BIGINT NOT NULL DEFAULT 0,          -- sinking / emergency fund
    target_date DATE,
    assigned_user_id UUID REFERENCES users(id),    -- untuk role member (uang saku)
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,   -- hierarki 2 level
    type VARCHAR(20) NOT NULL,                     -- income, expense
    name VARCHAR(255) NOT NULL,
    color VARCHAR(30) DEFAULT 'gray',
    is_zakat BOOLEAN DEFAULT FALSE,                -- penanda modul syariah
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Budget per periode (menggantikan budget_limit statis di categories)
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    period_month DATE NOT NULL,                    -- selalu tanggal 1
    amount_idr BIGINT NOT NULL,
    alerted_80_at TIMESTAMPTZ,
    alerted_100_at TIMESTAMPTZ,
    UNIQUE (tenant_id, category_id, period_month)
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    to_wallet_id UUID REFERENCES wallets(id),      -- hanya untuk type='transfer'
    category_id UUID REFERENCES categories(id),
    debt_id UUID REFERENCES debts(id),             -- bila transaksi ini pembayaran utang
    type VARCHAR(20) NOT NULL,                     -- income, expense, transfer, opening_balance
    amount_idr BIGINT NOT NULL CHECK (amount_idr > 0),
    description TEXT,
    transaction_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX ON transactions (tenant_id, transaction_date DESC);
CREATE INDEX ON transactions (tenant_id, category_id, transaction_date);

CREATE TABLE recurring_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    category_id UUID REFERENCES categories(id),
    type VARCHAR(20) NOT NULL,                     -- income, expense
    amount_idr BIGINT NOT NULL,
    description TEXT,
    frequency VARCHAR(20) NOT NULL DEFAULT 'monthly',
    day_of_month SMALLINT,
    next_run_date DATE NOT NULL,
    active BOOLEAN DEFAULT TRUE
);

CREATE TABLE debts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,                     -- debt, receivable
    title VARCHAR(255) NOT NULL,
    counterparty VARCHAR(255) NOT NULL,
    counterparty_phone VARCHAR(30),                -- untuk reminder wa.me
    total_idr BIGINT NOT NULL,
    remaining_idr BIGINT NOT NULL,
    interest_rate NUMERIC(5,2) DEFAULT 0,
    tenor_months SMALLINT,
    due_day SMALLINT,                              -- tanggal jatuh tempo bulanan
    due_date DATE,
    status VARCHAR(20) DEFAULT 'active',           -- active, paid, defaulted
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE debt_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    debt_id UUID NOT NULL REFERENCES debts(id) ON DELETE CASCADE,
    transaction_id UUID REFERENCES transactions(id),
    amount_idr BIGINT NOT NULL,
    paid_at DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE assets (                              -- aset non-komoditas
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,                 -- investment, property, vehicle, other
    name VARCHAR(255) NOT NULL,
    acquired_value_idr BIGINT NOT NULL,
    current_value_idr BIGINT NOT NULL,
    valued_at DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE commodity_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL,                     -- gold_bar, dinar, dirham, silver
    name VARCHAR(255) NOT NULL,
    pieces NUMERIC(10,2) DEFAULT 1,                -- jumlah keping (dinar/dirham)
    weight_gram NUMERIC(10,4) NOT NULL,            -- total gram
    karatage NUMERIC(5,2) DEFAULT 24.00,
    buy_price_idr BIGINT NOT NULL,
    bought_at DATE,
    manual_price_idr BIGINT,                       -- dipakai bila snapshot harga tidak tersedia
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE price_snapshots (                     -- global, bukan per tenant
    id BIGSERIAL PRIMARY KEY,
    commodity VARCHAR(30) NOT NULL,                -- gold_24k, gold_22k, silver
    price_per_gram_idr BIGINT NOT NULL,
    source VARCHAR(100) NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (commodity, fetched_at)
);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    entity VARCHAR(50) NOT NULL,
    entity_id UUID,
    action VARCHAR(20) NOT NULL,                   -- create, update, delete
    changes JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX ON audit_logs (tenant_id, created_at DESC);

CREATE TABLE vouchers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) UNIQUE NOT NULL,
    duration_days INT NOT NULL DEFAULT 365,
    used_by_tenant_id UUID REFERENCES tenants(id),
    used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    provider VARCHAR(30) NOT NULL DEFAULT 'mayar',
    external_id VARCHAR(255) UNIQUE NOT NULL,      -- idempotency webhook
    amount_idr BIGINT NOT NULL,
    status VARCHAR(30) NOT NULL,
    raw_payload JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Row-Level Security (contoh; diterapkan ke seluruh tabel ber-tenant_id)
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON transactions
    USING (tenant_id = current_setting('app.tenant_id', TRUE)::uuid);
```

---

## 9. DEFINITION OF DONE (berlaku untuk setiap fitur)

1. Route terdaftar + handler menangani error + RBAC & plan limit ditegakkan.
2. Query domain melalui scope tenant; RLS aktif untuk tabelnya.
3. Ada unit test untuk logika perhitungan, dan test RBAC bila menyentuh data sensitif.
4. UI punya state kosong, state error, dan konfirmasi untuk aksi destruktif.
5. Perubahan data menulis `audit_logs`.
6. `task.md` diperbarui **setelah** diverifikasi berjalan (bukan setelah kode ditulis).

---

*PRD v3.0.0 — sumber kebenaran teknis JurnalUmi. Perubahan lingkup wajib melalui revisi dokumen ini, bukan lewat catatan implementasi.*
