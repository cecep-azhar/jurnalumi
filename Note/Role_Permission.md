# Role & Menu Permission JurnalUmi

- **Diperbarui:** 8 September 2026 — diselaraskan dengan `PRD.md` v3.0.0
- **Status implementasi:** 🔴 **belum ada sama sekali di kode.** Saat ini hanya ada `RequireAuth` (autentikasi), tanpa satu pun pengecekan role. Dokumen ini adalah spesifikasi target middleware `RequireRole()`, bukan deskripsi kondisi sekarang. Lihat `review.md` temuan S4.

## Aturan Penegakan

1. **Role dicek di server pada setiap handler.** Menyembunyikan menu di UI **bukan** kontrol akses.
2. Kegagalan otorisasi → HTTP 403 (bukan redirect diam-diam).
3. `superadmin` adalah user tanpa `tenant_id`, dibuat lewat CLI seeder — **tidak pernah** lewat form publik.
4. `superadmin` **tidak boleh melihat isi transaksi tenant** — hanya metadata & agregat (nama keluarga, plan, jumlah user, tanggal aktif).
5. Setiap aksi tulis menulis `audit_logs` (siapa, kapan, entitas apa, dari nilai berapa ke berapa).

## Roles

| Kode | Sebutan |
|---|---|
| `superadmin` | Platform Owner |
| `owner` | Kepala Keluarga (Family Owner) |
| `spouse` | Pasangan (Family Co-Owner) |
| `member` | Anak / Tanggungan |
| `auditor` | Perencana Keuangan (Read-Only) |

## Matrix Hak Akses

**Legend:** `F` = full (lihat/tulis/hapus) · `W` = lihat & tulis, tidak boleh hapus · `R` = read-only · `L` = terbatas · `–` = tidak ada akses

| Modul / Aksi | superadmin | owner | spouse | member | auditor |
|---|:---:|:---:|:---:|:---:|:---:|
| **Platform** |
| Kelola tenant, billing, suspensi | F | – | – | – | – |
| Generate voucher | F | – | – | – | – |
| Variabel global (sumber harga logam mulia) | F | – | – | – | – |
| Melihat isi transaksi tenant | – | – | – | – | – |
| **Pengaturan Keluarga** |
| Kelola anggota keluarga | – | F | – | – | – |
| Setel budget per periode | – | F | W | – | R |
| Setel target dana darurat & sinking fund | – | F | W | – | R |
| Ubah plan / redeem voucher | – | F | – | – | – |
| Hapus akun keluarga & ekspor data | – | F | – | – | – |
| **Ledger Harian** |
| Catat pemasukan | – | F | F | – | R |
| Catat pengeluaran | – | F | F | L¹ | R |
| Edit / hapus transaksi | – | F | W² | L¹ | – |
| Transfer antar dompet | – | F | F | – | R |
| Kelola dompet | – | F | F | – | R |
| Kelola kategori | – | F | F | – | R |
| Import CSV/Excel | – | F | F | – | – |
| **Aset & Net Worth** |
| Dashboard net worth | – | F | F | – | R |
| Aset likuid & investasi | – | F | F | – | R |
| Logam mulia (emas/dinar/perak) | – | F | F | – | R |
| **Utang & Piutang** |
| Lihat utang & piutang | – | F | F | – | R |
| Catat utang/piutang & pembayaran | – | F | W³ | – | – |
| Hapus utang/piutang | – | F | – | – | – |
| Kalkulator Snowball/Avalanche | – | F | F | – | R |
| **Proteksi & Tujuan** |
| Dana darurat & health score | – | F | F | – | R |
| Sinking funds | – | F | F | – | R |
| **Syariah** |
| Kalkulator zakat maal | – | F | F | – | R |
| **Laporan** |
| Laporan & ekspor CSV/PDF | – | F | F | L¹ | R |
| Halaman Aktivitas Keluarga (audit log) | – | F | F | – | R |

### Catatan
1. **`member` (anak):** hanya pada dompet yang ditugaskan kepadanya (`wallets.assigned_user_id`). Hanya bisa mengedit/menghapus transaksi yang ia buat sendiri, dalam 24 jam. Tidak melihat utang, aset, maupun total keuangan keluarga.
2. **`spouse` edit/hapus transaksi:** boleh mengedit transaksi mana pun; penghapusan permanen hanya oleh `owner`. Semua perubahan tercatat di audit log.
3. **`spouse` pada utang/piutang:** *revisi dari versi sebelumnya yang menetapkan read-only.* Alasan: dalam praktik rumah tangga, pasangan yang membayar cicilan harus bisa mencatat pembayarannya. Penghapusan tetap eksklusif `owner` sebagai pengaman.
4. **`auditor`** tidak pernah bisa menulis apa pun, termasuk komentar. Akses diberikan oleh `owner` dan dapat dicabut kapan saja.
