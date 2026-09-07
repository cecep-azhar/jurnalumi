# Timeline & Roadmap JurnalUmi

- **Disusun:** 8 September 2026 (menggantikan tabel Fase 1-7 di PRD v2 yang seluruhnya ditandai selesai secara keliru)
- **Basis:** `review.md` — kondisi riil ±25-30% menuju produk yang bisa dijual
- **Kapasitas asumsi:** 1 developer paruh waktu (Prof punya pekerjaan tetap di Gerlink) ≈ **12-15 jam/minggu efektif**
- **Prinsip:** keamanan dulu → fondasi data → fitur yang dijanjikan → monetisasi → baru promosi besar

> Angka durasi di bawah sudah memperhitungkan paruh waktu. Bila ada minggu yang terpakai untuk pekerjaan kantor/klien, geser seluruh baris — jangan memotong lingkup P0.

---

## Ringkasan Fase

| Sprint | Periode | Fokus | Gate keluar |
|---|---|---|---|
| **S0** | 8 – 18 Sep 2026 | 🔴 P0: blocker keamanan & integritas data | Aman dipakai keluarga sendiri |
| **S1** | 21 Sep – 2 Okt 2026 | Fondasi data (int64, RLS, goose, tabel baru) | Semua test fondasi hijau |
| **S2** | 5 – 16 Okt 2026 | Ledger & UX inti (edit/hapus, transfer, quick entry, import) | Bisa dipakai harian tanpa frustrasi |
| **S3** | 19 – 30 Okt 2026 | Fitur yang selama ini cuma teks (budget, net worth, snowball, dana darurat) | Klaim di landing page jadi benar |
| **S4** | 2 – 13 Nov 2026 | Notifikasi, aset & harga logam mulia, PWA & dark mode | Reminder & valuasi berjalan nyata |
| **S5** | 16 – 27 Nov 2026 | Monetisasi + kepatuhan + deploy production | Bisa menerima uang |
| **Beta** | 30 Nov – 18 Des 2026 | Closed beta 10-15 keluarga | Retensi minggu-4 ≥ 40% |
| **Launch** | 5 Januari 2027 | Publik launch + promo Rp 190rb/tahun | — |
| **S6+** | Jan – Feb 2027 | P2: zakat maal, PDF, grafik aset, recurring | Diferensiasi berbayar |

---

## Detail Sprint

### S0 · 8 – 18 Sep 2026 — Blocker (WAJIB, tidak boleh dilewati)
- Proteksi `/admin/*` + role `superadmin` + seeder CLI
- `SESSION_SECRET` ke env, cookie `Secure`/`SameSite`, rotasi session
- Middleware `RequireRole()` + terapkan matrix `Role_Permission.md`
- CSRF global + rate limit login
- Ganti `DebtPayPOST` mock dengan pembayaran nyata
- Tangani semua error operasi tulis + validasi input server-side
- Perbaiki `Dockerfile.dev` (Go 1.25) + `docker-compose.yml` + `.env.example`

**Gate:** aplikasi boleh dipakai keluarga sendiri sehari-hari. Belum boleh dibagikan ke orang lain.

### S1 · 21 Sep – 2 Okt 2026 — Fondasi Data
- Migrasi ke goose, matikan AutoMigrate di production
- Konversi uang ke `int64`
- `category_id` benar-benar dipakai
- Tabel: `budgets`, `recurring_rules`, `debt_payments`, `assets`, `price_snapshots`, `audit_logs`, `payments`
- RLS + `SET LOCAL app.tenant_id` + helper `Scoped(c)`
- `opening_balance` + job rekonsiliasi saldo
- Test: uang, isolasi tenant, RBAC

**Gate:** `go test ./...` hijau; query tanpa konteks tenant mengembalikan 0 baris.

### S2 · 5 – 16 Okt 2026 — Ledger & UX Inti
- Edit & hapus transaksi (soft delete + audit log)
- Transfer antar dompet
- Pagination + filter periode + `FormatRupiah` format Indonesia
- Quick entry < 5 detik (FAB, default cerdas, shortcut nominal)
- Import CSV/Excel

**Gate:** Prof + istri memakai penuh selama 2 minggu tanpa perlu edit database manual.

### S3 · 19 – 30 Okt 2026 — Menepati Janji Produk
- Budget per periode + indikator server-side
- Net Worth (ganti kartu "Coming Soon")
- Snowball & Avalanche beneran menghitung + unit test
- Emergency Fund Health Score + Sinking Funds
- Audit log + halaman "Aktivitas Keluarga"

**Gate:** setiap klaim di landing page bisa ditunjukkan berjalan di aplikasi.

### S4 · 2 – 13 Nov 2026 — Notifikasi, Aset, Frontend
- Scheduler cron + DB lock
- Debt reminder H-3/H-1, budget alert 80/100%, ringkasan bulanan
- Verifikasi email + reset password
- Cron harga logam mulia → `price_snapshots` + fallback harga manual + perbaikan dinar/perak
- Tailwind CLI + self-host Alpine/HTMX + dark mode + dashboard pakai `@Layout`
- `sw.js` network-first + bersihkan cache saat logout + icon PWA sendiri

**Gate:** email reminder benar-benar masuk inbox; halaman aset < 400 ms.

### S5 · 16 – 27 Nov 2026 — Monetisasi & Production
- `RequirePlan()` + limit Free + trial 14 hari + cron penurunan plan
- `POST /activate-voucher` + webhook Mayar.id (idempotent, verifikasi signature)
- Kebijakan Privasi & S&K, hapus akun mandiri, ekspor data mandiri
- `pg_dump` harian terenkripsi + **uji restore** (catat tanggalnya di `task.md`)
- Dockerfile production, deploy via Coolify, domain + TLS, CI

**Gate:** satu transaksi pembayaran uji berhasil mengubah plan tenant secara otomatis.

### Beta · 30 Nov – 18 Des 2026 — Closed Beta
- 10-15 keluarga (komunitas & lingkaran dekat), voucher gratis 1 tahun sebagai imbalan feedback
- Ukur: aktivasi (catat transaksi pertama < 10 menit), retensi mingguan, transaksi/user/minggu
- Perbaikan bug + testimoni & screenshot untuk materi promosi (lihat `marketing.md`)

**Gate:** retensi minggu-4 ≥ 40% dan **tidak ada bug integritas data**. Bila belum tercapai, tunda launch — jangan promosikan produk yang belum dipakai orang.

### Launch · 5 Januari 2027
- Momentum awal tahun (resolusi keuangan keluarga) + promo Rp 190.000/tahun
- Eksekusi `marketing.md` & `content.md`

### S6+ · Januari – Februari 2027 — Diferensiasi
- Kalkulator Zakat Maal + rekap tahunan zakat/infaq
- PDF asli, grafik pertumbuhan aset, recurring auto-post, alokasi 50/30/20, reminder WA
- Onboarding wizard + push notification harian

---

## Target Bisnis (konservatif, untuk kalibrasi ekspektasi)

| Periode | Target |
|---|---|
| Des 2026 | 15 keluarga beta aktif, 0 churn karena bug data |
| Q1 2027 | 100 pendaftar free, **20 berbayar** (≈ Rp 3,8 jt) |
| Q2 2027 | 400 pendaftar free, **80 berbayar** kumulatif (≈ Rp 15,2 jt) |
| Q4 2027 | 250 berbayar (≈ Rp 47,5 jt/tahun) |

> Ini produk sampingan dengan pertumbuhan organik. Angka ini **belum** menjadi alasan resign dari pekerjaan tetap — sesuai aturan kerja: income remote/freelance harus stabil jauh di atas gaji tetap selama beberapa bulan berturut-turut lebih dulu.

---

## Risiko & Mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Kapasitas 12-15 jam/minggu tergerus proyek kantor | Timeline molor 2-4 minggu | Geser jadwal, jangan potong P0. Prioritaskan S0-S1 di minggu paling longgar |
| Sumber harga emas tidak tersedia | Fitur unggulan tidak bisa diiklankan | Fallback harga manual + jangan pasarkan "real-time" sebelum sumbernya nyata |
| Kebocoran data antar tenant | Fatal — produk mati | RLS + test isolasi + audit sebelum tiap rilis |
| Sepi konversi free → premium | Tidak ada pendapatan | Trial 14 hari + batasan pada user & histori (bukan jumlah transaksi) |
| Gagal restore backup saat insiden | Kehilangan kepercayaan permanen | Uji restore wajib sebelum launch, ulangi tiap kuartal |
