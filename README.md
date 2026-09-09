# JurnalUmi

Aplikasi manajemen keuangan keluarga (SaaS).

## Persyaratan
- Go 1.25+
- Node.js (untuk Tailwind CSS)
- Templ CLI
- Docker & Docker Compose (untuk lokal/dev)

## Menjalankan secara Lokal

1. **Siapkan konfigurasi:**
   ```bash
   cp .env.example .env
   ```
   *Edit isi `.env` bila perlu.*

2. **Jalankan dependensi lokal (PostgreSQL + Mailhog):**
   ```bash
   docker-compose up -d
   ```

3. **Install toolchain & generate kode (Templ + CSS):**
   ```bash
   npm install
   go mod download
   make build-tailwind # jika ada target makefile, atau manual: npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.css
   templ generate
   ```

4. **Jalankan aplikasi:**
   ```bash
   go run cmd/server/main.go
   ```
   Aplikasi akan berjalan di `http://localhost:8080`

## Menjalankan via Docker (Penuh)
Jika Anda menggunakan Dockerfile.dev:
```bash
docker build -t jurnalumi-dev -f Dockerfile.dev .
# (Atur port dan network sesuai kebutuhan)
```
