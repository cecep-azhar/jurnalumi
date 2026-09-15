#!/bin/sh
set -e

# Script Backup pg_dump terenkripsi ke MinIO / S3 compatible storage
# Sesuai Keputusan Prof. Cecep (QA-DECISIONS.md #4) & QA-P1-29

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_DIR=${BACKUP_DIR:-"/tmp/backups"}
DB_NAME=${POSTGRES_DB:-"jurnalumi"}
DB_USER=${POSTGRES_USER:-"postgres"}
DB_HOST=${POSTGRES_HOST:-"127.0.0.1"}
DB_PORT=${POSTGRES_PORT:-"5432"}
ENCRYPTION_KEY=${BACKUP_ENCRYPTION_KEY:-"default-secure-backup-key-jurnalumi-2026"}

mkdir -p "$BACKUP_DIR"

DUMP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"
ENC_FILE="${DUMP_FILE}.enc"

echo "[$(date)] Starting pg_dump for database ${DB_NAME}..."

# 1. pg_dump compressed with gzip
PGPASSWORD="${POSTGRES_PASSWORD}" pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" | gzip > "$DUMP_FILE"

# 2. Encrypt using openssl AES-256-CBC
echo "[$(date)] Encrypting backup file..."
openssl enc -aes-256-cbc -salt -pbkdf2 -in "$DUMP_FILE" -out "$ENC_FILE" -pass pass:"$ENCRYPTION_KEY"
rm -f "$DUMP_FILE"

echo "[$(date)] Encrypted backup created at: ${ENC_FILE}"

# 3. Upload to MinIO / S3 compatible storage if configured
if [ -n "$S3_ENDPOINT" ] && [ -n "$S3_BUCKET" ] && [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_ACCESS_KEY" ]; then
    echo "[$(date)] Uploading to S3/MinIO bucket ${S3_BUCKET}..."
    if command -v mc >/dev/null 2>&1; then
        mc alias set minio "$S3_ENDPOINT" "$AWS_ACCESS_KEY_ID" "$AWS_SECRET_ACCESS_KEY" --api S3v4
        mc cp "$ENC_FILE" "minio/${S3_BUCKET}/backups/"
    elif command -v aws >/dev/null 2>&1; then
        aws --endpoint-url "$S3_ENDPOINT" s3 cp "$ENC_FILE" "s3://${S3_BUCKET}/backups/"
    else
        echo "[WARNING] Neither 'mc' nor 'aws' CLI found. File kept locally at ${ENC_FILE}."
    fi
else
    echo "[INFO] S3/MinIO credentials not set. Encrypted backup kept locally at ${ENC_FILE}."
fi

echo "[$(date)] Backup process completed successfully."
