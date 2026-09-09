# JurnalUmi

Aplikasi pencatatan keuangan keluarga.

## Requirements
- Go 1.25+
- Docker & Docker Compose (untuk dev database & mail)

## Local Development

1. Salin `.env.example` ke `.env`
   ```bash
   cp .env.example .env
   ```

2. Jalankan database dan layanan pendukung
   ```bash
   docker-compose up -d
   ```

3. Generate UI dan jalankan aplikasi
   ```bash
   go run github.com/a-h/templ/cmd/templ@latest generate
   go run ./cmd/server
   ```

Aplikasi akan berjalan di http://localhost:8080
Mailhog akan berjalan di http://localhost:8025
