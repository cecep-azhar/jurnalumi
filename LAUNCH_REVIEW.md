# Launch Gate Review — JurnalUmi

> PR ringkasan sesuai `Note/prompt-qa.md` § "Kondisi STOP total": seluruh P0 + P1-KRITIS di `Note/task-qa.md` sudah `[x]`. Ini bukan PR kode — dibuka supaya Prof Cecep punya satu tempat untuk review & putuskan lanjut publish atau tidak. Setelah PR ini, cron QA berhenti (tidak lanjut otonom ke P1 non-kritis/P2) sampai ada instruksi baru dari Prof.

## Status checklist (per 2026-09-15)

- **P0 — BLOCKER:** 16/16 `[x]`
- **P1-KRITIS:** 31/31 `[x]` (6 item terakhir — QA-P1-17, 18, 23, 24, 29, 30 — baru selesai 2026-09-15 setelah Prof memberi keputusan di `QA-DECISIONS.md`)
- Detail lengkap per item: `Note/task-qa.md`. Keputusan bisnis/kredensial yang jadi dasar penyelesaian 6 item terakhir: `QA-DECISIONS.md`.

## Verifikasi dijalankan run ini (bukan cuma baca log lama)

- `go build ./...` — lulus.
- `go test ./...` — lulus (`cmd/server`, `internal/handlers`, `internal/middleware`, `internal/models`, `internal/services`).
- `templ generate` — 0 perubahan (generated code sudah sinkron dengan `.templ`).
- Spot-check kode untuk 6 item terakhir:
  - `.env.example` punya `MAYAR_API_KEY`, `MAYAR_WEBHOOK_SECRET` (kosong, diisi Prof saat deploy — sesuai Aturan Keras #1, tidak ada secret di source).
  - `internal/services/email.go` sudah membaca `SMTP_USERNAME`/`SMTP_PASSWORD` (fallback ke `SMTP_USER`/`SMTP_PASS`) — mismatch nama env var yang tercatat di "Temuan Baru Selama QA" (`task-qa.md`) sudah diperbaiki.
  - `internal/services/pricing.go` jadi satu sumber harga; `web/views/landing.html` konsisten "Rp 9.000/bln, 6 bulan pertama gratis" (sesuai `QA-DECISIONS.md`), tidak ada lagi angka 190rb/39k Mayar yang lama.
  - `scripts/backup.sh` + `scripts/restore.sh` ada (pg_dump terenkripsi AES-256-CBC → upload S3/MinIO compatible).
  - `Dockerfile` (production, multi-stage) ada di root, terpisah dari `Dockerfile.dev`.

## Yang PERLU dikonfirmasi/dilakukan Prof secara manual sebelum publish beneran (di luar apa yang bisa diverifikasi dari repo)

Ini bukan blocker kode — ini aksi infra/ops yang cuma bisa Prof lakukan:

1. **Kredensial produksi nyata** — isi `SMTP_USERNAME`/`SMTP_PASSWORD` (SMTP asli, bukan Mailhog) dan (kalau nanti Mayar diaktifkan lagi) `MAYAR_API_KEY`/`MAYAR_WEBHOOK_SECRET` di environment server produksi, bukan di repo.
2. **Deploy Coolify nyata** — pastikan Coolify benar-benar auto-deploy dari push ke `main` (per `QA-DECISIONS.md` poin 5: "cukup push ke GitHub"), domain custom sudah terpasang, dan TLS aktif (certificate valid, bukan self-signed/default).
3. **Uji restore produksi** — `scripts/restore.sh` sudah divalidasi pipeline enkripsi/dekripsi di sesi QA (dicatat "Uji restore: 2026-09-15" di `task-qa.md`), tapi belum pernah dijalankan melawan backup S3/MinIO produksi yang sesungguhnya — sarankan Prof jalankan sekali lagi end-to-end setelah kredensial S3 produksi terisi.
4. **Pembayaran manual transfer** — QA-P1-23 (webhook Mayar) tetap diimplementasikan tapi TIDAK dipakai untuk alur pembayaran saat ini (Mayar "DITUNDA" per `QA-DECISIONS.md`). Alur aktif adalah transfer manual ke rekening BSI yang tertera di landing page — pastikan ada proses manual (siapa yang cek mutasi & upgrade akun user) karena tidak ada otomatisasi untuk jalur ini.
5. **Konten hukum final** — halaman `/privacy` dan `/terms` (QA-P1-27) masih draft standar UU PDP, ditandai perlu review/finalisasi teks oleh Prof (dicatat di log `task-qa.md` 2026-09-12).

## Rekomendasi

Dari sisi kode & test, JurnalUmi memenuhi seluruh gate P0 + P1-KRITIS yang disepakati di `task-qa.md`. Cron QA berhenti di sini (kondisi STOP total, sesuai `prompt-qa.md`) — menunggu Prof review PR ini dan memutuskan langkah publish, termasuk 5 item ops di atas yang tidak bisa diverifikasi/dieksekusi dari sesi QA otonom ini.
