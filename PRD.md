# Product Requirement Document (PRD)
## JurnalUmi — Aplikasi Pencatatan & Manajemen Keuangan Rumah Tangga

---

### 1. RINGKASAN EKSEKUTIF
**JurnalUmi** adalah aplikasi pencatatan dan pengelolaan keuangan rumah tangga yang dirancang simpel, fungsional, dan fokus pada kejelasan arus kas (cashflow), pertumbuhan aset bersih (net worth), dan pengendalian utang/piutang keluarga. 

Aplikasi ini ditujukan untuk digunakan oleh pasangan/keluarga agar mampu mengontrol kesehatan finansial harian hingga perencanaan masa depan tanpa kerumitan pencatatan akuntansi tradisional.

---

### 2. TUJUAN PRODUK (PRODUCT GOALS)
1. **Pencatatan Cepat & Tanpa Beban (Zero Friction):** Memudahkan penginputan transaksi harian dalam kurang dari 5 detik.
2. **Keterbukaan & Transparansi Finansial:** Visibilitas mutlak atas pemasukan, pengeluaran, utang, dan saldo tabungan riil.
3. **Peringatan Kesehatan Keuangan (Early Warning System):** Mencegah defisit bulanan dan mendeteksi utang konsumtif yang membahayakan.
4. **Keamanan Data Keluarga:** Proteksi data sensitif dengan isolasi akun berbasis pasangan/keluarga.

---

### 3. FITUR UTAMA & MODUL APLIKASI

#### A. Pemasukan (Income Management)
- **Sumber Pemasukan:**
  - Pemasukan Utama (Gaji Suami/Istri, Hasil Bisnis).
  - Pemasukan Sampingan (Project Freelance, Dividen, Hasil Sewa).
  - Pemasukan Pasif / Non-Rutin (Bonus, THR, Hadiah).
- **Fitur:** Alokasi otomatis (Skema 50/30/20 atau Custom Budgeting), Rekues penerimaan terjadwal (Tanggal Gaji).

#### B. Pengeluaran (Expense Management)
- **Kategorisasi Berhierarki (Wajib vs Sukarela):**
  - **Kebutuhan Rutin (Fixed):** Belanja Dapur, Listrik/Air, Sekolah Anak, Servis Motor, BPJS/Asuransi.
  - **Keinginan (Variable):** Hiburan, Makan Luar, Jajan, Belanja Hobi.
  - **Kewajiban Sosial/Agama:** Zakat, Infaq, Sedekah, Uang Orang Tua/Mertua.
- **Budgeting Limit:** Alokasi batas maksimum per kategori dengan indikator warna (Hijau / Kuning / Merah).

#### C. Utang & Piutang (Debt & Receivable Tracking) — *SANGAT PENTING*
- **Pencatatan Utang (Debt):**
  - Utang Bank/Cicilan (KPR, Kredit Motor/Mobil).
  - Utang Kartu Kredit / Paylater.
  - Utang Perorangan (Kerabat/Teman).
  - Fitur: Jadwal jatuh tempo, jumlah sisa pokok, kalkulasi bunga, kalkulator pelunasan (Metode Snowball / Avalanche).
- **Pencatatan Piutang (Receivable):**
  - Pinjaman yang diberikan ke orang lain + status penagihan & histori pembayaran.

#### D. Aset & Liabilitas (Net Worth Tracker)
- **Aset Likuid:** Uang Tunai, Saldo Bank, E-Wallet (Gopay/OVO/ShopeePay).
- **Aset Investasi & Tabungan:** Emas/Logam Mulia, Reksadana, Saham, Deposito, Dana Darurat.
- **Aset Fisik:** Kendaraan (Motor/Mobil), Rumah/Tanah, Barang Elektronik Bernilai.
- **Net Worth Real-Time:** `Total Aset - Total Utang = Kekayaan Bersih Riil`.

#### E. Fitur Paling Krusial Rumah Tangga (Essential Core Features)
1. **Kalkulator & Tracker Dana Darurat (Emergency Fund):**
   - Menghitung kecukupan dana darurat (idealnya 6-12 bulan pengeluaran rutin).
   - Menampilkan status keamanan kas jika terjadi PHK / penurunan income mendadak.
2. **Post-Budget / Sinking Fund (Tabungan Alokasi Khusus):**
   - Pos dana khusus untuk kebutuhan tahunan (Kurban, Pajak Kendaraan, Mudik Lebaran, Liburan, Renovasi Rumah).
3. **Multi-User / Akses Bersama Pasangan (Suami-Istri Sync):**
   - Transaksi yang dicatat Suami atau Istri langsung sinkron real-time dalam satu dompet keluarga.

#### F. Laporan Finansial (Financial Reports)
- **Laporan Arus Kas (Cash Flow Report):** Pemasukan vs Pengeluaran Bulanan (Surplus/Defisit).
- **Laporan Kekayaan Bersih (Net Worth Evolution):** Grafik pertumbuhan aset dari bulan ke bulan.
- **Laporan Breakdown Pengeluaran:** Pie chart persentase konsumsi harian vs investasi.
- **Ekspor Data:** Export ke PDF & Excel (.xlsx) untuk evaluasi bulanan keluarga.

---

### 4. ARSITEKTUR TEKNIS & STACK APLIKASI
- **Backend Framework:** Go 1.22+ (Echo Framework)
- **Templating / Rendering:** Templ (Type-safe HTML templating for Go) — Pure SSR (Server-Side Rendering)
- **Database:** PostgreSQL 16+ (Podman Rootless Container)
- **Styling:** Tailwind CSS v3/v4
- **Live Reload Dev Tool:** Air (`.air.toml`)
- **Business Model / Architecture:** SaaS (Multi-Tenant Family Accounts, Role-based Access)

---

### 5. METRIK KEBERHASILAN (KPI)
- Waktu pencatatan transaksi < 5 detik.
- Ketepatan prediksi batas budget mingguan (0% overbudget tanpa notifikasi).
- Penggunaan aktif bersama oleh 2 user (Suami & Istri) dalam satu dompet keluarga.

---

### 6. RENCANA PENGEMBANGAN (ROADMAP)
- **Fase 1:** Core CRUD Transaction (Income, Expense, Wallet, Categories).
- **Fase 2:** Debt/Receivable Tracker & Net Worth Calculation.
- **Fase 3:** Emergency Fund Progress & Multi-User Couple Sync.
- **Fase 4:** Export PDF Report & PWA Mobile App.
