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

### 2026-09-09 13:35 WIB — Idle
Status: idle
Ringkasan: 3 PR terbuka (PR #4, #5, #6). Menunggu merge dari Prof.
