# Prosedur QA JurnalUmi

Prosedur setiap cron tick:
1. Baca status repositori.
2. Cek open pull requests dengan `gh pr list`. Jika ada 3 atau lebih PR yang belum di-merge, laporkan idle.
3. Cek log sebelumnya di file ini (bawah).
4. Baca `Note/task-qa.md`. Cari tugas berstatus `TODO`.
5. Buat branch baru dari `main` dengan format `qa/[id-task]`.
6. Lakukan perubahan sesuai tugas.
7. Verifikasi kompilasi: `templ generate` & `go build ./...`.
8. Commit, push branch, dan buat pull request (PR).
9. Tulis ringkasan hasil ke akhir file ini (append). Hentikan eksekusi.

STOP total jika:
- Error tak teratasi > 3 kali
- Ditemukan konflik berat di `main` yang mencegah rebase
Idle: 3 PRs open. Waiting for merge.

### $(date '+%Y-%m-%d %H:%M WIB') — QA-P0-14
Status: QA-P0-14 dikerjakan, PR terbuka
Ringkasan: Menambahkan `docker-compose.yml` untuk lingkungan pengembangan lokal (PostgreSQL 16 + Mailhog + App).
PR: https://github.com/cecep-azhar/jurnalumi/pull/16

Status: QA-P0-07 dikerjakan, PR terbuka
Ringkasan: Menghapus middleware CORS global karena tidak diperlukan.
PR: https://github.com/cecep-azhar/jurnalumi/pull/8

### 2026-09-09 13:50 WIB — QA-P0-08
Status: QA-P0-08 dikerjakan, PR terbuka
Ringkasan: Menambahkan security headers (HSTS, nosniff, Referrer-Policy, CSP).
PR: https://github.com/cecep-azhar/jurnalumi/pull/9
### 2026-09-09 14:05 WIB — QA-P0-09
Status: QA-P0-09 dikerjakan, PR terbuka
Ringkasan: Menambahkan CLI command di `cmd/seed/main.go` untuk membuat akun superadmin. `go build ./...` lulus.
PR: https://github.com/cecep-azhar/jurnalumi/pull/10

### $(date '+%Y-%m-%d %H:%M WIB') — QA-P0-10
Status: QA-P0-10 dikerjakan, PR terbuka
Ringkasan: Menghapus implementasi mock `DebtPayPOST` dan menggantinya dengan transaksi DB nyata (potong/tambah dompet, update sisa hutang, tambah ke riwayat transaksi). Form modal UI diubah untuk menanyakan dompet sumber/tujuan dan nominal. `go build ./...` lulus.
PR: https://github.com/cecep-azhar/jurnalumi/pull/11

### $(date '+%Y-%m-%d %H:%M WIB') — QA-P0-11
Status: QA-P0-11 dikerjakan, PR terbuka
Ringkasan: Tangani semua operasi `db.Create` dan `db.Save` di handler (assets, dashboard, debts, features, admin) dengan mengecek `Error` dan mengembalikan respon error 500 jika gagal.
PR: https://github.com/cecep-azhar/jurnalumi/pull/12
### $(date '+%Y-%m-%d %H:%M WIB') — QA-P0-11
Status: QA-P0-11 selesai, PR terbuka (re-created setelah merge conflict)
Ringkasan: Tangani semua operasi `db.Create` dan `db.Save` di handler dengan mengecek `Error`.
PR: https://github.com/cecep-azhar/jurnalumi/pull/13

### $(date '+%Y-%m-%d %H:%M WIB') — QA-P0-12
Status: QA-P0-12 dikerjakan, PR terbuka
Ringkasan: Menambahkan validasi input server-side pada AssetPOST, WalletPOST, CategoryPOST, DebtPOST, dan FamilyPOST (nominal > 0, field wajib, enum valid). `go build` sukses.
PR: https://github.com/cecep-azhar/jurnalumi/pull/14

### $(date '+%Y-%m-%d %H:%M WIB') — QA-P0-13
Status: QA-P0-13 dikerjakan, PR terbuka
Ringkasan: Memperbarui versi golang di Dockerfile.dev menjadi 1.25-alpine menyesuaikan dengan go.mod.
PR: https://github.com/cecep-azhar/jurnalumi/pull/15

### $(date '+%Y-%m-%d %H:%M WIB') — QA-P0-15
Status: QA-P0-15 dikerjakan, PR terbuka
Ringkasan: Membuat `.env.example` dan `README.md` dengan instruksi menjalankan secara lokal. `go build` lulus.
PR: https://github.com/cecep-azhar/jurnalumi/pull/17
### $(date '+%Y-%m-%d %H:%M WIB') — QA-P0-16
Status: QA-P0-16 dikerjakan, PR terbuka
Ringkasan: Menghapus duplikasi deklarasi `e.Static("/static", ...)` di `cmd/server/main.go`. Verifikasi kompilasi dan generation berhasil.
PR: https://github.com/cecep-azhar/jurnalumi/pull/18

### $(date '+%Y-%m-%d %H:%M WIB') — QA-P0-15
Status: QA-P0-15 dikerjakan, PR terbuka
Ringkasan: Membuat `.env.example` dan `README.md` dengan instruksi menjalankan secara lokal. `go build` lulus.
PR: https://github.com/cecep-azhar/jurnalumi/pull/19
### $(date '+%Y-%m-%d %H:%M WIB') — QA-P1-01
Status: QA-P1-01 dikerjakan, PR terbuka
Ringkasan: Mengubah tipe data nominal/keuangan dari `float64` menjadi `int64` di model, handler, template, dan service. Mengganti parser `strconv.ParseFloat` menjadi `strconv.ParseInt`. `go build` sukses.
PR: https://github.com/cecep-azhar/jurnalumi/pull/20
